package upnp

import (
	"net"
	"net/url"
	"time"
)

func localIP(remoteURL string) (net.IP, error) {
	parsed, err := url.Parse(remoteURL)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", parsed.Host, time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localIPAddress, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return nil, err
	}

	return net.ParseIP(localIPAddress), nil
}
