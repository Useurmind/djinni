package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/useurmind/djinni/pkg/log"
	"github.com/useurmind/djinni/pkg/utils"
)

// CloneToTemp clones a repository to a temporary directory
func CloneToTemp(sourceDir, baseDir, agentName, taskName string) (string, error) {
	repoName, err := GetRepoName(sourceDir)
	if err != nil {
		return "", err
	}

	destDir := filepath.Join(baseDir, repoName, agentName, taskName)

	stat, err := os.Stat(destDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to stat destination directory %s: %w", destDir, err)
		}
		// Directory doesn't exist, we'll create it
	} else if stat.IsDir() {
		log.Info(fmt.Sprintf("Using existing workspace: %s", destDir))
		return destDir, nil
	} else {
		return "", fmt.Errorf("destination path %s exists but is not a directory", destDir)
	}

	log.Info(fmt.Sprintf("Cloning repository from %s to %s", sourceDir, destDir))

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := utils.ExecCommand("git", []string{"clone", sourceDir, destDir}, sourceDir); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	return destDir, nil
}

// CheckoutNewBranch creates a new branch
func CheckoutNewBranch(repoPath, taskName string) error {
	log.Info(fmt.Sprintf("Creating feature branch for task: %s", taskName))

	branchName := fmt.Sprintf("feature/%s", taskName)

	if err := utils.ExecCommand("git", []string{"checkout", "-b", branchName}, repoPath); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Info(fmt.Sprintf("Branch %s already exists, skipping", branchName))
			return nil
		}
		return fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	log.Success(fmt.Sprintf("Created branch: %s", branchName))
	return nil
}

// GetRepoName returns the repository name from a directory
func GetRepoName(dir string) (string, error) {
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

// DeleteBranch deletes a branch
func DeleteBranch(repoPath, branchName string) error {
	log.Info(fmt.Sprintf("Deleting branch %s", branchName))

	if err := utils.ExecCommand("git", []string{"branch", "-D", branchName}, repoPath); err != nil {
		return fmt.Errorf("failed to delete branch %s: %w", branchName, err)
	}

	log.Success(fmt.Sprintf("Deleted branch %s", branchName))
	return nil
}

// AddFiles stages all files in a repository
func AddFiles(repoPath string) error {
	log.Info(fmt.Sprintf("Staging all changes in %s", repoPath))

	if err := utils.ExecCommand("git", []string{"add", "."}, repoPath); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	log.Success(fmt.Sprintf("Staged all changes in %s", repoPath))
	return nil
}
