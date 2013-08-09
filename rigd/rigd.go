package main

import (
	"os"
	"io/ioutil"
	"fmt"
	"sync"
	"encoding/json"
	"github.com/gocardless/rig/logging"
)

type ServiceConfig struct {
	Command string
	Dir     string
}

type Config struct {
	Services map[string]ServiceConfig
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

// Helper function to determine the longest service name, to help with
// displaying pretty terminal output
func maxNameWidth(services ...Service) int {
	max := 0
	for _, service := range(services) {
		if len(service.Name) > max {
			max = len(service.Name)
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

	// Initialise services from the configuration
	var services []Service
	for name, serviceConfig := range(config.Services) {
		service := Service{
			Name:   name,
			Cmd:    serviceConfig.Command,
			Dir:    serviceConfig.Dir,
			Logger: logging.NewLogger(d, name),
		}
		services = append(services, service)
	}

	// Spawn a terminal subscriber, so we se the logs in the terminal
	nameWidth := maxNameWidth(services...)
	go logging.NewTerminalSubscriber(d, nameWidth)

	var wg sync.WaitGroup
	wg.Add(len(services))

	// Kick off all services
	for _, service := range(services) {
		go func(s Service) { s.Start(&wg) }(service)
	}
	wg.Wait()

	// Wait for the log dispatcher to finish
	d.Stop()
	<- doneChan
}

