package docker

type Mount struct {
	Source      string
	Destination string
	ReadOnly    bool
}
