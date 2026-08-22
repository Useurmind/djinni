package docker

import "github.com/useurmind/djinni/pkg/config"

type TmpfsMount struct {
	Destination string
	Size        string
}

type ContainerCommands struct {
	PreCommands          []string
	PostCommands         []string
	FilesToCopy          []config.FilesToCopy
	ForceReadOnlyRootOff bool
	TmpfsMounts          []TmpfsMount
	WritablePaths        []config.WritablePath
	TempMount            *TempMount
}
