package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetChangedFilesWithDiffs(repoPath string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run git status in '%s': %w", repoPath, err)
	}

	return parsePorcelainOutputWithDiffs(repoPath, string(output)), nil
}

func parsePorcelainOutputWithDiffs(repoPath, output string) string {
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

		diff, _ := getFileDiff(repoPath, filePath)

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

func getFileDiff(repoPath, filePath string) (string, error) {
	cmd := exec.Command("git", "diff", "HEAD", "--", filePath)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
