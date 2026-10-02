package container

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/useurmind/djinni/pkg/log"
)

const (
	NetworkNamePrefix = "djinni-ai-"
)

// GetNetworkName generates the internal network name from agent name
func GetNetworkName(agentName string) string {
	return NetworkNamePrefix + strings.ReplaceAll(agentName, "-", "_")
}

// CreateInternalNetwork creates an internal podman network
func CreateInternalNetwork(client *Client, networkName string) error {
	args := []string{"network", "create", "--driver", "bridge", "--internal", networkName}

	cmd := exec.Command(client.Binary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create internal network %s: %w", networkName, err)
	}

	log.Info(string(output))
	log.Info(fmt.Sprintf("Created internal network: %s", networkName))
	return nil
}

// RemoveNetwork removes an internal podman network
func RemoveNetwork(client *Client, networkName string) error {
	args := []string{"network", "rm", networkName}

	cmd := exec.Command(client.Binary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if the error is because the network doesn't exist
		// podman returns exit status 125 for "network not found"
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 125 {
			log.Info(fmt.Sprintf("Network %s does not exist, nothing to remove", networkName))
			return nil
		}
		return fmt.Errorf("failed to remove network %s: %w", networkName, err)
	}

	log.Info(string(output))
	log.Info(fmt.Sprintf("Removed network: %s", networkName))
	return nil
}
