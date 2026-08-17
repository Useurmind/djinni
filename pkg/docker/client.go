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

const AgentImageNameFormat = "%s-%s:latest"

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

	return c.runContainer(image, cmd, name, mounts, commands)
}

func (c *Client) PrepareWritablePaths(repoName, agentName string, writablePaths []WritablePath, image string) error {
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
	lowerDir := GetLowerDir(repoName, agentName, writablePathName)
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

	err := exec.Command("mount", "-t", "overlay", "overlay",
		"-o", fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir),
		tempMount).Run()
	if err != nil {
		return "", fmt.Errorf("failed to mount overlayfs: %w", err)
	}

	return tempMount, nil
}

func (c *Client) CleanupOverlayMount(repoName, agentName, taskName, writablePathName string) error {
	workDir := GetWorkDir(repoName, agentName, writablePathName, taskName)

	tempMount := filepath.Join(workDir, "mnt")

	cmd := exec.Command("umount", tempMount)
	if err := cmd.Run(); err != nil {
		log.Error(fmt.Sprintf("Failed to unmount %s: %v", tempMount, err))
	}

	return CleanupOverlay(repoName, agentName, writablePathName, taskName)
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
			mountStr = fmt.Sprintf("%s:%s:Z,RO,U", m.Source, m.Destination)
		} else {
			mountStr = fmt.Sprintf("%s:%s:Z,U", m.Source, m.Destination)
		}
		args = append(args, "-v", mountStr)
	}

	args = append(args, "--entrypoint", "/bin/bash")
	args = append(args, image)
	args = append(args, "-i", "-c", entrypoint)

	log.Info(fmt.Sprintf("Running: %s %s", c.Binary, strings.Join(args, " ")))

	if len(commands.FilesToCopy) > 0 {
		copyErrChan := CopyFilesAsync(name, commands.FilesToCopy, c)
		cmd := exec.Command(c.Binary, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			exitErr, ok := err.(*exec.ExitError)
			if ok {
				<-copyErrChan
				return exitErr.ExitCode(), nil
			}
			<-copyErrChan
			return 1, fmt.Errorf("%s %s: %w", c.Binary, strings.Join(args, " "), err)
		}
		err = <-copyErrChan
		if err != nil {
			return 1, err
		}
		return 0, nil
	}

	return c.runCommand(args)
}

func (c *Client) generateEntrypoint(harnessCmd []string, commands *ContainerCommands) string {
	var builder strings.Builder

	builder.WriteString("set -e\n")

	if len(commands.FilesToCopy) > 0 {
		builder.WriteString("echo 'Waiting for files to be copied...'\n")
		builder.WriteString("while [ ! -e \"" + DefaultAgentMarker + "\" ]; do\n")
		builder.WriteString("  sleep 0.5\n")
		builder.WriteString("done\n")
		builder.WriteString("echo 'Files copied successfully'\n")
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
