package docker

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/useurmind/djinni/pkg/config"
)

func TestNewClient(t *testing.T) {
	if _, err := exec.LookPath("podman"); err != nil {
		t.Skip("podman not found, skipping test")
	}

	if _, err := exec.LookPath("docker"); err != nil {
		t.Log("Docker not found in PATH")
	}

	tests := []struct {
		name         string
		hasDocker    bool
		hasPodman    bool
		wantErr      bool
		expectedType string
	}{
		{
			name:         "docker found when podman not in system",
			hasDocker:    true,
			hasPodman:    false,
			wantErr:      false,
			expectedType: "docker",
		},
		{
			name:         "podman found",
			hasDocker:    false,
			hasPodman:    true,
			wantErr:      false,
			expectedType: "podman",
		},
		{
			name:         "both found (prefers podman)",
			hasDocker:    true,
			hasPodman:    true,
			wantErr:      false,
			expectedType: "podman",
		},
		{
			name:      "neither found",
			hasDocker: false,
			hasPodman: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			defer os.RemoveAll(tmpDir)
			binDir := filepath.Join(tmpDir, "bin")
			if err := os.MkdirAll(binDir, 0755); err != nil {
				t.Fatalf("Failed to create bin dir: %v", err)
			}

			oldPath := os.Getenv("PATH")
			defer os.Setenv("PATH", oldPath)
			os.Setenv("PATH", binDir+string(os.PathListSeparator))

			if tt.hasDocker {
				dockerPath := filepath.Join(binDir, "docker")
				if err := os.WriteFile(dockerPath, []byte("#!/bin/sh\nexit 0"), 0755); err != nil {
					t.Fatalf("Failed to create docker stub: %v", err)
				}
			}

			if tt.hasPodman {
				podmanPath := filepath.Join(binDir, "podman")
				if err := os.WriteFile(podmanPath, []byte("#!/bin/sh\nexit 0"), 0755); err != nil {
					t.Fatalf("Failed to create podman stub: %v", err)
				}
			}

			client, err := NewClient()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && client == nil {
				t.Fatal("NewClient() returned nil without error")
			}

			if client != nil && client.Type != tt.expectedType {
				t.Errorf("NewClient() type = %s, want %s", client.Type, tt.expectedType)
			}

			if client != nil && client.Binary != tt.expectedType {
				t.Errorf("NewClient() binary = %s, want %s", client.Binary, tt.expectedType)
			}
		})
	}
}

func TestRunContainer(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		mounts    []config.Mount
		hasDocker bool
	}{
		{
			name:      "no mounts",
			mounts:    []config.Mount{},
			hasDocker: true,
		},
		{
			name: "single read-write mount",
			mounts: []config.Mount{
				{Source: "/src", Destination: "/dst", ReadOnly: false},
			},
			hasDocker: true,
		},
		{
			name: "single read-only mount",
			mounts: []config.Mount{
				{Source: "/src", Destination: "/dst", ReadOnly: true},
			},
			hasDocker: true,
		},
		{
			name: "multiple mounts",
			mounts: []config.Mount{
				{Source: "/src1", Destination: "/dst1", ReadOnly: false},
				{Source: "/src2", Destination: "/dst2", ReadOnly: true},
			},
			hasDocker: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{Binary: "docker"}

			dockerPath := filepath.Join(tmpDir, "docker")
			if err := os.WriteFile(dockerPath, []byte("#!/bin/sh\nexit 0"), 0755); err != nil {
				t.Fatalf("Failed to create docker stub: %v", err)
			}

			oldPath := os.Getenv("PATH")
			defer os.Setenv("PATH", oldPath)
			os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)

			_, err := client.RunContainer("test-image", []string{"echo", "test"}, "test-container", tt.mounts)
			if err != nil {
				t.Errorf("RunContainer() error = %v", err)
			}
		})
	}
}

func TestBuildContainer(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name                 string
		containerfileContent []byte
		hasDocker            bool
		wantErr              bool
	}{
		{
			name:                 "successful build",
			containerfileContent: []byte("FROM alpine\nCMD echo hello"),
			hasDocker:            true,
			wantErr:              false,
		},
		{
			name:                 "containerfile not found - empty path",
			containerfileContent: nil,
			hasDocker:            true,
			wantErr:              true,
		},
		{
			name:                 "containerfile not found - nonexistent path",
			containerfileContent: nil,
			hasDocker:            true,
			wantErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{Binary: "docker"}

			containerfilePath := filepath.Join(tmpDir, "Dockerfile")
			if tt.containerfileContent != nil {
				if err := os.WriteFile(containerfilePath, tt.containerfileContent, 0644); err != nil {
					t.Fatalf("Failed to create Dockerfile: %v", err)
				}
			}

			dockerPath := filepath.Join(tmpDir, "docker")
			if err := os.WriteFile(dockerPath, []byte("#!/bin/sh\necho Built"), 0755); err != nil {
				t.Fatalf("Failed to create docker stub: %v", err)
			}

			oldPath := os.Getenv("PATH")
			defer os.Setenv("PATH", oldPath)
			os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)

			var path string
			if tt.containerfileContent != nil {
				path = containerfilePath
			} else {
				path = "/nonexistent/path/Dockerfile"
			}

			exitCode, err := client.BuildContainer("test", path)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildContainer() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && exitCode != 0 {
				t.Errorf("BuildContainer() exitCode = %d, want 0", exitCode)
			}
		})
	}
}

func TestBuildContainer_FileNotFound(t *testing.T) {
	client := &Client{Binary: "docker"}

	exitCode, err := client.BuildContainer("test", "/nonexistent/Dockerfile")
	if err == nil {
		t.Error("BuildContainer() expected error, got nil")
	}
	if exitCode != 1 {
		t.Errorf("BuildContainer() exitCode = %d, want 1", exitCode)
	}
}

func TestBuildContainer_ContainerfileDoesNotExist(t *testing.T) {
	client := &Client{Binary: "docker"}

	exitCode, err := client.BuildContainer("test", "/path/that/does/not/exist/Dockerfile")
	if err == nil {
		t.Error("BuildContainer() expected error for non-existent file, got nil")
	}
	if exitCode != 1 {
		t.Errorf("BuildContainer() exitCode = %d, want 1", exitCode)
	}
}
