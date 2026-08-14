package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/useurmind/djinni/pkg/config"
	"github.com/useurmind/djinni/pkg/docker"
	"github.com/useurmind/djinni/pkg/log"
)

var startAgentCmd = &cobra.Command{
	Use:   "start-agent <agent-name>",
	Short: "Start an agent by name from the config",
	Long:  `Start an agent container using the configuration from .djinni.yml`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		agentName := args[0]

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		agentCfg, ok := cfg.Agents[agentName]
		if !ok {
			return fmt.Errorf("agent '%s' not found in config", agentName)
		}

		log.Info("Initializing container client...")
		client, err := docker.NewClient()
		if err != nil {
			return fmt.Errorf("failed to initialize container client: %w", err)
		}

		log.Info("Starting container...")
		exitCode, err := client.RunContainer(agentCfg.Image, agentCfg.HarnessCommand, agentName, agentCfg.Mounts)
		if err != nil {
			return fmt.Errorf("failed to run container: %w", err)
		}

		os.Exit(exitCode)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startAgentCmd)
	startAgentCmd.Flags().StringP("config", "c", "", "Path to config file (default: .djinni.yml in current directory)")
}
