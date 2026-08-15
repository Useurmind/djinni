package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRepositoryClean(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "test-repo")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	gitInitCmd := exec.Command("git", "init")
	gitInitCmd.Dir = sourceDir
	output, err := gitInitCmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(output))

	clean, err := IsRepositoryClean(sourceDir)
	require.NoError(t, err)
	assert.True(t, clean)

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test"), 0644))

	clean, err = IsRepositoryClean(sourceDir)
	require.NoError(t, err)
	assert.False(t, clean)

	require.NoError(t, execCommand("git", []string{"add", "."}, sourceDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "test"}, sourceDir))

	clean, err = IsRepositoryClean(sourceDir)
	require.NoError(t, err)
	assert.True(t, clean)
}

func TestGetChangedFiles(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "test-repo")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	gitInitCmd := exec.Command("git", "init")
	gitInitCmd.Dir = sourceDir
	output, err := gitInitCmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(output))

	files, err := GetChangedFiles(sourceDir)
	require.NoError(t, err)
	assert.Empty(t, files)

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test"), 0644))

	files, err = GetChangedFiles(sourceDir)
	require.NoError(t, err)
	assert.Equal(t, []string{"test.txt"}, files)
}
