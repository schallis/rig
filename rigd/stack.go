package main

import (
	"github.com/gocardless/rig"
	"container/ring"
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

func (s *Stack) SubscribeToOutput(c chan rig.ProcessOutputMessage, num int) {
	var buffers []*ring.Ring
	for _, svc := range s.Services {
		for _, p := range svc.Processes {
			buffers = append(buffers, p.buffer)
		}
	}

	tailBuffer := MultiTail(buffers, num)
	for _, msg := range tailBuffer {
		c <- *msg
	}

	for _, svc := range s.Services {
		for _, p := range svc.Processes {
			p.outputDispatcher.Subscribe(c)
		}
	}
}
