package util

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GenericJsonResp struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	resp := &GenericJsonResp{
		Success: false,
		Error:   err.Error(),
		Data:    nil,
	}

	writeResp(w, statusCode, resp)
}

func WriteSuccessResponseWithData(w http.ResponseWriter, data interface{}) {
	resp := &GenericJsonResp{
		Success: true,
		Error:   "",
		Data:    data,
	}

	writeResp(w, http.StatusOK, resp)
}

func WriteSuccessResponse(w http.ResponseWriter) {
	WriteSuccessResponseWithData(w, nil)
}

func writeResp(w http.ResponseWriter, statusCode int, resp *GenericJsonResp) {
	b, _ := json.MarshalIndent(resp, "", "  ")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(b)))
	w.WriteHeader(statusCode)
	w.Write(b)
}
