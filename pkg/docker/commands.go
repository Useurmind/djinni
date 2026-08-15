package docker

type ContainerCommands struct {
	PreCommands  []string `yaml:"pre,omitempty"`
	PostCommands []string `yaml:"post,omitempty"`
	FilesToCopy  []FilesToCopy
}
