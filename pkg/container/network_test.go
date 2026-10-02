package container

import (
	"testing"
)

func TestGetNetworkName(t *testing.T) {
	tests := []struct {
		agentName string
		expected  string
	}{
		{"default", "djinni-ai-default"},
		{"test-agent", "djinni-ai-test_agent"},
		{"my_agent", "djinni-ai-my_agent"},
		{"agent-123", "djinni-ai-agent_123"},
	}

	for _, tt := range tests {
		t.Run(tt.agentName, func(t *testing.T) {
			result := GetNetworkName(tt.agentName)
			if result != tt.expected {
				t.Errorf("GetNetworkName(%q) = %q, want %q", tt.agentName, result, tt.expected)
			}
		})
	}
}
