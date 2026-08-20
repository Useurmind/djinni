package docker

import (
	"fmt"
	"os/exec"
	"path"
	"time"

	"github.com/useurmind/djinni/pkg/log"
)

const (
	DefaultAgentHome   = "/home/agent"
	DefaultAgentMarker = DefaultAgentHome + "/.copydone"
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
				errChan <- fmt.Errorf("container not ready after 30 seconds: %s", string(output))
				return
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
		return fmt.Errorf("failed to create directory %s: %w, output: %s", destDir, err, string(output))
	}

	cmd = exec.Command(client.Binary, "cp", source, containerID+":"+destination)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy %s to %s in container %s: %w, output: %s", source, destination, containerID, err, string(output))
	}

	return nil
}
func createMarkerFile(containerID string, client *Client) error {
	cmd := exec.Command(client.Binary, "exec", containerID, "mkdir", "-p", DefaultAgentHome, "&&", "touch", DefaultAgentMarker)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create marker file: %s", string(output)))
		return err
	}
	return nil
}
