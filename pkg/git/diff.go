package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/useurmind/djinni/pkg/log"
)

// Diff represents a file change with its status and diff content
type Diff struct {
	Path    string
	Status  string
	Content string
}

// GetChangedFilesWithDiffs returns a formatted string of all changed files with their diffs.
// Returns "No changes detected." if there are no changes.
func GetChangedFilesWithDiffs(repoPath string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run git status in '%s': %w", repoPath, err)
	}

	return formatChangedFilesWithDiffs(repoPath, string(output)), nil
}

// formatChangedFilesWithDiffs processes git porcelain output and formats files with their diffs
func formatChangedFilesWithDiffs(repoPath, output string) string {
	lines := strings.Split(output, "\n")
	var results []string

	for _, line := range lines {
		if line == "" {
			continue
		}

		if len(line) < 3 {
			continue
		}

		status := line[:2]
		filePath := strings.TrimSpace(line[2:])

		statusStr := ParseStatus(status)

		diff, err := GetFileDiff(repoPath, filePath)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to get diff for %s: %v", filePath, err))
			diff = ""
		}

		result := fmt.Sprintf("- %s %s", filePath, statusStr)
		if diff != "" {
			result += "\n" + diff
		}
		results = append(results, result)
	}

	if len(results) == 0 {
		return "No changes detected."
	}

	return strings.Join(results, "\n")
}

// GetFileDiff returns the diff content for a specific file
func GetFileDiff(repoPath, filePath string) (string, error) {
	cmd := exec.Command("git", "diff", "HEAD", "--", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// GetDiffs returns detailed diff information for all changed files
func GetDiffs(repoPath string) ([]Diff, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git status in '%s': %w", repoPath, err)
	}

	lines := strings.Split(string(output), "\n")
	var result []Diff

	for _, line := range lines {
		if line == "" {
			continue
		}

		if len(line) < 3 {
			continue
		}

		status := line[:2]
		filePath := strings.TrimSpace(line[2:])

		statusStr := ParseStatus(status)

		diff, err := GetFileDiff(repoPath, filePath)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to get diff for %s: %v", filePath, err))
			diff = ""
		}

		result = append(result, Diff{
			Path:    filePath,
			Status:  statusStr,
			Content: diff,
		})
	}

	return result, nil
}

// GetDiff returns the diff content for a specific file
func GetDiff(repoPath, filePath string) (string, error) {
	cmd := exec.Command("git", "diff", "HEAD", "--", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run git diff for '%s': %w", filePath, err)
	}

	return string(output), nil
}

// ParseStatus converts git status codes to human-readable strings
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
