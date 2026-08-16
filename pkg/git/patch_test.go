package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePatch(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "test-repo")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	gitInitCmd := exec.Command("git", "init")
	gitInitCmd.Dir = sourceDir
	output, err := gitInitCmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(output))

	require.NoError(t, execCommand("git", []string{"config", "user.name", "Test"}, sourceDir))
	require.NoError(t, execCommand("git", []string{"config", "user.email", "test@test.com"}, sourceDir))

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "README.md"), []byte("# Test"), 0644))

	require.NoError(t, execCommand("git", []string{"add", "."}, sourceDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "Initial commit"}, sourceDir))
	require.NoError(t, execCommand("git", []string{"branch", "-M", "main"}, sourceDir))

	require.NoError(t, execCommand("git", []string{"checkout", "-b", "feature/test"}, sourceDir))

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new_file.txt"), []byte("test content"), 0644))
	require.NoError(t, execCommand("git", []string{"add", "."}, sourceDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "Add new file"}, sourceDir))

	patchDir := filepath.Join(tempDir, "patches")
	require.NoError(t, os.MkdirAll(patchDir, 0755))
	err = CreatePatch(sourceDir, patchDir)
	require.NoError(t, err)

	patches, err := filepath.Glob(filepath.Join(patchDir, "*.patch"))
	require.NoError(t, err)
	assert.NotEmpty(t, patches)

	patchContent, err := os.ReadFile(patches[0])
	require.NoError(t, err)
	assert.Contains(t, string(patchContent), "new_file.txt")
}

func TestApplyPatch(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "source-repo")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	gitInitCmd := exec.Command("git", "init")
	gitInitCmd.Dir = sourceDir
	output, err := gitInitCmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(output))

	require.NoError(t, execCommand("git", []string{"config", "user.name", "Test"}, sourceDir))
	require.NoError(t, execCommand("git", []string{"config", "user.email", "test@test.com"}, sourceDir))

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "README.md"), []byte("# Test"), 0644))

	require.NoError(t, execCommand("git", []string{"add", "."}, sourceDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "Initial commit"}, sourceDir))
	require.NoError(t, execCommand("git", []string{"branch", "-M", "main"}, sourceDir))

	require.NoError(t, execCommand("git", []string{"checkout", "-b", "feature/test"}, sourceDir))

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new_file.txt"), []byte("test content"), 0644))
	require.NoError(t, execCommand("git", []string{"add", "."}, sourceDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "Add new file"}, sourceDir))

	patchDir := filepath.Join(tempDir, "patches")
	require.NoError(t, os.MkdirAll(patchDir, 0755))
	err = CreatePatch(sourceDir, patchDir)
	require.NoError(t, err)

	patches, err := filepath.Glob(filepath.Join(patchDir, "*.patch"))
	require.NoError(t, err)
	require.NotEmpty(t, patches)

	targetDir := filepath.Join(tempDir, "target-repo")
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	gitInitCmd = exec.Command("git", "init")
	gitInitCmd.Dir = targetDir
	output, err = gitInitCmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(output))

	require.NoError(t, execCommand("git", []string{"config", "user.name", "Test"}, targetDir))
	require.NoError(t, execCommand("git", []string{"config", "user.email", "test@test.com"}, targetDir))

	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "README.md"), []byte("# Target"), 0644))

	require.NoError(t, execCommand("git", []string{"add", "."}, targetDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "Initial commit"}, targetDir))
	require.NoError(t, execCommand("git", []string{"branch", "-M", "main"}, targetDir))

	err = ApplyPatch(targetDir, patches[0])
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(targetDir, "new_file.txt"))

	statusOutput, err := execCommandOutput("git", []string{"diff", "--name-only"}, targetDir)
	require.NoError(t, err)
	assert.Empty(t, statusOutput, "No staged changes should exist after ApplyPatch (user needs to staging/commit)")
}

func execCommandOutput(name string, args []string, workdir string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = workdir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestCreatePatchAndApplyIntegration(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "source-repo")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	gitInitCmd := exec.Command("git", "init")
	gitInitCmd.Dir = sourceDir
	output, err := gitInitCmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(output))

	require.NoError(t, execCommand("git", []string{"config", "user.name", "Test"}, sourceDir))
	require.NoError(t, execCommand("git", []string{"config", "user.email", "test@test.com"}, sourceDir))

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "README.md"), []byte("# Source"), 0644))

	require.NoError(t, execCommand("git", []string{"add", "."}, sourceDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "Initial commit"}, sourceDir))
	require.NoError(t, execCommand("git", []string{"branch", "-M", "main"}, sourceDir))

	require.NoError(t, execCommand("git", []string{"checkout", "-b", "feature/test"}, sourceDir))

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file1.txt"), []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file2.txt"), []byte("content2"), 0644))
	require.NoError(t, execCommand("git", []string{"add", "."}, sourceDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "Add two files"}, sourceDir))

	patchDir := filepath.Join(tempDir, "patches")
	require.NoError(t, os.MkdirAll(patchDir, 0755))
	err = CreatePatch(sourceDir, patchDir)
	require.NoError(t, err)

	patches, err := filepath.Glob(filepath.Join(patchDir, "*.patch"))
	require.NoError(t, err)
	require.NotEmpty(t, patches)

	targetDir := filepath.Join(tempDir, "target-repo")
	require.NoError(t, os.MkdirAll(targetDir, 0755))

	gitInitCmd = exec.Command("git", "init")
	gitInitCmd.Dir = targetDir
	output, err = gitInitCmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(output))

	require.NoError(t, execCommand("git", []string{"config", "user.name", "Test"}, targetDir))
	require.NoError(t, execCommand("git", []string{"config", "user.email", "test@test.com"}, targetDir))

	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "README.md"), []byte("# Target"), 0644))

	require.NoError(t, execCommand("git", []string{"add", "."}, targetDir))
	require.NoError(t, execCommand("git", []string{"commit", "-m", "Initial commit"}, targetDir))
	require.NoError(t, execCommand("git", []string{"branch", "-M", "main"}, targetDir))

	err = ApplyPatch(targetDir, patches[0])
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(targetDir, "file1.txt"))
	assert.FileExists(t, filepath.Join(targetDir, "file2.txt"))

	statusOutput, err := execCommandOutput("git", []string{"diff", "--name-only"}, targetDir)
	require.NoError(t, err)
	assert.Empty(t, statusOutput, "No staged changes should exist after ApplyPatch (user needs to staging/commit)")
}
