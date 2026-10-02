package container

type TempMount struct {
	Source      string
	Destination string
}

type ProxyContainerInfo struct {
	Name            string
	NetworkName     string
	SquidPort       int
	SquidConfigPath string
}
