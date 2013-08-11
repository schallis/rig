package main

import (
	"github.com/gocardless/rig"
)

type Server struct{}

func (srv *Server) Version() rig.ApiVersion {
	return rig.ApiVersion{
		rig.Version,
	}
}
