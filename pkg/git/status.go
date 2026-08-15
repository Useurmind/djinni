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

type ChangedFile struct {
	Path   string
	Status string
	Diff   string
}

func GetDiff(repoPath, filePath string) (string, error) {
	cmd := exec.Command("git", "diff", "HEAD", "--", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run git diff for '%s': %w", filePath, err)
	}

	return string(output), nil
}

func GetAllDiffs(repoPath string) ([]ChangedFile, error) {
	changedFiles, err := GetChangedFiles(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	var result []ChangedFile

	for _, filePath := range changedFiles {
		status, err := GetFileStatus(repoPath, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get status for '%s': %w", filePath, err)
		}

		diff, err := GetDiff(repoPath, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get diff for '%s': %w", filePath, err)
		}

		result = append(result, ChangedFile{
			Path:   filePath,
			Status: status,
			Diff:   diff,
		})
	}

	return result, nil
}

func GetFileStatus(repoPath, filePath string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain", "--", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run git status for '%s': %w", filePath, err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		if len(line) < 3 {
			continue
		}

		status := line[:2]
		statusStr := ParseStatus(status)
		return statusStr, nil
	}

	return "unknown", nil
}

func ParseStatus(status string) string {
	status = strings.TrimSpace(status)
	if len(status) == 0 {
		return "unknown()"
	}
	switch status[0] {
	case 'M':
		return "modified"
	case 'A':
		return "added"
	case 'D':
		return "deleted"
	case 'R':
		return "renamed"
	case 'C':
		return "copied"
	case 'U':
		return "unmerged"
	case '?':
		return "untracked"
	default:
		return fmt.Sprintf("unknown(%s)", status)
	}
}
