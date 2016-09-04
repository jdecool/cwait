package main

import (
	"database/sql"
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var (
		waitTimeoutFlag time.Duration
		wg              sync.WaitGroup
	)

	flag.DurationVar(&waitTimeoutFlag, "timeout", 1*time.Minute, "Host wait timeout")
	flag.Parse()

	dependencyChan := make(chan struct{})
	go func() {
		for _, arg := range flag.Args() {
			u, err := url.Parse(arg)
			if err != nil {
				log.Fatalf("Unable to parse: %s", u)
			}

			switch u.Scheme {
			case "tcp", "udp":
				wg.Add(1)
				go func() {
					defer wg.Done()

					for {
						conn, err := net.DialTimeout(u.Scheme, u.Host, waitTimeoutFlag)
						if err != nil {
							log.Printf("Unable to connect: %s", u)
							time.Sleep(5 * time.Second)
						}

						if conn != nil {
							return
						}
					}
				}()

			case "mysql":
				dsn := generateMysqlDsn(u)

				wg.Add(1)
				go func() {
					defer wg.Done()

					for {
						conn, err := sql.Open(u.Scheme, dsn)
						if err != nil {
							log.Printf("Unable to connect: %s", u)
							time.Sleep(5 * time.Second)
						}
						defer conn.Close()

						if conn != nil && conn.Ping() != nil {
							log.Printf("Unable to connect: %s (%s)", u, conn.Ping())
							time.Sleep(5 * time.Second)
						} else {
							return
						}
					}
				}()

			case "http", "https":
				wg.Add(1)

				go func() {
					defer wg.Done()

					for {
						client := &http.Client{}
						req, _ := http.NewRequest("GET", u.String(), nil)
						resp, err := client.Do(req)
						if err != nil || resp.StatusCode < 200 || resp.StatusCode > 300 {
							log.Printf("Unable to connect: %s", u)
							time.Sleep(5 * time.Second)
						}

						if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
							return
						}
					}
				}()

			default:
				log.Fatalf("Nnvalid host protocol provided: %s.", u.Scheme)
			}
		}

		wg.Wait()
		close(dependencyChan)
	}()

	select {
	case <-dependencyChan:
		break
	case <-time.After(waitTimeoutFlag):
		log.Fatalf("Timeout after %s waiting on dependencies to become available", waitTimeoutFlag)
	}
}

func generateMysqlDsn(u *url.URL) string {
	var (
		dsn  string
		host string
		port string
	)

	if u.User != nil {
		dsn += u.User.String()
	}

	if len(dsn) > 0 {
		dsn += "@"
	}

	if strings.Contains(u.Host, ":") {
		var parts = strings.Split(u.Host, ":")
		host = parts[0]
		port = parts[1]
	} else {
		host = u.Host
		port = "3306"
	}

	dsn += "tcp(" + host + ":" + port + ")/"

	return dsn
}
