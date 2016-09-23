package main

import (
	"flag"
	"log"
	"net"
	"os"
	"path"

	"github.com/MoZhonghua/mytools/tcpproxy"
)

var (
	port   int
	db     string
	noLoad bool
)

var logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

func main() {
	flag.IntVar(&port, "m", 3333, "admin port")
	flag.StringVar(&db, "d", "/var/lib/tcpproxy/mappings.db",
		"database to sync mappings")
	flag.BoolVar(&noLoad, "n", false, "don't load targets from database when start")
	flag.Parse()

	pdir := path.Dir(db)
	err := os.MkdirAll(pdir, 0755)
	if err != nil && !os.IsExist(err) {
		logger.Fatalf("failed to create db dir: %v", err)
	}

	s, err := tcpproxy.NewStore(db)
	if err != nil {
		logger.Fatalf("failed to open db: %v", err)
	}

	p := tcpproxy.NewProxy(logger)
	if !noLoad {
		list, err := s.GetAllPortMapping()
		if err != nil {
			logger.Fatalf("failed to load port mapping list: %v", err)
		}

		for _, pm := range list {
			remoteAddr, err := net.ResolveTCPAddr("tcp4", pm.RemoteAddr)
			if err != nil {
				logger.Printf("failed to resolve addr: %s - %v", pm.RemoteAddr, err)
				continue
			}

			err = p.AddPortMapping(pm.LocalPort, remoteAddr)
			if err != nil {
				logger.Printf("failed to map :%d -> %s - %v",
					pm.LocalPort, pm.RemoteAddr, err)
				continue
			} else {
				logger.Printf("map :%d -> %s OK", pm.LocalPort, pm.RemoteAddr)
				continue
			}
		}
	}

	d := tcpproxy.NewHttpd(p, s, logger)
	err = d.Serv(port)
	if err != nil {
		logger.Fatal(err)
	}
}
