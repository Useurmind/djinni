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
		return runStartAgent(cmd, args)
	},
}

func runStartAgent(cmd *cobra.Command, args []string) error {
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

	client, image, commands, _, workspacePath, err := prepareWorkspace(agentCfg, cwd, agentName, taskName)
	if err != nil {
		return err
	}

	mountSources := git.GetMountPaths(agentCfg.Mounts)
	exitCode, err := client.RunContainer(image, agentCfg.HarnessCommand, agentName, agentCfg.Mounts, commands)
	if err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}

	if len(mountSources) > 0 {
		log.Info("Restoring file ownership after container exit...")
		if err := git.RestoreOwnership(mountSources); err != nil {
			log.Error(fmt.Sprintf("Failed to restore ownership: %v", err))
		}
	}

	if workspacePath != "" {
		if err := handlePostExecution(agentCfg, cfg, cwd, taskName, workspacePath); err != nil {
			log.Error(fmt.Sprintf("Failed to handle post-execution: %v", err))
		}
	}

	os.Exit(exitCode)
	return nil
}

func prepareWorkspace(agentCfg *config.AgentConfig, cwd, agentName, taskName string) (*docker.Client, string, *docker.ContainerCommands, []docker.FilesToCopy, string, error) {
	log.Info("Initializing container client...")
	client, err := docker.NewClient()
	if err != nil {
		return nil, "", nil, nil, "", fmt.Errorf("failed to initialize container client: %w", err)
	}

	image, commands, filesToCopy, workspacePath, err := setupWorkspace(agentCfg, cwd, agentName, taskName)
	if err != nil {
		return nil, "", nil, nil, "", err
	}

	for _, fc := range agentCfg.FilesToCopy {
		filesToCopy = append(filesToCopy, docker.FilesToCopy{
			Source:      fc.Source,
			Destination: fc.Destination,
		})
	}

	if len(filesToCopy) > 0 {
		if commands == nil {
			commands = &docker.ContainerCommands{}
		}
		commands.FilesToCopy = filesToCopy
	}

	if workspacePath != "" {
		log.Info(fmt.Sprintf("Using local workspace: %s", workspacePath))
		if err := git.CheckoutNewBranch(workspacePath, taskName); err != nil {
			return nil, "", nil, nil, "", fmt.Errorf("failed to checkout branch: %w", err)
		}
	}

	log.Info(fmt.Sprintf("Using image: %s", image))

	return client, image, commands, filesToCopy, workspacePath, nil
}

func setupWorkspace(agentCfg *config.AgentConfig, cwd, agentName, taskName string) (string, *docker.ContainerCommands, []docker.FilesToCopy, string, error) {
	var image string
	var commands *docker.ContainerCommands
	filesToCopy := []docker.FilesToCopy{}
	var workspacePath string

	image = agentCfg.Image
	if agentCfg.Containerfile != "" {
		image = fmt.Sprintf("djinni-%s:latest", agentName)
		log.Info(fmt.Sprintf("Using local image: %s", image))
	}

	if taskName != "" {
		log.Info("Setting up git workspace mount...")

		baseDir := agentCfg.GitWorkspace.BaseDirectory

		var err error
		workspacePath, err = git.CloneToTemp(cwd, baseDir, agentName, taskName)
		if err != nil {
			return "", nil, nil, "", fmt.Errorf("failed to clone workspace: %w", err)
		}

		workspaceMountPath := filepath.Join("/workspace", fmt.Sprintf("%s-%s", filepath.Base(cwd), taskName))
		agentCfg.Mounts = append(agentCfg.Mounts, config.Mount{
			Source:      workspacePath,
			Destination: workspaceMountPath,
		})

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", nil, nil, "", fmt.Errorf("failed to get user home directory: %w", err)
		}

		gitconfigPath := filepath.Join(homeDir, ".gitconfig")
		if _, err := os.Stat(gitconfigPath); os.IsNotExist(err) {
			return "", nil, nil, "", fmt.Errorf("gitconfig file not found at %s", gitconfigPath)
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
	}

	return image, commands, filesToCopy, workspacePath, nil
}

func handlePostExecution(agentCfg *config.AgentConfig, cfg *config.Config, cwd, taskName, workspacePath string) error {
	log.Info("Checking for changes after container exit...")

	clean, err := git.IsRepositoryClean(workspacePath)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to check repository status: %v", err))
		return nil
	}

	if clean {
		log.Info("No changes detected, skipping commit/push")
		return nil
	}

	changedFiles, err := git.GetChangedFiles(workspacePath)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to get changed files: %v", err))
		return nil
	}

	log.Info(fmt.Sprintf("Detected %d changed files", len(changedFiles)))

	syncApproach := agentCfg.SyncApproach
	if syncApproach == "" {
		syncApproach, err = git.PromptSyncApproach()
		if err != nil {
			log.Error(fmt.Sprintf("Failed to prompt for sync approach: %v", err))
			return nil
		}
	}

	autodelete := agentCfg.AutoDeleteAgentBranch
	if !autodelete && syncApproach != "none" {
		autodelete, err = git.PromptAutoDeleteBranch()
		if err != nil {
			log.Error(fmt.Sprintf("Failed to prompt for autodelete: %v", err))
			return nil
		}
	}

	branchName := fmt.Sprintf("feature/%s", taskName)
	err = commitAndPushFromAgent(agentCfg, branchName, workspacePath, cwd)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to commit and push changes: %v", err))
		return nil
	}

	switch syncApproach {
	case "gitpatch":
		return syncWithPatch(agentCfg, branchName, workspacePath, cwd, autodelete)
	case "automerge":
		return syncWithBranch(agentCfg, branchName, workspacePath, cwd, autodelete)
	case "none":
		log.Info("No sync approach selected, leaving changes on feature branch")
	default:
		log.Error(fmt.Sprintf("Unknown sync approach: %s", syncApproach))
		return nil
	}

	return nil
}

func commitAndPushFromAgent(agentCfg *config.AgentConfig, branchName, workspacePath, cwd string) error {
	log.Info("Staging all changes...")
	if err := git.AddFiles(workspacePath); err != nil {
		log.Error(fmt.Sprintf("Failed to stage files: %v", err))
		return err
	}

	defaultModel := ""
	if agentCfg.DefaultModel != "" {
		defaultModel = agentCfg.DefaultModel
	} else {
		cfg, _ := config.LoadConfig("")
		defaultModel = cfg.DefaultModel
	}

	commitMsg, err := generateCommitMessage(workspacePath, defaultModel)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to generate commit message: %v", err))
		return err
	}

	if err := git.CommitAll(workspacePath, commitMsg); err != nil {
		log.Error(fmt.Sprintf("Failed to commit: %v", err))
		return err
	}

	if err := git.PushBranch(workspacePath, branchName); err != nil {
		log.Error(fmt.Sprintf("Failed to push branch: %v", err))
		return err
	}
	return nil
}

func syncWithPatch(agentCfg *config.AgentConfig, branchName, workspacePath, cwd string, autodelete bool) error {
	patchDir := "/tmp/djinni/patches"
	err := os.RemoveAll(patchDir)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to remove patch directory: %v", err))
		return err
	}

	if err := git.CreatePatch(workspacePath, patchDir); err != nil {
		log.Error(fmt.Sprintf("Failed to create patch: %v", err))
		return err
	}

	patches, _ := filepath.Glob(filepath.Join(patchDir, "*.patch"))
	if len(patches) > 0 {
		log.Info("Applying patch to user workspace (files only)...")
		if err := git.ApplyPatchNoIndex(cwd, patches[0]); err != nil {
			log.Error(fmt.Sprintf("Failed to apply patch: %v", err))
			return err
		}
	}

	if autodelete {
		if err := git.DeleteBranch(workspacePath, branchName); err != nil {
			log.Error(fmt.Sprintf("Failed to delete branch: %v", err))
		}
	}

	return nil
}

func syncWithBranch(agentCfg *config.AgentConfig, branchName, workspacePath, cwd string, autodelete bool) error {

	log.Info(fmt.Sprintf("Merging branch %s into current branch", branchName))
	if err := execCommand("git", []string{"merge", branchName}, cwd); err != nil {
		log.Error(fmt.Sprintf("Failed to merge branch: %v", err))
		return err
	}

	if autodelete {
		if err := git.DeleteBranch(workspacePath, branchName); err != nil {
			log.Error(fmt.Sprintf("Failed to delete branch: %v", err))
		}
	}

	return nil
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

func generateCommitMessage(workingDir, defaultModel string) (string, error) {
	globalCfg, err := config.LoadGlobalConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load global config: %w", err)
	}

	if len(globalCfg.ModelProviders) == 0 {
		return "", fmt.Errorf("no model providers configured in global config")
	}

	aiAgent := &ai.Agent{
		WorkingDir: workingDir,
		ReadPaths:  []string{workingDir},
		Provider:   &globalCfg.ModelProviders[0],
		ModelID:    defaultModel,
	}

	if defaultModel != "" {
		aiAgent.ModelID = defaultModel
	}

	commitMsg, err := aiAgent.Execute()
	if err != nil {
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	return strings.TrimSpace(commitMsg), nil
}

func init() {
	rootCmd.AddCommand(startAgentCmd)
	startAgentCmd.Flags().StringP("config", "c", "", "Path to config file (default: .djinni.yml in current directory)")
	startAgentCmd.Flags().StringP("task", "t", "", "Task name for workspace mount (creates feature/<taskname> branch)")
	if err := startAgentCmd.MarkFlagRequired("task"); err != nil {
		log.Error(fmt.Sprintf("Failed to mark flag required: %v", err))
	}
}
