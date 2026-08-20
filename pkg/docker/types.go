package docker

import (
	"github.com/useurmind/djinni/pkg/utils"
)

type Mount struct {
	Source      string
	Destination string
	ReadOnly    bool
}

type FilesToCopy struct {
	Source      string
	Destination string
	name        string
}

func (f *FilesToCopy) Name() string {
	if f.name == "" {
		f.name = utils.HashSourcePath(f.Source)
	}
	return f.name
}

type TempMount struct {
	Source      string
	Destination string
}

type WritablePath struct {
	Name        string
	Destination string
}
