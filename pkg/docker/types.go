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
