package docker

import (
	"testing"
)

func TestMount_Struct(t *testing.T) {
	m := Mount{
		Source:      "/src",
		Destination: "/dst",
		ReadOnly:    true,
	}

	if m.Source != "/src" {
		t.Errorf("Expected Source '/src', got '%s'", m.Source)
	}
	if m.Destination != "/dst" {
		t.Errorf("Expected Destination '/dst', got '%s'", m.Destination)
	}
	if !m.ReadOnly {
		t.Error("Expected ReadOnly to be true")
	}
}

func TestMount_Defaults(t *testing.T) {
	m := Mount{}
	if m.ReadOnly {
		t.Error("Expected ReadOnly to be false by default")
	}
}

func TestFilesToCopy_Struct(t *testing.T) {
	f := FilesToCopy{
		Source:      "/home/user/.gitconfig",
		Destination: "/home/agent/.gitconfig",
	}

	if f.Source != "/home/user/.gitconfig" {
		t.Errorf("Expected Source '/home/user/.gitconfig', got '%s'", f.Source)
	}
	if f.Destination != "/home/agent/.gitconfig" {
		t.Errorf("Expected Destination '/home/agent/.gitconfig', got '%s'", f.Destination)
	}
	if f.Name() != "f690a263187bdaa1" {
		t.Errorf("Expected Name 'f690a263187bdaa1', got '%s'", f.Name())
	}
}
