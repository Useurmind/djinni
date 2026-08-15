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

	require.NoError(t, os.WriteFile(filepath.Join(localDir, "README.md"), []byte("# Test"), 0644))

	require.NoError(t, execCommand("git", []string{"add", "."}, localDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "Initial commit"}, localDir))
	require.NoError(t, execCommand("git", []string{"branch", "-M", "main"}, localDir))
	require.NoError(t, execCommand("git", []string{"remote", "add", "origin", remoteDir}, localDir))

	err = PushBranch(localDir, "main")
	require.NoError(t, err)

	files, err := os.ReadDir(remoteDir)
	require.NoError(t, err)
	assert.NotEmpty(t, files)
}
