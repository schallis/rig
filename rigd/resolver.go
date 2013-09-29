package main

import (
	"fmt"
	"github.com/gocardless/rig"
	"path/filepath"
	"strings"
)

type Resolver struct {
	stacks  map[string]*Stack
	str     string
	dir     string
	stack   *Stack
	service *Service
	process *Process
}

func NewResolver(s map[string]*Stack, str string, dir string) *Resolver {
	return &Resolver{stacks: s, str: str, dir: dir}
}

func canonicalise(path string) string {
	// Hide errors, errors are boring
	path, _ = filepath.Abs(path)
	path, _ = filepath.EvalSymlinks(path)
	return path
}

func (r *Resolver) findServiceByDir() *Service {
	dirPath := canonicalise(r.dir)
	for _, stack := range r.stacks {
		for _, svc := range stack.Services {
			if canonicalise(svc.Dir) == dirPath {
				return svc
			}
		}
	}
	return nil
}

func (r *Resolver) possibleFirstParts() (map[string]Runnable, error) {
	possibilities := make(map[string]Runnable)

	for name, stack := range r.stacks {
		possibilities[name] = stack
	}

	for name, service := range r.stacks["default"].Services {
		possibilities[name] = service
	}

	if curSvc := r.findServiceByDir(); curSvc != nil {
		for name, process := range curSvc.Processes {
			possibilities[name] = process
		}
	}

	return possibilities, nil
}

func (r *Resolver) parseService(s *Stack, parts []string) error {
	if len(parts) > 0 {
		if svc, exists := s.Services[parts[0]]; exists {
			r.service = svc
			return r.parseProcess(svc, parts[1:])
		} else {
			return fmt.Errorf("Invalid process '%v'", parts)
		}
	}
	return nil
}

func (r *Resolver) parseProcess(s *Service, parts []string) error {
	if len(parts) > 0 {
		if p, exists := s.Processes[parts[0]]; exists {
			r.process = p
		} else {
			return fmt.Errorf("Invalid process '%v'", parts)
		}
	}
	return nil
}

func (r *Resolver) parseStack(parts []string) error {
	possibilities, err := r.possibleFirstParts()
	if err != nil {
		return err
	}

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
		return fmt.Errorf("Invalid descriptor '%v'", r.str)
	}
	return nil
}

func (r *Resolver) GetDescriptor() (*rig.Descriptor, error) {
	d := &rig.Descriptor{}

	parts := strings.SplitN(r.str, ":", 3)

	// Handle special case of colon as first character (sibling)
	if len(r.str) > 0 && r.str[0] == ':' {
		// Check that we're in a service directory
		if curSvc := r.findServiceByDir(); curSvc != nil {
			// We are, set the stack and parse the string from the servce down
			r.stack = curSvc.Stack
			if err := r.parseService(curSvc.Stack, parts[1:]); err != nil {
				return nil, err
			}
			// Everything went well, proceed to the build the descriptor below
		} else {
			// Trying to use the sibling specifier (: prefix) outside of a
			// service directory
			return nil, fmt.Errorf("Can't use sibling specifier here")
		}
	} else {
		// Check for empty descriptors
		if len(parts) == 1 && parts[0] == "" {
			// An empty descriptor in a service directory is allowed
			if curSvc := r.findServiceByDir(); curSvc != nil {
				d.Stack = curSvc.Stack.Name
				d.Service = curSvc.Name
				return d, nil
			}

			return nil, fmt.Errorf("Empty descriptor")
		}

		// Non-empty descriptor, parse away!
		if err := r.parseStack(parts); err != nil {
			return nil, err
		}
	}

	// Build a descriptor from the results of the resolver
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
