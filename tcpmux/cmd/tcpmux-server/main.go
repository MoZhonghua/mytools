package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/MoZhonghua/mytools/tcpmux"
)

var (
	servicePort int
	adminPort   int
	db          string
	noLoad      bool
)

var logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

func main() {
	flag.IntVar(&servicePort, "p", 6731, "service port")
	flag.IntVar(&adminPort, "m", 6732, "admin port")
	flag.StringVar(&db, "d", "/var/lib/tcpmux/targets.db",
		"database to sync targets")
	flag.BoolVar(&noLoad, "n", false, "don't load targets from database when start")
	flag.Parse()

	pdir := path.Dir(db)
	err := os.MkdirAll(pdir, 0755)
	if err != nil && !os.IsExist(err) {
		logger.Fatalf("failed to create db dir: %v", err)
	}

	s, err := tcpmux.NewStore(db)
	if err != nil {
		logger.Fatalf("failed to open db: %v", err)
	}

	m := tcpmux.NewTcpMux(logger)
	if !noLoad {
		list, err := s.GetAllTarget()
		if err != nil {
			logger.Fatalf("failed to load target list: %v", err)
		}
		for _, pm := range list {
			err = m.AddTarget(pm.Id, pm.Target)
			if err != nil {
				logger.Printf("failed to map %s -> %s - %v", pm.Id, pm.Target, err)
				continue
			} else {
				logger.Printf("map %s -> %s OK", pm.Id, pm.Target)
				continue
			}
		}
	}

	err = m.Start(servicePort)
	if err != nil {
		logger.Fatal(err)
	}

	d := tcpmux.NewHttpd(m, s, logger)
	err = d.Serv(adminPort)
	if err != nil {
		logger.Fatal(err)
	}
}
