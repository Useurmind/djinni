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
