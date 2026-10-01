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

	// Set git config in the repo
	configCmd := exec.Command("git", "config", "user.name", "Test")
	configCmd.Dir = sourceDir
	require.NoError(t, configCmd.Run())

	configCmd = exec.Command("git", "config", "user.email", "test@test.com")
	configCmd.Dir = sourceDir
	require.NoError(t, configCmd.Run())

	clean, err := IsRepositoryClean(sourceDir)
	require.NoError(t, err)
	assert.True(t, clean)

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test"), 0644))

	clean, err = IsRepositoryClean(sourceDir)
	require.NoError(t, err)
	assert.False(t, clean)

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = sourceDir
	require.NoError(t, addCmd.Run())

	commitCmd := exec.Command("git", "commit", "-m", "test")
	commitCmd.Dir = sourceDir
	require.NoError(t, commitCmd.Run())

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

	// Set git config in the repo
	configCmd := exec.Command("git", "config", "user.name", "Test")
	configCmd.Dir = sourceDir
	require.NoError(t, configCmd.Run())

	configCmd = exec.Command("git", "config", "user.email", "test@test.com")
	configCmd.Dir = sourceDir
	require.NoError(t, configCmd.Run())

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

	// Set git config in the repo
	configCmd := exec.Command("git", "config", "user.name", "Test")
	configCmd.Dir = sourceDir
	require.NoError(t, configCmd.Run())

	configCmd = exec.Command("git", "config", "user.email", "test@test.com")
	configCmd.Dir = sourceDir
	require.NoError(t, configCmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test"), 0644))

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = sourceDir
	require.NoError(t, addCmd.Run())

	commitCmd := exec.Command("git", "commit", "-m", "initial")
	commitCmd.Dir = sourceDir
	require.NoError(t, commitCmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test modified"), 0644))

	diff, err := GetDiff(sourceDir, "test.txt")
	require.NoError(t, err)
	assert.Contains(t, diff, "-test")
	assert.Contains(t, diff, "+test modified")
}

func TestParseStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"M ", "modified"},
		{"A ", "added"},
		{"D ", "deleted"},
		{"R ", "renamed"},
		{"C ", "copied"},
		{"U ", "unmerged"},
		{"?? ", "untracked"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDiffs(t *testing.T) {
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

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file1.go"), []byte("package main\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file2.md"), []byte("# Test\n"), 0644))

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = sourceDir
	require.NoError(t, addCmd.Run())

	commitCmd := exec.Command("git", "commit", "-m", "initial")
	commitCmd.Dir = sourceDir
	require.NoError(t, commitCmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file1.go"), []byte("package main\nfunc main() {}\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file3.go"), []byte("package main\n"), 0644))

	addCmd = exec.Command("git", "add", "file3.go")
	addCmd.Dir = sourceDir
	require.NoError(t, addCmd.Run())

	diffs, err := GetDiffs(sourceDir)
	require.NoError(t, err)

	foundModified := false
	foundAdded := false
	for _, d := range diffs {
		if d.Path == "file1.go" {
			assert.Equal(t, "modified", d.Status)
			assert.Contains(t, d.Content, "func main()")
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
