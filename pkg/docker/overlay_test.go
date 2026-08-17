package docker

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWritablePathDir(t *testing.T) {
	repoName := "test-repo"
	agentName := "test-agent"
	writablePathName := "home"

	expected := "/tmp/djinni/test-repo/test-agent/writablePaths/home"
	result := GetWritablePathDir(repoName, agentName, writablePathName)
	assert.Equal(t, expected, result)
}

func TestGetLowerDir(t *testing.T) {
	repoName := "test-repo"
	agentName := "test-agent"
	writablePathName := "home"

	expected := "/tmp/djinni/test-repo/test-agent/writablePaths/home/lower"
	result := GetLowerDir(repoName, agentName, writablePathName)
	assert.Equal(t, expected, result)
}

func TestGetUpperDir(t *testing.T) {
	repoName := "test-repo"
	agentName := "test-agent"
	writablePathName := "home"
	taskName := "task123"

	expected := "/tmp/djinni/test-repo/test-agent/writablePaths/home/upper/task123"
	result := GetUpperDir(repoName, agentName, writablePathName, taskName)
	assert.Equal(t, expected, result)
}

func TestGetWorkDir(t *testing.T) {
	repoName := "test-repo"
	agentName := "test-agent"
	writablePathName := "home"
	taskName := "task123"

	expected := "/tmp/djinni/test-repo/test-agent/writablePaths/home/work/task123"
	result := GetWorkDir(repoName, agentName, writablePathName, taskName)
	assert.Equal(t, expected, result)
}

func TestCreateOverlayStructure(t *testing.T) {
	repoName := "test-repo"
	agentName := "test-agent"
	writablePathName := "home"

	lowerDir := GetLowerDir(repoName, agentName, writablePathName)

	err := CreateOverlayStructure(repoName, agentName, writablePathName)
	require.NoError(t, err)

	defer os.RemoveAll(DefaultBaseDir + "-test")

	assert.DirExists(t, lowerDir)
}

func TestCleanupOverlay(t *testing.T) {
	repoName := "test-repo"
	agentName := "test-agent"
	writablePathName := "home"
	taskName := "task123"

	upperDir := GetUpperDir(repoName, agentName, writablePathName, taskName)
	workDir := GetWorkDir(repoName, agentName, writablePathName, taskName)

	err := os.MkdirAll(upperDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(workDir, 0755)
	require.NoError(t, err)

	err = CleanupOverlay(repoName, agentName, writablePathName, taskName)
	require.NoError(t, err)

	assert.NoDirExists(t, upperDir)
	assert.NoDirExists(t, workDir)
}
