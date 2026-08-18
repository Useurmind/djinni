package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/useurmind/djinni/pkg/log"
)

const TempContainerName = "djinni-temp-copy"

const DefaultBaseDir = "/tmp/djinni"

const (
	WritablePathsSubdir = "writablePaths"
	LowerSubdir         = "lower"
	UpperSubdir         = "upper"
	WorkSubdir          = "work"
)

func GetWritablePathDir(repoName, agentName, writablePathName string) string {
	return filepath.Join(DefaultBaseDir, repoName, agentName, WritablePathsSubdir, writablePathName)
}

func GetLowerDir(repoName, agentName, writablePathName string) string {
	return filepath.Join(GetWritablePathDir(repoName, agentName, writablePathName), LowerSubdir)
}

func GetUpperDir(repoName, agentName, writablePathName, taskName string) string {
	return filepath.Join(GetWritablePathDir(repoName, agentName, writablePathName), UpperSubdir, taskName)
}

func GetWorkDir(repoName, agentName, writablePathName, taskName string) string {
	return filepath.Join(GetWritablePathDir(repoName, agentName, writablePathName), WorkSubdir, taskName)
}

func CreateOverlayStructure(repoName, agentName, writablePathName string) error {
	baseDir := GetWritablePathDir(repoName, agentName, writablePathName)
	lowerDir := GetLowerDir(repoName, agentName, writablePathName)

	paths := []string{baseDir, lowerDir}
	for _, path := range paths {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}

	log.Info(fmt.Sprintf("Created overlayfs structure for %s: %s", writablePathName, baseDir))
	return nil
}

func CopyImageFolderToLower(client *Client, image, imageSourcePath, lowerDir string) error {
	log.Info(fmt.Sprintf("Copying %s from image %s to lower directory %s", imageSourcePath, image, lowerDir))

	cmd := exec.Command(client.Binary, "create", "--name", TempContainerName, image)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create temp container: %w, output: %s", err, strings.TrimSpace(string(output)))
	}
	tempContainerID := strings.TrimSpace(string(output))

	defer func() {
		_ = exec.Command(client.Binary, "rm", "-f", tempContainerID).Run()
	}()

	lowerParentDir := filepath.Dir(lowerDir)
	sourceBasename := filepath.Base(imageSourcePath)
	tempCopyDir := filepath.Join(lowerParentDir, ".tmp_copy")
	if err := os.MkdirAll(tempCopyDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp copy directory %s: %w", tempCopyDir, err)
	}
	defer os.RemoveAll(tempCopyDir)

	cmd = exec.Command(client.Binary, "cp", tempContainerID+":"+imageSourcePath, tempCopyDir)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy %s from container: %w, output: %s", imageSourcePath, err, strings.TrimSpace(string(output)))
	}

	sourceCopyPath := filepath.Join(tempCopyDir, sourceBasename)
	entries, err := os.ReadDir(sourceCopyPath)
	if err != nil {
		return fmt.Errorf("failed to read temp copy directory: %w", err)
	}

	for _, entry := range entries {
		src := filepath.Join(sourceCopyPath, entry.Name())
		dst := filepath.Join(lowerDir, entry.Name())
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("failed to move %s to lower directory: %w", entry.Name(), err)
		}
	}

	log.Info(fmt.Sprintf("Copied %s to lower directory", imageSourcePath))
	return nil
}

func CleanupOverlay(repoName, agentName, writablePathName, taskName string) error {
	upperDir := GetUpperDir(repoName, agentName, writablePathName, taskName)
	workDir := GetWorkDir(repoName, agentName, writablePathName, taskName)

	for _, dir := range []string{upperDir, workDir} {
		if err := os.RemoveAll(dir); err != nil {
			log.Error(fmt.Sprintf("Failed to cleanup %s: %v", dir, err))
		}
	}

	log.Info(fmt.Sprintf("Cleaned up overlayfs directories for %s", writablePathName))
	return nil
}
