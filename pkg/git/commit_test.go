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

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "README.md"), []byte("# Test"), 0644))

	require.NoError(t, execCommand("git", []string{"add", "."}, sourceDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "Initial commit"}, sourceDir))

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
