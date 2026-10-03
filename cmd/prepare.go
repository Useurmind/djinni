package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/useurmind/djinni/pkg/config"
	container "github.com/useurmind/djinni/pkg/container"
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

		globalCfg, err := config.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("failed to load global config: %w", err)
		}

		config.ResolveStorageBaseDirectory(cfg, globalCfg)

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

		client, err := container.NewClient(agentCfg.GitWorkspace.BaseDirectory)
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

		// Setup network and proxy for the agent
		log.Info("Setting up network and proxy...")
		if agentCfg.Network.Internal {
			// Cleanup any existing network and proxy first for idempotency
			if err := client.CleanupExistingProxy(agentName); err != nil {
				return fmt.Errorf("failed to cleanup existing proxy: %w", err)
			}
			if err := client.CleanupExistingNetwork(agentName); err != nil {
				return fmt.Errorf("failed to cleanup existing network: %w", err)
			}

			// Setup new network and proxy
			proxyInfo, err := client.SetupNetwork(agentName, &agentCfg.Network)
			if err != nil {
				return fmt.Errorf("failed to setup network: %w", err)
			}

			// Save proxy info to file if proxy is configured
			if proxyInfo != nil && proxyInfo.SquidAddress != "" {
				proxyDir := filepath.Join(agentCfg.GitWorkspace.BaseDirectory, "proxyInfo", repoName, agentName)
				if err := os.MkdirAll(proxyDir, 0755); err != nil {
					return fmt.Errorf("failed to create proxy info directory: %w", err)
				}

				proxyFilePath := filepath.Join(proxyDir, "proxy.json")
				proxyData, err := json.MarshalIndent(proxyInfo, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal proxy info: %w", err)
				}

				if err := os.WriteFile(proxyFilePath, proxyData, 0644); err != nil {
					return fmt.Errorf("failed to write proxy info: %w", err)
				}

				log.Info(fmt.Sprintf("Proxy info saved to: %s", proxyFilePath))
			}
		}

		if len(agentCfg.WritablePaths) > 0 {
			log.Info("Setting up writable paths with overlayfs...")

			image := fmt.Sprintf(container.AgentImageNameFormat, repoName, agentName)
			if agentCfg.Image != "" {
				image = agentCfg.Image
			}

			for _, wp := range agentCfg.WritablePaths {
				wpObj := config.WritablePath{
					Name:        wp.Name,
					Destination: wp.Destination,
				}
				if err := client.PrepareWritablePaths(repoName, agentName, []config.WritablePath{wpObj}, image); err != nil {
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
