package container

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

	tests := []struct {
		name         string
		hasPodman    bool
		wantErr      bool
		expectedType string
	}{
		{
			name:         "podman found",
			hasPodman:    true,
			wantErr:      false,
			expectedType: "podman",
		},
		{
			name:      "podman not found",
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
		hasPodman bool
	}{
		{
			name:      "no mounts",
			mounts:    []config.Mount{},
			hasPodman: true,
		},
		{
			name: "single read-write mount",
			mounts: []config.Mount{
				{Source: "/src", Destination: "/dst", ReadOnly: false},
			},
			hasPodman: true,
		},
		{
			name: "single read-only mount",
			mounts: []config.Mount{
				{Source: "/src", Destination: "/dst", ReadOnly: true},
			},
			hasPodman: true,
		},
		{
			name: "multiple mounts",
			mounts: []config.Mount{
				{Source: "/src1", Destination: "/dst1", ReadOnly: false},
				{Source: "/src2", Destination: "/dst2", ReadOnly: true},
			},
			hasPodman: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{Binary: "podman"}

			podmanPath := filepath.Join(tmpDir, "podman")
			if err := os.WriteFile(podmanPath, []byte("#!/bin/sh\nexit 0"), 0755); err != nil {
				t.Fatalf("Failed to create podman stub: %v", err)
			}

			oldPath := os.Getenv("PATH")
			defer os.Setenv("PATH", oldPath)
			os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)

			_, err := client.RunContainer("test-image", []string{"echo", "test"}, "test-container", tt.mounts, nil)
			if err != nil {
				t.Errorf("RunContainer() error = %v", err)
			}
		})
	}
}

func TestRunContainer_ReadOnlyRoot(t *testing.T) {
	tmpDir := t.TempDir()

	client := &Client{Binary: "podman"}

	podmanPath := filepath.Join(tmpDir, "podman")
	if err := os.WriteFile(podmanPath, []byte("#!/bin/sh\nexit 0"), 0755); err != nil {
		t.Fatalf("Failed to create podman stub: %v", err)
	}

	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)

	commands := &ContainerCommands{
		ForceReadOnlyRootOff: false,
		TmpfsMounts: []TmpfsMount{
			{Destination: "/tmp"},
			{Destination: "/cache", Size: "512m"},
		},
	}

	_, err := client.RunContainer("test-image", []string{"echo", "test"}, "test-container", []config.Mount{}, commands)
	if err != nil {
		t.Errorf("RunContainer() error = %v", err)
	}
}

func TestRunContainer_TmpfsMounts(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		tmpfsMounts []TmpfsMount
	}{
		{
			name:        "no tmpfs mounts",
			tmpfsMounts: []TmpfsMount{},
		},
		{
			name: "single tmpfs mount",
			tmpfsMounts: []TmpfsMount{
				{Destination: "/tmp"},
			},
		},
		{
			name: "tmpfs mount with size",
			tmpfsMounts: []TmpfsMount{
				{Destination: "/tmp", Size: "1g"},
			},
		},
		{
			name: "multiple tmpfs mounts",
			tmpfsMounts: []TmpfsMount{
				{Destination: "/tmp"},
				{Destination: "/cache", Size: "512m"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{Binary: "podman"}

			podmanPath := filepath.Join(tmpDir, "podman")
			if err := os.WriteFile(podmanPath, []byte("#!/bin/sh\nexit 0"), 0755); err != nil {
				t.Fatalf("Failed to create podman stub: %v", err)
			}

			oldPath := os.Getenv("PATH")
			defer os.Setenv("PATH", oldPath)
			os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)

			commands := &ContainerCommands{
				TmpfsMounts: tt.tmpfsMounts,
			}

			_, err := client.RunContainer("test-image", []string{"echo", "test"}, "test-container", []config.Mount{}, commands)
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
		hasPodman            bool
		wantErr              bool
	}{
		{
			name:                 "successful build",
			containerfileContent: []byte("FROM alpine\nCMD echo hello"),
			hasPodman:            true,
			wantErr:              false,
		},
		{
			name:                 "containerfile not found - empty path",
			containerfileContent: nil,
			hasPodman:            true,
			wantErr:              true,
		},
		{
			name:                 "containerfile not found - nonexistent path",
			containerfileContent: nil,
			hasPodman:            true,
			wantErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{Binary: "podman"}

			containerfilePath := filepath.Join(tmpDir, "Containerfile")
			if tt.containerfileContent != nil {
				if err := os.WriteFile(containerfilePath, tt.containerfileContent, 0644); err != nil {
					t.Fatalf("Failed to create Containerfile: %v", err)
				}
			}

			podmanPath := filepath.Join(tmpDir, "podman")
			if err := os.WriteFile(podmanPath, []byte("#!/bin/sh\necho Built"), 0755); err != nil {
				t.Fatalf("Failed to create podman stub: %v", err)
			}

			oldPath := os.Getenv("PATH")
			defer os.Setenv("PATH", oldPath)
			os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)

			var path string
			if tt.containerfileContent != nil {
				path = containerfilePath
			} else {
				path = "/nonexistent/path/Containerfile"
			}

			exitCode, err := client.BuildContainer("test", "test", path)
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
	client := &Client{Binary: "podman"}

	exitCode, err := client.BuildContainer("test", "test", "/nonexistent/Containerfile")
	if err == nil {
		t.Error("BuildContainer() expected error, got nil")
	}
	if exitCode != 1 {
		t.Errorf("BuildContainer() exitCode = %d, want 1", exitCode)
	}
}

func TestBuildContainer_ContainerfileDoesNotExist(t *testing.T) {
	client := &Client{Binary: "podman"}

	exitCode, err := client.BuildContainer("test", "test", "/path/that/does/not/exist/Containerfile")
	if err == nil {
		t.Error("BuildContainer() expected error for non-existent file, got nil")
	}
	if exitCode != 1 {
		t.Errorf("BuildContainer() exitCode = %d, want 1", exitCode)
	}
}

func TestRunContainer_WithOverlayMounts(t *testing.T) {
	tmpDir := t.TempDir()

	client := &Client{Binary: "podman"}

	podmanPath := filepath.Join(tmpDir, "podman")
	if err := os.WriteFile(podmanPath, []byte("#!/bin/sh\nexit 0"), 0755); err != nil {
		t.Fatalf("Failed to create podman stub: %v", err)
	}

	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)

	mounts := []config.Mount{
		{
			Source:      "/tmp/djinni/test-repo/test-agent/writablePaths/home/mnt",
			Destination: "/home/agent",
			ReadOnly:    false,
		},
	}

	_, err := client.RunContainer("test-image", []string{"echo", "test"}, "test-container", mounts, nil)
	if err != nil {
		t.Errorf("RunContainer() error = %v", err)
	}
}
