package tcpproxy

import (
	"fmt"

	"github.com/MoZhonghua/mytools/util"
)

type Client struct {
	server string
}

func NewClient(server string) (*Client, error) {
	c := &Client{
		server: server,
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
	return util.DefaultHttpClient.DoJsonPostAndParseResult(url, req, resp)
}

func (c *Client) DeletePortMapping(localPort int) error {
	resp := &util.GenericJsonResp{}
	url := util.JoinURL(c.server, fmt.Sprintf("/delete?localPort=%d", localPort))
	return util.DefaultHttpClient.DoRequestParseResult("DELETE", url, resp)
}

type portMappingListResp struct {
	util.GenericJsonResp
	Data []*PortMappingInfo `json:"data"`
}

func (c *Client) ListPortMapping() ([]*PortMappingInfo, error) {
	resp := &portMappingListResp{}
	url := util.JoinURL(c.server, "/list")
	err := util.DefaultHttpClient.DoRequestParseResult("GET", url, resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
