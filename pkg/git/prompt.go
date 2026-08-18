package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/useurmind/djinni/pkg/log"
)

type CommitMode int

const (
	CommitModeNone CommitMode = iota
	CommitModeAI
	CommitModeManual
)

func PromptCommitChanges(repoPath string, configPath string) (CommitMode, string, error) {
	clean, err := IsRepositoryClean(repoPath)
	if err != nil {
		return CommitModeNone, "", fmt.Errorf("failed to check repository status: %w", err)
	}

	if clean {
		log.Info("No uncommitted changes detected.")
		return CommitModeNone, "", nil
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("WARNING: Uncommitted changes detected!")
	fmt.Println("========================================")
	fmt.Println()

	log.Info("Git status:")
	statusOutput, err := getGitStatusOutput(repoPath)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to get git status: %v", err))
		fmt.Println("Unable to show git status.")
	} else {
		fmt.Println(statusOutput)
	}

	fmt.Println()
	log.Info("Changes will not be included in the container unless committed.")

	var confirm bool
	confirmField := huh.NewConfirm().
		Title("Commit changes before starting the agent?").
		Affirmative("Yes").Negative("No").
		Value(&confirm)

	if err := confirmField.Run(); err != nil {
		return CommitModeNone, "", fmt.Errorf("failed to prompt for commit: %w", err)
	}

	if !confirm {
		log.Info("Proceeding without committing changes.")
		return CommitModeNone, "", nil
	}

	var mode string
	modeField := huh.NewSelect[string]().
		Title("How to generate commit message?").
		Options(
			huh.NewOption("Use AI to generate", "ai"),
			huh.NewOption("Write manually", "manual"),
		).
		Value(&mode)

	if err := modeField.Run(); err != nil {
		return CommitModeNone, "", fmt.Errorf("failed to prompt for commit mode: %w", err)
	}

	if mode == "ai" {
		return CommitModeAI, "", nil
	}

	var msg string
	inputField := huh.NewInput().
		Title("Enter commit message").
		Prompt("> ").
		Value(&msg).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("message cannot be empty")
			}
			return nil
		})

	if err := inputField.Run(); err != nil {
		return CommitModeNone, "", fmt.Errorf("failed to enter commit message: %w", err)
	}

	return CommitModeManual, msg, nil
}

func PromptSyncApproach() (string, error) {
	var syncApproach string
	syncField := huh.NewSelect[string]().
		Title("Select sync approach").
		Options(
			huh.NewOption("none", "none"),
			huh.NewOption("gitpatch", "gitpatch"),
			huh.NewOption("automerge", "automerge"),
		).
		Value(&syncApproach)

	if err := syncField.Run(); err != nil {
		return "none", fmt.Errorf("failed to select sync approach: %w", err)
	}

	return syncApproach, nil
}

func PromptAutoDeleteBranch() (bool, error) {
	var autodelete bool
	deleteField := huh.NewConfirm().
		Title("Delete agent feature branch after sync?").
		Affirmative("Yes").Negative("No").
		Value(&autodelete)

	if err := deleteField.Run(); err != nil {
		return false, fmt.Errorf("failed to prompt for autodelete: %w", err)
	}

	return autodelete, nil
}

func PromptDeleteOnExit() (string, error) {
	var deleteMode string
	deleteField := huh.NewSelect[string]().
		Title("Delete agent workspace and writable paths content?").
		Options(
			huh.NewOption("Keep workspace", "none"),
			huh.NewOption("Delete everything", "all"),
		).
		Value(&deleteMode)

	if err := deleteField.Run(); err != nil {
		return "none", fmt.Errorf("failed to prompt for delete on exit: %w", err)
	}

	return deleteMode, nil
}

func getGitStatusOutput(repoPath string) (string, error) {
	cmd := exec.Command("git", "status")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get git status: %w", err)
	}
	return string(output), nil
}
