package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/useurmind/djinni/pkg/config"
	"github.com/useurmind/djinni/pkg/docker"
	"github.com/useurmind/djinni/pkg/git"
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
		taskName, _ := cmd.Flags().GetString("task")

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

		image := agentCfg.Image
		if agentCfg.Containerfile != "" {
			image = fmt.Sprintf("djinni-%s:latest", agentName)
			log.Info(fmt.Sprintf("Using local image: %s", image))
		}

		log.Info(fmt.Sprintf("Using image: %s", image))

		var workspacePath string
		var commands *docker.ContainerCommands
		if taskName != "" {
			log.Info("Setting up git workspace mount...")
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			baseDir := agentCfg.GitWorkspace.BaseDirectory
			if baseDir == "" {
				baseDir = fmt.Sprintf("/tmp/%s", agentName)
			}

			workspacePath, err = git.CloneToTemp(cwd, baseDir, taskName)
			if err != nil {
				return fmt.Errorf("failed to clone workspace: %w", err)
			}

			if err := git.CheckoutNewBranch(workspacePath, taskName); err != nil {
				defer git.CleanupWorkspace(workspacePath)
				return fmt.Errorf("failed to checkout branch: %w", err)
			}

			repoName := filepath.Base(cwd)
			mountPath := filepath.Join("/workspace", fmt.Sprintf("%s-%s", repoName, taskName))
			workspaceMount := config.Mount{
				Source:      workspacePath,
				Destination: mountPath,
			}
			agentCfg.Mounts = append(agentCfg.Mounts, workspaceMount)

			commands = &docker.ContainerCommands{
				PreCommands: []string{
					fmt.Sprintf("git config --global --add safe.directory /workspace/%s-%s", repoName, taskName),
				},
			}

			defer git.CleanupWorkspace(workspacePath)
		} else if taskName != "" && agentCfg.GitWorkspace.BaseDirectory == "" {
			return fmt.Errorf("git_workspace not configured for agent '%s'", agentName)
		}

		exitCode, err := client.RunContainer(image, agentCfg.HarnessCommand, agentName, agentCfg.Mounts, commands)
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
	startAgentCmd.Flags().StringP("task", "t", "", "Task name for workspace mount (creates feature/<taskname> branch)")
	startAgentCmd.MarkFlagRequired("task")
}
