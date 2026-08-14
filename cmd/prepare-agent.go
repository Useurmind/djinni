package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/useurmind/djinni/pkg/config"
	"github.com/useurmind/djinni/pkg/docker"
	"github.com/useurmind/djinni/pkg/log"
)

var prepareAgentCmd = &cobra.Command{
	Use:   "prepare-agent <agent-name>",
	Short: "Build a local container image for an agent",
	Long:  `Build a local container image for an agent using the configuration from .djinni.yml`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		agentName := args[0]

		log.Info("Preparing agent...")
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		agentCfg, ok := cfg.Agents[agentName]
		if !ok {
			return fmt.Errorf("agent '%s' not found in config", agentName)
		}

		if agentCfg.Containerfile != "" {
			log.Info(fmt.Sprintf("Building from: %s", agentCfg.Containerfile))

			client, err := docker.NewClient()
			if err != nil {
				return fmt.Errorf("failed to initialize container client: %w", err)
			}

			exitCode, err := client.BuildContainer(agentName, agentCfg.Containerfile)
			if err != nil {
				return fmt.Errorf("failed to build container: %w", err)
			}
			if exitCode != 0 {
				os.Exit(exitCode)
				return nil
			}
		}

		log.Info(fmt.Sprintf("Image specified, nothing to prepare: %s", agentCfg.Image))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(prepareAgentCmd)
	prepareAgentCmd.Flags().StringP("config", "c", "", "Path to config file (default: .djinni.yml in current directory)")
}
