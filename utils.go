package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

// LookupAddr looks up an ip or hostname
func lookupAddr(lookup string) {

	var ciso string
	var cname string
	var mmdb string
	var output []string
	var response string
	var ipraw string

	// convert to ip if hostname
	addresses, err := net.LookupHost(lookup)

	if len(addresses) > 0 {
		verbose(fmt.Sprintf("Ip search for: %s", addresses[0]))
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

	verbose(fmt.Sprintf("Opening %s", mmdb))

	db, err := geoip2.Open(mmdb)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	ip := net.ParseIP(ipraw)

	record, err := db.Country(ip)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if record.Traits.IsAnonymousProxy {
		verbose("Anonymous IP detected")
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
		response = strings.Join(output, ", ")
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
func downloadToFile(filepath string, uri string) error {
	debugURI := uri
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}
	q := u.Query()
	if q.Get("license_key") != "" {
		// remove key from debug
		debugURI = strings.ReplaceAll(debugURI, q.Get("license_key"), "xxxxxxxxxx")
	}

	verbose(fmt.Sprintf("Downloading %s", debugURI))

	// Get the data
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// IsFile returns if a path is a file
func isFile(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) || !info.Mode().IsRegular() {
		return false
	}

	return true
}

// IsDir returns whether a path is a directory
func isDir(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) || !info.IsDir() {
		return false
	}

	return true
}

// Verbose displays debug information with `-v`
func verbose(m string) {
	if verboseOutput {
		fmt.Println(m)
	}
}
