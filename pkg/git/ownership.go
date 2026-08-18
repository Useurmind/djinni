package git

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/useurmind/djinni/pkg/config"
	"github.com/useurmind/djinni/pkg/log"
)

func RestoreOwnership(paths []string) error {
	for _, path := range paths {
		stat, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				log.Error(fmt.Sprintf("Path does not exist: %s", path))
				continue
			}
			log.Error(fmt.Sprintf("Failed to stat path %s: %v", path, err))
			continue
		}
		if stat.IsDir() {
			log.Error(fmt.Sprintf("Path is a directory, expected a file: %s", path))
			continue
		}

		log.Info(fmt.Sprintf("Restoring ownership for: %s", path))

		// the root id in the container is the user id of the user running the container, so we need to chown to 0
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

func GetMountPaths(mounts []config.Mount) []string {
	var paths []string
	for _, m := range mounts {
		if m.Source != "" {
			paths = append(paths, m.Source)
		}
	}
	return paths
}
