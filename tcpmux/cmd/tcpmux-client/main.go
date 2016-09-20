package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/MoZhonghua/mytools/tcpmux"
	"github.com/MoZhonghua/mytools/util"
)

var listenPort int
var remoteAddr string
var id string

func servConn(c, s *net.TCPConn) {
	defer c.Close()
	defer s.Close()

	err := tcpmux.WriteHeader(id, s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to send id: %v\n", err)
		return
	}

	code2, err := tcpmux.ReadHeader(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to receive echo id: %v\n", err)
		return
	}
	if code2 != id {
		fmt.Fprintf(os.Stderr, "receive invalid echo id: %v\n", code2)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go tcpmux.Pipeline(c, s, &wg)
	go tcpmux.Pipeline(s, c, &wg)
	wg.Wait()
}

func main() {
	flag.IntVar(&listenPort, "p", 2233, "local listening port")
	flag.StringVar(&remoteAddr, "s", "", "tcpmux server address")
	flag.StringVar(&id, "t", "", "tcpmux authentication id")
	flag.Parse()

	if len(remoteAddr) == 0 {
		fmt.Fprintf(os.Stderr, "null server address\n")
		os.Exit(1)
	}

	if len(id) == 0 {
		fmt.Fprintf(os.Stderr, "null tcpmux authentication id\n")
		os.Exit(1)
	}

	l, err := net.Listen("tcp4", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}

	log.Printf("tcpmux client listens at %v\n", l.Addr())
	ips := util.GetIPList()
	for _, ip := range ips {
		log.Printf("%s:%d", ip, listenPort)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to accept: %v\n", err)
			continue
		}

		s, err := net.Dial("tcp", remoteAddr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect tcpmux server: %v\n", err)
			conn.Close()
			continue
		}

		log.Printf("%v -> %v\n", conn.RemoteAddr(), s.RemoteAddr())

		go servConn(conn.(*net.TCPConn), s.(*net.TCPConn))
	}
}
