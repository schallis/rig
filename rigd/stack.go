package main

import "sync"

type Stack struct {
	Name     string
	Services map[string]*Service
}

func (s *Stack) Start() {
	var wg sync.WaitGroup
	for _, svc := range(s.Services) {
		wg.Add(1)
		go func(svc *Service) {
			svc.Start()
			wg.Done()
		}(svc)
	}
	wg.Wait()
}

