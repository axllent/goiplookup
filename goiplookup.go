package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/axllent/ghru"
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
	licenseKey    string // GeoLite2 license key for updating
	version       = "dev"
)

// URLs
const (
	releaseURL = "https://api.github.com/repos/axllent/goiplookup/releases/latest"
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

		latest, _, _, err := ghru.Latest("axllent/goiplookup", "goiplookup")
		if err == nil && ghru.GreaterThan(latest, version) {
			fmt.Printf("Update available: %s\nRun `%s self-update` to update\n", latest, os.Args[0])
		}
		os.Exit(0)
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
		rel, err := ghru.Update("axllent/goiplookup", "goiplookup", version)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Updated %s to version %s\n", os.Args[0], rel)
		os.Exit(0)
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
