package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// global flags
var verbose = flag.Bool("v", false, "verbose output")
var country = flag.Bool("c", false, "return country name")
var iso = flag.Bool("i", false, "return country iso code")
var data_dir = flag.String("d", "/usr/share/GeoIP", "database directory or file")

/**
 * Main function
 */
func main() {
	showhelp := flag.Bool("h", false, "show help")
	flag.Parse()

	if len(flag.Args()) != 1 || *showhelp {
		Usage()
		os.Exit(1)
	}

	lookup := flag.Args()[0]

	if lookup == "update" {
		UpdateGeoLite2Country()
	} else {
		Lookup(lookup)
	}
}

/**
 * Lookup ip string
 */
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
	case mode.IsDir():
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

func UpdateGeoLite2Country() {
	url := "https://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.tar.gz"

	Debug("Updating GeoLite2-Country.mmdb")

	// check the output directory is writeable
	if _, err := os.Stat(*data_dir); os.IsNotExist(err) {
		os.MkdirAll(*data_dir, os.ModePerm)
	}

	_, err := os.Stat(*data_dir)
	if err != nil {
		fmt.Println("Error: Cannot create", *data_dir)
		os.Exit(1)
	}

	if err := DownloadFile("/tmp/GeoLite2-Country.tar.gz", url); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := ExtractDatabase(*data_dir, "/tmp/GeoLite2-Country.tar.gz"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := os.Remove("/tmp/GeoLite2-Country.tar.gz"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func DownloadFile(filepath string, url string) error {
	// Get the data
	Debug(fmt.Sprintf("Downloading %s", url))
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

// Extract the database from the tar.gz
func ExtractDatabase(dst string, targz string) error {
	Debug(fmt.Sprintf("Opening %s", targz))
	re, _ := regexp.Compile(`GeoLite2\-Country\.mmdb$`)

	r, err := os.Open(targz)
	if err != nil {
		return err
	}

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// check the file type
		switch header.Typeflag {

		case tar.TypeReg:

			if re.Match([]byte(target)) {
				outfile := filepath.Join(dst, "GeoLite2-Country.mmdb")

				f, err := os.OpenFile(outfile, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil {
					return err
				}

				Debug(fmt.Sprintf("Copy GeoLite2-Country.mmdb to %s", outfile))
				if _, err := io.Copy(f, tr); err != nil {
					return err
				}

				f.Close()
			}
		}
	}
}

/**
 * Print the help function
 */
var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-i] [-c] [-d <database directory>] <ipaddress|hostname|update>\n", os.Args[0])
	flag.PrintDefaults()
}

/**
 * Return debug information `-v`
 */
func Debug(m string) {
	if *verbose {
		fmt.Println(m)
	}
}
