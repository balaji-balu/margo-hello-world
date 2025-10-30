package cfffg

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Domain string `yaml:"domain"`
		Site string `yaml:"site"`
		Port int    `yaml:"port"`
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
