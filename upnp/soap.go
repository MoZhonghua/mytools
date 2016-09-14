package upnp

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/MoZhonghua/mytools/util"
)

func soapRequest(c *util.HttpClient, url, service, function, message string) ([]byte, error) {
	tpl := `<?xml version="1.0" ?>
	<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
	<s:Body>%s</s:Body>
	</s:Envelope>
`
	var resp []byte

	body := fmt.Sprintf(tpl, message)

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return resp, err
	}
	req.Close = true
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("User-Agent", "syncthing/1.0")
	req.Header["SOAPAction"] = []string{fmt.Sprintf(`"%s#%s"`, service, function)} // Enforce capitalization in header-entry for sensitive routers. See issue #1696
	req.Header.Set("Connection", "Close")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	// l.Debugln("SOAP Request URL: " + url)
	// l.Debugln("SOAP Action: " + req.Header.Get("SOAPAction"))
	// l.Debugln("SOAP Request:\n\n" + body)

	r, err := c.Do(req)
	if err != nil {
		// l.Debugln(err)
		return resp, err
	}

	resp, _ = ioutil.ReadAll(r.Body)
	// l.Debugf("SOAP Response: %s\n\n%s\n\n", r.Status, resp)

	r.Body.Close()

	if r.StatusCode >= 400 {
		return resp, errors.New(function + ": " + r.Status)
	}

	return resp, nil
}

type soapGetExternalIPAddressResponseEnvelope struct {
	XMLName xml.Name
	Body    soapGetExternalIPAddressResponseBody `xml:"Body"`
}

type soapGetExternalIPAddressResponseBody struct {
	XMLName                      xml.Name
	GetExternalIPAddressResponse getExternalIPAddressResponse `xml:"GetExternalIPAddressResponse"`
}

type getExternalIPAddressResponse struct {
	NewExternalIPAddress string `xml:"NewExternalIPAddress"`
}

type soapErrorResponse struct {
	ErrorCode        int    `xml:"Body>Fault>detail>UPnPError>errorCode"`
	ErrorDescription string `xml:"Body>Fault>detail>UPnPError>errorDescription"`
}
