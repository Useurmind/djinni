package git

import (
	"fmt"
	"os/exec"

	"github.com/useurmind/djinni/pkg/log"
	"github.com/useurmind/djinni/pkg/utils"
)

func CreatePatch(repoPath, branchName, patchDir string) error {
	log.Info(fmt.Sprintf("Creating patch from branch %s", branchName))

	branchInfo, err := getBranchInfo(repoPath, branchName)
	if err != nil {
		return fmt.Errorf("failed to get branch info: %w", err)
	}

	baseHash := branchInfo.BaseHash
	headHash := branchInfo.HeadHash

	args := []string{"format-patch", "-o", patchDir}
	if baseHash == "" || baseHash == headHash {
		args = append(args, "HEAD~1..HEAD")
	} else {
		args = append(args, baseHash+".."+headHash)
	}

	if err := utils.ExecCommand("git", args, repoPath); err != nil {
		return fmt.Errorf("failed to generate patch: %w", err)
	}

	log.Success(fmt.Sprintf("Created patches in %s", patchDir))
	return nil
}

func ApplyPatch(repoPath, patchPath string) error {
	log.Info(fmt.Sprintf("Applying patch %s to %s", patchPath, repoPath))

	args := []string{"apply", "-p1", patchPath}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply patch: %w, output: %s", err, string(output))
	}

	log.Success(fmt.Sprintf("Applied patch to %s", repoPath))
	return nil
}

type BranchInfo struct {
	BranchName string
	BaseHash   string
	HeadHash   string
}

func getBranchInfo(repoPath, branchName string) (BranchInfo, error) {
	cmd := exec.Command("git", "rev-parse", branchName)
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return BranchInfo{}, fmt.Errorf("failed to get branch head: %w", err)
	}
	headHash := string(output)

	mergeBaseCmd := exec.Command("git", "merge-base", "HEAD", branchName)
	mergeBaseCmd.Dir = repoPath
	mergeBaseOutput, err := mergeBaseCmd.Output()
	if err != nil {
		return BranchInfo{}, fmt.Errorf("failed to get merge base: %w", err)
	}
	baseHash := string(mergeBaseOutput)

	return BranchInfo{
		BranchName: branchName,
		BaseHash:   baseHash,
		HeadHash:   headHash,
	}, nil
}
