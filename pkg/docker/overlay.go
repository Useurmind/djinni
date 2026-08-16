package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/useurmind/djinni/pkg/log"
)

const djinniBaseDir = "/tmp/djinni"

func GetWritablePathDir(repoName, agentName, writablePathName string) string {
	return filepath.Join(djinniBaseDir, repoName, agentName, "writablePaths", writablePathName)
}

func GetLowerDir(repoName, agentName, writablePathName string) string {
	return filepath.Join(GetWritablePathDir(repoName, agentName, writablePathName), "lower")
}

func GetUpperDir(repoName, agentName, writablePathName, taskName string) string {
	return filepath.Join(GetWritablePathDir(repoName, agentName, writablePathName), "upper", taskName)
}

func GetWorkDir(repoName, agentName, writablePathName, taskName string) string {
	return filepath.Join(GetWritablePathDir(repoName, agentName, writablePathName), "work", taskName)
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

	log.Success(fmt.Sprintf("Created overlayfs structure for %s", writablePathName))
	return nil
}

func CopyImageFolderToLower(client *Client, image, imageSourcePath, lowerDir string) error {
	log.Info(fmt.Sprintf("Copying %s from image %s to lower directory %s", imageSourcePath, image, lowerDir))

	err := os.MkdirAll(lowerDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create lower directory %s: %w", lowerDir, err)
	}

	cmd := exec.Command(client.Binary, "create", "--name", "djinni-temp-copy", image)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create temp container: %w, output: %s", err, strings.TrimSpace(string(output)))
	}
	tempContainerID := strings.TrimSpace(string(output))

	defer func() {
		_ = exec.Command(client.Binary, "rm", "-f", tempContainerID).Run()
	}()

	cmd = exec.Command(client.Binary, "cp", tempContainerID+":"+imageSourcePath, lowerDir)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy %s from container: %w, output: %s", imageSourcePath, err, strings.TrimSpace(string(output)))
	}

	log.Success(fmt.Sprintf("Copied %s to lower directory", imageSourcePath))
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
