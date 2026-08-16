package docker

type TmpfsMount struct {
	Destination string `yaml:"destination"`
	Size        string `yaml:"size,omitempty"`
}

type ContainerCommands struct {
	PreCommands          []string `yaml:"pre,omitempty"`
	PostCommands         []string `yaml:"post,omitempty"`
	FilesToCopy          []FilesToCopy
	ForceReadOnlyRootOff bool
	TmpfsMounts          []TmpfsMount
}
