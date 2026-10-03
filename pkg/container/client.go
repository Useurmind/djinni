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
	Type    string
	Binary  string
	BaseDir string
}

func NewClient(baseDir string) (*Client, error) {
	if baseDir == "" {
		baseDir = DefaultBaseDir
	}
	binary := "podman"
	_, err := exec.LookPath(binary)
	if err != nil {
		return nil, fmt.Errorf("container runtime %s not found", binary)
	}
	log.Info(fmt.Sprintf("Detected container runtime: %s", binary))
	return &Client{
		Type:    binary,
		Binary:  binary,
		BaseDir: baseDir,
	}, nil
}

func (c *Client) SetupNetwork(agentName string, networkCfg *config.AgentNetworkConfig) (*ProxyContainerInfo, error) {
	if !networkCfg.Internal {
		return nil, nil
	}

	// Cleanup any existing resources before setting up new ones
	if err := c.CleanupExistingProxy(agentName); err != nil {
		log.Error(fmt.Sprintf("Failed to cleanup existing proxy: %v", err))
	}
	if err := c.CleanupExistingNetwork(agentName); err != nil {
		log.Error(fmt.Sprintf("Failed to cleanup existing network: %v", err))
	}

	networkName := GetNetworkName(agentName)

	// Create internal network
	if err := CreateInternalNetwork(c, networkName); err != nil {
		return nil, fmt.Errorf("failed to create internal network: %w", err)
	}

	// Create proxy container info
	proxyInfo := &ProxyContainerInfo{
		Name:        fmt.Sprintf("djinni-proxy-%s", strings.ReplaceAll(agentName, "-", "_")),
		NetworkName: networkName,
		SquidPort:   DefaultSquidPort,
	}

	if networkCfg.Proxy != nil && networkCfg.Proxy.Enabled {
		// Generate squid config
		configPath, err := GenerateSquidConfig(
			true,
			networkCfg.Proxy.AllowList,
			DefaultSquidPort,
			c.BaseDir,
			"repo",
			agentName,
			"task",
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate squid config: %w", err)
		}

		proxyInfo.SquidConfigPath = configPath

		// Start proxy container
		if _, err := StartProxyContainer(
			c,
			configPath,
			networkName,
			proxyInfo.Name,
			DefaultSquidPort,
		); err != nil {
			return nil, fmt.Errorf("failed to start proxy container: %w", err)
		}

		// Connect proxy to internet (bridge network)
		if err := ConnectProxyToInternet(c, proxyInfo.Name); err != nil {
			return nil, fmt.Errorf("failed to connect proxy to internet: %w", err)
		}

		// Use container name as address (Podman DNS resolves container names)
		proxyInfo.SquidAddress = proxyInfo.Name
	}

	return proxyInfo, nil
}

func (c *Client) CleanupNetwork(proxyInfo *ProxyContainerInfo) error {
	if proxyInfo == nil {
		return nil
	}

	// Remove proxy container
	if err := RemoveProxyContainer(c, proxyInfo.Name); err != nil {
		log.Error(fmt.Sprintf("Failed to remove proxy container: %v", err))
	}

	// Remove internal network
	if err := RemoveNetwork(c, proxyInfo.NetworkName); err != nil {
		log.Error(fmt.Sprintf("Failed to remove network: %v", err))
	}

	// Cleanup proxy config files
	if proxyInfo.SquidConfigPath != "" {
		if err := CleanupProxyConfig(c.BaseDir, "repo", "agent", "task"); err != nil {
			log.Error(fmt.Sprintf("Failed to cleanup proxy config: %v", err))
		}
	}

	return nil
}

// CleanupExistingProxy stops and removes any existing proxy container for the agent
func (c *Client) CleanupExistingProxy(agentName string) error {
	proxyContainerName := fmt.Sprintf("djinni-proxy-%s", strings.ReplaceAll(agentName, "-", "_"))

	// Stop the container first if it exists
	if err := StopProxyContainer(c, proxyContainerName); err != nil {
		// Container might not exist, which is fine
		log.Info(fmt.Sprintf("No existing proxy container to stop: %s", proxyContainerName))
	}

	// Remove the container
	if err := RemoveProxyContainer(c, proxyContainerName); err != nil {
		// Container might not exist, which is fine
		log.Info(fmt.Sprintf("No existing proxy container to remove: %s", proxyContainerName))
	}

	return nil
}

// CleanupExistingNetwork removes any existing internal network for the agent
func (c *Client) CleanupExistingNetwork(agentName string) error {
	networkName := GetNetworkName(agentName)

	// Disconnect any containers from the network first
	// Then remove the network
	if err := RemoveNetwork(c, networkName); err != nil {
		// Network might not exist, which is fine
		log.Info(fmt.Sprintf("No existing network to remove: %s", networkName))
	}

	return nil
}

// CleanupExistingWorkspace removes any existing workspace for the task
func (c *Client) CleanupExistingWorkspace(baseDir, agentName, taskName string) error {
	// Workspace directory pattern used by git.CloneToTemp
	workspacePattern := filepath.Join(baseDir, "temp-clones", fmt.Sprintf("%s-%s-*", agentName, taskName))

	// Find and remove any matching workspace directories
	matches, err := filepath.Glob(workspacePattern)
	if err != nil {
		return fmt.Errorf("failed to glob workspace pattern: %w", err)
	}

	for _, match := range matches {
		if err := os.RemoveAll(match); err != nil {
			log.Error(fmt.Sprintf("Failed to remove existing workspace %s: %v", match, err))
		} else {
			log.Info(fmt.Sprintf("Removed existing workspace: %s", match))
		}
	}

	return nil
}

func (c *Client) RunContainer(image string, cmd []string, name string, mounts []config.Mount, commands *ContainerCommands) (int, error) {
	if commands == nil {
		commands = &ContainerCommands{}
	}

	return c.runContainer(image, cmd, name, mounts, commands)
}

func (c *Client) PrepareWritablePaths(repoName, agentName string, writablePaths []config.WritablePath, image string) error {
	for _, wp := range writablePaths {
		if err := CreateOverlayStructure(c.BaseDir, repoName, agentName, wp.Name); err != nil {
			return fmt.Errorf("failed to create overlay structure for %s: %w", wp.Name, err)
		}

		lowerDir := GetLowerDir(c.BaseDir, repoName, agentName, wp.Name)
		if err := CopyImageFolderToLower(c, image, wp.Destination, lowerDir); err != nil {
			return fmt.Errorf("failed to copy image folder to lower for %s: %w", wp.Name, err)
		}
	}

	return nil
}

func (c *Client) SetupOverlayMount(repoName, agentName, taskName, writablePathName, destination string) (string, error) {
	upperDir := GetUpperDir(c.BaseDir, repoName, agentName, writablePathName, taskName)
	workDir := GetWorkDir(c.BaseDir, repoName, agentName, writablePathName, taskName)

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
			// Capture stderr from the ExitError for better debugging
			if stderr := exitErr.Stderr; len(stderr) > 0 {
				log.Error(fmt.Sprintf("Podman command %s %s failed: %s", c.Binary, strings.Join(args, " "), string(stderr)))
			} else {
				log.Error(fmt.Sprintf("Podman command %s %s failed with exit code %d", c.Binary, strings.Join(args, " "), exitErr.ExitCode()))
			}
			return exitErr.ExitCode(), nil
		}
		log.Error(fmt.Sprintf("Podman command %s %s failed: %v", c.Binary, strings.Join(args, " "), err))
		return 1, fmt.Errorf("%s %s: %w", c.Binary, strings.Join(args, " "), err)
	}

	return 0, nil
}

func (c *Client) runContainer(image string, cmd []string, name string, mounts []config.Mount, commands *ContainerCommands) (int, error) {
	entrypoint := c.generateEntrypoint(cmd, commands)

	// Determine network mode
	networkMode := "bridge" // default
	if commands != nil && commands.Proxy != nil {
		networkMode = commands.Proxy.NetworkName
	}

	args := []string{"run", "--rm", "-it", "--network", networkMode, "--name", name}

	// Add environment variables from commands
	if commands.EnvVars != nil {
		for key, value := range commands.EnvVars {
			args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
		}
	}

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
