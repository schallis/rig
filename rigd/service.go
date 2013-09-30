package main

import (
	"bufio"
	"fmt"
	"github.com/gocardless/rig"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

type Service struct {
	Name      string
	Dir       string
	Stack     *Stack
	Processes map[string]*Process
}

func NewService(name string, dir string, stack *Stack) (*Service, error) {
	s := &Service{
		Name:      name,
		Dir:       dir,
		Stack:     stack,
		Processes: make(map[string]*Process),
	}

	if err := s.parseProcfile(path.Join(dir, "Procfile")); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Service) Start() error {
	var wg sync.WaitGroup
	for _, p := range s.Processes {
		wg.Add(1)
		go func(p *Process) {
			if err := p.Start(); err != nil {
				log.Printf("[S] %v\n", err)
			}
			wg.Done()
		}(p)
	}
	wg.Wait()
	return nil
}

func (s *Service) Stop() error {
	for _, p := range s.Processes {
		if err := p.Stop(); err != nil {
			log.Printf("[S] %v\n", err)
		}
	}
	return nil
}

func (s *Service) SubscribeToOutput(c chan rig.ProcessOutputMessage) {
	for _, p := range s.Processes {
		p.outputDispatcher.Subscribe(c)
	}
}

func (s *Service) parseProcfile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	lineNo := 1
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)

		if len(parts) != 2 {
			return fmt.Errorf("[S] Error parsing %v (line %v)", path, lineNo)
		}

		name := strings.TrimSpace(parts[0])
		cmd := strings.TrimSpace(parts[1])
		if name == "" || cmd == "" {
			return fmt.Errorf("[S] Error in procfile %v (line %v)", path, lineNo)
		}

		s.Processes[name] = NewProcess(name, cmd, s)

		lineNo += 1
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
