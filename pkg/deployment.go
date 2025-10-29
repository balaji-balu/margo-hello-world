package pkg

type Deployment struct {
    ApiVersion string `yaml:"apiVersion"`
    Kind       string `yaml:"kind"`
    Metadata   struct {
        Name string `yaml:"name"`
    } `yaml:"metadata"`
    Spec struct {
        Nodes []struct {
            Id       string    `yaml:"id"`
            Services []Service `yaml:"services"`
        } `yaml:"nodes"`
    } `yaml:"spec"`
}

type Service struct {
    Name      string `yaml:"name"`
    Container struct {
        Image   string   `yaml:"image"`
        Command []string `yaml:"command"`
    } `yaml:"container"`
}
