package main

import (
	"io/ioutil"
	"os"
	"log"
	"path"
)

type Config struct {
	Stacks map[string]*Stack
}

func addServiceToStack(path string, info os.FileInfo, stack *Stack) error {
	dir, err := os.Readlink(path)
	if err != nil {
		log.Panicf("error: reading service link %v (%v)\n", path, err)
	}

	service, err := NewService(info.Name(), dir, stack)
	if err != nil {
		return err
	}
	stack.Services[service.Name] = service

	return nil
}

func NewConfig() *Config {
	return &Config{Stacks: make(map[string]*Stack)}
}

func NewConfigFromDir(configDir string) *Config {
	config := NewConfig()
	defaultStack := NewStack("default")
	config.Stacks[defaultStack.Name] = defaultStack

	entries, err := ioutil.ReadDir(configDir)
	if err != nil {
		log.Panicf("error: listing config dir %v (%v)\n", configDir, err)
	}

	for _, entry := range(entries) {
		entryPath := path.Join(configDir, entry.Name())

		// If we find a symlink in the config dir, add to the default stack
		// as a service
		if entry.Mode() & os.ModeSymlink != 0 {
			err := addServiceToStack(entryPath, entry, defaultStack)
			if err != nil {
				log.Panicln("error:", err)
			}
		}
	}

	return config
}

