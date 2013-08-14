package main

import (
	"fmt"
	"github.com/gocardless/rig"
	"path/filepath"
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

func canonicalise(path string) string {
	// Hide errors, errors are boring
	path, _ = filepath.Abs(path)
	path, _ = filepath.EvalSymlinks(path)
	return path
}

func (r *Resolver) findServiceByDir() *Service {
	dirPath := canonicalise(r.dir)
	for _, stack := range r.config.Stacks {
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

	for name, stack := range r.config.Stacks {
		possibilities[name] = stack
	}

	for name, service := range r.config.Stacks["default"].Services {
		possibilities[name] = service
	}

	if curSvc := r.findServiceByDir(); curSvc != nil {
		for name, process := range curSvc.Processes {
			fmt.Printf("canon %q\n", name)
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
		return fmt.Errorf("Invalid descriptor '%q'", r.str)
	}
	return nil
}

func (r *Resolver) GetDescriptor() (*rig.Descriptor, error) {
	d := &rig.Descriptor{}

	// TODO handle special case of colon as first character (sibling)
	parts := strings.SplitN(r.str, ":", 3)

	if len(parts) == 1 && parts[0] == "" {
		if curSvc := r.findServiceByDir(); curSvc != nil {
			d.Stack = curSvc.Stack.Name
			d.Service = curSvc.Name
			return d, nil
		}

		return nil, fmt.Errorf("Empty descriptor")
	}

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
