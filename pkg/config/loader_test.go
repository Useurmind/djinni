package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		path        string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "load from provided path",
			yamlContent: "agents:\n  test:\n    image: test-image\n    harness_command: [\"echo\", \"test\"]",
			path:        "test-config.yml",
			wantErr:     false,
		},
		{
			name:        "missing file",
			yamlContent: "",
			path:        "non-existent.yml",
			wantErr:     true,
		},
		{
			name:        "empty config",
			yamlContent: "",
			path:        "empty.yml",
			wantErr:     true,
		},
		{
			name:        "agent without image or containerfile",
			yamlContent: "agents:\n  test:\n    harness_command: [\"echo\", \"test\"]",
			path:        "invalid-no-image.yml",
			wantErr:     true,
		},
		{
			name:        "agent with both image and containerfile",
			yamlContent: "agents:\n  test:\n    image: test-image\n    containerfile: test-file\n    harness_command: [\"echo\", \"test\"]",
			path:        "invalid-both.yml",
			wantErr:     true,
		},
		{
			name:        "agent without harness_command",
			yamlContent: "agents:\n  test:\n    image: test-image",
			path:        "invalid-no-command.yml",
			wantErr:     true,
		},
		{
			name:        "valid containerfile config",
			yamlContent: "agents:\n  test:\n    containerfile: test-file\n    harness_command: [\"echo\", \"test\"]",
			path:        "valid-containerfile.yml",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if tt.yamlContent != "" {
				configPath := filepath.Join(tmpDir, tt.path)
				if err := os.WriteFile(configPath, []byte(tt.yamlContent), 0600); err != nil {
					t.Fatalf("Failed to write test config: %v", err)
				}
				tt.path = configPath
			}

			cfg, err := LoadConfig(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && cfg == nil {
				t.Fatal("LoadConfig() returned nil config without error")
			}

			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestLoadConfig_DefaultPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	yamlContent := "agents:\n  test:\n    image: test-image\n    harness_command: [\"echo\", \"test\"]"
	configPath := filepath.Join(tmpDir, ".djinni.yml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}
	if len(cfg.Agents) == 0 {
		t.Fatal("Config has no agents")
	}
}
