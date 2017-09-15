package util

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

type httpClientGlobalConfig struct {
	debug bool
	mu    sync.Mutex
}

func newDefaultGlobalConfig() *httpClientGlobalConfig {
	gcfg := &httpClientGlobalConfig{}
	gcfg.debug = false
	return gcfg
}

var (
	gcfg = newDefaultGlobalConfig()
)

func SetHttpClientDebugMode(flag bool) {
	gcfg.mu.Lock()
	defer gcfg.mu.Unlock()
	gcfg.debug = flag
}

func GetHttpClientDebugMode() bool {
	gcfg.mu.Lock()
	defer gcfg.mu.Unlock()
	return gcfg.debug
}

type HttpClientConfig struct {
	Proxy            string
	NoFollowRedirect bool
	NoTLSVerify      bool
	DialTimeout      time.Duration
}

type HttpClient struct {
	tr               *http.Transport
	client           *http.Client
	noFollowRedirect bool
	debug            bool
	debugLock        sync.Mutex
}

var DefaultHttpClient *HttpClient

func init() {
	DefaultHttpClient, _ = NewHttpClient(&HttpClientConfig{})
}

func NewHttpClient(cfg *HttpClientConfig) (*HttpClient, error) {
	tr := &http.Transport{}
	tr.MaxIdleConnsPerHost = 1
	timeout := cfg.DialTimeout
	tr.Dial = func(network string, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, timeout)
	}
	httpClient := &http.Client{
		Transport: tr,
	}
	httpClient.Timeout = 0
	if len(cfg.Proxy) != 0 {
		proxyUrl, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy: %s", err.Error())
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	}

	if cfg.NoTLSVerify {
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	c := &HttpClient{
		tr:               tr,
		client:           httpClient,
		noFollowRedirect: cfg.NoFollowRedirect,
	}

	return c, nil
}

func (c *HttpClient) DoJsonPostAndParseResult(url string,
	data interface{}, result interface{}) error {
	resp, err := c.DoJsonPost(url, data)
	if err != nil {
		return err
	}
	return c.ParseJsonResp(resp, result)
}

func (c *HttpClient) DoRequestParseResult(method, url string, result interface{}) error {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	return c.ParseJsonResp(resp, result)
}

func (c *HttpClient) DoGetAndParseResult(url string, result interface{}) error {
	return c.DoRequestParseResult("GET", url, result)
}

func (c *HttpClient) Do(req *http.Request) (*http.Response, error) {
	if c.debug || GetHttpClientDebugMode() {
		c.DumpRequest(req, os.Stdout)
	}

	var resp *http.Response
	var err error

	if c.noFollowRedirect {
		resp, err = c.tr.RoundTrip(req)
	} else {
		resp, err = c.client.Do(req)
	}
	if err != nil {
		return nil, err
	}

	if c.debug || GetHttpClientDebugMode() {
		c.DumpResponse(resp, os.Stdout)
	}

	return resp, nil
}

func (c *HttpClient) DoJsonPost(url string, data interface{}) (*http.Response, error) {
	var r io.Reader
	if data != nil {
		rdata, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(rdata)
	}
	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.Do(req)
}

func (c *HttpClient) ParseJsonResp(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to read request body: %s", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Debugf("server returns: %s - %s", resp.Status, string(body))
		return fmt.Errorf("http error %s: %s", resp.Status, string(body))
	}

	if err != nil {
		return err
	}
	return json.Unmarshal(body, result)
}

func printLinesWithPrefix(data []byte, prefix string, w io.Writer) {
	r := bufio.NewReader(bytes.NewReader(data))
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			break
		}

		fmt.Fprintf(w, "%s%s\n", prefix, string(line))
	}
}

func (c *HttpClient) DumpRequest(req *http.Request, w io.Writer) {
	c.debugLock.Lock()
	defer c.debugLock.Unlock()
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return
	}

	rnrn := bytes.Index(dump, []byte("\r\n\r\n"))
	if rnrn < 0 {
		return
	}

	printLinesWithPrefix(dump[:rnrn], "> ", w)
	fmt.Print("> \n")
	fmt.Printf("%s\n", string(dump[rnrn+4:]))
}

func (c *HttpClient) DumpResponse(resp *http.Response, w io.Writer) {
	c.debugLock.Lock()
	defer c.debugLock.Unlock()
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return
	}

	rnrn := bytes.Index(dump, []byte("\r\n\r\n"))
	if rnrn < 0 {
		return
	}

	printLinesWithPrefix(dump[:rnrn], "< ", w)
	fmt.Print("< \n")
	fmt.Printf("%s\n", string(dump[rnrn+4:]))
}
