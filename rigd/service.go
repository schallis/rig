package main

import (
	"sync"
	"log"
)

type Service struct {
	Name      string
	Dir       string
	Processes []*Process
}

func (s *Service) Start() {
	var wg sync.WaitGroup
	for _, p := range(s.Processes) {
		wg.Add(1)
		go func(p *Process) {
			if err := p.Start(); err != nil {
				log.Printf("service: error from process %v (%v)", p.Name, err)
			}
			wg.Done()
		}(p)
	}
	wg.Wait()
}

