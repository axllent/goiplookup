
GOPATH := ${PWD}
export GOPATH
TAG=`git describe --tags`
VERSION ?= `git describe --tags`
LDFLAGS=-ldflags "-s -extldflags \"--static\" -w -X main.version=${VERSION}"

geoiplookup: goiplookup.go
	go get github.com/oschwald/geoip2-golang
	CGO_ENABLED=0 go build ${LDFLAGS}

clean:
	rm -rf pkg src goiplookup

release:
	go get github.com/oschwald/geoip2-golang
	mkdir -p dist
	rm -f dist/goiplookup_${VERSION}_*

	echo "Building binaries for ${VERSION}"

	echo "- linux-amd64"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/goiplookup_${VERSION}_linux_amd64
	bzip2 dist/goiplookup_${VERSION}_linux_amd64

	echo "- linux-386"
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build ${LDFLAGS} -o dist/goiplookup_${VERSION}_linux_386
	bzip2 dist/goiplookup_${VERSION}_linux_386

	echo "- linux-arm"
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build ${LDFLAGS} -o dist/goiplookup_${VERSION}_linux_arm
	bzip2 dist/goiplookup_${VERSION}_linux_arm

	echo "- linux-arm64"
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o dist/goiplookup_${VERSION}_linux_arm64
	bzip2 dist/goiplookup_${VERSION}_linux_arm64

	echo "- darwin-amd64"
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o dist/goiplookup_${VERSION}_darwin_amd64
	bzip2 dist/goiplookup_${VERSION}_darwin_amd64

	echo "- darwin-386"
	CGO_ENABLED=0 GOOS=darwin GOARCH=386 go build ${LDFLAGS} -o dist/goiplookup_${VERSION}_darwin_386
	bzip2 dist/goiplookup_${VERSION}_darwin_386
