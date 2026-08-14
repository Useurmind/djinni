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
}
