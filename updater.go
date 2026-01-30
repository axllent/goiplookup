package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"

	"github.com/oschwald/geoip2-golang"
)

// UpdateGeoLite2Country updates GeoLite2-Country.mmdb
func updateGeoLite2Country() {

	key := os.Getenv("LICENSEKEY")
	if key == "" && licenseKey != "" {
		key = licenseKey
	}

	if key == "" {
		fmt.Println("Error: GeoIP License Key not set.\nPlease see https://github.com/axllent/goiplookup#database-updates")
		os.Exit(1)
	}

	dbUpdateURL := fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&license_key=%s&suffix=tar.gz", key)

	updateRequired, err := requiresDBUpdate(dbUpdateURL)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if !updateRequired {
		verbose("No database update available")
		os.Exit(0)
	}

	verbose("Updating GeoLite2-Country.mmdb")

	tmpDir := os.TempDir()
	gzFile := filepath.Join(tmpDir, "GeoLite2-Country.tar.gz")

	// check the output directory is writeable
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
			fmt.Println("Error: Cannot create", dataDir)
			os.Exit(1)
		}
	}

	if _, err := os.Stat(dataDir); err != nil {
		fmt.Println("Error: Cannot create", dataDir)
		os.Exit(1)
	}

	if err := downloadToFile(gzFile, dbUpdateURL); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := extractDatabaseFile(dataDir, gzFile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := os.Remove(gzFile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// get last-modified header to see if it is an update
func requiresDBUpdate(updateURL string) (bool, error) {
	dstFile := path.Join(dataDir, "GeoLite2-Country.mmdb")
	if !isFile(dstFile) {
		// missing local file, update
		return true, nil
	}

	info, err := os.Stat(dstFile)
	if err != nil {
		return false, err
	}

	lastModifiedLocal := info.ModTime()

	res, err := http.Head(updateURL)
	if err != nil {
		return false, err
	}

	lmHdr := res.Header.Get("last-modified")
	if lmHdr == "" {
		return false, errors.New("update server returned unexpected response")
	}

	lastModifiedRemote, err := time.Parse(time.RFC1123, lmHdr)
	if err != nil {
		return false, err
	}

	return lastModifiedRemote.After(lastModifiedLocal), nil
}

// func getLastModifiedFromHeader(h string) time.Time {
// 	var t time.Time
// 	if h == "" {
// 		return t
// 	}
// 	t, _ = time.Parse(time.RFC1123, h)
// 	return t
// }

// ExtractDatabaseFile extracts just the GeoLite2-Country.mmdb from the tar.gz
func extractDatabaseFile(dst string, tarGz string) error {
	verbose(fmt.Sprintf("Opening %s", tarGz))

	re, _ := regexp.Compile(`GeoLite2\-Country\.mmdb$`)

	r, err := os.Open(tarGz)
	if err != nil {
		return err
	}

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() { _ = gzr.Close() }()

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
				outFile := filepath.Join(dst, "GeoLite2-Country.mmdb")

				// tmpFile is used to first ensure the extracted database is valid before replacing the previous one
				tmpFile, err := os.CreateTemp("", "testDBFile")
				if err != nil {
					log.Fatal(err)
				}
				defer func() { _ = os.Remove(tmpFile.Name()) /* clean up*/ }()

				verbose(fmt.Sprintf("Copy GeoLite2-Country.mmdb to %s for testing", tmpFile.Name()))
				if _, err := io.Copy(tmpFile, tr); err != nil {
					return err
				}

				db, err := geoip2.Open(tmpFile.Name())
				if err != nil {
					return fmt.Errorf("downloaded GeoLite2-Country.mmdb database (%s) corrupt, aborting updating", tmpFile.Name())
				}
				_ = db.Close()

				verbose(fmt.Sprintf("Copy %s to %s", tmpFile.Name(), outFile))

				input, err := os.ReadFile(tmpFile.Name())
				if err != nil {
					return err
				}

				err = os.WriteFile(outFile, input, 0644)
				if err != nil {
					return err
				}
			}
		}
	}
}
