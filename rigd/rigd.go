package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocardless/rig/logging"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// This is a copy-paste from rig.go
var (
	defaultProto string = "http"
	defaultAddr  string = "0.0.0.0:9696"
)

type ProcessConfig struct {
	Command string
	Dir     string
}

type Config struct {
	Processes map[string]ProcessConfig
}

func loadConfig(configFile string) Config {
	file, e := ioutil.ReadFile(configFile)
	if e != nil {
		fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", e)
		os.Exit(1)
	}

	var config Config
	err := json.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}

	return config
}

// Helper function to determine the longest process name, to help with
// displaying pretty terminal output
func maxNameWidth(processes ...Process) int {
	max := 0
	for _, process := range processes {
		if len(process.Name) > max {
			max = len(process.Name)
		}
	}
	return max
}

func main() {
	launchServer()

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %v config.json\n", os.Args[0])
		os.Exit(1)
	}
	config := loadConfig(os.Args[1])

	doneChan := make(chan bool)

	d := logging.NewDispatcher()
	go func() {
		d.Start()
		doneChan <- true
	}()

	// Initialise processes from the configuration
	var processes []Process
	for name, processConfig := range config.Processes {
		process := Process{
			Name:   name,
			Cmd:    processConfig.Command,
			Dir:    processConfig.Dir,
			Logger: logging.NewLogger(d, name),
		}
		processes = append(processes, process)
	}

	// Spawn a terminal subscriber, so we se the logs in the terminal
	nameWidth := maxNameWidth(processes...)
	go logging.NewTerminalSubscriber(d, nameWidth)

	var wg sync.WaitGroup

	// Kick off all processes
	for _, process := range processes {
		wg.Add(1)
		go func(s Process) {
			s.Start()
			wg.Done()
		}(process)
	}
	wg.Wait()

	// Wait for the log dispatcher to finish
	d.Stop()
	<-doneChan
}

func launchServer() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, os.Signal(syscall.SIGTERM))
	go func() {
		sig := <-c
		log.Printf("Received signal %v. Exiting...\n", sig)
		os.Exit(0)
	}()

	srv, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}

	if err := ListenAndServe(defaultAddr, srv); err != nil {
		log.Fatal(err)
	}
}
