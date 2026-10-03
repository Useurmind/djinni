# Djinni

A Go application for running AI agent coding harnesses inside Podman containers for isolation and security during software development.

## TL;DR

### Installation

```bash
go install github.com/useurmind/djinni@latest
```

Requires Podman (Docker is not supported).

### Commands

```bash
# Build container image and prepare network/proxy (if configured)
djinni prepare <agent-name>

# Start an agent with task (creates feature/<taskname> branch)
djinni start <agent-name> --task <task-name>
  --cmd <command>    Override harness command
  --rm               Delete workspace and overlay on exit

# Clean up network and proxy containers
djinni clean
```

### Configuration

See [Configuration Guide](docs/configuration.md) for detailed documentation.

Create `.djinni.yml` in your project root:

```yaml
default_model: qwen-qwen3-coder-next-fp8

agents:
  default:
    harness_command:
      - opencode
    containerfile: ./Containerfile
    # sync_approach: git_patch
    network:
      internal: true
      proxy:
        enabled: true
        allowList:
          - github.com
          - google.com
    mounts:
      # for opencode you must provide all folders including the database folder
      # while agent is running it is chowned to the podman user, later it is chowned back to you
      - source: ~/.config/opencode
        destination: /home/agent/.config/opencode
      - source: ~/.local/state/opencode
        destination: /home/agent/.local/state/opencode
      - source: ~/.local/share/opencode
        destination: /home/agent/.local/share/opencode
    tmpfsMounts:
      - destination: /cache
        size: "512m"
    writablePaths:
      - name: home
        destination: /home/agent
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

### Image Naming

When specifying container images in your agent configuration, **always use fully-qualified image names**. Short names (e.g., `ubuntu`, `python:3.11`) may fail in non-interactive environments.

**Correct (fully-qualified):**
```yaml
image: docker.io/library/ubuntu:latest
image: docker.io/library/python:3.11-slim
image: quay.io/bitnami/postgresql:15
```

**Incorrect (short names - avoid):**
```yaml
image: ubuntu:latest        # May fail - no registry prefix
image: python:3.11-slim     # May fail - no registry prefix
```

**Why?** Podman enforces short-name resolution in non-interactive environments. Without a fully-qualified name (`docker.io/library/image:tag`), Podman cannot determine which registry to use, leading to errors like "short-name resolution enforced".

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
| `delete_on_exit` | Delete workspace and overlay on exit: `none` or `all` |
| `forceReadOnlyRootOff` | Disable read-only root filesystem |
| `tmpfsMounts` | Tmpfs mounts for writable paths in read-only mode |
| `default_model` | Override default LLM model for this agent |
| `network.internal` | Enable internal network isolation for agent |
| `network.proxy.enabled` | Enable HTTP proxy via Squid for agent |
| `network.proxy.allowList` | Domains allowed through proxy (via Squid ACL) |

### Network Isolation

Agents can be configured with internal network isolation and HTTP proxy control:

- **Internal Network**: Creates an isolated network for the agent (prevents direct internet access)
- **HTTP Proxy**: Squid proxy sits between agent and internet with ACL-based access control
- **Proxy Environment Variables**: Agent container automatically configured with proxy environment variables

**Example:**
```yaml
network:
  internal: true
  proxy:
    enabled: true
    allowList:
      - github.com
      - google.com
```

### Workflow

1. Define agents in `.djinni.yml`
2. Run `djinni prepare <name>` to build container image (sets up network/proxy if configured)
3. Run `djinni start <name> --task <task>` to execute
4. Agent runs in container, makes changes to git working directory
5. Changes are committed and pushed to `feature/<task>` branch
6. Changes sync back per `sync_approach` setting
7. Run `djinni clean` to remove network and proxy containers

## Overview

Djinni provides a framework to manage AI agents in isolated Podman environments. Each agent runs in its own container, ensuring security isolation, resource limits, and clean dependencies.

## Security and Isolation

See [Security and Isolation](docs/security.md) for detailed documentation on container-based security, non-root user execution, volume mounts, and comparison with VS Code agent isolation.

## Features

- **Container isolation**: Each agent runs in a separate Podman container
- **Resource management**: CPU/memory limits per agent
- **Environment configuration**: Secure environment variable injection
- **Network isolation**: Internal network support with optional Squid proxy
- **Proxy control**: ACL-based access control for outbound traffic
- **Read-only filesystem**: Default read-only root filesystem for enhanced security

## License

MIT