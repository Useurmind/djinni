package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/useurmind/djinni/pkg/config"
	"github.com/useurmind/djinni/pkg/log"
)

type Client struct {
	Type   string
	Binary string
}

func NewClient() (*Client, error) {
	for _, binary := range []string{"podman", "docker"} {
		if _, err := exec.LookPath(binary); err == nil {
			log.Info(fmt.Sprintf("Detected container runtime: %s", binary))
			return &Client{
				Type:   binary,
				Binary: binary,
			}, nil
		}
	}
	return nil, fmt.Errorf("no container runtime found (podman or docker)")
}

func (c *Client) RunContainer(image string, cmd []string, name string, mounts []config.Mount) (int, error) {
	args := []string{"run", "--rm", "-it", "--network", "bridge", "--name", name}

	for _, m := range mounts {
		var mountStr string
		if m.ReadOnly {
			mountStr = fmt.Sprintf("%s:%s:ro", m.Source, m.Destination)
		} else {
			mountStr = fmt.Sprintf("%s:%s", m.Source, m.Destination)
		}
		args = append(args, "-v", mountStr)
	}

	args = append(args, image)
	args = append(args, cmd...)

	log.Info(fmt.Sprintf("Running: %s %s", c.Binary, strings.Join(args, " ")))
	return c.runCommand(args)
}

func (c *Client) runCommand(args []string) (int, error) {
	cmd := exec.Command(c.Binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			return exitErr.ExitCode(), nil
		}
		return 1, fmt.Errorf("%s %s: %w", c.Binary, strings.Join(args, " "), err)
	}

	return 0, nil
}
