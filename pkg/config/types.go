package config

import "fmt"

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

func (c *Config) Validate() error {
	for name, agent := range c.Agents {
		if agent.Image == "" {
			return fmt.Errorf("agent '%s': image is required", name)
		}
		if len(agent.HarnessCommand) == 0 {
			return fmt.Errorf("agent '%s': harness_command is required", name)
		}
	}
	return nil
}
