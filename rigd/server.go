package main

import (
	"github.com/gocardless/rig"
)

type Server struct{}

func NewServer() (*Server, error) {
	srv := &Server{}
	return srv, nil
}

func (srv *Server) Version() rig.ApiVersion {
	return rig.ApiVersion{
		rig.Version,
	}
}
