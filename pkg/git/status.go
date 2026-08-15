package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func IsRepositoryClean(repoPath string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to run git status in '%s': %w", repoPath, err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			return false, nil
		}
	}

	return true, nil
}

func GetChangedFiles(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git status in '%s': %w", repoPath, err)
	}

	return parsePorcelainOutput(string(output)), nil
}

func parsePorcelainOutput(output string) []string {
	lines := strings.Split(output, "\n")
	var results []string

	for _, line := range lines {
		if line == "" {
			continue
		}

		if len(line) < 3 {
			continue
		}

		filePath := strings.TrimSpace(line[2:])
		results = append(results, filePath)
	}

	return results
}
