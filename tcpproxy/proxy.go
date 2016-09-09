package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type proxy struct {
	listenAddr string
	remoteAddr []*net.TCPAddr
}

func newProxy(port int, remoteAddr string) (*proxy, error) {
	addrList := strings.Split(remoteAddr, ",")
	if len(addrList) == 0 {
		return nil, errors.New("empty remote address")
	}

	ra := make([]*net.TCPAddr, len(addrList))
	for i, v := range addrList {
		r, e := net.ResolveTCPAddr("tcp", v)
		if e != nil {
			return nil, e
		}
		ra[i] = r
	}

	return &proxy{
		listenAddr: fmt.Sprintf(":%d", port),
		remoteAddr: ra,
	}, nil
}

func (p *proxy) serv() {
	ln, err := net.Listen("tcp4", p.listenAddr)
	if err != nil {
		panic(err)
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go p.connectRemoteAndServ(conn)

	}
}

func (p *proxy) connectRemote() (net.Conn, error) {
	for _, addr := range p.remoteAddr {
		// fmt.Println("try to connect: ", addr.String())
		rconn, err := net.DialTimeout("tcp", addr.String(), time.Second*3)
		if err == nil {
			return rconn, err
		} else {
			fmt.Println(err)
		}
	}

	return nil, errors.New("can't connect to remote")
}

func (p *proxy) connectRemoteAndServ(lconn net.Conn) {
	defer lconn.Close()

	rconn, err := p.connectRemote()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rconn.Close()

	t := &tunnel{
		lconn: lconn,
		rconn: rconn,
		key:   []byte("1111"),
	}

	err = t.handshakeWithRemote()
	if err != nil {
		fmt.Println(err)
		return
	}

	t.loop()
}
