package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/oschwald/geoip2-golang"
)

// UpdateGeoLite2Country updates GeoLite2-Country.mmdb
func UpdateGeoLite2Country() {

	key := os.Getenv("LICENSEKEY")
	if key == "" && licenseKey != "" {
		key = licenseKey
	}

	if key == "" {
		fmt.Println("Error: GeoIP License Key not set.\nPlease see https://github.com/axllent/goiplookup#database-updates")
		os.Exit(1)
	}

	dbUpdateURL := fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&license_key=%s&suffix=tar.gz", key)

	Verbose("Updating GeoLite2-Country.mmdb")

	tmpDir := os.TempDir()
	gzFile := filepath.Join(tmpDir, "GeoLite2-Country.tar.gz")

	// check the output directory is writeable
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.MkdirAll(dataDir, os.ModePerm)
	}

	if _, err := os.Stat(dataDir); err != nil {
		fmt.Println("Error: Cannot create", dataDir)
		os.Exit(1)
	}

	if err := DownloadToFile(gzFile, dbUpdateURL); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := ExtractDatabaseFile(dataDir, gzFile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := os.Remove(gzFile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// ExtractDatabaseFile extracts just the GeoLite2-Country.mmdb from the tar.gz
func ExtractDatabaseFile(dst string, tarGz string) error {
	Verbose(fmt.Sprintf("Opening %s", tarGz))

	re, _ := regexp.Compile(`GeoLite2\-Country\.mmdb$`)

	r, err := os.Open(tarGz)
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
				outFile := filepath.Join(dst, "GeoLite2-Country.mmdb")

				// tmpFile is used to first ensure the extracted database is valid before replacing the previous one
				tmpFile, err := os.CreateTemp("", "testDBFile")
				if err != nil {
					log.Fatal(err)
				}
				defer os.Remove(tmpFile.Name()) // clean up

				Verbose(fmt.Sprintf("Copy GeoLite2-Country.mmdb to %s for testing", tmpFile.Name()))
				if _, err := io.Copy(tmpFile, tr); err != nil {
					return err
				}

				db, err := geoip2.Open(tmpFile.Name())
				if err != nil {
					return fmt.Errorf("Downloaded GeoLite2-Country.mmdb database (%s) corrupt, aborting updating", tmpFile.Name())
				}
				db.Close()

				Verbose(fmt.Sprintf("Copy %s to %s", tmpFile.Name(), outFile))

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
