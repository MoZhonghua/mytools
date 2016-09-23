package main

import (
    "log"
    "net/http"
	"flag"
	"fmt"

    "github.com/elazarl/goproxy"
)

func main() {
	var port int
	flag.IntVar(&port, "p", 8118, "port")
	flag.Parse()

    proxy := goproxy.NewProxyHttpServer()
    proxy.Verbose = true

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), proxy))
}
