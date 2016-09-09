package main

import "flag"

var (
	port           int
	remoteAddrList string
)

func main() {
	flag.IntVar(&port, "p", 3333, "listen port")
	flag.StringVar(&remoteAddrList, "r", "",
		"remote address, seperated with ',', e.g. 10.0.0.95:3306,10.0.0.96:3306")
	flag.Parse()

	if len(remoteAddrList) == 0 {
		flag.Usage()
		return
	}

	p, err := newProxy(port, remoteAddrList)
	if err != nil {
		panic(err)
	}
	p.serv()
}
