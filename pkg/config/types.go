package config

type AgentConfig struct {
	HarnessCommand []string `yaml:"harness_command"`
	Image          string   `yaml:"image"`
	Mounts         []Mount  `yaml:"mounts"`
}

type Mount struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
	ReadOnly    bool   `yaml:"readOnly"`
}

type Config struct {
	Agents map[string]*AgentConfig `yaml:"agents"`
}
