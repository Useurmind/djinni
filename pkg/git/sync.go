package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/useurmind/djinni/pkg/log"
	"github.com/useurmind/djinni/pkg/utils"
)

// SyncStrategy defines how changes are synced back to the user's repository
type SyncStrategy int

const (
	// SyncNone - no sync, changes remain on agent branch
	SyncNone SyncStrategy = iota
	// SyncPatch - sync via patch file
	SyncPatch
	// SyncBranch - sync via branch merge
	SyncBranch
)

// PushBranch pushes the specified branch to origin
func PushBranch(repoPath, branchName string) error {
	log.Info(fmt.Sprintf("Pushing branch %s to origin", branchName))

	if err := utils.ExecCommand("git", []string{"push", "origin", branchName}, repoPath); err != nil {
		return fmt.Errorf("failed to push branch %s: %w", branchName, err)
	}

	log.Success(fmt.Sprintf("Successfully pushed branch %s", branchName))
	return nil
}

// CreatePatch creates a git patch from HEAD~1 to HEAD
// The patch is written to patchDir/content.patch
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

// ApplyPatch applies a patch to a repository
func ApplyPatch(repoPath, patchPath string) error {
	log.Info(fmt.Sprintf("Applying patch %s to %s", patchPath, repoPath))

	args := []string{"apply", patchPath}

	if err := utils.ExecCommand("git", args, repoPath); err != nil {
		return fmt.Errorf("failed to apply patch: %w", err)
	}

	log.Success(fmt.Sprintf("Applied patch to %s", repoPath))
	os.Remove(patchPath)
	return nil
}

// ApplyPatchNoIndex applies a patch to a repository without updating the index
func ApplyPatchNoIndex(repoPath, patchPath string) error {
	log.Info(fmt.Sprintf("Applying patch %s to %s (files only, not index)", patchPath, repoPath))

	args := []string{"apply", "--whitespace=nowarn", patchPath}

	if err := utils.ExecCommand("git", args, repoPath); err != nil {
		return fmt.Errorf("failed to apply patch (no index): %w", err)
	}

	log.Success(fmt.Sprintf("Applied patch to %s (files only)", repoPath))
	os.Remove(patchPath)
	return nil
}
