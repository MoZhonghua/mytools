package util

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

var (
	ErrParamNotFound = errors.New("param not found")
)

func ParseJsonRequest(r *http.Request, out interface{}) error {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}

func HasQueryParam(req *http.Request, param string) (bool, error) {
	_, err := QueryParam(req, param)
	if err != nil {
		if err == ErrParamNotFound {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func QueryParam(req *http.Request, param string) (string, error) {
	err := req.ParseForm()
	if err != nil {
		return "", err
	}
	if v, ok := req.Form[param]; ok {
		return v[0], nil
	}
	return "", ErrParamNotFound
}
