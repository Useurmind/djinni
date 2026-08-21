package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/useurmind/djinni/pkg/config"
	"github.com/useurmind/djinni/pkg/docker"
	"github.com/useurmind/djinni/pkg/git"
	"github.com/useurmind/djinni/pkg/log"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare <agent-name>",
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

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		repoName, err := git.GetRepoName(cwd)
		if err != nil {
			return fmt.Errorf("failed to get repo name: %w", err)
		}

		client, err := docker.NewClient()
		if err != nil {
			return fmt.Errorf("failed to initialize container client: %w", err)
		}

		if agentCfg.Containerfile != "" {
			log.Info(fmt.Sprintf("Building from: %s", agentCfg.Containerfile))

			exitCode, err := client.BuildContainer(repoName, agentName, agentCfg.Containerfile)
			if err != nil {
				return fmt.Errorf("failed to build container: %w", err)
			}
			if exitCode != 0 {
				os.Exit(exitCode)
				return nil
			}
		}

		if len(agentCfg.WritablePaths) > 0 {
			log.Info("Setting up writable paths with overlayfs...")

			image := fmt.Sprintf(docker.AgentImageNameFormat, repoName, agentName)
			if agentCfg.Image != "" {
				image = agentCfg.Image
			}

			for _, wp := range agentCfg.WritablePaths {
				wpObj := docker.WritablePath{
					Name:        wp.Name,
					Destination: wp.Destination,
				}
				if err := client.PrepareWritablePaths(repoName, agentName, []docker.WritablePath{wpObj}, image); err != nil {
					return fmt.Errorf("failed to prepare writable path %s: %w", wp.Name, err)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(prepareCmd)
	prepareCmd.Flags().StringP("config", "c", "", "Path to config file (default: .djinni.yml in current directory)")
}
