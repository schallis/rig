package main

import (
	"os"
	"io/ioutil"
	"fmt"
	"sync"
	"encoding/json"
	"github.com/gocardless/rig/logging"
)

type TaskConfig struct {
	Command string
	Dir     string
}

type Config struct {
	Tasks map[string]TaskConfig
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

// Helper function to determine the longest task name, to help with
// displaying pretty terminal output
func maxNameWidth(tasks ...Task) int {
	max := 0
	for _, task := range(tasks) {
		if len(task.Name) > max {
			max = len(task.Name)
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

	// Initialise tasks from the configuration
	var tasks []Task
	for name, taskConfig := range(config.Tasks) {
		task := Task{
			Name:   name,
			Cmd:    taskConfig.Command,
			Dir:    taskConfig.Dir,
			Logger: logging.NewLogger(d, name),
		}
		tasks = append(tasks, task)
	}

	// Spawn a terminal subscriber, so we se the logs in the terminal
	nameWidth := maxNameWidth(tasks...)
	go logging.NewTerminalSubscriber(d, nameWidth)

	var wg sync.WaitGroup
	wg.Add(len(tasks))

	// Kick off all tasks
	for _, task := range(tasks) {
		go func(s Task) { s.Start(&wg) }(task)
	}
	wg.Wait()

	// Wait for the log dispatcher to finish
	d.Stop()
	<- doneChan
}

