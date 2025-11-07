package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type TriggerConfig struct {
	Type      string        `yaml:"type"`
	Interval  time.Duration `yaml:"interval"`
	RepoOwner string        `yaml:"repo_owner"`
	RepoName  string        `yaml:"repo_name"`
	YamlPath  string        `yaml:"yaml_path"`
	TokenEnv  string        `yaml:"token_env"`
}

type BrokerConfig struct {
	URL   string `yaml:"url"`
	Topic string `yaml:"topic"`
}

type Config struct {
	Trigger TriggerConfig `yaml:"trigger"`
	Broker  BrokerConfig  `yaml:"broker"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}
