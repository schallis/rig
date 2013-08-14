package main

import (
	"fmt"
	"github.com/gocardless/rig"
	"strings"
)

type Resolver struct {
	config  *Config
	str     string
	dir     string
	stack   *Stack
	service *Service
	process *Process
}

func NewResolver(c *Config, str string, dir string) *Resolver {
	return &Resolver{config: c, str: str, dir: dir}
}

func (r *Resolver) possibleFirstParts() map[string]Runnable {
	possibilities := make(map[string]Runnable)

	for name, stack := range r.config.Stacks {
		possibilities[name] = stack
	}

	for name, service := range r.config.Stacks["default"].Services {
		possibilities[name] = service
	}

	return possibilities
}

func (r *Resolver) parseService(s *Stack, parts []string) error {
	if len(parts) > 0 {
		if svc, exists := s.Services[parts[0]]; exists {
			r.service = svc
			return r.parseProcess(svc, parts[1:])
		} else {
			return fmt.Errorf("Invalid process '%q'", parts)
		}
	}
	return nil
}

func (r *Resolver) parseProcess(s *Service, parts []string) error {
	if len(parts) > 0 {
		if p, exists := s.Processes[parts[0]]; exists {
			r.process = p
		} else {
			return fmt.Errorf("Invalid process '%q'", parts)
		}
	}
	return nil
}

func (r *Resolver) parseStack(parts []string) error {
	possibilities := r.possibleFirstParts()
	switch obj := possibilities[parts[0]].(type) {
	case *Stack:
		r.stack = obj
		if err := r.parseService(obj, parts[1:]); err != nil {
			return err
		}
	case *Service:
		r.stack = obj.Stack
		r.service = obj
		if err := r.parseProcess(obj, parts[1:]); err != nil {
			return err
		}
	case *Process:
		r.stack = obj.Service.Stack
		r.service = obj.Service
		r.process = obj
	case nil:
		// Non-empty, invalid first part. There's nothing we can do.
		return fmt.Errorf("Invalid descriptor '%q'", r.str)
	}
	return nil
}

func (r *Resolver) GetDescriptor() (*rig.Descriptor, error) {
	d := &rig.Descriptor{}

	// TODO handle special case of colon as first character (sibling)
	parts := strings.SplitN(r.str, ":", 3)
	if err := r.parseStack(parts); err != nil {
		return nil, err
	}
	if r.stack != nil {
		d.Stack = r.stack.Name
	}
	if r.service != nil {
		d.Service = r.service.Name
	}
	if r.process != nil {
		d.Process = r.process.Name
	}
	return d, nil
}
