package upnp

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"

	"github.com/MoZhonghua/mytools/util"
)

type UPnpService struct {
	ID         string `xml:"serviceId"`
	Type       string `xml:"serviceType"`
	ControlURL string `xml:"controlURL"`
}

type UPnpDevice struct {
	DeviceType   string        `xml:"deviceType"`
	FriendlyName string        `xml:"friendlyName"`
	UDN          string        `xml:"UDN"`
	Devices      []UPnpDevice  `xml:"deviceList>device"`
	Services     []UPnpService `xml:"serviceList>service"`
}

type UPnPRoot struct {
	Device UPnpDevice `xml:"device"`
}

func (d *UPnpDevice) GetChildDevices(deviceType string) []UPnpDevice {
	var result []UPnpDevice
	for _, dev := range d.Devices {
		if dev.DeviceType == deviceType {
			result = append(result, dev)
		}
	}
	return result
}

func (d *UPnpDevice) GetChildServices(serviceType string) []UPnpService {
	var result []UPnpService
	for _, service := range d.Services {
		if service.Type == serviceType {
			result = append(result, service)
		}
	}
	return result
}

func GetUPnPData(deviceDescriptionLocation string) (*UPnPRoot, error) {
	req, err := http.NewRequest("GET", deviceDescriptionLocation, nil)
	if err != nil {
		return nil, err
	}
	response, err := util.DefaultHttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, errors.New("bad status code:" + response.Status)
	}

	var upnpRoot UPnPRoot
	err = xml.NewDecoder(response.Body).Decode(&upnpRoot)
	if err != nil {
		return nil, err
	}

	return &upnpRoot, nil
}

func AddUpnpTcpPortMapping(igdServer string, externalPort int,
	internalIp string, internalPort int) (string, error) {
	igdURL := fmt.Sprintf("http://%s:1900/igd.xml", igdServer)
	root, err := GetUPnPData(igdURL)
	if err != nil {
		return "", err
	}

	igd, err := GetIGDDevice(root, igdURL)
	if err != nil {
		return "", err
	}

	for _, s := range igd.Services {
		externalIP, err := s.GetExternalIPAddress()
		if err != nil {
			return "", err
		}

		err = s.AddPortMapping("TCP", externalPort, internalIp, internalPort, 0, "")
		if err != nil {
			return "", err
		}

		return externalIP.String(), nil
	}

	return "", errors.New("can't create port mapping")
}

func DeleteUpnpTcpPortMapping(igdServer string, externalPort int) error {
	igdURL := fmt.Sprintf("http://%s:1900/igd.xml", igdServer)
	root, err := GetUPnPData(igdURL)
	if err != nil {
		return err
	}

	igd, err := GetIGDDevice(root, igdURL)
	if err != nil {
		return err
	}

	for _, s := range igd.Services {
		err = s.DeletePortMapping("TCP", externalPort)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}
