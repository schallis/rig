package main

import (
	"log"
	"io"
	"bufio"
	"os/exec"
)

type Process struct {
	Name   string
	Cmd	   string
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

	if err := cmd.Wait(); err != nil {
		log.Printf("Process '%v' failed: %v\n", p.Name, err)
		return err
	} else {
		log.Printf("Process '%v' stopped\n", p.Name)
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
		log.Println(p.Name, "|", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error reading %v: %v\n", streamName, err)
	}
}

