package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	//"sync"
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

func main() {
	//launchServer()

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %v configdir\n", os.Args[0])
		os.Exit(1)
	}
	config := NewConfigFromDir(os.Args[1])

	for _, s := range(config.Stacks) {
		s.Start()
	}
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
