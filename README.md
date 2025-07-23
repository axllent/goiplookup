# GoIPLookup - geoiplookup for GeoLite2 written in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/axllent/goiplookup)](https://goreportcard.com/report/github.com/axllent/goiplookup)

GoIPLookup is a geoiplookup replacement for the [free MaxMind GeoLite2-Country](https://dev.maxmind.com/geoip/geoip2/geolite2/) and [GeoLite2-City](https://dev.maxmind.com/geoip/geoip2/geolite2/),
written in [Go](https://golang.org/).

It currently supports the free GeoLite2-Country and GeoLite2-City databases. There is no planned support for other types.

## Features

-   Drop-in replacement for the now defunct `geoiplookup` utility, simply rename it
-   Works with the current MaxMind database format (mmdb)
-   IPv4, IPv6 and fully qualified domain name (FQDN) support
-   Supports both GeoLite2-Country and GeoLite2-City databases
-   City database returns city, subdivision, country ISO and name
-   Options to return just the country iso (`NZ`) or country name (`New Zealand`), rather than the full `GeoIP Country Edition: NZ, New Zealand`
-   Built-in database update support (see [Database updates](#database-updates))
-   Built-in self updater (if new release is available)

## Installing

Multiple OS/Architecture binaries are supplied with releases. Extract the binary, make it executable, and move it to a location such as `/usr/local/bin`.

If you wish to replace an existing defunct implementation of geoiplookup, then simply name the file `geoiplookup`.

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
goiplookup 8.8.8.8               Return the country/city ISO code and name
goiplookup -d ~/GeoIP 8.8.8.8    Use a different database directory
goiplookup -i 8.8.8.8            Return just the country ISO code
goiplookup -c 8.8.8.8            Return just the country name
goiplookup db-update             Update the GeoLite2-Country database (do not run more than once a month)
goiplookup db-update-city        Update the GeoLite2-City database (do not run more than once a month)
goiplookup self-update           Update the GoIpLookup binary with the latest release
```

If the GeoLite2-City database is present, output will include city and subdivision information.

## GoIPLookup updates

GoIPLookup comes with a built-in self-updater:

```
goiplookup self-update
```

Version checked (`goiplookup -V`) will tell you if your version is out of date.

## Database updates

GoIPLookup is able to update your GeoLite2 Country and City databases. As of 01/01/2020 MaxMind require a (free) License Key in order to download these updates. The release (binary) versions of goiplookup (>= 0.2.2) already contain a key for this, however if you are compiling from source you will need to set your own licence key in your environment (see below).

### Binary release database updates

```
goiplookup db-update         # Update Country database
goiplookup db-update-city    # Update City database
```

### Self-compiled database updates

If you wish to use your own MaxMind license key, or you are compiling from source, then you must provide a key in your environment.
To generate your own license key from MaxMind you must first [register a free account](https://www.maxmind.com/en/geolite2/signup) and follow the instructions.

```
LICENSEKEY="xxxxxxxx" goiplookup db-update
```

or

```
export LICENSEKEY="xxxxxxxx"
goiplookup db-update
```

or if you wish to compile the source code yourself with your key:

```
make LICENSEKEY="xxxxxxxx"
```

## Compiling from source

Go >= 1.23 required:

```
go install github.com/axllent/goiplookup@latest
```
