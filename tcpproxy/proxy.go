package tcpproxy

import (
	"errors"
	"log"
	"net"
	"sync"
)

var (
	ErrLocalPortUsed       = errors.New("local port used")
	ErrPortMappingNotFound = errors.New("port mapping not found")
)

type Proxy struct {
	mu       sync.Mutex
	mappings map[int]*portMapping
	logger   *log.Logger
}

func NewProxy(logger *log.Logger) *Proxy {
	p := &Proxy{
		mappings: make(map[int]*portMapping),
		logger:   logger,
	}
	return p
}

func (p *Proxy) AddPortMapping(localPort int, remoteAddr *net.TCPAddr) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, found := p.mappings[localPort]
	if found {
		return ErrLocalPortUsed
	}

	m := newPortMapping(localPort, remoteAddr)
	err := m.start()
	if err != nil {
		return err
	}

	p.mappings[localPort] = m
	return err
}

func (p *Proxy) DeletePortMapping(localPort int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	m, found := p.mappings[localPort]
	if !found {
		return ErrPortMappingNotFound
	}

	m.stop()
	delete(p.mappings, localPort)
	return nil
}

func (p *Proxy) ListPortMapping() []*PortMappingInfo {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := make([]*PortMappingInfo, 0)
	for _, m := range p.mappings {
		result = append(result, &PortMappingInfo{
			LocalPort:  m.localPort,
			RemoteAddr: m.remoteAddr.String(),
		})
	}
	return result
}
