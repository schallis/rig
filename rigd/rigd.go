package main

import (
	"os"
	"io/ioutil"
	"fmt"
	"sync"
	"encoding/json"
	"github.com/gocardless/rig/logging"
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
	for _, process := range(processes) {
		if len(process.Name) > max {
			max = len(process.Name)
		}
	}
	return max
}

func main() {
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
	for name, processConfig := range(config.Processes) {
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
	wg.Add(len(processes))

	// Kick off all processes
	for _, process := range(processes) {
		go func(s Process) { s.Start(&wg) }(process)
	}
	wg.Wait()

	// Wait for the log dispatcher to finish
	d.Stop()
	<- doneChan
}

