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

func (c *Client) RunContainer(image string, cmd []string, name string, mounts []config.Mount, commands *ContainerCommands) (int, error) {
	if commands == nil {
		commands = &ContainerCommands{}
	}

	if len(commands.PreCommands) == 0 && len(commands.PostCommands) == 0 {
		return c.runContainerDirect(image, cmd, name, mounts)
	}

	return c.runContainerWithCommands(image, cmd, name, mounts, commands)
}

func (c *Client) runContainerDirect(image string, cmd []string, name string, mounts []config.Mount) (int, error) {
	args := []string{"run", "--rm", "-it", "--network", "bridge", "--name", name}

	for _, m := range mounts {
		var mountStr string
		if m.ReadOnly {
			mountStr = fmt.Sprintf("%s:%s:zro", m.Source, m.Destination)
		} else {
			mountStr = fmt.Sprintf("%s:%s:z", m.Source, m.Destination)
		}
		args = append(args, "-v", mountStr)
	}

	args = append(args, image)
	args = append(args, cmd...)

	log.Info(fmt.Sprintf("Running: %s %s", c.Binary, strings.Join(args, " ")))
	return c.runCommand(args)
}

func (c *Client) BuildContainer(name string, containerfile string) (int, error) {
	if _, err := os.Stat(containerfile); os.IsNotExist(err) {
		return 1, fmt.Errorf("containerfile '%s' does not exist", containerfile)
	}

	args := []string{"build", "-f", containerfile, "-t", fmt.Sprintf("djinni-%s:latest", name), "."}

	log.Info(fmt.Sprintf("Building container: %s", name))
	log.Info(fmt.Sprintf("Building from: %s", containerfile))
	log.Info(fmt.Sprintf("Running: %s %s", c.Binary, strings.Join(args, " ")))

	cmd := exec.Command(c.Binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			return exitErr.ExitCode(), nil
		}
		return 1, fmt.Errorf("build failed: %w", err)
	}
	log.Success(fmt.Sprintf("Built image: djinni-%s:latest", name))
	return 0, nil
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

func (c *Client) runContainerWithCommands(image string, cmd []string, name string, mounts []config.Mount, commands *ContainerCommands) (int, error) {
	args := []string{"run", "--rm", "-it", "--network", "bridge", "--name", name}

	for _, m := range mounts {
		var mountStr string
		if m.ReadOnly {
			mountStr = fmt.Sprintf("%s:%s:zro", m.Source, m.Destination)
		} else {
			mountStr = fmt.Sprintf("%s:%s:z", m.Source, m.Destination)
		}
		args = append(args, "-v", mountStr)
	}

	entrypoint := c.generateEntrypoint(cmd, commands)
	args = append(args, "--entrypoint", "/bin/sh")
	args = append(args, image)
	args = append(args, "-c", entrypoint)

	log.Info(fmt.Sprintf("Running: %s %s", c.Binary, strings.Join(args, " ")))
	return c.runCommand(args)
}

func (c *Client) generateEntrypoint(harnessCmd []string, commands *ContainerCommands) string {
	var builder strings.Builder

	builder.WriteString("set -e\n")

	for _, preCmd := range commands.PreCommands {
		builder.WriteString(preCmd)
		builder.WriteString("\n")
	}

	builder.WriteString("exec ")
	for i, arg := range harnessCmd {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(arg)
	}

	builder.WriteString("\n")

	for _, postCmd := range commands.PostCommands {
		builder.WriteString(postCmd)
		builder.WriteString("\n")
	}

	return builder.String()
}
