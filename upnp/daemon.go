package upnp

import (
	"errors"
	"sync"
)

var (
	ErrLocalPortUsed       = errors.New("local port used")
	ErrPortMappingNotFound = errors.New("port mapping not found")
)

type Daemon struct {
	mu       sync.Mutex
	mappings map[string]*portMapping
}

func NewDaemon() *Daemon {
	p := &Daemon{
		mappings: make(map[string]*portMapping),
	}
	return p
}

func (p *Daemon) AddPortMapping(igdServer string, externalPort int,
	internalIP string, internalPort int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	info := &PortMappingInfo{
		IGDServer:    igdServer,
		ExternalPort: externalPort,
		InternalIP:   internalIP,
		InternalPort: internalPort,
	}

	id := genId(info.IGDServer, info.ExternalPort)
	_, found := p.mappings[id]
	if found {
		return ErrLocalPortUsed
	}

	m := newPortMapping(info)
	err := m.start()
	if err != nil {
		return err
	}

	p.mappings[id] = m
	return err
}

func (p *Daemon) DeletePortMapping(igdServer string, externalPort int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := genId(igdServer, externalPort)
	m, found := p.mappings[id]
	if !found {
		return ErrPortMappingNotFound
	}

	m.stop()
	delete(p.mappings, id)
	return nil
}

func (p *Daemon) ListPortMapping() []*PortMappingInfo {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := make([]*PortMappingInfo, 0)
	for _, m := range p.mappings {
		result = append(result, m.info)
	}
	return result
}
