package main

import "flag"

var (
	listenAddr     string
	remoteAddrList string
)

func main() {
	flag.StringVar(&listenAddr, "listen", ":3306", "listen address")
	flag.StringVar(&remoteAddrList, "remote", "",
		"remote address, seperated with ',', e.g. 10.0.0.95:3306,10.0.0.96:3306")
	flag.Parse()

	if len(listenAddr) == 0 || len(remoteAddrList) == 0 {
		flag.Usage()
		return
	}

	p, _ := newProxy(listenAddr, remoteAddrList)
	p.serv()
}
