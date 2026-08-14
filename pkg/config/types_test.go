package config

import (
	"testing"
)

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

func TestAgentConfig_YAMLTags(t *testing.T) {
	c := &AgentConfig{
		HarnessCommand: []string{"cmd"},
		Image:          "image",
		Containerfile:  "file",
		Mounts:         []Mount{{"/src", "/dst", true}},
		DefaultModel:   "mymodel",
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
	if c.DefaultModel != "mymodel" {
		t.Errorf("Expected default model 'mymodel', got '%s'", c.DefaultModel)
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
	c := &AuthConfig{
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
		authConfig *AuthConfig
		wantErr    bool
	}{
		{
			name:       "empty providers",
			authConfig: &AuthConfig{ModelProviders: []ModelProvider{}},
			wantErr:    false,
		},
		{
			name: "valid providers",
			authConfig: &AuthConfig{
				ModelProviders: []ModelProvider{
					{Name: "litellm", APIBase: "http://localhost:8000", Models: []Model{{ID: "model1"}}},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid provider",
			authConfig: &AuthConfig{
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
