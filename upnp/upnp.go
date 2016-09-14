package upnp

import (
	"encoding/xml"
	"errors"
	"net/http"
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
	response, err := http.Get(deviceDescriptionLocation)
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
