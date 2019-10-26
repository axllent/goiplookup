# GoipLookup - geoiplookup for GeoLite2 written in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/axllent/goiplookup)](https://goreportcard.com/report/github.com/axllent/goiplookup)

GoipLookup is a geoiplookup replacement for the [free GeoLite2-Country](https://dev.maxmind.com/geoip/geoip2/geolite2/),
written in [Go](https://golang.org/).

It currently only supports the free GeoLite2-Country database, and there is no planned support for the other types.


## Features

- Drop-in replacement for the now defunt `geoiplookup` utility, simply rename it
- Works with the current Maxmind database format (mmdd)
- IPv4, IPv6 and fully qualified domain name (FQDN) support
- Options to return just the country iso (`NZ`) or country name (`New Zealand`), rather than the full `GeoIP Country Edition: NZ, New Zealand`
- Built-in database update support
- Built-in self updater (if new release is available)


## Installing

Multiple OS/Architecture binaries are supplied with releases. Extract the binary, make it executable, and move it to a location such as `/usr/local/bin`.

If you wish to replace an existing defunct implementation of geoiplookup, then simply name the file `geoiplookup`.


## Updating

GoipLookup comes with a built-in self-updater:

```
goiplookup self-update
```


## Compiling from source

Go >= 1.11 required:

```
go get github.com:axllent/goiplookup.git
```

## Basic usage

```
Usage: goiplookup [-i] [-c] [-d <database directory>] <ipaddress|hostname|db-update|self-update>

Options:
  -V	show version number
  -c	return country name
  -d string
    	database directory or file (default "/usr/share/GeoIP")
  -h	show help
  -i	return country iso code
  -v	verbose/debug output

Examples:
goiplookup 8.8.8.8			Return the country ISO code and name
goiplookup -d ~/GeoIP 8.8.8.8		Use a different database directory
goiplookup -i 8.8.8.8			Return just the country ISO code
goiplookup -c 8.8.8.8			Return just the country name
goiplookup db-update			Update the GeoLite2-Country database (do not run more than once a month)
goiplookup self-update			Update the GoIpLookup binary with the latest release
```
