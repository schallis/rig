package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"fmt"
	"os/exec"
	"syscall"
)

const (
	Stopped = iota
	Running
)

type ProcessStatus int

type Process struct {
	Name    string
	Cmd     string
	Status  ProcessStatus
	Process *os.Process
}

func NewProcess(name, cmd string) *Process {
	return &Process{Name: name, Cmd: cmd, Status: Stopped}
}

func (p *Process) Start(dir string) error {
	cmd := exec.Command("/bin/sh", "-c", p.Cmd)
	cmd.Dir = dir

	p.logOutputStreams(cmd)

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

func (p *Process) setStatus(status ProcessStatus) {
	p.Status = status
}

func (p *Process) logOutputStreams(cmd *exec.Cmd) {
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
		log.Println(p.Name, "|", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error reading %v: %v\n", streamName, err)
	}
}
