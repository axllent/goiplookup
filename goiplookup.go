// Package main is the main application
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/axllent/ghru"
	flag "github.com/spf13/pflag"
)

// Flags
var (
	country       bool
	iso           bool
	showHelp      bool
	verboseOutput bool
	showVersion   bool
	dataDir       string
	licenseKey    string // GeoLite2 license key for updating
	batchFile     string // file containing batch IPs
	version       = "dev"
)

// Main function
func main() {

	p := "/usr/share/GeoIP"
	if isDir("/usr/local/share/GeoIP") {
		// alternate default path for OSX or custom
		p = "/usr/local/share/GeoIP"
	}
	flag.StringVarP(&dataDir, "dir", "d", p, "database directory or file")
	flag.StringVarP(&batchFile, "file", "f", "", "file containing IPs (one per line)")

	flag.BoolVarP(&country, "country", "c", false, "return country name")
	flag.BoolVarP(&iso, "iso", "i", false, "return country iso code")
	flag.BoolVarP(&showHelp, "help", "h", false, "show help")
	flag.BoolVarP(&verboseOutput, "verbose", "v", false, "verbose/debug output")
	flag.BoolVarP(&showVersion, "version", "V", false, "show version number")

	// parse flags
	flag.Parse()

	if batchFile != "" {
		file, err := os.Open(batchFile)
		if err != nil {
			fmt.Println("Error opening batch file:", err)
			os.Exit(1)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			ip := strings.TrimSpace(scanner.Text())
			if ip != "" {
				lookupAddr(ip)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading batch file:", err)
		}
		return
	}

	if len(flag.Args()) != 1 || showHelp {
		showUsage()
		return
	}

	lookup := flag.Args()[0]

	switch lookup {
	case "db-update":
		updateGeoLite2Country()
	case "db-update-city":
		updateGeoLite2City()
	case "self-update":
		rel, err := ghru.Update("axllent/goiplookup", "goiplookup", version)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Printf("Updated %s to version %s\n", os.Args[0], rel)
		os.Exit(0)
	default:
		lookupAddr(lookup)
	}
}

// ShowUsage prints the help function
var showUsage = func() {
	fmt.Printf("Usage: %s [-i] [-c] [-d <database directory>] [-f <file>] <ipaddress|hostname|db-update|db-update-city|self-update>\n", os.Args[0])
	fmt.Println("\nGoiplookup uses the GeoLite2-Country or GeoLite2-City database to find the Country or City that an IP address or hostname originates from.")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Printf("%s 8.8.8.8               # Return the country/city ISO code and name\n", os.Args[0])
	fmt.Printf("%s -d ~/GeoIP 8.8.8.8    # Use a different database directory\n", os.Args[0])
	fmt.Printf("%s -i 8.8.8.8            # Return just the country ISO code\n", os.Args[0])
	fmt.Printf("%s -c 8.8.8.8            # Return just the country name\n", os.Args[0])
	fmt.Printf("%s -f ips.txt            # Lookup a batch of IPs from a file, output JSON per line\n", os.Args[0])
	fmt.Printf("%s db-update             # Update the GeoLite2-Country database (do not run more than once a month)\n", os.Args[0])
	fmt.Printf("%s db-update-city        # Update the GeoLite2-City database (do not run more than once a month)\n", os.Args[0])
	fmt.Printf("%s self-update           # Update the GoIpLookup binary with the latest release\n", os.Args[0])
}
