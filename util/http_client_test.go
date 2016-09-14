package util

import (
	"log"
	"net/http"
	"os"
	"testing"
)

func TestHttpClientRedirect(t *testing.T) {
	cfg := &HttpClientConfig{}
	cfg.Logger = log.New(os.Stdout, "", log.LstdFlags)
	cfg.Debug = true

	c, err := NewHttpClient(cfg)
	if err != nil {
		t.Fatal(err)
		return
	}

	req, _ := http.NewRequest("GET", "http://127.0.0.1:9334/pool/status?sync", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
		return
	}

	r := &GenericJsonResp{}
	c.ParseJsonResp(resp, r)
}
