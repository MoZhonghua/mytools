package tcpmux

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

func (c *Client) AddTarget(id string, target string) error {
	resp := &util.GenericJsonResp{}
	req := &TargetInfo{
		Id:     id,
		Target: target,
	}

	url := util.JoinURL(c.server, "/add")
	return c.hc.DoJsonPostAndParseResult(url, req, resp)
}

func (c *Client) DeleteTarget(id string) error {
	resp := &util.GenericJsonResp{}
	url := util.JoinURL(c.server, fmt.Sprintf("/delete?id=%s", id))
	return c.hc.DoRequestParseResult("DELETE", url, resp)
}

type targetListResp struct {
	util.GenericJsonResp
	Data []*TargetInfo `json:"data"`
}

func (c *Client) ListTarget() ([]*TargetInfo, error) {
	resp := &targetListResp{}
	url := util.JoinURL(c.server, "/list")
	err := c.hc.DoRequestParseResult("GET", url, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
