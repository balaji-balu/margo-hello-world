package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		//Domain string `yaml:"domain"`
		Site    string `yaml:"site"`
		Node    string `yaml:"node"`
		Runtime string `yaml:"runtime"`
		Region  string `yaml:"region"`
		Port    int    `yaml:"port"`
	} `yaml:"server"`
	CO struct {
		URL string `yaml:"url"`
	} `yaml:"co"`
	Database struct {
		URL string `yaml:"url"`
	} `yaml:"database"`
	LO struct {
		URL string `yaml:"url"`
	} `yaml:"lo"`
	NATS struct {
		URL string `yaml:"url"`
	} `yaml:"nats"`
	Git struct {
		Repo   string `yaml:"repo"`
		Owner  string `yaml:"owner"`
		Branch string `yaml:"branch"`
	}
	Appregistry struct {
		Repo string `yaml:"repo"`
		Branch string `yaml:"branch"`
	}
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}
