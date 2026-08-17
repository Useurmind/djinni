package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/useurmind/djinni/pkg/config"
	"github.com/useurmind/djinni/pkg/log"
)

func RunContainerWithUnshare(c *Client, repoName, agentName, taskName, image string, cmd []string, mounts []config.Mount, commands *ContainerCommands) (int, error) {
	if commands == nil {
		commands = &ContainerCommands{}
	}

	writablePaths := commands.WritablePaths

	if len(writablePaths) > 0 {
		tempMounts := make([]string, 0, len(writablePaths))
		for _, wp := range writablePaths {
			workDir := GetWorkDir(repoName, agentName, wp.Name, taskName)
			tempMount := filepath.Join(workDir, "mnt")
			tempMounts = append(tempMounts, tempMount)

			log.Info(fmt.Sprintf("Mounting overlayfs for %s at %s", wp.Name, tempMount))
			if err := MountOverlay(repoName, agentName, wp.Name, taskName, tempMount); err != nil {
				return 1, fmt.Errorf("failed to mount overlay for %s: %w", wp.Name, err)
			}
		}

		mountArgs := make([]string, 0, len(tempMounts)*2)
		for i, wp := range writablePaths {
			mountArgs = append(mountArgs, "-v", fmt.Sprintf("%s:%s", tempMounts[i], wp.Destination))
		}

		args := append([]string{"run", "--rm", "-it", "--network", "bridge", "--name", agentName}, mountArgs...)
		args = append(args, "--entrypoint", "/bin/bash")
		args = append(args, image)
		args = append(args, "-i", "-c", strings.Join(cmd, " "))

		log.Info(fmt.Sprintf("Running: %s %s", c.Binary, strings.Join(args, " ")))

		runCmd := exec.Command(c.Binary, args...)
		runCmd.Stdin = os.Stdin
		runCmd.Stdout = os.Stdout
		runCmd.Stderr = os.Stderr

		err := runCmd.Run()
		if err != nil {
			exitErr, ok := err.(*exec.ExitError)
			if ok {
				for _, m := range tempMounts {
					if err := UnmountOverlay(m); err != nil {
						log.Error(fmt.Sprintf("Failed to unmount overlay at %s: %v", m, err))
					}
				}
				return exitErr.ExitCode(), nil
			}
			for _, m := range tempMounts {
				if err := UnmountOverlay(m); err != nil {
					log.Error(fmt.Sprintf("Failed to unmount overlay at %s: %v", m, err))
				}
			}
			return 1, fmt.Errorf("%s %s: %w", c.Binary, strings.Join(mountArgs, " "), err)
		}

		for _, m := range tempMounts {
			if err := UnmountOverlay(m); err != nil {
				log.Error(fmt.Sprintf("Failed to unmount overlay at %s: %v", m, err))
			}
		}

		return 0, nil
	}

	return c.RunContainer(image, cmd, agentName, mounts, commands)
}

func MountOverlay(repoName, agentName, writablePathName, taskName, tempMount string) error {
	lowerDir := GetLowerDir(repoName, agentName, writablePathName)
	upperDir := GetUpperDir(repoName, agentName, writablePathName, taskName)
	workDir := GetWorkDir(repoName, agentName, writablePathName, taskName)

	if err := os.MkdirAll(upperDir, 0755); err != nil {
		return fmt.Errorf("failed to create upper directory %s: %w", upperDir, err)
	}

	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create work directory %s: %w", workDir, err)
	}

	if err := os.MkdirAll(tempMount, 0755); err != nil {
		return fmt.Errorf("failed to create temp mount directory %s: %w", tempMount, err)
	}

	cmd := exec.Command("podman", "unshare", "mount", "-t", "overlay", "overlay",
		"-o", fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir),
		tempMount)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to mount overlayfs: %w, output: %s", err, string(output))
	}

	log.Info(fmt.Sprintf("Mounted overlayfs for %s at %s", writablePathName, tempMount))
	return nil
}

func UnmountOverlay(mountPath string) error {
	cmd := exec.Command("podman", "unshare", "umount", mountPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to unmount overlayfs at %s: %w, output: %s", mountPath, err, string(output))
	}

	log.Info(fmt.Sprintf("Unmounted overlayfs at %s", mountPath))
	return nil
}
