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
	logger      *log.Logger
}

func newPortMapping(localPort int, remoteAddr *net.TCPAddr,
	logger *log.Logger) *portMapping {
	m := &portMapping{
		localPort:  localPort,
		remoteAddr: remoteAddr,
		stopCh:     make(chan int),
		logger:     logger,
	}
	return m
}

func (m *portMapping) stop() {
	close(m.stopCh)
	m.waitStopped.Wait()
	m.logger.Printf("close port mapping :%d -> %v", m.localPort, m.remoteAddr)
}

func (m *portMapping) servLoop(l net.Listener) {
	defer l.Close()
	defer m.waitStopped.Done()
	connCh := make(chan net.Conn, 64)
	go func() {
		for {
			select {
			case <-m.stopCh:
				return
			default:
			}
			conn, err := l.Accept()
			if err != nil {
				m.logger.Printf("%v", err)
				continue
			}
			m.logger.Printf("new connection @ %v", conn.LocalAddr())
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

	m.logger.Printf("new port mapping :%d -> %v", m.localPort, m.remoteAddr)

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

	m.logger.Printf("pipe %v -> %v ", l.LocalAddr(), r.RemoteAddr())

	var wg sync.WaitGroup
	wg.Add(2)
	go pipe(l, r, &wg)
	go pipe(r, l, &wg)
	wg.Wait()
	return nil
}
