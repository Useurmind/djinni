package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/useurmind/djinni/pkg/ai"
	"github.com/useurmind/djinni/pkg/config"
	"github.com/useurmind/djinni/pkg/docker"
	"github.com/useurmind/djinni/pkg/git"
	"github.com/useurmind/djinni/pkg/log"
)

func execCommand(name string, args []string, workdir string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = workdir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %v, output: %s", name, err, string(output))
	}
	return nil
}

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

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		if err := commitUncommittedChanges(cwd, configPath); err != nil {
			return err
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
		filesToCopy := []docker.FilesToCopy{}
		for _, fc := range agentCfg.FilesToCopy {
			filesToCopy = append(filesToCopy, docker.FilesToCopy{
				Source:      fc.Source,
				Destination: fc.Destination,
			})
		}

		if taskName != "" {
			log.Info("Setting up git workspace mount...")

			baseDir := agentCfg.GitWorkspace.BaseDirectory
			if baseDir == "" {
				baseDir = "/tmp/djinni"
			}

			workspacePath, err = git.CloneToTemp(cwd, baseDir, agentName, taskName)
			if err != nil {
				return fmt.Errorf("failed to clone workspace: %w", err)
			}

			if err := git.CheckoutNewBranch(workspacePath, taskName); err != nil {
				defer git.CleanupWorkspace(workspacePath)
				return fmt.Errorf("failed to checkout branch: %w", err)
			}

			mountPath := filepath.Join("/workspace", fmt.Sprintf("%s-%s", filepath.Base(cwd), taskName))
			agentCfg.Mounts = append(agentCfg.Mounts, config.Mount{
				Source:      workspacePath,
				Destination: mountPath,
			})

			// Get user's home directory and add gitconfig to files to copy
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get user home directory: %w", err)
			}

			gitconfigPath := filepath.Join(homeDir, ".gitconfig")
			if _, err := os.Stat(gitconfigPath); os.IsNotExist(err) {
				return fmt.Errorf("gitconfig file not found at %s", gitconfigPath)
			}

			gitconfigCopy := docker.FilesToCopy{
				Source:      gitconfigPath,
				Destination: "/home/agent/.gitconfig",
			}
			filesToCopy = append(filesToCopy, gitconfigCopy)

			commands = &docker.ContainerCommands{
				PreCommands: []string{
					fmt.Sprintf("cd /workspace/%s-%s", filepath.Base(cwd), taskName),
					fmt.Sprintf("git config --global --add safe.directory /workspace/%s-%s", filepath.Base(cwd), taskName),
				},
			}

			defer git.CleanupWorkspace(workspacePath)
		} else if taskName != "" && agentCfg.GitWorkspace.BaseDirectory == "" {
			return fmt.Errorf("git_workspace not configured for agent '%s'", agentName)
		}

		if len(filesToCopy) > 0 {
			if commands == nil {
				commands = &docker.ContainerCommands{}
			}
			commands.FilesToCopy = filesToCopy
		}

		exitCode, err := client.RunContainer(image, agentCfg.HarnessCommand, agentName, agentCfg.Mounts, commands)
		if err != nil {
			return fmt.Errorf("failed to run container: %w", err)
		}

		mountSources := git.GetMountPaths(agentCfg.Mounts)
		if len(mountSources) > 0 {
			log.Info("Restoring file ownership after container exit...")
			if err := git.RestoreOwnership(mountSources); err != nil {
				log.Error(fmt.Sprintf("Failed to restore ownership: %v", err))
			}
		}

		if workspacePath != "" {
			log.Info("Checking for changes after container exit...")

			clean, err := git.IsRepositoryClean(workspacePath)
			if err != nil {
				log.Error(fmt.Sprintf("Failed to check repository status: %v", err))
			} else if clean {
				log.Info("No changes detected, skipping commit/push")
			} else {
				changedFiles, err := git.GetChangedFiles(workspacePath)
				if err != nil {
					log.Error(fmt.Sprintf("Failed to get changed files: %v", err))
				} else {
					log.Info(fmt.Sprintf("Detected %d changed files", len(changedFiles)))

					syncApproach := "branch_sync"
					if agentCfg.SyncApproach != "" {
						syncApproach = agentCfg.SyncApproach
					}

					if syncApproach == "git_patch" {
						patchDir := filepath.Join(filepath.Dir(workspacePath), "patches")
						if err := git.CreatePatch(workspacePath, fmt.Sprintf("feature/%s", taskName), patchDir); err != nil {
							log.Error(fmt.Sprintf("Failed to create patch: %v", err))
						} else {
							patches, _ := filepath.Glob(filepath.Join(patchDir, "*.patch"))
							if len(patches) > 0 {
								if err := git.ApplyPatch(cwd, patches[0]); err != nil {
									log.Error(fmt.Sprintf("Failed to apply patch: %v", err))
								}
							}
						}
					} else {
						log.Info("Staging all changes...")
						if err := git.AddFiles(workspacePath); err != nil {
							log.Error(fmt.Sprintf("Failed to stage files: %v", err))
						}

						globalCfg, err := config.LoadGlobalConfig()
						if err != nil {
							log.Error(fmt.Sprintf("Failed to load global config: %v", err))
						} else {
							defaultModel := ""
							if agentCfg.DefaultModel != "" {
								defaultModel = agentCfg.DefaultModel
							} else {
								defaultModel = cfg.DefaultModel
							}

							aiAgent := &ai.Agent{
								WorkingDir: workspacePath,
								ReadPaths:  []string{workspacePath},
								Provider:   &globalCfg.ModelProviders[0],
								ModelID:    defaultModel,
							}

							if defaultModel != "" {
								aiAgent.ModelID = defaultModel
							}

							commitMsg, err := aiAgent.Execute()
							if err != nil {
								log.Error(fmt.Sprintf("Failed to generate commit message: %v", err))
							} else {
								if err := git.CommitAll(workspacePath, commitMsg); err != nil {
									log.Error(fmt.Sprintf("Failed to commit: %v", err))
								} else {
									branchName := fmt.Sprintf("feature/%s", taskName)
									if err := git.PushBranch(workspacePath, branchName); err != nil {
										log.Error(fmt.Sprintf("Failed to push branch: %v", err))
									} else if agentCfg.AutomergeAgentBranch {
										log.Info(fmt.Sprintf("Merging branch %s into current branch", branchName))
										if err := execCommand("git", []string{"merge", branchName}, cwd); err != nil {
											log.Error(fmt.Sprintf("Failed to merge branch: %v", err))
										}
									}
								}
							}
						}
					}
				}
			}
		}

		os.Exit(exitCode)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startAgentCmd)
	startAgentCmd.Flags().StringP("config", "c", "", "Path to config file (default: .djinni.yml in current directory)")
	startAgentCmd.Flags().StringP("task", "t", "", "Task name for workspace mount (creates feature/<taskname> branch)")
	if err := startAgentCmd.MarkFlagRequired("task"); err != nil {
		log.Error(fmt.Sprintf("Failed to mark flag required: %v", err))
	}
}

func commitUncommittedChanges(cwd, configPath string) error {
	mode, msg, err := git.PromptCommitChanges(cwd, configPath)
	if err != nil {
		return fmt.Errorf("failed to prompt for commit: %w", err)
	}

	if mode == git.CommitModeNone {
		return nil
	}

	if mode == git.CommitModeAI {
		globalCfg, err := config.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("failed to load global config: %w", err)
		}

		if len(globalCfg.ModelProviders) == 0 {
			return fmt.Errorf("no model providers configured in global config")
		}

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		defaultModel := ""
		if cfg.DefaultModel != "" {
			defaultModel = cfg.DefaultModel
		}

		aiAgent := &ai.Agent{
			WorkingDir: cwd,
			ReadPaths:  []string{cwd},
			Provider:   &globalCfg.ModelProviders[0],
			ModelID:    defaultModel,
		}

		msg, err = aiAgent.Execute()
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}
		msg = strings.TrimSpace(msg)
	} else if mode == git.CommitModeManual && msg == "" {
		return fmt.Errorf("manual commit message is empty")
	}

	log.Info("Staging all changes...")
	if err := git.AddFiles(cwd); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	log.Info(fmt.Sprintf("Committing changes with message: %s", msg))
	if err := git.CommitAll(cwd, msg); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Success("Changes committed successfully.")
	return nil
}
