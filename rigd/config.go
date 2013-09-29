package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const defaultConfig = `
{
	"stacks": {
		"default": {}
	}
}
`

type Config struct {
	Filename string
	Stacks   map[string]*StackConfig `json:"stacks,omitempty"`
}

type StackConfig struct {
	Services map[string]*ServiceConfig `json:"services,omitempty"`
}

type ServiceConfig struct {
	Dir string `json:"dir,omitempty"`
}

func LoadConfigFromFile(filename string) (*Config, error) {
	log.Printf("Loading config from '%s'\n", filename)

	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			log.Printf("No config found, creating one\n")
			if err = createDefaultConfig(filename); err != nil {
				return nil, err
			}
		}
	}

	configData, err := ioutil.ReadFile(filename)
	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}
	config.Filename = filename

	_, exist := config.Stacks["default"]
	if !exist {
		config.Stacks["default"] = &StackConfig{}
		// Not very happy about this, maybe the defaultConfig should be
		// a Config struct so this logic is written into writeConfig
		b, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}
		writeConfig(filename, string(b))
	}

	return &config, nil
}

func createDefaultConfig(filename string) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	if err := writeConfig(filename, defaultConfig); err != nil {
		return err
	}
	return nil
}

func writeConfig(filename string, config string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = f.WriteString(config)
	if err != nil {
		return err
	}
	return f.Close()
}
