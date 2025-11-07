package application

type Configuration struct {
	Sections []Section `yaml:"sections"`
	Schema   []Schema  `yaml:"schema"`
}

type Section struct {
	Name     string    `yaml:"name"`
	Settings []Setting `yaml:"settings"`
}

type Setting struct {
	Parameter   string `yaml:"parameter"`
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Immutable   bool   `yaml:"immutable,omitempty"`
	Schema      string `yaml:"schema"`
}

type Schema struct {
	Name       string   `yaml:"name"`
	DataType   string   `yaml:"dataType"`
	AllowEmpty bool     `yaml:"allowEmpty,omitempty"`
	MinValue   *float64 `yaml:"minValue,omitempty"`
	MaxValue   *float64 `yaml:"maxValue,omitempty"`
}
