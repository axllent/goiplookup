package main

import (
	"fmt"
	"os"
	"runtime"

	flag "github.com/spf13/pflag"
)

// Flags
var (
	country       bool
	iso           bool
	showhelp      bool
	verboseoutput bool
	showversion   bool
	dataDir       string
	version       = "dev"
)

// we set this in `main()` based on OS
// var dataDir (*string)

// URLs
const (
	dbUpdateURL = "https://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.tar.gz"
	releaseURL  = "https://api.github.com/repos/axllent/goiplookup/releases/latest"
)

// Main function
func main() {
	// alternate default path for OSX
	if runtime.GOOS == "darwin" {
		flag.StringVarP(&dataDir, "dir", "d", "/usr/local/share/GeoIP", "database directory or file")
	} else {
		flag.StringVarP(&dataDir, "dir", "d", "/usr/share/GeoIP", "database directory or file")
	}

	flag.BoolVarP(&country, "country", "c", false, "return country name")
	flag.BoolVarP(&iso, "iso", "i", false, "return country iso code")
	flag.BoolVarP(&showhelp, "help", "h", false, "show help")
	flag.BoolVarP(&verboseoutput, "verbose", "v", false, "verbose/debug output")
	flag.BoolVarP(&showversion, "version", "V", false, "show version number")

	// parse flags
	flag.Parse()

	if showversion {
		fmt.Println(fmt.Sprintf("Version %s", version))
		latest, err := LatestRelease()
		if err == nil && version != latest {
			fmt.Println(fmt.Sprintf("Version %s available", latest))
			if _, err := GetUpdateURL(); err == nil {
				fmt.Println(fmt.Sprintf("Run `%s self-update` to update", os.Args[0]))
			}
		} else {
			fmt.Println("You have the latest version")
		}
		return
	}

	if len(flag.Args()) != 1 || showhelp {
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

// ShowUsage prints the help function
var ShowUsage = func() {
	fmt.Printf("Usage: %s [-i] [-c] [-d <database directory>] <ipaddress|hostname|db-update|self-update>\n", os.Args[0])
	fmt.Println("\nGoiplookup uses the GeoLite2-Country database to find the Country that an IP address or hostname originates from.")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Printf("%s 8.8.8.8               # Return the country ISO code and name\n", os.Args[0])
	fmt.Printf("%s -d ~/GeoIP 8.8.8.8    # Use a different database directory\n", os.Args[0])
	fmt.Printf("%s -i 8.8.8.8            # Return just the country ISO code\n", os.Args[0])
	fmt.Printf("%s -c 8.8.8.8            # Return just the country name\n", os.Args[0])
	fmt.Printf("%s db-update             # Update the GeoLite2-Country database (do not run more than once a month)\n", os.Args[0])
	fmt.Printf("%s self-update           # Update the GoIpLookup binary with the latest release\n", os.Args[0])
}
