
GOPATH := ${PWD}
export GOPATH

geoiplookup: goiplookup.go
	go get github.com/oschwald/geoip2-golang
	go build goiplookup.go
	strip goiplookup

clean:
	rm -rf pkg src goiplookup
