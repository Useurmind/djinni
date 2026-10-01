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

	branchCmd := exec.Command("git", "branch", "-M", "main")
	branchCmd.Dir = sourceDir
	require.NoError(t, branchCmd.Run())

	branchCmd = exec.Command("git", "checkout", "-b", "feature/test")
	branchCmd.Dir = sourceDir
	require.NoError(t, branchCmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new_file.txt"), []byte("test content"), 0644))

	addCmd = exec.Command("git", "add", ".")
	addCmd.Dir = sourceDir
	require.NoError(t, addCmd.Run())

	commitCmd = exec.Command("git", "commit", "-m", "Add new file")
	commitCmd.Dir = sourceDir
	require.NoError(t, commitCmd.Run())

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

	branchCmd := exec.Command("git", "branch", "-M", "main")
	branchCmd.Dir = sourceDir
	require.NoError(t, branchCmd.Run())

	branchCmd = exec.Command("git", "checkout", "-b", "feature/test")
	branchCmd.Dir = sourceDir
	require.NoError(t, branchCmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new_file.txt"), []byte("test content"), 0644))

	addCmd = exec.Command("git", "add", ".")
	addCmd.Dir = sourceDir
	require.NoError(t, addCmd.Run())

	commitCmd = exec.Command("git", "commit", "-m", "Add new file")
	commitCmd.Dir = sourceDir
	require.NoError(t, commitCmd.Run())

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

	// Set git config in the repo
	configCmd = exec.Command("git", "config", "user.name", "Test")
	configCmd.Dir = targetDir
	require.NoError(t, configCmd.Run())

	configCmd = exec.Command("git", "config", "user.email", "test@test.com")
	configCmd.Dir = targetDir
	require.NoError(t, configCmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "README.md"), []byte("# Target"), 0644))

	addCmd = exec.Command("git", "add", ".")
	addCmd.Dir = targetDir
	require.NoError(t, addCmd.Run())

	commitCmd = exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = targetDir
	require.NoError(t, commitCmd.Run())

	branchCmd = exec.Command("git", "branch", "-M", "main")
	branchCmd.Dir = targetDir
	require.NoError(t, branchCmd.Run())

	err = ApplyPatch(targetDir, patches[0])
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(targetDir, "new_file.txt"))

	statusCmd := exec.Command("git", "diff", "--name-only")
	statusCmd.Dir = targetDir
	statusOutput, err := statusCmd.CombinedOutput()
	require.NoError(t, err)
	assert.Empty(t, string(statusOutput), "No staged changes should exist after ApplyPatch (user needs to staging/commit)")
}

func TestCreatePatchAndApplyIntegration(t *testing.T) {
	tempDir := t.TempDir()

	sourceDir := filepath.Join(tempDir, "source-repo")
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

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "README.md"), []byte("# Source"), 0644))

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = sourceDir
	require.NoError(t, addCmd.Run())

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = sourceDir
	require.NoError(t, commitCmd.Run())

	branchCmd := exec.Command("git", "branch", "-M", "main")
	branchCmd.Dir = sourceDir
	require.NoError(t, branchCmd.Run())

	branchCmd = exec.Command("git", "checkout", "-b", "feature/test")
	branchCmd.Dir = sourceDir
	require.NoError(t, branchCmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file1.txt"), []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file2.txt"), []byte("content2"), 0644))

	addCmd = exec.Command("git", "add", ".")
	addCmd.Dir = sourceDir
	require.NoError(t, addCmd.Run())

	commitCmd = exec.Command("git", "commit", "-m", "Add two files")
	commitCmd.Dir = sourceDir
	require.NoError(t, commitCmd.Run())

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

	// Set git config in the repo
	configCmd = exec.Command("git", "config", "user.name", "Test")
	configCmd.Dir = targetDir
	require.NoError(t, configCmd.Run())

	configCmd = exec.Command("git", "config", "user.email", "test@test.com")
	configCmd.Dir = targetDir
	require.NoError(t, configCmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(targetDir, "README.md"), []byte("# Target"), 0644))

	addCmd = exec.Command("git", "add", ".")
	addCmd.Dir = targetDir
	require.NoError(t, addCmd.Run())

	commitCmd = exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = targetDir
	require.NoError(t, commitCmd.Run())

	branchCmd = exec.Command("git", "branch", "-M", "main")
	branchCmd.Dir = targetDir
	require.NoError(t, branchCmd.Run())

	err = ApplyPatch(targetDir, patches[0])
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(targetDir, "file1.txt"))
	assert.FileExists(t, filepath.Join(targetDir, "file2.txt"))

	statusCmd := exec.Command("git", "diff", "--name-only")
	statusCmd.Dir = targetDir
	statusOutput, err := statusCmd.CombinedOutput()
	require.NoError(t, err)
	assert.Empty(t, string(statusOutput), "No staged changes should exist after ApplyPatch (user needs to staging/commit)")
}
