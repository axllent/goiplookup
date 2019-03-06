package main

import (
	"encoding/json"
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// Github release json struct
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

// Download a URL to a file
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

// Fetch the latest release
func LatestRelease() (string, error) {
	resp, err := http.Get(release_url)
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

// Return a download URL based on OS & architecture
func GetUpdateURL() (string, error) {
	Verbose(fmt.Sprintf("Fetching %s", release_url))
	resp, err := http.Get(release_url)
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
		return "", fmt.Errorf("You already have the latest version: %s", version)
	}

	link_os := runtime.GOOS
	link_arch := runtime.GOARCH
	release_name := fmt.Sprintf("goiplookup_%s_%s_%s.bz2", result.TagName, link_os, link_arch)

	Verbose(fmt.Sprintf("Searching %s", release_name))

	for _, v := range result.Assets {
		if v.Name == release_name {
			Verbose(fmt.Sprintf("Found download URL %s", v.BrowserDownloadURL))
			return v.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("No downlodable update found for %s", release_name) // nothing found
}

// Replace one file with another
func ReplaceFile(dst string, src string) error {
	// open the source file for reading
	Verbose(fmt.Sprintf("Opening %s", src))
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// destination directory eg: /usr/local/bin
	dst_dir := filepath.Dir(dst)
	// binary filename eg: goiplookup
	binary_filename := filepath.Base(dst)
	// old tmp file name
	dst_old := fmt.Sprintf("%s.old", binary_filename)
	// new tmp file name
	dst_new := fmt.Sprintf("%s.new", binary_filename)
	// absolute path of new tmp file
	new_tmp_abs := filepath.Join(dst_dir, dst_new)
	// absolute path of old tmp file
	old_tmp_abs := filepath.Join(dst_dir, dst_old)

	// create the new file
	tmp_new, err := os.OpenFile(new_tmp_abs, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer tmp_new.Close()

	// copy new binary to <binary>.new
	Verbose(fmt.Sprintf("Copying %s to %s", src, new_tmp_abs))
	if _, err := io.Copy(tmp_new, source); err != nil {
		return err
	}

	// rename the current executable to <binary>.old
	Verbose(fmt.Sprintf("Renaming %s to %s", dst, old_tmp_abs))
	if err := os.Rename(dst, old_tmp_abs); err != nil {
		return err
	}

	// rename the <binary>.new to current executable
	Verbose(fmt.Sprintf("Renaming %s to %s", new_tmp_abs, dst))
	if err := os.Rename(new_tmp_abs, dst); err != nil {
		return err
	}

	// delete the old binary
	Verbose(fmt.Sprintf("Deleting %s", old_tmp_abs))
	if err := os.Remove(old_tmp_abs); err != nil {
		return err
	}

	// remove the src file
	Verbose(fmt.Sprintf("Deleting %s", src))
	if err := os.Remove(src); err != nil {
		return err
	}

	return nil
}

// Display debug information with `-v`
func Verbose(m string) {
	if *verbose_output {
		fmt.Println(m)
	}
}
