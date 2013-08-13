package main

import (
	"bufio"
	"fmt"
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

func (s *Service) Start() {
	var wg sync.WaitGroup
	for _, p := range s.Processes {
		wg.Add(1)
		go func(p *Process) {
			if err := p.Start(); err != nil {
				log.Printf("service: error from process %v (%v)\n", p.Name, err)
			}
			wg.Done()
		}(p)
	}
	wg.Wait()
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
			return fmt.Errorf("error parsing %v (line %v)", path, lineNo)
		}

		name := strings.TrimSpace(parts[0])
		cmd := strings.TrimSpace(parts[1])
		if name == "" || cmd == "" {
			return fmt.Errorf("error in procfile %v (line %v)", path, lineNo)
		}

		s.Processes[name] = NewProcess(name, cmd, s)

		lineNo += 1
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
