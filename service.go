package main

import (
	"log"
	"io"
	"bufio"
	"sync"
	"os/exec"
	"github.com/hmarr/ignition/logging"
)

type Service struct {
	Name   string
	Cmd	   string
	Dir    string
	logger *logging.Logger
}

func (s *Service) Start(wg *sync.WaitGroup) {
	cmd := exec.Command("/bin/sh", "-c", s.Cmd)
	cmd.Dir = s.Dir

	s.logOutputStreams(cmd)

	s.logger.Logf("Starting service '%v'", s.Name)
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Wait(); err != nil {
		s.logger.Logf("Service '%v' failed: %v", s.Name, err)
	} else {
		s.logger.Logf("Service '%v' stopped", s.Name)
	}

	wg.Done()
}

func (s *Service) logOutputStreams(cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	go s.logStream(stdout, "stdout")
	go s.logStream(stderr, "stderr")
}

func (s *Service) logStream(stream io.ReadCloser, streamName string) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		s.logger.Log(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		s.logger.Logf("error reading %v: %v", streamName, err)
	}
}

