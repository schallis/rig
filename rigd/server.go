package main

import (
	"fmt"
	"github.com/gocardless/rig"
	"log"
)

type Server struct {
	Config *Config
	Stacks map[string]*Stack
}

func NewServer() *Server {
	return &Server{
		Stacks: map[string]*Stack{},
	}
}

func (srv *Server) LoadConfig(configFilename string) error {
	config, err := LoadConfigFromFile(configFilename)
	if err != nil {
		return err
	}
	srv.Config = config

	for name, config := range srv.Config.Stacks {
		stack := NewStack(name)
		if err := loadServices(stack, config); err != nil {
			return err
		}
		srv.Stacks[name] = stack
	}
	return nil
}

func loadServices(stack *Stack, stackConfig *StackConfig) error {
	for name, config := range stackConfig.Services {
		service, err := NewService(name, config.Dir, stack)
		if err != nil {
			return err
		}
		stack.Services[name] = service
	}
	return nil
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

func (srv *Server) TailStack(d *rig.Descriptor, c chan rig.ProcessOutputMessage, num int) error {
	s, err := srv.GetStack(d)
	if err != nil {
		return err
	}

	s.SubscribeToOutput(c, num)
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

func (srv *Server) TailService(d *rig.Descriptor, c chan rig.ProcessOutputMessage, num int) error {
	svc, err := srv.GetService(d)
	if err != nil {
		return err
	}

	svc.SubscribeToOutput(c, num)
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

func (srv *Server) TailProcess(d *rig.Descriptor, c chan rig.ProcessOutputMessage, num int) error {
	p, err := srv.GetProcess(d)
	if err != nil {
		return err
	}

	p.SubscribeToOutput(c, num)
	return nil
}

func (srv *Server) Resolve(str, pwd string) (*rig.Descriptor, error) {
	res := NewResolver(srv.Stacks, str, pwd)

	d, err := res.GetDescriptor()
	if err != nil {
		return nil, err
	}

	return d, err
}

func (srv *Server) ReloadConfig() error {
	log.Printf("Reloading config...\n")
	srv.Stacks = map[string]*Stack{}
	if err := srv.LoadConfig(srv.Config.Filename); err != nil {
		return err
	}
	return nil
}

func (srv *Server) Version() rig.ApiVersion {
	return rig.ApiVersion{
		rig.Version,
	}
}
