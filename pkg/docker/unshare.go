package docker

import (
	"fmt"
	"github.com/useurmind/djinni/pkg/log"
	"os"
	"os/exec"
)

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

func MountOverlayFsWithMountPoint(repoName, agentName, writablePathName, taskName, mountPoint string) error {
	upperDir := GetUpperDir(repoName, agentName, writablePathName, taskName)
	workDir := GetWorkDir(repoName, agentName, writablePathName, taskName)

	if err := os.MkdirAll(upperDir, 0755); err != nil {
		return fmt.Errorf("failed to create upper directory %s: %w", upperDir, err)
	}

	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create work directory %s: %w", workDir, err)
	}

	return MountOverlay(repoName, agentName, writablePathName, taskName, mountPoint)
}
