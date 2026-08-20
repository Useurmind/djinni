package docker

type Mount struct {
	Source      string
	Destination string
	ReadOnly    bool
}

type FilesToCopy struct {
	Source      string
	Destination string
}

type TempMount struct {
	Source      string
	Destination string
}

type WritablePath struct {
	Name        string
	Destination string
}
