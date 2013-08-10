package main

import (
	"log"
	"io"
	"bufio"
	"sync"
	"os/exec"
	"github.com/gocardless/rig/logging"
)

type Task struct {
	Name   string
	Cmd	   string
	Dir    string
	Logger *logging.Logger
}

func (s *Task) Start(wg *sync.WaitGroup) {
	cmd := exec.Command("/bin/sh", "-c", s.Cmd)
	cmd.Dir = s.Dir

	s.logOutputStreams(cmd)

	s.Logger.Logf("Starting task '%v'", s.Name)
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Wait(); err != nil {
		s.Logger.Logf("Task '%v' failed: %v", s.Name, err)
	} else {
		s.Logger.Logf("Task '%v' stopped", s.Name)
	}

	wg.Done()
}

func (s *Task) logOutputStreams(cmd *exec.Cmd) {
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

func (s *Task) logStream(stream io.ReadCloser, streamName string) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		s.Logger.Log(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		s.Logger.Logf("error reading %v: %v", streamName, err)
	}
}

