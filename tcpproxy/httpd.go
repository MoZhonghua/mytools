package tcpproxy

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/MoZhonghua/mytools/util"
	"github.com/gorilla/mux"
)

type Httpd struct {
	p      *Proxy
	s      *Store
	logger *log.Logger
}

func NewHttpd(p *Proxy, s *Store, logger *log.Logger) *Httpd {
	d := &Httpd{
		p:      p,
		s:      s,
		logger: logger,
	}
	return d
}

func (d *Httpd) Serv(port int) error {
	m := mux.NewRouter()
	m.Methods("POST").Path("/add").HandlerFunc(d.handleAddPortMapping)
	m.Methods("DELETE").Path("/delete").HandlerFunc(d.handleDeletePortMapping)
	m.Methods("GET").Path("/list").HandlerFunc(d.handleListPortMapping)

	l, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	return http.Serve(l, m)
}

func (d *Httpd) handleAddPortMapping(w http.ResponseWriter, r *http.Request) {
	pmInfo := &PortMappingInfo{}
	err := util.ParseJsonRequest(r, pmInfo)
	if err != nil {
		util.WriteErrorResponse(w, 400, err)
		return
	}

	remoteAddr, err := net.ResolveTCPAddr("tcp4", pmInfo.RemoteAddr)
	if err != nil {
		util.WriteErrorResponse(w, 400, err)
		return
	}

	err = d.s.AddPortMapping(pmInfo)
	if err != nil {
		util.WriteErrorResponse(w, 500, err)
		return
	}

	err = d.p.AddPortMapping(pmInfo.LocalPort, remoteAddr)
	if err != nil {
		d.s.DeletePortMapping(pmInfo.LocalPort)
		util.WriteErrorResponse(w, 500, err)
		return
	}

	util.WriteSuccessResponse(w)
}

func (d *Httpd) handleDeletePortMapping(w http.ResponseWriter, r *http.Request) {
	localPortStr, err := util.QueryParam(r, "localPort")
	if err != nil {
		util.WriteErrorResponse(w, 400, err)
		return
	}

	localPort, err := strconv.ParseInt(localPortStr, 10, 32)
	if err != nil {
		util.WriteErrorResponse(w, 400, err)
		return
	}

	err = d.s.DeletePortMapping(int(localPort))
	if err != nil {
		util.WriteErrorResponse(w, 500, err)
		return
	}

	err = d.p.DeletePortMapping(int(localPort))
	if err != nil {
		util.WriteErrorResponse(w, 500, err)
		return
	}

	util.WriteSuccessResponse(w)
}

func (d *Httpd) handleListPortMapping(w http.ResponseWriter, r *http.Request) {
	list := d.p.ListPortMapping()
	util.WriteSuccessResponseWithData(w, list)
}
