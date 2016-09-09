package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"

	"github.com/gorilla/context"
	"github.com/keep94/weblogs"
)

func getIPList() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	ips := make([]string, 0)
	for _, i := range ifaces {
		if i.Flags&net.FlagUp == 0 {
			continue
		}

		if i.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}

			ips = append(ips, ip.String())
		}
	}

	return ips
}

func main() {
	var port int
	var rootdir string

	flag.StringVar(&rootdir, "root", ".", "root directory")
	flag.IntVar(&port, "port", 80, "listening port")
	flag.Parse()

	absDir, err := filepath.Abs(rootdir)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Serve %s at :%d ...\n", absDir, port)
	for _, ip := range getIPList() {
		log.Printf("http://%s:%d", ip, port)
	}
	http.Handle("/", http.FileServer(http.Dir(absDir)))
	handler := context.ClearHandler(weblogs.Handler(http.DefaultServeMux))
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
	if err != nil {
		log.Fatal(err)
	}
}
