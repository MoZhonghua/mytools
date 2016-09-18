package main

import (
	"flag"
	"log"
	"os"

	"github.com/MoZhonghua/mytools/tcpproxy"
)

var (
	port int
)

func main() {
	flag.IntVar(&port, "p", 3333, "managment port")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	p := tcpproxy.NewProxy(logger)
	d := tcpproxy.NewHttpd(p, logger)

	err := d.Serv(port)
	if err != nil {
		logger.Fatal(err)
	}
}
