.SILENT :
.PHONY : cwait clean fmt

TAG:=`git rev-parse HEAD`
LDFLAGS:=-X main.buildVersion=$(TAG)

all: cwait

cwait:
	echo "Building cwait"
	go build -ldflags "$(LDFLAGS)"

dist-clean:
	rm -rf dist

dist: dist-clean
	mkdir -p dist/linux/386 && GOOS=linux GOARCH=386 go build -ldflags "$(LDFLAGS)" -o dist/linux/386/cwait
	mkdir -p dist/linux/amd64 && GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/linux/amd64/cwait

release: dist
	mkdir -p dist/releases
	tar -cvzf dist/releases/cwait-linux-386-$(TAG).tar.gz -C dist/linux/386 cwait
	tar -cvzf dist/releases/cwait-linux-amd64-$(TAG).tar.gz -C dist/linux/amd64 cwait
