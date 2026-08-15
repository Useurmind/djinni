package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/useurmind/djinni/pkg/log"
	"github.com/useurmind/djinni/pkg/utils"
)

func CreatePatch(repoPath, branchName, patchDir string) error {
	log.Info(fmt.Sprintf("Creating patch from branch %s", branchName))

	args := []string{"diff"}

	output, err := utils.ExecCommandWithOutput("git", args, repoPath)
	if err != nil {
		return fmt.Errorf("failed to generate patch: %w", err)
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

type BranchInfo struct {
	BranchName string
	BaseHash   string
	HeadHash   string
}
