package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/MoZhonghua/mytools/tunnel"
	"github.com/MoZhonghua/mytools/util"
)

var listenPort int
var remoteAddr string
var code string

func servConn(c, s *net.TCPConn) {
	defer c.Close()
	defer s.Close()

	err := tunnel.WriteCode(code, s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to send code: %v\n", err)
		return
	}

	code2, err := tunnel.ReadCode(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to receive echo code: %v\n", err)
		return
	}
	if code2 != code {
		fmt.Fprintf(os.Stderr, "receive invalid echo code: %v\n", code2)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go tunnel.Pipeline(c, s, &wg)
	go tunnel.Pipeline(s, c, &wg)
	wg.Wait()
}

func main() {
	flag.IntVar(&listenPort, "p", 7421, "local listening port")
	flag.StringVar(&remoteAddr, "s", "", "tunnel server address")
	flag.StringVar(&code, "c", "", "tunnel authentication code")
	flag.Parse()

	if len(remoteAddr) == 0 {
		fmt.Fprintf(os.Stderr, "null server address\n")
		os.Exit(1)
	}

	if len(code) == 0 {
		fmt.Fprintf(os.Stderr, "null tunnel authentication code\n")
		os.Exit(1)
	}

	l, err := net.Listen("tcp4", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}

	log.Printf("tunnel client listens at %v\n", l.Addr())
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
			fmt.Fprintf(os.Stderr, "failed to connect tunnel server: %v\n", err)
			conn.Close()
			continue
		}

		log.Printf("%v -> %v\n", conn.RemoteAddr(), s.RemoteAddr())

		go servConn(conn.(*net.TCPConn), s.(*net.TCPConn))
	}
}
