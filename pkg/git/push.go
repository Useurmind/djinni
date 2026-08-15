package git

import (
	"fmt"

	"github.com/useurmind/djinni/pkg/log"
)

func PushBranch(repoPath, branchName string) error {
	log.Info(fmt.Sprintf("Pushing branch %s to origin", branchName))

	if err := execCommand("git", []string{"push", "origin", branchName}, repoPath); err != nil {
		return fmt.Errorf("failed to push branch %s: %w", branchName, err)
	}

	log.Success(fmt.Sprintf("Successfully pushed branch %s", branchName))
	return nil
}
