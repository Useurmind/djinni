package git

import (
	"fmt"

	"github.com/useurmind/djinni/pkg/log"
	"github.com/useurmind/djinni/pkg/utils"
)

// CommitAll stages and commits all changes in the repository
func CommitAll(repoPath, message string) error {
	log.Info(fmt.Sprintf("Committing changes in %s", repoPath))

	if err := utils.ExecCommand("git", []string{"add", "."}, repoPath); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	if err := utils.ExecCommand("git", []string{"commit", "-m", message}, repoPath); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Success(fmt.Sprintf("Committed changes to %s", repoPath))
	return nil
}
