package container

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/useurmind/djinni/pkg/config"
	"github.com/useurmind/djinni/pkg/log"
)

// Mount represents a volume mount from host to container
type Mount = config.Mount

// RestoreOwnership restores file ownership after container execution
// It changes ownership to 0:0 (root) for files that were modified inside the container
func RestoreOwnership(paths []string) error {
	for _, path := range paths {
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				log.Error(fmt.Sprintf("Path does not exist: %s", path))
				continue
			}
			log.Error(fmt.Sprintf("Failed to stat path %s: %v", path, err))
			continue
		}

		log.Info(fmt.Sprintf("Restoring ownership for: %s", path))

		cmd := exec.Command("podman", "unshare", "chown", "-R", "0:0", path)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(fmt.Sprintf("Failed to restore ownership for %s: %v, output: %s", path, err, string(output)))
		} else {
			log.Success(fmt.Sprintf("Restored ownership for: %s", path))
		}
	}

	return nil
}

// GetMountPaths returns the source paths from mounts
func GetMountPaths(mounts []config.Mount) []string {
	var paths []string
	for _, m := range mounts {
		if m.Source != "" {
			paths = append(paths, m.Source)
		}
	}
	return paths
}
