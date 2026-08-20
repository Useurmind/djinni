package docker

type TmpfsMount struct {
	Destination string
	Size        string
}

type ContainerCommands struct {
	PreCommands          []string
	PostCommands         []string
	FilesToCopy          []FilesToCopy
	ForceReadOnlyRootOff bool
	TmpfsMounts          []TmpfsMount
	WritablePaths        []WritablePath
	TempMount            *TempMount
}
