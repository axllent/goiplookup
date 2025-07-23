package main

import (
	"encoding/json"
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
	var cityName string
	var subdivisionName string
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
	case mode.IsDir(): // if dataDir is dir, try City then Country
		cityPath := path.Join(dataDir, "GeoLite2-City.mmdb")
		countryPath := path.Join(dataDir, "GeoLite2-Country.mmdb")
		cityExists := isFile(cityPath)
		countryExists := isFile(countryPath)
		if !cityExists && !countryExists {
			fmt.Printf("Error: No GeoLite2-City.mmdb or GeoLite2-Country.mmdb found in %s\n", dataDir)
			os.Exit(1)
		}
		if cityExists && countryExists {
			dbCity, err := geoip2.Open(cityPath)
			if err != nil {
				fmt.Println("Error opening City database:", err)
			} else {
				defer func() { _ = dbCity.Close() }()
				ip := net.ParseIP(ipraw)
				record, err := dbCity.City(ip)
				if err != nil {
					fmt.Println("Error:", err)
				} else {
					cityName = record.City.Names["en"]
					countryName := record.Country.Names["en"]
					result := map[string]interface{}{
						"ip": ipraw,
						"location": map[string]string{
							"city":    cityName,
							"country": countryName,
						},
					}
					jsonBytes, _ := json.MarshalIndent(result, "", "  ")
					fmt.Println(string(jsonBytes))
				}
			}
			return
		}
		if cityExists {
			mmdb = cityPath
		} else {
			mmdb = countryPath
		}
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

	if strings.Contains(mmdb, "City.mmdb") {
		record, err := db.City(ip)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		ciso = record.Country.IsoCode
		cname = record.Country.Names["en"]
		cityName = record.City.Names["en"]
		if len(record.Subdivisions) > 0 {
			subdivisionName = record.Subdivisions[0].Names["en"]
		}
		if cityName != "" {
			output = append(output, cityName)
		}
		if subdivisionName != "" {
			output = append(output, subdivisionName)
		}
		if ciso != "" {
			output = append(output, ciso)
		}
		if cname != "" {
			output = append(output, cname)
		}
		response = strings.Join(output, ", ")
	} else {
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
	}

	if strings.TrimSpace(response) == "" {
		response = "GeoIP Country Edition: IP Address not found"
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
