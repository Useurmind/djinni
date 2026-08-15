package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	} else {
		fmt.Println(statusOutput)
	}

	fmt.Println()
	log.Info("Changes will not be included in the container unless committed.")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Commit changes before starting the agent? (yes/no): ")
		response, err := reader.ReadString('\n')
		if err != nil {
			return CommitModeNone, "", fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))

		switch response {
		case "yes", "y":
			mode, msg, err := promptCommitMode(repoPath, configPath)
			if err != nil {
				return CommitModeNone, "", err
			}
			return mode, msg, nil
		case "no", "n":
			log.Info("Proceeding without committing changes.")
			return CommitModeNone, "", nil
		default:
			fmt.Println("Please answer 'yes' or 'no'.")
		}
	}
}

func promptCommitMode(repoPath string, configPath string) (CommitMode, string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Use AI to generate commit message? (yes/no): ")
		response, err := reader.ReadString('\n')
		if err != nil {
			return CommitModeNone, "", fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))

		switch response {
		case "yes", "y":
			return CommitModeAI, "", nil
		case "no", "n":
			log.Info("Please enter your commit message:")
			fmt.Print("> ")
			commitMsg, err := reader.ReadString('\n')
			if err != nil {
				return CommitModeNone, "", fmt.Errorf("failed to read commit message: %w", err)
			}
			commitMsg = strings.TrimSpace(commitMsg)
			if commitMsg == "" {
				fmt.Println("Commit message cannot be empty. Please try again.")
				continue
			}
			return CommitModeManual, commitMsg, nil
		default:
			fmt.Println("Please answer 'yes' or 'no'.")
		}
	}
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
