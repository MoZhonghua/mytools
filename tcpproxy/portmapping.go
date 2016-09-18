package tcpproxy

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type portMapping struct {
	localPort   int
	remoteAddr  *net.TCPAddr
	stopCh      chan int
	waitStopped sync.WaitGroup
}

func newPortMapping(localPort int, remoteAddr *net.TCPAddr) *portMapping {
	m := &portMapping{
		localPort:  localPort,
		remoteAddr: remoteAddr,
		stopCh:     make(chan int),
	}
	return m
}

func (m *portMapping) stop() {
	close(m.stopCh)
	m.waitStopped.Wait()
}

func (m *portMapping) servLoop(l net.Listener) {
	defer l.Close()
	defer m.waitStopped.Done()
	connCh := make(chan net.Conn, 64)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Printf("accept() error: %v", err)
				continue
			}
			connCh <- conn
		}
	}()

	for {
		select {
		case <-m.stopCh:
			return
		case conn := <-connCh:
			if conn == nil {
				return
			}
			go m.handleConn(conn)
		}
	}
}

func (m *portMapping) start() error {
	m.waitStopped.Add(1)
	l, err := net.Listen("tcp4", fmt.Sprintf(":%d", m.localPort))
	if err != nil {
		return err
	}

	go m.servLoop(l)
	return nil
}

func (m *portMapping) handleConn(l net.Conn) error {
	defer l.Close()
	r, err := net.Dial("tcp4", m.remoteAddr.String())
	if err != nil {
		return err
	}
	defer r.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go pipe(l, r, &wg)
	go pipe(r, l, &wg)
	wg.Wait()
	return nil
}
