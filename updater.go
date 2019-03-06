package main

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

// Update GeoLite2-Country.mmdb
func UpdateGeoLite2Country() {
	Verbose("Updating GeoLite2-Country.mmdb")

	tmp_dir := os.TempDir()
	gzfile := filepath.Join(tmp_dir, "GeoLite2-Country.tar.gz")

	// check the output directory is writeable
	if _, err := os.Stat(*data_dir); os.IsNotExist(err) {
		os.MkdirAll(*data_dir, os.ModePerm)
	}

	if _, err := os.Stat(*data_dir); err != nil {
		fmt.Println("Error: Cannot create", *data_dir)
		os.Exit(1)
	}

	if err := DownloadToFile(gzfile, db_update_url); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := ExtractDatabaseFile(*data_dir, gzfile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := os.Remove(gzfile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Extract just the GeoLite2-Country.mmdb from the tar.gz
func ExtractDatabaseFile(dst string, targz string) error {
	Verbose(fmt.Sprintf("Opening %s", targz))

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

				Verbose(fmt.Sprintf("Copy GeoLite2-Country.mmdb to %s", outfile))

				if _, err := io.Copy(f, tr); err != nil {
					return err
				}

				f.Close()
			}
		}
	}
}

// Built-in updater
func SelfUpdate() {
	tmp_dir := os.TempDir()
	bz2file := filepath.Join(tmp_dir, "goiplookup.bz2")
        newexec := filepath.Join(tmp_dir, "goiplookup.tmp")

	download_url, err := GetUpdateURL()
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}

	if err := DownloadToFile(bz2file, download_url); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	Verbose(fmt.Sprintf("Opening %s", bz2file))
	f, err := os.OpenFile(bz2file, 0, 0)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}
	defer f.Close()

	// create a bzip2 reader
	br := bzip2.NewReader(f)

	// write the file
	out, err := os.OpenFile(newexec, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}

	Verbose(fmt.Sprintf("Extracting %s", newexec))

	_, err = io.Copy(out, br)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}

	// replace os.Args[0] with new file
	// cannot overwrite open file so rename then delete
        // get executable's absolute path
        oldexec, _ := os.Readlink("/proc/self/exe")

        err = ReplaceFile(oldexec, newexec)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}

	// remove the src file
	Verbose(fmt.Sprintf("Deleting %s", bz2file))
	if err := os.Remove(bz2file); err != nil {
		fmt.Println(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}
}
