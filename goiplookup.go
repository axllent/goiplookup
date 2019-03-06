package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

// Flags
var (
	country        = flag.Bool("c", false, "return country name")
	iso            = flag.Bool("i", false, "return country iso code")
	showhelp       = flag.Bool("h", false, "show help")
	verbose_output = flag.Bool("v", false, "verbose/debug output")
	showversion    = flag.Bool("V", false, "show version number")
	version        = "dev"
)

// we set this in `main()` based on OS
var data_dir (*string)

// URLs
const (
	db_update_url = "https://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.tar.gz"
	repo_url      = "https://github.com/axllent/goiplookup/releases"
	release_url   = "https://api.github.com/repos/axllent/goiplookup/releases/latest"
)

// Main function
func main() {
	// alternate default path for OSX
	if runtime.GOOS == "darwin" {
		data_dir = flag.String("d", "/usr/local/share/GeoIP", "database directory or file")
	} else {
		data_dir = flag.String("d", "/usr/share/GeoIP", "database directory or file")
        }

        // parse flags
	flag.Parse()

	if *showversion {
		fmt.Println(fmt.Sprintf("Current: %s", version))
		latest, err := LatestRelease()
		if err == nil && version != latest {
			fmt.Println(fmt.Sprintf("Latest:  %s", latest))
		}
		return
	}

	if len(flag.Args()) != 1 || *showhelp {
		ShowUsage()
		return
	}

	lookup := flag.Args()[0]

	if lookup == "db-update" {
		// update database
		UpdateGeoLite2Country()
	} else if lookup == "self-update" {
		// update app if needed
		SelfUpdate()
	} else {
		// lookup ip/hostname
		Lookup(lookup)
	}
}

// Print the help function
var ShowUsage = func() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-i] [-c] [-d <database directory>] <ipaddress|hostname|db-update|self-update>\n", os.Args[0])
	fmt.Println("\nGoiplookup uses the GeoLite2-Country database to find the Country that an IP address or hostname originates from.")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Fprintf(os.Stderr, "%s 8.8.8.8\t\t\tReturn the country ISO code and name\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s -d ~/GeoIP 8.8.8.8\t\tUse a different database directory\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s -i 8.8.8.8\t\t\tReturn just the country ISO code\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s -c 8.8.8.8\t\t\tReturn just the country name\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s db-update\t\t\tUpdate the GeoLite2-Country database (do not run more than once a month)\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s self-update\t\t\tUpdate the GoIpLookup binary with the latest release\n", os.Args[0])
}
