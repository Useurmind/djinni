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

func TestGetDiff(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "test-repo")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	gitInitCmd := exec.Command("git", "init")
	gitInitCmd.Dir = sourceDir
	output, err := gitInitCmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(output))

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test"), 0644))
	require.NoError(t, execCommand("git", []string{"add", "."}, sourceDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "initial"}, sourceDir))

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test modified"), 0644))

	diff, err := GetDiff(sourceDir, "test.txt")
	require.NoError(t, err)
	assert.Contains(t, diff, "-test")
	assert.Contains(t, diff, "+test modified")
}

func TestGetFileStatus(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "test-repo")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	gitInitCmd := exec.Command("git", "init")
	gitInitCmd.Dir = sourceDir
	output, err := gitInitCmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(output))

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test"), 0644))

	status, err := GetFileStatus(sourceDir, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, "untracked", status)

	require.NoError(t, execCommand("git", []string{"add", "."}, sourceDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "initial"}, sourceDir))

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test modified"), 0644))

	status, err = GetFileStatus(sourceDir, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, "modified", status)
}

func TestGetAllDiffs(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "test-repo")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	gitInitCmd := exec.Command("git", "init")
	gitInitCmd.Dir = sourceDir
	output, err := gitInitCmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(output))

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file1.go"), []byte("package main\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file2.md"), []byte("# Test\n"), 0644))

	require.NoError(t, execCommand("git", []string{"add", "."}, sourceDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "initial"}, sourceDir))

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file1.go"), []byte("package main\nfunc main() {}\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file3.go"), []byte("package main\n"), 0644))
	require.NoError(t, execCommand("git", []string{"add", "file3.go"}, sourceDir))

	diffs, err := GetAllDiffs(sourceDir)
	require.NoError(t, err)

	foundModified := false
	foundAdded := false
	for _, d := range diffs {
		if d.Path == "file1.go" {
			assert.Equal(t, "modified", d.Status)
			assert.Contains(t, d.Diff, "func main()")
			foundModified = true
		}
		if d.Path == "file3.go" {
			assert.Equal(t, "added", d.Status)
			foundAdded = true
		}
	}
	assert.True(t, foundModified, "should find modified file1.go")
	assert.True(t, foundAdded, "should find added file3.go")
}
