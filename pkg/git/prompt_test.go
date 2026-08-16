package git

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptCommitChangesRepoClean(t *testing.T) {
	repoPath := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	assert.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoPath
	assert.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	assert.NoError(t, cmd.Run())

	// Create an initial commit
	filePath := repoPath + "/test.txt"
	err := exec.Command("touch", filePath).Run()
	assert.NoError(t, err)

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	assert.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoPath
	assert.NoError(t, cmd.Run())

	clean, err := IsRepositoryClean(repoPath)
	assert.NoError(t, err)
	assert.True(t, clean)

	mode, msg, err := PromptCommitChanges(repoPath, "")
	assert.NoError(t, err)
	assert.Equal(t, CommitModeNone, mode)
	assert.Equal(t, "", msg)
}
