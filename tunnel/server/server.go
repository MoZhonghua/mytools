package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/MoZhonghua/mytools/tunnel"
	"github.com/MoZhonghua/mytools/util"
)

var listenPort int

func servConn(c *net.TCPConn) {
	defer c.Close()

	code, err := tunnel.ReadCode(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to receive echo code: %v\n", err)
		return
	}

	upstream, found := upstreams[string(code)]
	if found == false {
		fmt.Fprintf(os.Stderr, "invalid code: %s\n", string(code))
		return
	}

	s, err := net.Dial("tcp", upstream.Addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect tunnel server: %v\n", err)
		return
	}
	defer s.Close()

	// log.Println("done read code")
	err = tunnel.WriteCode(code, c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to send code: %v\n", err)
		return
	}

	// log.Println("done write echo code")
	log.Printf("%v -> %v", c.RemoteAddr(), s.RemoteAddr())

	var wg sync.WaitGroup
	wg.Add(2)
	go tunnel.Pipeline(c, s.(*net.TCPConn), &wg)
	go tunnel.Pipeline(s.(*net.TCPConn), c, &wg)
	wg.Wait()
}

type UpstreamServer struct {
	Code string
	Addr string
}

var upstreams = make(map[string]*UpstreamServer)

func main() {
	flag.IntVar(&listenPort, "p", 7422, "local listening port")
	flag.Parse()

	servers := flag.Args()

	if len(servers) == 0 {
		fmt.Fprintf(os.Stderr, "null server address\n")
		os.Exit(1)
	}

	for _, s := range servers {
		parts := strings.Split(s, ":")
		if len(parts) != 3 {
			fmt.Fprintf(os.Stderr, "invalid server address: %s\n", s)
			os.Exit(1)
		}
		code := parts[0]
		addr := fmt.Sprintf("%s:%s", parts[1], parts[2])

		upstreams[code] = &UpstreamServer{code, addr}
	}

	l, err := net.Listen("tcp4", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}

	log.Printf("tunnel server listens at %v\n", l.Addr())
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

		go servConn(conn.(*net.TCPConn))
	}
}
