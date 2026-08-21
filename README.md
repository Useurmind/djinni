# Djinni

A Go application for running AI agent coding harnesses inside Podman containers for isolation and security during software development.

## TL;DR

### Installation

```bash
go install github.com/useurmind/djinni@latest
```

Requires Podman.

### Commands

```bash
# Start an agent (requires task name, creates feature/<taskname> branch)
djinni start <agent-name> --task <task-name>

# Build a local container image from Containerfile
djinni prepare <agent-name>

# Enable debug mode
djinni --debug
```

### Configuration

See [Configuration Guide](docs/configuration.md) for detailed documentation.

Create `.djinni.yml` in your project root:

```yaml
default_model: qwen-qwen3-coder-next-fp8

agents:
  local-agent:
    harness_command:
      - opencode
    containerfile: ./Containerfile
    # sync_approach: git_patch
    mounts:
      # for opencode you must provide all folders including the database folder
      # while agent is running it is chowned to the podman user, later it is chowned back to you
      - source: /home/jgruen/.config/opencode
        destination: /home/agent/.config/opencode
      - source: /home/jgruen/.local/state/opencode
        destination: /home/agent/.local/state/opencode
      - source: /home/jgruen/.local/share/opencode
        destination: /home/agent/.local/share/opencode
```

### Global Configuration

See [Configuration Guide](docs/configuration.md#global-configuration-djinniconfigconfigyaml) for detailed documentation.

Create `~/.config/djinni/config.yaml` (or `$DJINNI_CONFIG_DIR/config.yaml`) for model providers:

```yaml
modelProviders:
  - name: local
    apiBase: http://localhost:11434/v1
    apiKey: dummy
    models:
      - id: qwen-qwen3-coder-next-fp8
```

### Sync Approaches

See [Configuration Guide](docs/configuration.md#agent-configurationsync_approach-string-optional) for detailed documentation.

| Approach | Description |
|----------|-------------|
| `none` | No sync; leave changes on `feature/<task>` branch in your workspace |
| `gitpatch` | Generate patch file from agent workspace and apply to your workspace (files only) |
| `automerge` | Merge `feature/<task>` branch directly into current local branch |

### Agent Options

See [Configuration Guide](docs/configuration.md#agent-configurations) for detailed documentation.

| Option | Description |
|--------|-------------|
| `harness_command` | Command to run in container (required) |
| `image` | Container image to use (mutually exclusive with `containerfile`) |
| `containerfile` | Build image from local Containerfile |
| `mounts` | Volume mounts (source → destination) |
| `files_to_copy` | Files to copy into container (e.g., `.gitconfig`) |
| `git_workspace` | Git workspace configuration for task-based work |
| `sync_approach` | How to sync changes back: `none`, `gitpatch`, `automerge` |
| `autodelete_agent_branch` | Auto-delete feature branch after sync |
| `forceReadOnlyRootOff` | Disable read-only root filesystem |
| `tmpfsMounts` | Tmpfs mounts for writable paths in read-only mode |
| `default_model` | Override default LLM model for this agent |

### Workflow

1. Define agents in `.djinni.yml`
2. Run `djinni start <name> --task <task>` to execute
3. Agent runs in container, makes changes to git working directory
4. Changes are committed and pushed to `feature/<task>` branch
5. Changes sync back per `sync_approach` setting

## Overview

Djinni provides a framework to manage AI agents in isolated Podman environments. Each agent runs in its own container, ensuring security isolation, resource limits, and clean dependencies.

## Security and Isolation

See [Security and Isolation](docs/security.md) for detailed documentation on container-based security, non-root user execution, volume mounts, and comparison with VS Code agent isolation.

## Features

- **Container isolation**: Each agent runs in a separate Podman container
- **Resource management**: CPU/memory limits per agent
- **Environment configuration**: Secure environment variable injection

## License

MIT