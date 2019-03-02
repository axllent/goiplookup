
GOPATH := ${PWD}
export GOPATH
VERSION=`git describe --tags`
LDFLAGS=-ldflags "-X main.version=${VERSION}"

geoiplookup: goiplookup.go
	go get github.com/oschwald/geoip2-golang
	go build ${LDFLAGS} goiplookup.go
	strip goiplookup

clean:
	rm -rf pkg src goiplookup
