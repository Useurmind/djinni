package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushBranch(t *testing.T) {
	tempDir := t.TempDir()

	remoteDir := filepath.Join(tempDir, "remote-repo")
	require.NoError(t, os.MkdirAll(remoteDir, 0755))

	remoteInitCmd := exec.Command("git", "init", "--bare")
	remoteInitCmd.Dir = remoteDir
	output, err := remoteInitCmd.CombinedOutput()
	require.NoError(t, err, "git init --bare failed: %s", string(output))

	localDir := filepath.Join(tempDir, "local-repo")
	require.NoError(t, os.MkdirAll(localDir, 0755))

	gitInitCmd := exec.Command("git", "init")
	gitInitCmd.Dir = localDir
	output, err = gitInitCmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(output))

	// Set git config in the repo
	configCmd := exec.Command("git", "config", "user.name", "Test")
	configCmd.Dir = localDir
	require.NoError(t, configCmd.Run())

	configCmd = exec.Command("git", "config", "user.email", "test@test.com")
	configCmd.Dir = localDir
	require.NoError(t, configCmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(localDir, "README.md"), []byte("# Test"), 0644))

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = localDir
	require.NoError(t, addCmd.Run())

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = localDir
	require.NoError(t, commitCmd.Run())

	branchCmd := exec.Command("git", "branch", "-M", "main")
	branchCmd.Dir = localDir
	require.NoError(t, branchCmd.Run())

	remoteCmd := exec.Command("git", "remote", "add", "origin", remoteDir)
	remoteCmd.Dir = localDir
	require.NoError(t, remoteCmd.Run())

	err = PushBranch(localDir, "main")
	require.NoError(t, err)

	files, err := os.ReadDir(remoteDir)
	require.NoError(t, err)
	assert.NotEmpty(t, files)
}
