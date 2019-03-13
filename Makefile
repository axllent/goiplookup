
GOPATH := ${PWD}
export GOPATH
VERSION=`git describe --tags`
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION}"

geoiplookup: goiplookup.go
	go get github.com/oschwald/geoip2-golang
	go build ${LDFLAGS}

clean:
	rm -rf pkg src goiplookup
