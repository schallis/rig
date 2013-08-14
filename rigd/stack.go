package main

import (
	"github.com/gocardless/rig"

	"sync"
)

type Stack struct {
	Name     string
	Services map[string]*Service
}

func NewStack(name string) *Stack {
	return &Stack{Name: name, Services: make(map[string]*Service)}
}

func (s *Stack) Start() error {
	var wg sync.WaitGroup
	for _, svc := range s.Services {
		wg.Add(1)
		go func(svc *Service) {
			svc.Start()
			wg.Done()
		}(svc)
	}
	wg.Wait()

	return nil
}

func (s *Stack) Stop() error {
	for _, svc := range s.Services {
		svc.Stop()
	}
	return nil
}

func (s *Stack) SubscribeToOutput(c chan rig.ProcessOutputMessage) {
	for _, svc := range s.Services {
		svc.SubscribeToOutput(c)
	}
}
