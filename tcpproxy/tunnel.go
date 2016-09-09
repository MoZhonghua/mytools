package main

import "net"
import "sync"

// import "errors"
// import "io"

type tunnel struct {
	lconn       net.Conn
	rconn       net.Conn
	key         []byte
	waitStopped sync.WaitGroup
}

func (t *tunnel) handshakeWithRemote() error {
	return nil
}

func pipe(r net.Conn, w net.Conn, stop *sync.WaitGroup) {
	defer stop.Done()

	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		remain := buf[:n]
		for len(remain) > 0 {
			n2, err := w.Write(remain)
			if err != nil {
				return
			}

			remain = remain[n2:]
		}
		if err != nil {
			w.(*net.TCPConn).CloseWrite()
			return
		}
	}
}

func (t *tunnel) loop() {
	t.waitStopped.Add(2)
	go pipe(t.lconn, t.rconn, &t.waitStopped)
	pipe(t.rconn, t.lconn, &t.waitStopped)

	t.waitStopped.Wait()
}
