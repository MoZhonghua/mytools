package upnp

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
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

func parseInt(s string) (int, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}

	return int(v), nil
}

func parseIPPort(s string) (net.IP, int, error) {
	f := strings.Split(s, ":")
	if len(f) != 2 {
		return nil, 0, fmt.Errorf("invalid addr: %s", s)
	}

	ip := net.ParseIP(f[0])
	port, err := parseInt(f[1])
	if err != nil {
		return nil, 0, err
	}

	return ip, port, nil
}
