package upnp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type SSDPNotify struct {
	DeviceUUID                string `json:"uuid"`
	RespondingDeviceType      string `json:"deviceType"`
	DeviceUSN                 string `json:"usn"`
	DeviceDescriptionLocation string `json:"location"`
}

func ParseSSDPNotify(deviceType string, data []byte) (*SSDPNotify, error) {
	reader := bufio.NewReader(bytes.NewBuffer(data))
	request := &http.Request{}
	response, err := http.ReadResponse(reader, request)
	if err != nil {
		return nil, err
	}

	respondingDeviceType := response.Header.Get("St")
	if respondingDeviceType != deviceType {
		return nil, errors.New("unrecognized UPnP device of type " + respondingDeviceType)
	}

	deviceDescriptionLocation := response.Header.Get("Location")
	if deviceDescriptionLocation == "" {
		return nil, errors.New("invalid IGD response: no location specified")
	}

	deviceUSN := response.Header.Get("USN")
	if deviceUSN == "" {
		return nil, errors.New("invalid IGD response: USN not specified")
	}

	deviceUUID := strings.TrimPrefix(strings.Split(deviceUSN, "::")[0], "uuid:")

	r := &SSDPNotify{}
	r.DeviceUSN = deviceUSN
	r.RespondingDeviceType = respondingDeviceType
	r.DeviceUUID = deviceUUID
	r.DeviceDescriptionLocation = deviceDescriptionLocation
	return r, nil
}

func buildSSDPSearchPackage(deviceType string, timeout time.Duration) []byte {
	var tpl = "M-SEARCH * HTTP/1.1 \r\n" +
		"HOST: 239.255.255.250:1900\r\n" +
		"ST: %s\r\n" +
		"MAN: \"ssdp:discover\"\r\n" +
		"MX: %d\r\n" +
		"USER-AGENT: UPnP/1.0\r\n" +
		"\r\n"

	pkt := fmt.Sprintf(tpl, deviceType, timeout/time.Second)
	return []byte(pkt)
}

func SSDPSearch(intf *net.Interface, deviceType string,
	timeout time.Duration) (<-chan *SSDPNotify, error) {

	result := make(chan *SSDPNotify, 16)
	search := buildSSDPSearchPackage(deviceType, timeout)

	local := &net.UDPAddr{IP: []byte{239, 255, 255, 250}, Port: 0}
	conn, err := net.ListenMulticastUDP("udp4", intf, local)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(timeout))

	dst := &net.UDPAddr{IP: []byte{239, 255, 255, 250}, Port: 1900}
	_, err = conn.WriteTo(search, dst)
	if err != nil {
		conn.Close()
		return nil, err
	}

	go func() {
		defer conn.Close()
		defer close(result)
		for {
			resp := make([]byte, 65536)
			n, _, err := conn.ReadFrom(resp)
			if err != nil {
				if e, ok := err.(net.Error); !ok || !e.Timeout() {
					log.Printf("failed to recv notify: %v\n", err)
				}
				break
			}
			notify, err := ParseSSDPNotify(deviceType, resp[:n])
			if err != nil {
				continue
			}
			result <- notify
		}
	}()
	return result, nil
}
