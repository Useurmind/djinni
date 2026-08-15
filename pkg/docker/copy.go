package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/useurmind/djinni/pkg/log"
)

func CopyFilesAsync(containerID string, files []FilesToCopy, client *Client) chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)

		log.Info("Starting async file copy to container")

		if len(files) == 0 {
			log.Info("No files to copy, creating marker file immediately")
			if err := createMarkerFile(containerID, client); err != nil {
				errChan <- fmt.Errorf("failed to create marker file: %w", err)
				return
			}
			errChan <- nil
			return
		}

		log.Info("Waiting for container to be ready...")

		for i := 0; i < 60; i++ {
			output, err := exec.Command(client.Binary, "exec", containerID, "echo", "ready").CombinedOutput()
			if err == nil {
				log.Success("Container is ready")
				break
			}
			if i == 59 {
				log.Error(fmt.Sprintf("Container not ready: %s", string(output)))
			}
			time.Sleep(500 * time.Millisecond)
		}

		for _, file := range files {
			if err := copyFileToContainer(containerID, file.Source, file.Destination, client); err != nil {
				errChan <- fmt.Errorf("failed to copy %s to container: %w", file.Source, err)
				return
			}
			log.Success(fmt.Sprintf("Copied %s to %s in container", file.Source, file.Destination))
		}

		if err := createMarkerFile(containerID, client); err != nil {
			errChan <- fmt.Errorf("failed to create marker file: %w", err)
			return
		}

		log.Success("File copy complete, marker file created")
		errChan <- nil
	}()

	return errChan
}

func copyFileToContainer(containerID, source, destination string, client *Client) error {
	destDir := path.Dir(destination)

	cmd := exec.Command(client.Binary, "exec", containerID, "mkdir", "-p", destDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create directory: %s", string(output)))
		return err
	}

	hostSourcePath, err := getHostPathForSource(source)
	if err != nil {
		return err
	}

	cmd = exec.Command(client.Binary, "cp", hostSourcePath, containerID+":"+destination)
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Error(fmt.Sprintf("Failed to copy file: %s", string(output)))
		return fmt.Errorf("failed to execute %s cp %s %s:%s: %w", client.Binary, hostSourcePath, containerID, destination, err)
	}

	return nil
}

func getHostPathForSource(source string) (string, error) {
	if strings.HasPrefix(source, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return homeDir + source[1:], nil
	}
	return source, nil
}

func createMarkerFile(containerID string, client *Client) error {
	cmd := exec.Command(client.Binary, "exec", containerID, "mkdir", "-p", "/home/agent", "&&", "touch", "/home/agent/.copydone")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create marker file: %s", string(output)))
		return err
	}
	return nil
}
