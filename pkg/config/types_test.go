package config

import (
	"os"
	"testing"
)

func TestFilesToCopy_Name(t *testing.T) {
	f := FilesToCopy{
		Source:      "/home/user/.gitconfig",
		Destination: "/home/agent/.gitconfig",
	}

	expected := "f690a263187bdaa1"
	if f.Name() != expected {
		t.Errorf("Expected Name '%s', got '%s'", expected, f.Name())
	}
}

func TestFilesToCopy_Name_Caching(t *testing.T) {
	f := &FilesToCopy{
		Source:      "/home/user/.gitconfig",
		Destination: "/home/agent/.gitconfig",
	}

	firstCall := f.Name()
	secondCall := f.Name()

	if firstCall != secondCall {
		t.Errorf("Name() should be cached, got different results: %s vs %s", firstCall, secondCall)
	}

	if firstCall != "f690a263187bdaa1" {
		t.Errorf("Expected Name 'f690a263187bdaa1', got '%s'", firstCall)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid agent with image",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Image:          "test-image",
						HarnessCommand: []string{"echo", "test"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid agent with containerfile",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Containerfile:  "Dockerfile",
						HarnessCommand: []string{"echo", "test"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid agent with all fields",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Image:          "test-image",
						HarnessCommand: []string{"echo", "test"},
						Mounts: []Mount{
							{
								Source:      "/source",
								Destination: "/dest",
								ReadOnly:    true,
							},
						},
						FilesToCopy: []FilesToCopy{
							{
								Source:      "/source/file.txt",
								Destination: "/dest/file.txt",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "agent with neither image nor containerfile",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						HarnessCommand: []string{"echo", "test"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "agent with both image and containerfile",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Image:          "test-image",
						Containerfile:  "Dockerfile",
						HarnessCommand: []string{"echo", "test"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "agent without harness_command",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Image: "test-image",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "agent without harness_command empty array",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Image:          "test-image",
						HarnessCommand: []string{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "multiple agents - all valid",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"agent1": {
						Image:          "image1",
						HarnessCommand: []string{"cmd1"},
					},
					"agent2": {
						Containerfile:  "Dockerfile2",
						HarnessCommand: []string{"cmd2"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple agents - one invalid",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"agent1": {
						Image:          "image1",
						HarnessCommand: []string{"cmd1"},
					},
					"agent2": {
						Image:          "image2",
						Containerfile:  "Dockerfile2",
						HarnessCommand: []string{"cmd2"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "agent with empty name",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"": {
						Image:          "test-image",
						HarnessCommand: []string{"echo", "test"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "agent with valid writablePaths",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Image:          "test-image",
						HarnessCommand: []string{"echo", "test"},
						WritablePaths: []WritablePath{
							{
								Name:        "home",
								Destination: "/home/agent",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "agent with writablePaths missing name",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Image:          "test-image",
						HarnessCommand: []string{"echo", "test"},
						WritablePaths: []WritablePath{
							{
								Name:        "",
								Destination: "/home/agent",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "agent with writablePaths missing destination",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Image:          "test-image",
						HarnessCommand: []string{"echo", "test"},
						WritablePaths: []WritablePath{
							{
								Name:        "home",
								Destination: "",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "agent with writablePaths but empty array",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Image:          "test-image",
						HarnessCommand: []string{"echo", "test"},
						WritablePaths:  []WritablePath{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "agent with valid delete_on_exit none",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Image:          "test-image",
						HarnessCommand: []string{"echo", "test"},
						DeleteOnExit:   "none",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "agent with valid delete_on_exit all",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Image:          "test-image",
						HarnessCommand: []string{"echo", "test"},
						DeleteOnExit:   "all",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "agent with invalid delete_on_exit",
			config: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Image:          "test-image",
						HarnessCommand: []string{"echo", "test"},
						DeleteOnExit:   "invalid",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Expected error '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

func TestMount_YAMLTags(t *testing.T) {
	m := Mount{
		Source:      "/src",
		Destination: "/dst",
		ReadOnly:    true,
	}
	if m.Source != "/src" {
		t.Errorf("Expected source '/src', got '%s'", m.Source)
	}
	if m.Destination != "/dst" {
		t.Errorf("Expected destination '/dst', got '%s'", m.Destination)
	}
	if !m.ReadOnly {
		t.Error("Expected readOnly true")
	}
}

func TestFilesToCopy_YAMLTags(t *testing.T) {
	f := FilesToCopy{
		Source:      "/src/file.txt",
		Destination: "/dst/file.txt",
	}
	if f.Source != "/src/file.txt" {
		t.Errorf("Expected source '/src/file.txt', got '%s'", f.Source)
	}
	if f.Destination != "/dst/file.txt" {
		t.Errorf("Expected destination '/dst/file.txt', got '%s'", f.Destination)
	}
}

func TestAgentConfig_YAMLTags(t *testing.T) {
	c := &AgentConfig{
		HarnessCommand: []string{"cmd"},
		Image:          "image",
		Containerfile:  "file",
		Mounts:         []Mount{{"/src", "/dst", true}},
		FilesToCopy:    []FilesToCopy{{"/src", "/dst"}},
		DefaultModel:   "mymodel",
		WritablePaths:  []WritablePath{{"home", "/home/agent"}},
		TmpfsMounts:    []TmpfsMount{{"/tmp", "1g"}},
	}
	if len(c.HarnessCommand) != 1 {
		t.Errorf("Expected 1 harness command, got %d", len(c.HarnessCommand))
	}
	if c.Image != "image" {
		t.Errorf("Expected image 'image', got '%s'", c.Image)
	}
	if c.Containerfile != "file" {
		t.Errorf("Expected containerfile 'file', got '%s'", c.Containerfile)
	}
	if len(c.Mounts) != 1 {
		t.Errorf("Expected 1 mount, got %d", len(c.Mounts))
	}
	if len(c.FilesToCopy) != 1 {
		t.Errorf("Expected 1 files_to_copy, got %d", len(c.FilesToCopy))
	}
	if c.DefaultModel != "mymodel" {
		t.Errorf("Expected default model 'mymodel', got '%s'", c.DefaultModel)
	}
	if len(c.WritablePaths) != 1 {
		t.Errorf("Expected 1 writablePath, got %d", len(c.WritablePaths))
	}
	if len(c.TmpfsMounts) != 1 {
		t.Errorf("Expected 1 tmpfsMount, got %d", len(c.TmpfsMounts))
	}
}

func TestModel_YAMLTags(t *testing.T) {
	m := Model{ID: "mymodel"}
	if m.ID != "mymodel" {
		t.Errorf("Expected ID 'mymodel', got '%s'", m.ID)
	}
}

func TestModelProvider_YAMLTags(t *testing.T) {
	p := ModelProvider{
		Name:    "litellm",
		APIBase: "http://localhost:8000",
		APIKey:  "key123",
		Models:  []Model{{ID: "model1"}, {ID: "model2"}},
	}
	if p.Name != "litellm" {
		t.Errorf("Expected name 'litellm', got '%s'", p.Name)
	}
	if p.APIBase != "http://localhost:8000" {
		t.Errorf("Expected APIBase 'http://localhost:8000', got '%s'", p.APIBase)
	}
	if p.APIKey != "key123" {
		t.Errorf("Expected APIKey 'key123', got '%s'", p.APIKey)
	}
	if len(p.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(p.Models))
	}
}

func TestAuthConfig_YAMLTags(t *testing.T) {
	c := &GlobalConfig{
		ModelProviders: []ModelProvider{
			{
				Name:    "litellm",
				APIBase: "http://localhost:8000",
				APIKey:  "key123",
				Models:  []Model{{ID: "model1"}},
			},
		},
	}
	if len(c.ModelProviders) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(c.ModelProviders))
	}
	if c.ModelProviders[0].Name != "litellm" {
		t.Errorf("Expected provider name 'litellm', got '%s'", c.ModelProviders[0].Name)
	}
}

func TestModelProvider_Validate(t *testing.T) {
	tests := []struct {
		name     string
		provider *ModelProvider
		wantErr  bool
	}{
		{
			name:     "valid provider",
			provider: &ModelProvider{Name: "litellm", Models: []Model{{ID: "model1"}}},
			wantErr:  false,
		},
		{
			name:     "empty name",
			provider: &ModelProvider{Name: "", Models: []Model{{ID: "model1"}}},
			wantErr:  true,
		},
		{
			name:     "no models",
			provider: &ModelProvider{Name: "litellm", Models: []Model{}},
			wantErr:  true,
		},
		{
			name:     "empty provider",
			provider: &ModelProvider{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.provider.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthConfig_Validate(t *testing.T) {
	tests := []struct {
		name       string
		authConfig *GlobalConfig
		wantErr    bool
	}{
		{
			name:       "empty providers",
			authConfig: &GlobalConfig{ModelProviders: []ModelProvider{}},
			wantErr:    false,
		},
		{
			name: "valid providers",
			authConfig: &GlobalConfig{
				ModelProviders: []ModelProvider{
					{Name: "litellm", APIBase: "http://localhost:8000", Models: []Model{{ID: "model1"}}},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid provider",
			authConfig: &GlobalConfig{
				ModelProviders: []ModelProvider{
					{Name: "", Models: []Model{{ID: "model1"}}},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.authConfig.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "path with ~",
			input:    "~/.config/opencode",
			expected: os.Getenv("HOME") + "/.config/opencode",
		},
		{
			name:     "path with env var",
			input:    "$HOME/.config/opencode",
			expected: os.Getenv("HOME") + "/.config/opencode",
		},
		{
			name:     "path with ${var} syntax",
			input:    "${HOME}/.config/opencode",
			expected: os.Getenv("HOME") + "/.config/opencode",
		},
		{
			name:     "absolute path",
			input:    "/home/user/.config/opencode",
			expected: "/home/user/.config/opencode",
		},
		{
			name:     "relative path",
			input:    "./some/path",
			expected: "./some/path",
		},
		{
			name:     "mixed ~ and env var",
			input:    "~/$USER",
			expected: os.Getenv("HOME") + "/" + os.Getenv("USER"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandConfigPaths(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		input    *Config
		expected *Config
	}{
		{
			name: "expand paths in mounts",
			input: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Mounts: []Mount{
							{Source: "~/.config", Destination: "/dest"},
							{Source: "$HOME/data", Destination: "/data"},
						},
					},
				},
			},
			expected: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Mounts: []Mount{
							{Source: homeDir + "/.config", Destination: "/dest"},
							{Source: homeDir + "/data", Destination: "/data"},
						},
					},
				},
			},
		},
		{
			name: "expand paths in files_to_copy",
			input: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						FilesToCopy: []FilesToCopy{
							{Source: "~/.gitconfig", Destination: "/dest"},
						},
					},
				},
			},
			expected: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						FilesToCopy: []FilesToCopy{
							{Source: homeDir + "/.gitconfig", Destination: "/dest"},
						},
					},
				},
			},
		},
		{
			name: "expand paths in tmpfs_mounts",
			input: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						TmpfsMounts: []TmpfsMount{
							{Destination: "/$VAR"},
						},
					},
				},
			},
			expected: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						TmpfsMounts: []TmpfsMount{
							{Destination: "/$VAR"},
						},
					},
				},
			},
		},
		{
			name: "expand paths in all fields",
			input: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Mounts: []Mount{
							{Source: "~/.config", Destination: "/dest"},
						},
						FilesToCopy: []FilesToCopy{
							{Source: "~/.gitconfig", Destination: "/dest"},
						},
					},
				},
			},
			expected: &Config{
				Agents: map[string]*AgentConfig{
					"test": {
						Mounts: []Mount{
							{Source: homeDir + "/.config", Destination: "/dest"},
						},
						FilesToCopy: []FilesToCopy{
							{Source: homeDir + "/.gitconfig", Destination: "/dest"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ExpandConfigPaths(tt.input)
			result := tt.input

			if len(result.Agents["test"].Mounts) != len(tt.expected.Agents["test"].Mounts) {
				t.Errorf("Mounts length = %d, want %d", len(result.Agents["test"].Mounts), len(tt.expected.Agents["test"].Mounts))
			}

			for i := range result.Agents["test"].Mounts {
				if result.Agents["test"].Mounts[i].Source != tt.expected.Agents["test"].Mounts[i].Source {
					t.Errorf("Mounts[%d].Source = %q, want %q", i, result.Agents["test"].Mounts[i].Source, tt.expected.Agents["test"].Mounts[i].Source)
				}
				if result.Agents["test"].Mounts[i].Destination != tt.expected.Agents["test"].Mounts[i].Destination {
					t.Errorf("Mounts[%d].Destination = %q, want %q", i, result.Agents["test"].Mounts[i].Destination, tt.expected.Agents["test"].Mounts[i].Destination)
				}
			}

			if len(result.Agents["test"].FilesToCopy) != len(tt.expected.Agents["test"].FilesToCopy) {
				t.Errorf("FilesToCopy length = %d, want %d", len(result.Agents["test"].FilesToCopy), len(tt.expected.Agents["test"].FilesToCopy))
			}

			for i := range result.Agents["test"].FilesToCopy {
				if result.Agents["test"].FilesToCopy[i].Source != tt.expected.Agents["test"].FilesToCopy[i].Source {
					t.Errorf("FilesToCopy[%d].Source = %q, want %q", i, result.Agents["test"].FilesToCopy[i].Source, tt.expected.Agents["test"].FilesToCopy[i].Source)
				}
				if result.Agents["test"].FilesToCopy[i].Destination != tt.expected.Agents["test"].FilesToCopy[i].Destination {
					t.Errorf("FilesToCopy[%d].Destination = %q, want %q", i, result.Agents["test"].FilesToCopy[i].Destination, tt.expected.Agents["test"].FilesToCopy[i].Destination)
				}
			}

			if len(result.Agents["test"].WritablePaths) != len(tt.expected.Agents["test"].WritablePaths) {
				t.Errorf("WritablePaths length = %d, want %d", len(result.Agents["test"].WritablePaths), len(tt.expected.Agents["test"].WritablePaths))
			}

			for i := range result.Agents["test"].WritablePaths {
				if result.Agents["test"].WritablePaths[i].Destination != tt.expected.Agents["test"].WritablePaths[i].Destination {
					t.Errorf("WritablePaths[%d].Destination = %q, want %q", i, result.Agents["test"].WritablePaths[i].Destination, tt.expected.Agents["test"].WritablePaths[i].Destination)
				}
			}
		})
	}
}
