package main

import (
	"flag"
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"net"
	"os"
	"path"
	"strings"
)

// Flags
var (
	data_dir       = flag.String("d", "/usr/share/GeoIP", "database directory or file")
	country        = flag.Bool("c", false, "return country name")
	iso            = flag.Bool("i", false, "return country iso code")
	showhelp       = flag.Bool("h", false, "show help")
	verbose_output = flag.Bool("v", false, "verbose/debug output")
	showversion    = flag.Bool("V", false, "show version number")
	version        = "dev"
)

// URLs
const (
	db_update_url = "https://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.tar.gz"
	repo_url      = "https://github.com/axllent/goiplookup/releases"
	version_url   = "https://api.github.com/repos/axllent/goiplookup/releases/latest"
)

// Main function
func main() {

	flag.Parse()

	if *showversion {
		fmt.Println(fmt.Sprintf("Version: %s", version))
		os.Exit(1)
	}

	if len(flag.Args()) != 1 || *showhelp {
		Usage()
		os.Exit(1)
	}

	lookup := flag.Args()[0]

	if lookup == "db-update" {
		UpdateGeoLite2Country()
	} else {
		Lookup(lookup)
	}
}

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
		Debug(fmt.Sprintf("Ip search for: %s", addresses[0]))
		ipraw = addresses[0]
	} else {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fi, err := os.Stat(*data_dir)
	if err != nil {
		fmt.Println("Error: Directory does not exist", *data_dir)
		os.Exit(1)
	}

	switch mode := fi.Mode(); {
	case mode.IsDir(): // if data_dir is dir, append GeoLite2-Country.mmdb
		mmdb = path.Join(*data_dir, "GeoLite2-Country.mmdb")
	case mode.IsRegular():
		mmdb = *data_dir
	}

	Debug(fmt.Sprintf("Opening %s", mmdb))

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
		Debug("Anonymous IP detected")
		ciso = "A1"
		cname = "Anonymous Proxy"
	} else {
		ciso = record.Country.IsoCode
		cname = record.Country.Names["en"]
	}

	if *country || *iso {
		if *iso && ciso != "" {
			output = append(output, ciso)
		}
		if *country && cname != "" {
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

// Print the help function
var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-i] [-c] [-d <database directory>] <ipaddress|hostname|db-update>\n", os.Args[0])
	fmt.Println("\nGoiplookup uses the GeoLite2-Country database to find the Country that an IP address or hostname originates from.")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Fprintf(os.Stderr, "%s 8.8.8.8\t\t\tReturn the country ISO code and name\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s -d ~/GeoIP 8.8.8.8\t\tUse a different database directory\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s -i 8.8.8.8\t\t\tReturn just the country ISO code\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s -c 8.8.8.8\t\t\tReturn just the country name\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s db-update\t\t\tUpdate the GeoLite2-Country database (do not run more than once a month)\n", os.Args[0])
}

// Display debug information with `-v`
func Debug(m string) {
	if *verbose_output {
		fmt.Println(m)
	}
}
