package main

import (
	"bufio"
	"container/ring"
	"fmt"
	"github.com/gocardless/rig"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sync"
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
	buffer           *ring.Ring
	bufferMutex      sync.Mutex
}

func NewProcess(name, cmd string, service *Service) *Process {
	return &Process{
		Name:             name,
		Cmd:              cmd,
		Service:          service,
		Status:           Stopped,
		outputDispatcher: NewProcessOutputDispatcher(),
		buffer:           ring.New(100),
	}
}

func getUserShell() string {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	pw := Getpwnam(user.Name)
	return pw.Shell
}

func (p *Process) Start() error {
	if p.Status != Stopped {
		return fmt.Errorf("Process '%s' is already running", p.Sqd())
	}

	shell := getUserShell()
	var opts []string
	switch filepath.Base(shell) {
	case "zsh":
		opts = []string{"-i", "-l", "-c", p.Cmd}
	default:
		opts = []string{"-l", "-c", p.Cmd}
	}
	cmd := exec.Command(shell, opts...)
	cmd.Dir = p.Service.Dir
	cmd.Env = os.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[P] Starting process %s\n", p.Sqd())
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Error starting process %s: %v", p.Sqd(), err)
	}
	p.Process = cmd.Process
	p.Status = Running
	defer p.setStatus(Stopped)

	// Cmd.Wait() closes the fds, so we need to wait for reading to finish first
	var wg sync.WaitGroup
	wg.Add(2)
	go p.logStream(stdout, "stdout", &wg)
	go p.logStream(stderr, "stderr", &wg)
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s failed: %v", p.Sqd(), err)
	} else {
		log.Printf("[P] Process %s stopped\n", p.Sqd())
	}

	return nil
}

func (p *Process) Stop() error {
	if p.Status != Running {
		return fmt.Errorf("Can't stop: %s isn't running", p.Sqd())
	}

	p.Process.Signal(syscall.SIGTERM)

	return nil
}

func (p *Process) SubscribeToOutput(c chan rig.ProcessOutputMessage, num int) {
	tailBuffer := MultiTail([]*ring.Ring{p.buffer}, num)
	for _, msg := range tailBuffer {
		c <- *msg
	}
	p.outputDispatcher.Subscribe(c)
}

func (p *Process) setStatus(status ProcessStatus) {
	p.Status = status
}

func (p *Process) appendToBuffer(msg rig.ProcessOutputMessage) {
	p.bufferMutex.Lock()
	defer p.bufferMutex.Unlock()

	if p.buffer.Value != nil {
		p.buffer = p.buffer.Next()
	}
	p.buffer.Value = msg
}

func (p *Process) logStream(stream io.ReadCloser, name string, wg *sync.WaitGroup) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		msg := rig.ProcessOutputMessage{
			Content: scanner.Text(),
			Stack:   p.Service.Stack.Name,
			Service: p.Service.Name,
			Process: p.Name,
			Time:    time.Now(),
		}
		p.outputDispatcher.Publish(msg)
		p.appendToBuffer(msg)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading stdout for %s: %v\n", p.Sqd(), err)
	}

	wg.Done()
}

// Fully qualified descriptor: stack:service:process
func (p *Process) Fqd() string {
	return fmt.Sprintf("%s:%s:%s", p.Service.Stack.Name, p.Service.Name, p.Name)
}

// Semi qualified descriptor: service:process
func (p *Process) Sqd() string {
	return fmt.Sprintf("%s:%s", p.Service.Name, p.Name)
}
