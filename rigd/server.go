package main

import (
	"fmt"
	"github.com/gocardless/rig"
)

type Server struct {
	stacks map[string]*Stack
}

func NewServer(stacks map[string]*Stack) (*Server, error) {
	srv := &Server{stacks: stacks}
	return srv, nil
}

func (srv *Server) StartStack(stack string) error {
	s := srv.stacks[stack]
	if s == nil {
		return fmt.Errorf("stack '%v' does not exist", stack)
	}

	go s.Start()

	return nil
}

func (srv *Server) StartService(stack, service string) error {
	s := srv.stacks[stack]
	if s == nil {
		return fmt.Errorf("stack '%v' does not exist", stack)
	}

	svc := s.Services[service]
	if svc == nil {
		return fmt.Errorf("service '%v' does not exist", service)
	}

	go svc.Start()

	return nil
}

func (srv *Server) StartProcess(stack, service, process string) error {
	s := srv.stacks[stack]
	if s == nil {
		return fmt.Errorf("stack '%v' does not exist", stack)
	}

	svc := s.Services[service]
	if svc == nil {
		return fmt.Errorf("service '%v' does not exist", service)
	}

	p := svc.Processes[process]
	if p == nil {
		return fmt.Errorf("process '%v' does not exist", process)
	}

	go p.Start(svc.Dir)

	return nil
}

func (srv *Server) StopProcess(stack, service, process string) error {
	s := srv.stacks[stack]
	if s == nil {
		return fmt.Errorf("stack '%v' does not exist", stack)
	}

	svc := s.Services[service]
	if svc == nil {
		return fmt.Errorf("service '%v' does not exist", service)
	}

	p := svc.Processes[process]
	if p == nil {
		return fmt.Errorf("process '%v' does not exist", process)
	}

	p.Stop()

	return nil
}

func (s *Server) Resolve() error {
	return nil
}

func (srv *Server) Version() rig.ApiVersion {
	return rig.ApiVersion{
		rig.Version,
	}
}
