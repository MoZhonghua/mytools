package main

import (
	"flag"
	"log"
	"os"

	"github.com/MoZhonghua/mytools/tcpmux"
)

var (
	servicePort int
	adminPort   int
)

func main() {
	flag.IntVar(&servicePort, "p", 6731, "service port")
	flag.IntVar(&adminPort, "m", 6732, "admin port")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	m := tcpmux.NewTcpMux(logger)
	err := m.Start(servicePort)
	if err != nil {
		logger.Fatal(err)
	}

	d := tcpmux.NewHttpd(m, logger)
	err = d.Serv(adminPort)
	if err != nil {
		logger.Fatal(err)
	}
}
