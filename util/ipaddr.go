package util

import (
	"net"
)

func GetIPList() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	ips := make([]string, 0)
	for _, i := range ifaces {
		if i.Flags&net.FlagUp == 0 {
			continue
		}

		if i.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}

			ips = append(ips, ip.String())
		}
	}

	return ips
}
