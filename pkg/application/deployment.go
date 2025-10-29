package application

type DeploymentProfile struct {
    Type             string        `yaml:"type"`
    ID               string        `yaml:"id"`
    Description      string        `yaml:"description,omitempty"`
    Components       []Component   `yaml:"components"`
    RequiredResources *Resources   `yaml:"requiredResources,omitempty"`
}

type Component struct {
    Name       string              `yaml:"name"`
    Properties ComponentProperties `yaml:"properties"`
}

type ComponentProperties struct {
    Repository     string `yaml:"repository,omitempty"`
    Revision       string `yaml:"revision,omitempty"`
    Wait           bool   `yaml:"wait,omitempty"`
    Timeout        string `yaml:"timeout,omitempty"`
    PackageLocation string `yaml:"packageLocation,omitempty"`
    KeyLocation     string `yaml:"keyLocation,omitempty"`
}

type Resources struct {
    CPU       CPUInfo     `yaml:"cpu,omitempty"`
    Memory    string   `yaml:"memory,omitempty"`
    Storage   string   `yaml:"storage,omitempty"`
    Peripherals []Peripheral `yaml:"peripherals,omitempty"`
    Interfaces  []Interface  `yaml:"interfaces,omitempty"`
}

type CPUInfo struct {
    Cores        float64  `yaml:"cores"`
    Architectures []string `yaml:"architectures,omitempty"`
}

type Peripheral struct {
    Type         string `yaml:"type"`
    Manufacturer string `yaml:"manufacturer,omitempty"`
    Model        string `yaml:"model,omitempty"`
}

type Interface struct {
    Type string `yaml:"type"`
}
