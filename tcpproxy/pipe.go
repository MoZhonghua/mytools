package tcpproxy

import (
	"net"
	"sync"
)

func pipe(r net.Conn, w net.Conn, done *sync.WaitGroup) {
	defer done.Done()
	w.(*net.TCPConn).SetKeepAlive(true)
	r.(*net.TCPConn).SetKeepAlive(true)

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
