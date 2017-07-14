package main

import (
	"flag"
	"net"
	"os"
	"path"
	"runtime"

	"github.com/MoZhonghua/mytools/tcpproxy"
	log "github.com/Sirupsen/logrus"
)

var (
	adminAddr string
	db        string
	noLoad    bool
)

func getDefaultDatabaseFile() string {
	if runtime.GOOS == "windows" {
		return "data.db"
	} else {
		return "/var/lib/tcpproxy/data.db"
	}
}

func main() {
	flag.StringVar(&adminAddr, "m", "127.0.0.1:3333", "admin api address")
	flag.StringVar(&db, "d", getDefaultDatabaseFile(), "database to save mappings")
	flag.BoolVar(&noLoad, "n", false, "don't load targets from database when start")
	flag.Parse()

	pdir := path.Dir(db)
	err := os.MkdirAll(pdir, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatalf("failed to create db dir: %v", err)
	}

	s, err := tcpproxy.NewStore(db)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	p := tcpproxy.NewProxy()
	if !noLoad {
		list, err := s.GetAllPortMapping()
		if err != nil {
			log.Fatalf("failed to load port mapping list: %v", err)
		}

		for _, pm := range list {
			remoteAddr, err := net.ResolveTCPAddr("tcp4", pm.RemoteAddr)
			if err != nil {
				log.Printf("failed to resolve addr: %s - %v", pm.RemoteAddr, err)
				continue
			}

			err = p.AddPortMapping(pm.LocalPort, remoteAddr)
			if err != nil {
				log.Printf("failed to map :%d -> %s - %v",
					pm.LocalPort, pm.RemoteAddr, err)
				continue
			} else {
				log.Printf("map :%d -> %s OK", pm.LocalPort, pm.RemoteAddr)
				continue
			}
		}
	}

	d := tcpproxy.NewHttpd(p, s)
	l, err := net.Listen("tcp4", adminAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	err = d.Serv(l)
	if err != nil {
		log.Fatal(err)
	}
}
