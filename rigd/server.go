package main

import (
	"fmt"
	"github.com/gocardless/rig"
)

type Server struct {
	Stacks map[string]*Stack
}

func NewServer(stacks map[string]*Stack) (*Server, error) {
	srv := &Server{Stacks: stacks}
	return srv, nil
}

func (srv *Server) GetStack(d *rig.Descriptor) (*Stack, error) {
	s := srv.Stacks[d.Stack]
	if s == nil {
		return nil, fmt.Errorf("stack '%v' does not exist", d.Stack)
	}

	return s, nil
}

func (srv *Server) GetService(d *rig.Descriptor) (*Service, error) {
	s := srv.Stacks[d.Stack]
	if s == nil {
		return nil, fmt.Errorf("stack '%v' does not exist", d.Stack)
	}

	svc := s.Services[d.Service]
	if svc == nil {
		return nil, fmt.Errorf("service '%v' does not exist", d.Service)
	}

	return svc, nil
}

func (srv *Server) GetProcess(d *rig.Descriptor) (*Process, error) {
	s := srv.Stacks[d.Stack]
	if s == nil {
		return nil, fmt.Errorf("stack '%v' does not exist", d.Stack)
	}

	svc := s.Services[d.Service]
	if svc == nil {
		return nil, fmt.Errorf("service '%v' does not exist", d.Service)
	}

	p := svc.Processes[d.Process]
	if p == nil {
		return nil, fmt.Errorf("process '%v' does not exist", d.Process)
	}

	return p, nil
}

func (srv *Server) StartStack(d *rig.Descriptor) error {
	s, err := srv.GetStack(d)
	if err != nil {
		return err
	}

	go s.Start()

	return nil
}

func (srv *Server) StopStack(d *rig.Descriptor) error {
	s, err := srv.GetStack(d)
	if err != nil {
		return err
	}

	s.Stop()

	return nil
}

func (srv *Server) TailStack(d *rig.Descriptor, c chan rig.ProcessOutputMessage) error {
	s, err := srv.GetStack(d)
	if err != nil {
		return err
	}

	s.SubscribeToOutput(c)
	return nil
}

func (srv *Server) StartService(d *rig.Descriptor) error {
	svc, err := srv.GetService(d)
	if err != nil {
		return err
	}

	go svc.Start()

	return nil
}

func (srv *Server) StopService(d *rig.Descriptor) error {
	svc, err := srv.GetService(d)
	if err != nil {
		return err
	}

	svc.Stop()

	return nil
}

func (srv *Server) TailService(d *rig.Descriptor, c chan rig.ProcessOutputMessage) error {
	svc, err := srv.GetService(d)
	if err != nil {
		return err
	}

	svc.SubscribeToOutput(c)
	return nil
}

func (srv *Server) StartProcess(d *rig.Descriptor) error {
	p, err := srv.GetProcess(d)
	if err != nil {
		return err
	}

	go p.Start()

	return nil
}

func (srv *Server) StopProcess(d *rig.Descriptor) error {
	p, err := srv.GetProcess(d)
	if err != nil {
		return err
	}

	return p.Stop()
}

func (srv *Server) TailProcess(d *rig.Descriptor, c chan rig.ProcessOutputMessage) error {
	p, err := srv.GetProcess(d)
	if err != nil {
		return err
	}

	p.SubscribeToOutput(c)
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
