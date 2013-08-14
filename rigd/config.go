package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
)

type Config struct {
	Stacks map[string]*Stack
	Dir    string
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

func loadStackServices(stack *Stack, dir string) {
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Panicf("error: listing config dir %v (%v)\n", dir, err)
	}

	for _, entry := range entries {
		entryPath := path.Join(dir, entry.Name())

		// If we find a symlink, add to the stack as a service
		if entry.Mode()&os.ModeSymlink != 0 {
			err := addServiceToStack(entryPath, entry, stack)
			if err != nil {
				log.Panicln("error:", err)
			}
		}
	}
}

func NewConfig() *Config {
	return &Config{Stacks: make(map[string]*Stack)}
}

func NewConfigFromDir(configDir string) *Config {
	config := NewConfig()
	config.Dir = configDir

	defaultStack := NewStack("default")
	config.Stacks[defaultStack.Name] = defaultStack
	loadStackServices(defaultStack, configDir)

	entries, err := ioutil.ReadDir(configDir)
	if err != nil {
		log.Panicf("error: listing config dir %v (%v)\n", configDir, err)
	}

	for _, entry := range entries {
		entryPath := path.Join(configDir, entry.Name())

		// If we find a directory in the config dir, add it as a stack
		if entry.IsDir() {
			stack := NewStack(entry.Name())
			config.Stacks[stack.Name] = stack
			loadStackServices(stack, entryPath)
		}
	}

	return config
}
