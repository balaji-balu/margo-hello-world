package application

type Parameter struct {
	name string `yaml:"name"`
    Value   string   `yaml:"value,omitempty"`
    Targets []Target `yaml:"targets,omitempty"`
}

type Target struct {
    Pointer    string   `yaml:"pointer"`
    Components []string `yaml:"components"`
}