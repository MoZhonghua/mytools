package tcpmux

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

var (
	ErrIdNotFound = errors.New("id not found")
)

type TargetInfo struct {
	Id     string `json:"id"`
	Target string `json:"target"`
}

type Tcpmux struct {
	mu      sync.Mutex
	targets map[string]*TargetInfo

	connCh chan net.Conn

	stopCh      chan int
	waitStopped sync.WaitGroup
	logger      *log.Logger
}

func NewTcpMux(logger *log.Logger) *Tcpmux {
	m := &Tcpmux{
		targets: make(map[string]*TargetInfo),
		connCh:  make(chan net.Conn, 8),
		stopCh:  make(chan int),

		logger: logger,
	}
	return m
}

func (m *Tcpmux) getTarget(id string) (*TargetInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	target, found := m.targets[id]
	if found == false {
		return nil, ErrIdNotFound
	}
	return target, nil
}

func (m *Tcpmux) AddTarget(id, target string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.targets[id] = &TargetInfo{id, target}
	return nil
}

func (m *Tcpmux) DeleteTarget(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, found := m.targets[id]
	if !found {
		return ErrIdNotFound
	}
	delete(m.targets, id)
	return nil
}

func (m *Tcpmux) ListTarget() []TargetInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	targets := make([]TargetInfo, 0)
	for _, v := range m.targets {
		targets = append(targets, *v)
	}
	return targets
}

func (m *Tcpmux) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	close(m.stopCh)
	m.waitStopped.Wait()
	close(m.connCh)
	m.stopCh = nil
	m.connCh = nil
}

func (m *Tcpmux) Start(port int) error {
	l, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stopCh == nil {
		m.stopCh = make(chan int, 1)
	}

	if m.connCh == nil {
		m.connCh = make(chan net.Conn, 1)
	}

	m.waitStopped.Add(1)
	go m.acceptLoop(l)

	m.waitStopped.Add(1)
	go m.servLoop()

	return nil
}

func (m *Tcpmux) servLoop() {
	defer m.waitStopped.Done()
	for {
		select {
		case <-m.stopCh:
			return
		case conn := <-m.connCh:
			go m.servConn(conn.(*net.TCPConn))
		}
	}
}

func (m *Tcpmux) acceptLoop(l net.Listener) {
	defer l.Close()
	defer m.waitStopped.Done()
	for {
		select {
		case <-m.stopCh:
			return
		default:
		}

		conn, err := l.Accept()
		if err != nil {
			m.logger.Print(err)
			continue
		}

		m.connCh <- conn
	}
}

func (m *Tcpmux) servConn(c *net.TCPConn) {
	defer c.Close()

	id, err := ReadHeader(c)
	if err != nil {
		m.logger.Printf("failed to receive echo id: %v", err)
		return
	}

	target, err := m.getTarget(string(id))
	if err != nil {
		return
	}

	s, err := net.Dial("tcp", target.Target)
	if err != nil {
		m.logger.Printf("failed to connect tunnel server: %v", err)
		return
	}
	defer s.Close()

	// log.Println("done read id")
	err = WriteHeader(id, c)
	if err != nil {
		m.logger.Printf("failed to send id: %v", err)
		return
	}

	// log.Println("done write echo id")
	m.logger.Printf("%v -> %v", c.RemoteAddr(), s.RemoteAddr())

	var wg sync.WaitGroup
	wg.Add(2)
	go Pipeline(c, s.(*net.TCPConn), &wg)
	go Pipeline(s.(*net.TCPConn), c, &wg)
	wg.Wait()
}
