package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

// Repository Struct for Github release json
type Repository struct {
	Assets []struct {
		BrowserDownloadURL string `json:"browser_download_url"`
		CreatedAt          string `json:"created_at"`
		ID                 int64  `json:"id"`
		Name               string `json:"name"`
		Size               int64  `json:"size"`
	} `json:"assets"`
	Name        string `json:"name"`
	Prerelease  bool   `json:"prerelease"`
	PublishedAt string `json:"published_at"`
	TagName     string `json:"tag_name"`
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

// LatestRelease fetches the latest release
func LatestRelease() (string, error) {
	resp, err := http.Get(releaseURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var result Repository

	json.Unmarshal(body, &result)

	return result.TagName, nil
}

// GetUpdateURL returns a download URL based on OS & architecture
func GetUpdateURL() (string, error) {
	Verbose(fmt.Sprintf("Fetching %s", releaseURL))
	resp, err := http.Get(releaseURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var result Repository

	json.Unmarshal(body, &result)

	Verbose(fmt.Sprintf("Latest release is %s", result.TagName))
	if version == result.TagName {
		return "", fmt.Errorf("You already have the latest version (%s)", version)
	}

	linkOS := runtime.GOOS
	linkArch := runtime.GOARCH
	releaseName := fmt.Sprintf("goiplookup_%s_%s_%s.bz2", result.TagName, linkOS, linkArch)

	Verbose(fmt.Sprintf("Searching %s", releaseName))

	for _, v := range result.Assets {
		if v.Name == releaseName {
			Verbose(fmt.Sprintf("Found download URL %s", v.BrowserDownloadURL))
			return v.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("No downlodable update found for %s", releaseName) // nothing found
}

// ReplaceFile replaces one file with another
func ReplaceFile(dst string, src string) error {
	// open the source file for reading
	Verbose(fmt.Sprintf("Opening %s", src))
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// destination directory eg: /usr/local/bin
	dstDir := filepath.Dir(dst)
	// binary filename eg: goiplookup
	binaryFilename := filepath.Base(dst)
	// old tmp file name
	dstOld := fmt.Sprintf("%s.old", binaryFilename)
	// new tmp file name
	dstNew := fmt.Sprintf("%s.new", binaryFilename)
	// absolute path of new tmp file
	newTmpAbs := filepath.Join(dstDir, dstNew)
	// absolute path of old tmp file
	oldTmpAbs := filepath.Join(dstDir, dstOld)

	// create the new file
	tmpNew, err := os.OpenFile(newTmpAbs, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer tmpNew.Close()

	// copy new binary to <binary>.new
	Verbose(fmt.Sprintf("Copying %s to %s", src, newTmpAbs))
	if _, err := io.Copy(tmpNew, source); err != nil {
		return err
	}

	// rename the current executable to <binary>.old
	Verbose(fmt.Sprintf("Renaming %s to %s", dst, oldTmpAbs))
	if err := os.Rename(dst, oldTmpAbs); err != nil {
		return err
	}

	// rename the <binary>.new to current executable
	Verbose(fmt.Sprintf("Renaming %s to %s", newTmpAbs, dst))
	if err := os.Rename(newTmpAbs, dst); err != nil {
		return err
	}

	// delete the old binary
	Verbose(fmt.Sprintf("Deleting %s", oldTmpAbs))
	if err := os.Remove(oldTmpAbs); err != nil {
		return err
	}

	// remove the src file
	Verbose(fmt.Sprintf("Deleting %s", src))
	if err := os.Remove(src); err != nil {
		return err
	}

	return nil
}

// Verbose displays debug information with `-v`
func Verbose(m string) {
	if verboseoutput {
		fmt.Println(m)
	}
}
