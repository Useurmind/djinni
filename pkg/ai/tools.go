package ai

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ReadFileTool struct {
	readPaths []string
}

func NewReadFileTool(readPaths []string) *ReadFileTool {
	return &ReadFileTool{
		readPaths: readPaths,
	}
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read content from files in allowed paths. Input: file path relative to allowed directories."
}

func (t *ReadFileTool) Call(ctx context.Context, input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("file path is required")
	}

	for _, path := range t.readPaths {
		fullPath := filepath.Join(path, input)
		if _, err := os.Stat(fullPath); err == nil {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				return "", fmt.Errorf("failed to read file '%s': %w", fullPath, err)
			}
			return string(content), nil
		}
	}

	return "", fmt.Errorf("file '%s' not found in allowed paths", input)
}

type GitChangedFilesTool struct {
	workingDir string
}

func NewGitChangedFilesTool(workingDir string) *GitChangedFilesTool {
	return &GitChangedFilesTool{
		workingDir: workingDir,
	}
}

func (t *GitChangedFilesTool) Name() string {
	return "git_changed_files"
}

func (t *GitChangedFilesTool) Description() string {
	return "Get all files changed in the repository relative to HEAD. Output includes file path and status (modified, untracked, added, deleted, renamed, copied, unmerged). Input: unused."
}

func (t *GitChangedFilesTool) Call(ctx context.Context, input string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = t.workingDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run git status in '%s': %w", t.workingDir, err)
	}

	return t.parsePorcelainOutput(string(output)), nil
}

func (t *GitChangedFilesTool) parsePorcelainOutput(output string) string {
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

		statusStr := t.parseStatus(status)
		results = append(results, fmt.Sprintf("%s: %s", statusStr, filePath))
	}

	if len(results) == 0 {
		return "No changes detected."
	}

	return strings.Join(results, "\n")
}

func (t *GitChangedFilesTool) parseStatus(status string) string {
	switch status {
	case "M ":
		return "modified"
	case "A ":
		return "added"
	case "D ":
		return "deleted"
	case "R ":
		return "renamed"
	case "C ":
		return "copied"
	case "U ":
		return "unmerged"
	case "??":
		return "untracked"
	default:
		return fmt.Sprintf("unknown(%s)", status)
	}
}
