package main

import (
	"log"
	"io"
	"bufio"
	"os/exec"
	"github.com/gocardless/rig/logging"
)

type Process struct {
	Name   string
	Cmd	   string
	Dir    string
	Logger *logging.Logger
}

func (p *Process) Start() error {
	cmd := exec.Command("/bin/sh", "-c", p.Cmd)
	cmd.Dir = p.Dir

	p.logOutputStreams(cmd)

	p.Logger.Logf("Starting process '%v'", p.Name)
	if err := cmd.Start(); err != nil {
		p.Logger.Logf("Error starting process '%v': %v", p.Name, err)
		return err
	}

	if err := cmd.Wait(); err != nil {
		p.Logger.Logf("Process '%v' failed: %v", p.Name, err)
		return err
	} else {
		p.Logger.Logf("Process '%v' stopped", p.Name)
	}

	return nil
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
		p.Logger.Log(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		p.Logger.Logf("error reading %v: %v", streamName, err)
	}
}

