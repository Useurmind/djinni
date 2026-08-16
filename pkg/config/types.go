package config

import "fmt"

type AgentConfig struct {
	HarnessCommand        []string          `yaml:"harness_command"`
	Image                 string            `yaml:"image"`
	Containerfile         string            `yaml:"containerfile"`
	Mounts                []Mount           `yaml:"mounts"`
	FilesToCopy           []FilesToCopy     `yaml:"files_to_copy,omitempty"`
	DefaultModel          string            `yaml:"default_model"`
	GitWorkspace          GitWorkspaceMount `yaml:"git_workspace"`
	SyncApproach          string            `yaml:"sync_approach,omitempty"`
	AutoDeleteAgentBranch bool              `yaml:"autodelete_agent_branch,omitempty"`
}

type Mount struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
	ReadOnly    bool   `yaml:"readOnly"`
}

type GitWorkspaceMount struct {
	BaseDirectory string `yaml:"base_directory"`
}

type FilesToCopy struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
}

type Model struct {
	ID string `yaml:"id"`
}

type ModelProvider struct {
	Name    string  `yaml:"name"`
	APIBase string  `yaml:"apiBase"`
	APIKey  string  `yaml:"apiKey"`
	Models  []Model `yaml:"models"`
}

type GlobalConfig struct {
	ModelProviders []ModelProvider `yaml:"modelProviders"`
}

type Config struct {
	Agents       map[string]*AgentConfig `yaml:"agents"`
	DefaultModel string                  `yaml:"default_model"`
}

func (c *Config) Validate() error {
	if c.Agents == nil {
		c.Agents = make(map[string]*AgentConfig)
	}
	for name, agent := range c.Agents {
		if name == "" {
			return fmt.Errorf("agent name cannot be empty")
		}
		if agent.Image == "" && agent.Containerfile == "" {
			return fmt.Errorf("agent '%s': either image or containerfile is required", name)
		}
		if agent.Image != "" && agent.Containerfile != "" {
			return fmt.Errorf("agent '%s': cannot specify both image and containerfile", name)
		}
		if len(agent.HarnessCommand) == 0 {
			return fmt.Errorf("agent '%s': harness_command is required", name)
		}
		if agent.GitWorkspace.BaseDirectory == "" {
			agent.GitWorkspace.BaseDirectory = "/tmp/djinni"
		}
		if agent.SyncApproach != "" && agent.SyncApproach != "none" && agent.SyncApproach != "gitpatch" && agent.SyncApproach != "automerge" {
			return fmt.Errorf("agent '%s': sync_approach must be 'none', 'gitpatch', or 'automerge'", name)
		}
	}
	return nil
}

func (p *ModelProvider) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("model provider name cannot be empty")
	}
	if len(p.Models) == 0 {
		return fmt.Errorf("model provider '%s': at least one model is required", p.Name)
	}
	return nil
}
