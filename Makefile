
GOPATH := ${PWD}
export GOPATH
VERSION=$(git describe --tags)

geoiplookup: goiplookup.go
	go get github.com/oschwald/geoip2-golang
	go build -ldflags "-X main.version=${VERSION}" goiplookup.go
	strip goiplookup

clean:
	rm -rf pkg src goiplookup
