package tcpproxy

import (
	"fmt"
	"log"

	"github.com/MoZhonghua/mytools/util"
)

type Client struct {
	server string
	hc     *util.HttpClient
	logger *log.Logger
}

func NewClient(server string, logger *log.Logger,
	proxy string, debug bool) (*Client, error) {
	cfg := &util.HttpClientConfig{}
	cfg.Debug = debug
	cfg.NoTLSVerify = true
	cfg.Proxy = proxy
	cfg.Logger = logger

	hc, err := util.NewHttpClient(cfg)
	if err != nil {
		return nil, err
	}

	c := &Client{
		server: server,
		hc:     hc,
		logger: logger,
	}
	return c, nil
}

func (c *Client) AddPortMapping(localPort int, remoteAddr string) error {
	resp := &util.GenericJsonResp{}
	req := &PortMappingInfo{
		LocalPort:  localPort,
		RemoteAddr: remoteAddr,
	}

	url := util.JoinURL(c.server, "/add")
	return c.hc.DoJsonPostAndParseResult(url, req, resp)
}

func (c *Client) DeletePortMapping(localPort int) error {
	resp := &util.GenericJsonResp{}
	url := util.JoinURL(c.server, fmt.Sprintf("/delete?localPort=%d", localPort))
	return c.hc.DoRequestParseResult("DELETE", url, resp)
}

type portMappingListResp struct {
	util.GenericJsonResp
	Data []*PortMappingInfo `json:"data"`
}

func (c *Client) ListPortMapping() ([]*PortMappingInfo, error) {
	resp := &portMappingListResp{}
	url := util.JoinURL(c.server, "/list")
	err := c.hc.DoRequestParseResult("GET", url, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
