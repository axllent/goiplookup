package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

// Lookup ip or hostname
func Lookup(lookup string) {

	var ciso string
	var cname string
	var mmdb string
	var output []string
	var response string
	var ipraw string

	// convert to ip if hostname
	addresses, err := net.LookupHost(lookup)

	if len(addresses) > 0 {
		Verbose(fmt.Sprintf("Ip search for: %s", addresses[0]))
		ipraw = addresses[0]
	} else {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fi, err := os.Stat(dataDir)
	if err != nil {
		fmt.Println("Error: Directory does not exist", dataDir)
		os.Exit(1)
	}

	switch mode := fi.Mode(); {
	case mode.IsDir(): // if dataDir is dir, append GeoLite2-Country.mmdb
		mmdb = path.Join(dataDir, "GeoLite2-Country.mmdb")
	case mode.IsRegular():
		mmdb = dataDir
	}

	Verbose(fmt.Sprintf("Opening %s", mmdb))

	db, err := geoip2.Open(mmdb)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer db.Close()

	ip := net.ParseIP(ipraw)

	record, err := db.Country(ip)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if record.Traits.IsAnonymousProxy {
		Verbose("Anonymous IP detected")
		ciso = "A1"
		cname = "Anonymous Proxy"
	} else {
		ciso = record.Country.IsoCode
		cname = record.Country.Names["en"]
	}

	if country || iso {
		if iso && ciso != "" {
			output = append(output, ciso)
		}
		if country && cname != "" {
			output = append(output, cname)
		}
		response = fmt.Sprintf(strings.Join(output, ", "))
	} else {
		if ciso == "" {
			response = "GeoIP Country Edition: IP Address not found"
		} else {
			response = fmt.Sprintf("GeoIP Country Edition: %s, %s", ciso, cname)
		}
	}

	fmt.Println(response)
}

// DownloadToFile downloads a URL to a file
func DownloadToFile(filepath string, url string) error {

	Verbose(fmt.Sprintf("Downloading %s", url))

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// Verbose displays debug information with `-v`
func Verbose(m string) {
	if verboseoutput {
		fmt.Println(m)
	}
}
