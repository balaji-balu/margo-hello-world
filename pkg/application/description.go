package application

import (
    //"gopkg.in/yaml.v3"
    "fmt"
    //"errors"
    "github.com/goccy/go-yaml"
    "os"
)

type ApplicationDescription struct {
    APIVersion        string               `yaml:"apiVersion"`
    Kind              string               `yaml:"kind"`
    Metadata          Metadata             `yaml:"metadata"`
    DeploymentProfiles []DeploymentProfile `yaml:"deploymentProfiles"`
    Parameters        map[string]Parameter `yaml:"parameters,omitempty"`
    Configuration     *Configuration       `yaml:"configuration,omitempty"`
}

type Metadata struct {
    ID          string     `yaml:"id"`
    Name        string     `yaml:"name"`
    Description string     `yaml:"description,omitempty"`
    Version     string     `yaml:"version"`
    Catalog     Catalog    `yaml:"catalog"`
}

func ParseFromFile(path string) (*ApplicationDescription, error) {

    cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return nil,err
	}

	// Print the current working directory
	fmt.Println("Current working directory:", cwd)
    fmt.Printf("Parsing %s\n", path)

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var ad ApplicationDescription
    //var raw map[string]interface{}
    if err := yaml.Unmarshal([]byte(data), &ad); err != nil {
        return nil, err
    }
    //fmt.Printf("%#v\n", raw)
    //return nil, errors.New("not implemented")
    return &ad, nil
}
