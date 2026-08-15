package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneToTemp(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "test-repo")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	gitInitCmd := exec.Command("git", "init")
	gitInitCmd.Dir = sourceDir
	output, err := gitInitCmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(output))

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "README.md"), []byte("# Test"), 0644))

	baseDir := filepath.Join(tempDir, "bases")
	agentName := "test-agent"
	taskName := "test-task"

	destDir, err := CloneToTemp(sourceDir, baseDir, agentName, taskName)
	require.NoError(t, err)

	assert.NotEmpty(t, destDir)
	assert.Contains(t, destDir, "test-repo/test-agent/test-task")

	info, err := os.Stat(destDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestSetPermissions(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test-perms")
	require.NoError(t, os.MkdirAll(testDir, 0755))

	filePath := filepath.Join(testDir, "test.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("test"), 0644))

	err := SetPermissions(testDir)
	require.NoError(t, err)

	info, err := os.Stat(testDir)
	require.NoError(t, err)

	perms := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0777), perms)
}

func TestGetRepoName(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "my-repo")
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	name, err := getRepoName(repoPath)
	require.NoError(t, err)
	assert.Equal(t, "my-repo", name)
}

func TestGetRepoNameWithNestedPath(t *testing.T) {
	tempDir := t.TempDir()
	repoPath := filepath.Join(tempDir, "parent", "my-repo")
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	name, err := getRepoName(repoPath)
	require.NoError(t, err)
	assert.Equal(t, "my-repo", name)
}

func TestGetRepoNameInvalidPath(t *testing.T) {
	tempDir := t.TempDir()
	nonExistent := filepath.Join(tempDir, "does-not-exist")

	_, err := getRepoName(nonExistent)
	assert.Error(t, err)
}

func TestGetRepoNameFileNotDir(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("test"), 0644))

	_, err := getRepoName(filePath)
	assert.Error(t, err)
}
