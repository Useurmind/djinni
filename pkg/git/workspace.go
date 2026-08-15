package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/useurmind/djinni/pkg/log"
	"github.com/useurmind/djinni/pkg/utils"
)

func CloneToTemp(sourceDir, baseDir, agentName, taskName string) (string, error) {
	repoName, err := getRepoName(sourceDir)
	if err != nil {
		return "", err
	}

	destDir := filepath.Join(baseDir, repoName, agentName, taskName)

	if _, err := os.Stat(destDir); err == nil {
		log.Info(fmt.Sprintf("Using existing workspace: %s", destDir))
		return destDir, nil
	}

	log.Info(fmt.Sprintf("Cloning repository from %s to %s", sourceDir, destDir))

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := execCommand("git", []string{"clone", sourceDir, destDir}, sourceDir); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	return destDir, nil
}

func CheckoutNewBranch(repoPath, taskName string) error {
	log.Info(fmt.Sprintf("Creating feature branch for task: %s", taskName))

	branchName := fmt.Sprintf("feature/%s", taskName)

	if err := execCommand("git", []string{"checkout", "-b", branchName}, repoPath); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Info(fmt.Sprintf("Branch %s already exists, skipping", branchName))
			return nil
		}
		return fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	log.Success(fmt.Sprintf("Created branch: %s", branchName))
	return nil
}

func CleanupWorkspace(workspacePath string) {
	if workspacePath == "" {
		return
	}

	log.Info(fmt.Sprintf("Cleaning up workspace: %s", workspacePath))

	if err := os.RemoveAll(workspacePath); err != nil {
		log.Error(fmt.Sprintf("Failed to cleanup workspace %s: %v", workspacePath, err))
	}
}

func getRepoName(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory")
	}

	return filepath.Base(absDir), nil
}

func execCommand(name string, args []string, workdir string) error {
	if err := utils.ExecCommand(name, args, workdir); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Executed: %s %s", name, strings.Join(args, " ")))
	return nil
}
