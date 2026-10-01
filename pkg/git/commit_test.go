package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommitAll(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "test-repo")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	gitInitCmd := exec.Command("git", "init")
	gitInitCmd.Dir = sourceDir
	output, err := gitInitCmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(output))

	// Set git config in the repo
	configCmd := exec.Command("git", "config", "user.name", "Test")
	configCmd.Dir = sourceDir
	require.NoError(t, configCmd.Run())

	configCmd = exec.Command("git", "config", "user.email", "test@test.com")
	configCmd.Dir = sourceDir
	require.NoError(t, configCmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "README.md"), []byte("# Test"), 0644))

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = sourceDir
	require.NoError(t, addCmd.Run())

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = sourceDir
	require.NoError(t, commitCmd.Run())

	newFile := filepath.Join(sourceDir, "new_file.txt")
	require.NoError(t, os.WriteFile(newFile, []byte("test content"), 0644))

	err = CommitAll(sourceDir, "Add new file")
	require.NoError(t, err)

	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = sourceDir
	output, err = statusCmd.CombinedOutput()
	require.NoError(t, err)
	assert.Empty(t, strings.TrimSpace(string(output)))
}
