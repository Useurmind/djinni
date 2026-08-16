package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/useurmind/djinni/pkg/log"
)

func RunInMountNamespace(cmd []string, mounts []WritablePath, destination string) error {
	log.Info("Running in mount namespace with unshare -rm")

	args := []string{"-rm", "bash", "-c"}

	overlayCmd := &strings.Builder{}
	overlayCmd.WriteString("set -e\n")

	for _, wp := range mounts {
		upperDir := GetUpperDir("repo", "agent", wp.Name, "task")
		workDir := GetWorkDir("repo", "agent", wp.Name, "task")

		fmt.Fprintf(overlayCmd, "mkdir -p %s %s\n", upperDir, workDir)

		lowerDir := GetLowerDir("repo", "agent", wp.Name)
		fmt.Fprintf(overlayCmd, "mount -t overlay overlay -o lowerdir=%s,upperdir=%s,workdir=%s %s\n",
			lowerDir, upperDir, workDir, destination)
	}

	overlayCmd.WriteString("exec ")
	for i, arg := range cmd {
		if i > 0 {
			overlayCmd.WriteString(" ")
		}
		overlayCmd.WriteString(arg)
	}

	args = append(args, overlayCmd.String())

	log.Info(fmt.Sprintf("Running: unshare %s", strings.Join(args[1:], " ")))

	unshareCmd := exec.Command("unshare", args...)
	unshareCmd.Stdin = os.Stdin
	unshareCmd.Stdout = os.Stdout
	unshareCmd.Stderr = os.Stderr

	err := unshareCmd.Run()
	if err != nil {
		return fmt.Errorf("unshare failed: %w", err)
	}

	return nil
}
