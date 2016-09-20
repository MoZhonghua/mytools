package tcpmux

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/MoZhonghua/mytools/util"
	"github.com/gorilla/mux"
)

type Httpd struct {
	m      *Tcpmux
	logger *log.Logger
}

func NewHttpd(m *Tcpmux, logger *log.Logger) *Httpd {
	d := &Httpd{
		m:      m,
		logger: logger,
	}
	return d
}

func (d *Httpd) Serv(port int) error {
	m := mux.NewRouter()
	m.Methods("POST").Path("/add").HandlerFunc(d.handleAddTarget)
	m.Methods("DELETE").Path("/delete").HandlerFunc(d.handleDeleteTarget)
	m.Methods("GET").Path("/list").HandlerFunc(d.handleListTarget)

	l, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	return http.Serve(l, m)
}

func (d *Httpd) handleAddTarget(w http.ResponseWriter, r *http.Request) {
	ti := &TargetInfo{}
	err := util.ParseJsonRequest(r, ti)
	if err != nil {
		util.WriteErrorResponse(w, 400, err)
		return
	}

	_, err = net.ResolveTCPAddr("tcp4", ti.Target)
	if err != nil {
		util.WriteErrorResponse(w, 400, err)
		return
	}

	err = d.m.AddTarget(ti.Id, ti.Target)
	if err != nil {
		util.WriteErrorResponse(w, 500, err)
		return
	}

	util.WriteSuccessResponse(w)
}

func (d *Httpd) handleDeleteTarget(w http.ResponseWriter, r *http.Request) {
	id, err := util.QueryParam(r, "id")
	if err != nil {
		util.WriteErrorResponse(w, 400, err)
		return
	}

	err = d.m.DeleteTarget(id)
	if err != nil {
		util.WriteErrorResponse(w, 500, err)
		return
	}

	util.WriteSuccessResponse(w)
}

func (d *Httpd) handleListTarget(w http.ResponseWriter, r *http.Request) {
	list := d.m.ListTarget()
	util.WriteSuccessResponseWithData(w, list)
}
