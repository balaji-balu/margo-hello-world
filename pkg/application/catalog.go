package application

type Catalog struct {
    Application   *ApplicationMetadata `yaml:"application,omitempty"`
    Author        []Author             `yaml:"author,omitempty"`
    Organization  []Organization       `yaml:"organization"`
}

type ApplicationMetadata struct {
    DescriptionFile string   `yaml:"descriptionFile,omitempty"`
    Icon            string   `yaml:"icon,omitempty"`
    LicenseFile     string   `yaml:"licenseFile,omitempty"`
    ReleaseNotes    string   `yaml:"releaseNotes,omitempty"`
    Site            string   `yaml:"site,omitempty"`
    Tagline         string   `yaml:"tagline,omitempty"`
    Tags            []string `yaml:"tags,omitempty"`
}

type Author struct {
    Name  string `yaml:"name,omitempty"`
    Email string `yaml:"email,omitempty"`
}

type Organization struct {
    Name string `yaml:"name"`
    Site string `yaml:"site,omitempty"`
}
