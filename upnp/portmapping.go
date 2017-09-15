package upnp

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

type portMapping struct {
	info        *PortMappingInfo
	stopCh      chan int
	waitStopped sync.WaitGroup
}

func newPortMapping(info *PortMappingInfo) *portMapping {
	m := &portMapping{
		info:   info,
		stopCh: make(chan int),
	}
	return m
}

func (m *portMapping) stop() {
	close(m.stopCh)
	m.waitStopped.Wait()
	DeleteUpnpTcpPortMapping(m.info.IGDServer, m.info.ExternalPort)
	log.Infof("delete port mapping :%d -> %s",
		m.info.ExternalPort, m.info.internalAddr())
}

func (m *portMapping) servLoop() {
	ticker := time.NewTicker(time.Second * 120)
	defer ticker.Stop()
	defer m.waitStopped.Done()

	for {
		m.ensureUpnpPortMapping()
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
		}
	}
}

func (m *portMapping) ensureUpnpPortMapping() (string, error) {
	return AddUpnpTcpPortMapping(m.info.IGDServer, m.info.ExternalPort,
		m.info.InternalIP, m.info.InternalPort)
}

func (m *portMapping) start() error {
	externalIP, err := m.ensureUpnpPortMapping()
	if err != nil {
		return err
	}

	log.Infof("new port mapping %s:%d -> %s", externalIP,
		m.info.ExternalPort,
		m.info.internalAddr())

	m.info.ExternalIP = externalIP
	m.waitStopped.Add(1)
	go m.servLoop()
	return nil
}
