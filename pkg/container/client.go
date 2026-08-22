package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/useurmind/djinni/pkg/config"
	"github.com/useurmind/djinni/pkg/log"
)

const AgentImageNameFormat = "%s-%s:latest"

type Client struct {
	Type   string
	Binary string
}

func NewClient() (*Client, error) {
	binary := "podman"
	_, err := exec.LookPath(binary)
	if err != nil {
		return nil, fmt.Errorf("container runtime %s not found", binary)
	}
	log.Info(fmt.Sprintf("Detected container runtime: %s", binary))
	return &Client{
		Type:   binary,
		Binary: binary,
	}, nil
}

func (c *Client) RunContainer(image string, cmd []string, name string, mounts []config.Mount, commands *ContainerCommands) (int, error) {
	if commands == nil {
		commands = &ContainerCommands{}
	}

	return c.runContainer(image, cmd, name, mounts, commands)
}

func (c *Client) PrepareWritablePaths(repoName, agentName string, writablePaths []config.WritablePath, image string) error {
	for _, wp := range writablePaths {
		if err := CreateOverlayStructure(repoName, agentName, wp.Name); err != nil {
			return fmt.Errorf("failed to create overlay structure for %s: %w", wp.Name, err)
		}

		lowerDir := GetLowerDir(repoName, agentName, wp.Name)
		if err := CopyImageFolderToLower(c, image, wp.Destination, lowerDir); err != nil {
			return fmt.Errorf("failed to copy image folder to lower for %s: %w", wp.Name, err)
		}
	}

	return nil
}

func (c *Client) SetupOverlayMount(repoName, agentName, taskName, writablePathName, destination string) (string, error) {
	upperDir := GetUpperDir(repoName, agentName, writablePathName, taskName)
	workDir := GetWorkDir(repoName, agentName, writablePathName, taskName)

	if err := os.MkdirAll(upperDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upper directory %s: %w", upperDir, err)
	}

	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create work directory %s: %w", workDir, err)
	}

	tempMount := filepath.Join(workDir, "mnt")
	if err := os.MkdirAll(tempMount, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp mount directory %s: %w", tempMount, err)
	}

	return tempMount, nil
}

func (c *Client) BuildContainer(repoName, agentName string, containerfile string) (int, error) {
	if _, err := os.Stat(containerfile); os.IsNotExist(err) {
		return 1, fmt.Errorf("containerfile '%s' does not exist", containerfile)
	}

	args := []string{"build", "-f", containerfile, "-t", fmt.Sprintf(AgentImageNameFormat, repoName, agentName), "."}

	log.Info(fmt.Sprintf("Building container: %s", agentName))
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
	log.Success(fmt.Sprintf("Built image: "+AgentImageNameFormat, repoName, agentName))
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

func (c *Client) runContainer(image string, cmd []string, name string, mounts []config.Mount, commands *ContainerCommands) (int, error) {
	entrypoint := c.generateEntrypoint(cmd, commands)
	args := []string{"run", "--rm", "-it", "--network", "bridge", "--name", name}

	if commands == nil || !commands.ForceReadOnlyRootOff {
		args = append(args, "--read-only")
	}

	for _, tmpfs := range commands.TmpfsMounts {
		var tmpfsArg string
		if tmpfs.Size != "" {
			tmpfsArg = fmt.Sprintf("%s:mode=1777,size=%s", tmpfs.Destination, tmpfs.Size)
		} else {
			tmpfsArg = fmt.Sprintf("%s:mode=1777", tmpfs.Destination)
		}
		args = append(args, "--tmpfs", tmpfsArg)
	}

	for _, m := range mounts {
		var mountStr string
		if m.ReadOnly {
			mountStr = fmt.Sprintf("%s:%s:Z,ro,U", m.Source, m.Destination)
		} else {
			mountStr = fmt.Sprintf("%s:%s:Z,U", m.Source, m.Destination)
		}
		args = append(args, "-v", mountStr)
	}

	if commands.TempMount != nil {
		args = append(args, "-v", fmt.Sprintf("%s:%s:Z,ro,U", commands.TempMount.Source, commands.TempMount.Destination))
	}

	args = append(args, "--entrypoint", "/bin/bash")
	args = append(args, image)
	args = append(args, "-i", "-c", entrypoint)

	log.Info(fmt.Sprintf("Running: %s %s", c.Binary, strings.Join(args, " ")))

	return c.runCommand(args)
}

func (c *Client) generateEntrypoint(harnessCmd []string, commands *ContainerCommands) string {
	var builder strings.Builder

	builder.WriteString("set -e\n")

	if commands.TempMount != nil && len(commands.FilesToCopy) > 0 {
		builder.WriteString("echo 'Copying files from temp mount to destinations...'\n")
		fmt.Fprintf(&builder, "TEMP_MOUNT_DIR=\"%s\"\n", commands.TempMount.Destination)
		for _, file := range commands.FilesToCopy {
			fmt.Fprintf(&builder, "mkdir -p \"%s\"\n", filepath.Dir(file.Destination))
			fmt.Fprintf(&builder, "cp \"%s/%s\" \"%s\"\n", commands.TempMount.Destination, file.Name(), file.Destination)
		}
		builder.WriteString("echo 'File copy complete'\n")
	}

	for _, preCmd := range commands.PreCommands {
		builder.WriteString(preCmd)
		builder.WriteString("\n")
	}

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

func (c *Client) ExecInContainer(containerName string, cmd []string) (int, error) {
	args := []string{"exec", "-it"}

	if len(cmd) == 0 {
		cmd = []string{"bash"}
	}

	args = append(args, containerName)
	args = append(args, cmd...)

	log.Info(fmt.Sprintf("Running: %s %s", c.Binary, strings.Join(args, " ")))

	return c.runCommand(args)
}
