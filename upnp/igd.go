package upnp

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
	log "github.com/Sirupsen/logrus"
)

type IGDService struct {
	ID  string
	URL string
	URN string
}

type IGD struct {
	UUID           string
	FriendlyName   string
	DescriptionURL string
	internalIP     net.IP
	Services       []IGDService
}

var (
	urnInternetGatewayDevice1 = "urn:schemas-upnp-org:device:InternetGatewayDevice:1"
	urnWANDevice1             = "urn:schemas-upnp-org:device:WANDevice:1"
	urnWanConnectionDevice1   = "urn:schemas-upnp-org:device:WANConnectionDevice:1"
	urnWANIPConnection1       = "urn:schemas-upnp-org:service:WANIPConnection:1"
	urnWANPPPConnection1      = "urn:schemas-upnp-org:service:WANPPPConnection:1"

	urnInternetGatewayDevice2 = "urn:schemas-upnp-org:device:InternetGatewayDevice:2"
	urnWANDevice2             = "urn:schemas-upnp-org:device:WANDevice:2"
	urnWanConnectionDevice2   = "urn:schemas-upnp-org:device:WANConnectionDevice:2"
	urnWANIPConnection2       = "urn:schemas-upnp-org:service:WANIPConnection:2"
	urnWANPPPConnection2      = "urn:schemas-upnp-org:service:WANPPPConnection:2"
)

func GetIGDDevice(root *UPnPRoot, rootURL string) (*IGD, error) {
	services, err := GetIGDServices(&root.Device, rootURL)
	if err != nil {
		return nil, err
	}

	// Figure out our IP number, on the network used to reach the IGD.
	// We do this in a fairly roundabout way by connecting to the IGD and
	// checking the address of the local end of the socket. I'm open to
	// suggestions on a better way to do this...
	internalIP, err := localIP(rootURL)
	if err != nil {
		return nil, err
	}

	return &IGD{
		UUID:           root.Device.UDN,
		FriendlyName:   root.Device.FriendlyName,
		Services:       services,
		internalIP:     internalIP,
		DescriptionURL: rootURL,
	}, nil
}

func GetIGDServices(device *UPnpDevice, rootURL string) ([]IGDService, error) {
	var result []IGDService

	if device.DeviceType == urnInternetGatewayDevice1 {
		descriptions := getServices(device, rootURL,
			urnWANDevice1, urnWanConnectionDevice1,
			[]string{urnWANIPConnection1, urnWANPPPConnection1})

		result = append(result, descriptions...)
	} else if device.DeviceType == urnInternetGatewayDevice1 {
		descriptions := getServices(device, rootURL,
			urnWANDevice2, urnWanConnectionDevice2,
			[]string{urnWANIPConnection2, urnWANPPPConnection2})
		result = append(result, descriptions...)
	} else {
		return result, errors.New("not an InternetGatewayDevice")
	}

	if len(result) < 1 {
		return result, errors.New("no compatible service descriptions found.")
	}
	return result, nil
}

func getServices(device *UPnpDevice, rootURL string, wanDeviceURN string, wanConnectionURN string, URNs []string) []IGDService {
	var result []IGDService
	log.Infof("-----------------")

	devices := device.GetChildDevices(wanDeviceURN)
	if len(devices) < 1 {
		return result
	}

	for _, device := range devices {
		connections := device.GetChildDevices(wanConnectionURN)
		for _, connection := range connections {
			for _, URN := range URNs {
				services := connection.GetChildServices(URN)
				for _, service := range services {
					if len(service.ControlURL) == 0 {
						continue
					}
					u, _ := url.Parse(rootURL)
					replaceRawPath(u, service.ControlURL)
					service := IGDService{ID: service.ID,
						URL: u.String(),
						URN: service.Type}
					result = append(result, service)
				}
			}
		}
	}

	return result
}

func replaceRawPath(u *url.URL, rp string) {
	asURL, err := url.Parse(rp)
	if err != nil {
		return
	} else if asURL.IsAbs() {
		u.Path = asURL.Path
		u.RawQuery = asURL.RawQuery
	} else {
		var p, q string
		fs := strings.Split(rp, "?")
		p = fs[0]
		if len(fs) > 1 {
			q = fs[1]
		}

		if p[0] == '/' {
			u.Path = p
		} else {
			u.Path += p
		}
		u.RawQuery = q
	}
}

// AddPortMapping adds a port mapping to the specified IGD service.
func (s *IGDService) AddPortMapping(
	protocol string, externalPort int,
	internalIP string, internalPort int, duration time.Duration,
	description string) error {
	tpl := `<u:AddPortMapping xmlns:u="%s">
	<NewRemoteHost></NewRemoteHost>
	<NewExternalPort>%d</NewExternalPort>
	<NewProtocol>%s</NewProtocol>
	<NewInternalPort>%d</NewInternalPort>
	<NewInternalClient>%s</NewInternalClient>
	<NewEnabled>1</NewEnabled>
	<NewPortMappingDescription>%s</NewPortMappingDescription>
	<NewLeaseDuration>%d</NewLeaseDuration>
	</u:AddPortMapping>`
	body := fmt.Sprintf(tpl, s.URN, externalPort, protocol, internalPort, internalIP, description, duration/time.Second)

	response, err := soapRequest(s.URL, s.URN, "AddPortMapping", body)
	if err != nil && duration > 0 {
		// Try to repair error code 725 - OnlyPermanentLeasesSupported
		envelope := &soapErrorResponse{}
		if unmarshalErr := xml.Unmarshal(response, envelope); unmarshalErr != nil {
			return unmarshalErr
		}
		if envelope.ErrorCode == 725 {
			return s.AddPortMapping(protocol, externalPort,
				internalIP, internalPort, 0, description)
		}
	}

	return err
}

// DeletePortMapping deletes a port mapping from the specified IGD service.
func (s *IGDService) DeletePortMapping(protocol string, externalPort int) error {
	tpl := `<u:DeletePortMapping xmlns:u="%s">
	<NewRemoteHost></NewRemoteHost>
	<NewExternalPort>%d</NewExternalPort>
	<NewProtocol>%s</NewProtocol>
	</u:DeletePortMapping>`
	body := fmt.Sprintf(tpl, s.URN, externalPort, protocol)

	_, err := soapRequest(s.URL, s.URN, "DeletePortMapping", body)
	return err
}

// GetExternalIPAddress queries the IGD service for its external IP address.
// Returns nil if the external IP address is invalid or undefined, along with
// any relevant errors
func (s *IGDService) GetExternalIPAddress() (net.IP, error) {
	tpl := `<u:GetExternalIPAddress xmlns:u="%s" />`

	body := fmt.Sprintf(tpl, s.URN)

	response, err := soapRequest(s.URL, s.URN, "GetExternalIPAddress", body)

	if err != nil {
		return nil, err
	}

	envelope := &soapGetExternalIPAddressResponseEnvelope{}
	err = xml.Unmarshal(response, envelope)
	if err != nil {
		return nil, err
	}

	result := net.ParseIP(envelope.Body.GetExternalIPAddressResponse.NewExternalIPAddress)

	return result, nil
}
