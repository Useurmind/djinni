package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/useurmind/djinni/pkg/log"
	"github.com/useurmind/djinni/pkg/utils"
)

func CreatePatch(repoPath, patchDir string) error {
	args := []string{"diff", "HEAD~1...HEAD"}

	output, err := utils.ExecCommandWithOutput("git", args, repoPath)
	if err != nil {
		return fmt.Errorf("failed to generate patch: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		log.Info("No changes detected, skipping patch creation")
		return nil
	}

	if err := os.MkdirAll(patchDir, 0755); err != nil {
		return fmt.Errorf("failed to create patch directory: %w", err)
	}

	log.Success(fmt.Sprintf("Created patches in %s", patchDir))
	return os.WriteFile(filepath.Join(patchDir, "content.patch"), []byte(output), 0644)
}

func ApplyPatch(repoPath, patchPath string) error {
	log.Info(fmt.Sprintf("Applying patch %s to %s", patchPath, repoPath))

	args := []string{"apply", patchPath}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply patch: %w, output: %s", err, string(output))
	}

	log.Success(fmt.Sprintf("Applied patch to %s", repoPath))
	os.Remove(patchPath)
	return nil
}

func ApplyPatchNoIndex(repoPath, patchPath string) error {
	log.Info(fmt.Sprintf("Applying patch %s to %s (files only, not index)", patchPath, repoPath))

	args := []string{"apply", "--whitespace=nowarn", patchPath}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply patch (no index): %w, output: %s", err, string(output))
	}

	log.Success(fmt.Sprintf("Applied patch to %s (files only)", repoPath))
	os.Remove(patchPath)
	return nil
}

type BranchInfo struct {
	BranchName string
	BaseHash   string
	HeadHash   string
}
