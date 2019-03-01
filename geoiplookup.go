package main

import (
    "flag"
    "fmt"
    "github.com/oschwald/geoip2-golang"
    "net"
    "os"
    "path"
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

    Lookup(lookup)
}

/**
 * Lookup ip string
 */
func Lookup(lookup string) {

    var ciso     string
    var cname    string
    var mmdb     string
    var output []string
    var response string
    var ipraw    string

    // convert to ip if hostname
    addresses, err := net.LookupHost(lookup)

    if len(addresses) > 0 {
        Debug(fmt.Sprintf("Ip search for: %s", addresses[0]))
        ipraw = addresses[0];
    } else {
        fmt.Println("Error:", err)
        os.Exit(1)
    }

    fi, err := os.Stat(*data_dir)
    if err != nil {
        fmt.Println("Error: File does not exist", *data_dir)
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
        ciso  = "A1"
        cname = "Anonymous Proxy"
    } else {
        ciso  = record.Country.IsoCode
        cname = record.Country.Names["en"]
    }

    if ( *country || *iso ) {
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

/**
 * Print the help function
 */
var Usage = func() {
    fmt.Fprintf(os.Stderr, "Usage: %s [-i|-c] <ipaddress|hostname>\n", os.Args[0])
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
