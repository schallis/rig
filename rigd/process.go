package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const (
	Stopped = iota
	Running
)

type ProcessStatus int

type Process struct {
	Name             string
	Cmd              string
	Service          *Service
	Status           ProcessStatus
	Process          *os.Process
	outputDispatcher *ProcessOutputDispatcher
}

func NewProcess(name, cmd string, service *Service) *Process {
	return &Process{
		Name:             name,
		Cmd:              cmd,
		Service:          service,
		Status:           Stopped,
		outputDispatcher: NewProcessOutputDispatcher(),
	}
}

func (p *Process) Start(dir string) error {
	if p.Status != Stopped {
		return fmt.Errorf("process %v is already running", p.Name)
	}

	cmd := exec.Command("/bin/sh", "-c", p.Cmd)
	cmd.Dir = dir

	p.streamOutput(cmd)

	log.Printf("Starting process '%v'\n", p.Name)
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting process '%v': %v\n", p.Name, err)
		return err
	}
	p.Process = cmd.Process
	p.Status = Running
	defer p.setStatus(Stopped)

	if err := cmd.Wait(); err != nil {
		log.Printf("Process '%v' failed: %v\n", p.Name, err)
		return err
	} else {
		log.Printf("Process '%v' stopped\n", p.Name)
	}

	return nil
}

func (p *Process) Stop() error {
	if p.Status != Running {
		return fmt.Errorf("can't stop a process that isn't running")
	}

	p.Process.Signal(syscall.SIGTERM)

	return nil
}

func (p *Process) SubscribeToOutput(c chan ProcessOutputMessage) *ProcessOutputSubscription {
	return p.outputDispatcher.Subscribe(c)
}

func (p *Process) setStatus(status ProcessStatus) {
	p.Status = status
}

func (p *Process) streamOutput(cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	go p.logStream(stdout, "stdout")
	go p.logStream(stderr, "stderr")
}

func (p *Process) logStream(stream io.ReadCloser, streamName string) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		msg := ProcessOutputMessage{
			Content: scanner.Text(),
			Stack:   p.Service.Stack.Name,
			Service: p.Service.Name,
			Process: p.Name,
			Time:    time.Now(),
		}
		p.outputDispatcher.Publish(msg)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error reading %v: %v\n", streamName, err)
	}
}
