package main

import (
	"flag"
	"github.com/gocardless/rig/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// This is a copy-paste from rig.go
var (
	defaultProto string = "http"
	defaultAddr  string = "0.0.0.0:9696"
)

func main() {
	// config flag
	configFlag := flag.String("-c", "~/.config/rig/config.json", "Path to config")
	flag.Parse()

	configFilename := utils.ExpandPath(*configFlag)
	launchServer(configFilename)
}

func launchServer(configFilename string) {
	config, err := LoadConfigFromFile(configFilename)
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, os.Signal(syscall.SIGTERM))
	go func() {
		sig := <-c
		log.Printf("Received signal %v. Exiting...\n", sig)
		os.Exit(0)
	}()

	srv, err := NewServerFromConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	if err := ListenAndServe(defaultAddr, srv); err != nil {
		log.Fatal(err)
	}
}
