package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/useurmind/djinni/pkg/docker"
)

var attachCmd = &cobra.Command{
	Use:   "attach <agent-name>",
	Short: "Attach to a running agent container",
	Long:  `Attach to a running agent container and execute a command (default: bash)`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAttachAgent(cmd, args)
	},
}

func runAttachAgent(cmd *cobra.Command, args []string) error {
	agentName := args[0]
	cmdFlag, err := cmd.Flags().GetString("cmd")
	if err != nil {
		return fmt.Errorf("failed to get cmd flag: %w", err)
	}

	if err := ensureContainerExists(agentName); err != nil {
		return err
	}

	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to initialize container client: %w", err)
	}

	var execCmd []string
	if cmdFlag != "" {
		execCmd = []string{cmdFlag}
	} else {
		execCmd = []string{"bash"}
	}

	_, err = client.ExecInContainer(agentName, execCmd)
	if err != nil {
		return fmt.Errorf("failed to exec in container: %w", err)
	}

	return nil
}

func ensureContainerExists(containerName string) error {
	cmd := exec.Command("podman", "ps", "-q", "-f", "name="+containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}

	if len(output) == 0 {
		return fmt.Errorf("container '%s' not found or not running", containerName)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(attachCmd)
	attachCmd.Flags().StringP("cmd", "c", "", "Command to execute in container (default: bash)")
}
