package git

import (
	"fmt"

	"github.com/useurmind/djinni/pkg/log"
)

func CommitAll(repoPath, message string) error {
	log.Info(fmt.Sprintf("Committing changes in %s", repoPath))

	if err := execCommand("git", []string{"add", "."}, repoPath); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	if err := execCommand("git", []string{"commit", "-m", message}, repoPath); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	log.Success(fmt.Sprintf("Committed changes to %s", repoPath))
	return nil
}
