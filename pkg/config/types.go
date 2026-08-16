package config

import "fmt"

// AgentConfig defines the configuration for an AI agent
type AgentConfig struct {
	// HarnessCommand is the command to run inside the container (required)
	HarnessCommand []string `yaml:"harness_command"`
	// Image is the Docker image to use for the agent (mutually exclusive with Containerfile)
	Image string `yaml:"image"`
	// Containerfile is the path to a local Containerfile to build an image from (mutually exclusive with Image)
	Containerfile string `yaml:"containerfile"`
	// Mounts are volume mounts to make files and directories available in the container
	Mounts []Mount `yaml:"mounts"`
	// FilesToCopy are files to copy into the container (e.g., .gitconfig)
	FilesToCopy []FilesToCopy `yaml:"files_to_copy,omitempty"`
	// DefaultModel overrides the global default LLM model for this agent
	DefaultModel string `yaml:"default_model"`
	// GitWorkspace configures the git workspace directory inside the container
	GitWorkspace GitWorkspaceMount `yaml:"git_workspace"`
	// SyncApproach is the strategy for syncing changes back to the workspace: none, gitpatch, or automerge
	SyncApproach string `yaml:"sync_approach,omitempty"`
	// AutoDeleteAgentBranch automatically deletes the feature branch after sync
	AutoDeleteAgentBranch bool `yaml:"autodelete_agent_branch,omitempty"`
}

// Mount represents a volume mount from host to container
type Mount struct {
	// Source is the path on the host machine (required)
	Source string `yaml:"source"`
	// Destination is the path inside the container (required)
	Destination string `yaml:"destination"`
	// ReadOnly mounts the volume as read-only (default: false)
	ReadOnly bool `yaml:"readOnly"`
}

// GitWorkspaceMount configures the git workspace directory inside the container
type GitWorkspaceMount struct {
	// BaseDirectory is the base directory for git operations (default: /tmp/djinni)
	BaseDirectory string `yaml:"base_directory"`
}

// FilesToCopy represents a file to copy into the container
type FilesToCopy struct {
	// Source is the path on the host machine (required)
	Source string `yaml:"source"`
	// Destination is the path inside the container (required)
	Destination string `yaml:"destination"`
}

// Model represents a model available from a provider
type Model struct {
	// ID is the model identifier used in agent configuration
	ID string `yaml:"id"`
}

// ModelProvider configures a model provider for LLM integration
type ModelProvider struct {
	// Name is the unique identifier for the provider (required)
	Name string `yaml:"name"`
	// APIBase is the base URL for the provider's API endpoint (required)
	APIBase string `yaml:"apiBase"`
	// APIKey is the API key or token for authentication (optional)
	APIKey string `yaml:"apiKey"`
	// Models is the list of available models from this provider
	Models []Model `yaml:"models"`
}

// GlobalConfig holds global configuration for model providers
type GlobalConfig struct {
	// ModelProviders is the list of configured model providers
	ModelProviders []ModelProvider `yaml:"modelProviders"`
}

// Config holds the local project configuration
type Config struct {
	// Agents is the map of agent names to their configurations
	Agents map[string]*AgentConfig `yaml:"agents"`
	// DefaultModel is the default LLM model for agents that don't specify their own
	DefaultModel string `yaml:"default_model"`
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
