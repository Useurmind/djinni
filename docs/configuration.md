# Configuration Guide

Djinni uses a YAML-based configuration system with two levels of configuration:

1. **Local Configuration** (`.djinni.yml`): Project-specific agent definitions
2. **Global Configuration** (`~/.config/djinni/config.yaml`): Global model providers

## Local Configuration (`.djinni.yml`)

The local configuration file defines agents and their execution settings. Place this file in your project root.

### Top-Level Fields

#### `default_model` (string)

Default LLM model to use for agents that don't specify their own. This can be overridden per-agent using the `default_model` field in the agent configuration.

**Example:**
```yaml
default_model: qwen-qwen3-coder-next-fp8
```

#### `agents` (map[string]*AgentConfig)

Map of agent names to their configuration objects. Each agent defines how it should be executed.

**Example:**
```yaml
agents:
  local-agent:
    # agent configuration
  registry-agent:
    # agent configuration
```

### Agent Configuration (`AgentConfig`)

Each agent supports the following fields:

#### `harness_command` ([]string) **Required**

Command to run inside the container. This is the main entrypoint for the agent's work.

**Example:**
```yaml
harness_command:
  - opencode
```

Or with arguments:
```yaml
harness_command:
  - python
  - -m
  - agent.harness
```

#### `image` (string) **Conditional**

Docker image to use for the agent. Either `image` or `containerfile` must be specified, but not both.

**Example:**
```yaml
image: python:3.11-slim
```

#### `containerfile` (string) **Conditional**

Path to a local Containerfile/Dockerfile to build an image from. Either `image` or `containerfile` must be specified, but not both.

**Example:**
```yaml
containerfile: ./Containerfile
```

#### `mounts` ([]Mount)

Volume mounts to make files and directories available in the container.

**Mount Fields:**

- `source` (string) **Required**: Path on the host machine
- `destination` (string) **Required**: Path inside the container
- `readOnly` (bool, optional): Mount the volume as read-only (default: `false`)

**Example:**
```yaml
mounts:
  - source: /home/user/.config/opencode
    destination: /home/agent/.config/opencode
  - source: ./specs
    destination: /app/specs
    readOnly: true
```

#### `files_to_copy` ([]FilesToCopy)

Files to copy into the container (e.g., `.gitconfig`, SSH keys). Unlike mounts, these are copied once at startup.

**FilesToCopy Fields:**

- `source` (string) **Required**: Path on the host machine
- `destination` (string) **Required**: Path inside the container

**Example:**
```yaml
files_to_copy:
  - source: ~/.gitconfig
    destination: /home/agent/.gitconfig
```

#### `default_model` (string, optional)

Override the global `default_model` for this specific agent.

**Example:**
```yaml
default_model: gpt-4
```

#### `git_workspace` (GitWorkspaceMount)

Configure the git workspace directory inside the container. Used when the agent needs to make git commits.

**GitWorkspaceMount Fields:**

- `base_directory` (string, optional): Base directory for git operations (default: `/tmp/djinni`)

**Example:**
```yaml
git_workspace:
  base_directory: /tmp/git-agent
```

#### `sync_approach` (string, optional)

Strategy for syncing changes back to your workspace after the agent completes.

**Valid values:**
- `none`: No sync; leave changes on `feature/<task>` branch in your workspace
- `gitpatch`: Generate patch file from agent workspace and apply to your workspace (files only)
- `automerge`: Merge `feature/<task>` branch directly into current local branch

**Example:**
```yaml
sync_approach: gitpatch
```

#### `autodelete_agent_branch` (bool, optional)

Automatically delete the `feature/<task>` branch after syncing changes back to your workspace. Only applicable when using `gitpatch` or `automerge` sync approaches.

**Example:**
```yaml
autodelete_agent_branch: true
```

## Global Configuration (`~/.config/djinni/config.yaml`)

The global configuration file defines available model providers for LLM integration.

### GlobalConfig Fields

#### `modelProviders` ([]ModelProvider)

List of model providers configured for use with Djinni agents.

### ModelProvider Fields

#### `name` (string) **Required**

Unique identifier for the model provider.

**Example:**
```yaml
name: local
```

#### `apiBase` (string) **Required**

Base URL for the provider's API endpoint.

**Example:**
```yaml
apiBase: http://localhost:11434/v1
```

#### `apiKey` (string, optional)

API key or token for authentication. Not all providers require this.

**Example:**
```yaml
apiKey: dummy
```

#### `models` ([]Model)

List of available models from this provider.

**Model Fields:**

- `id` (string) **Required**: Model identifier used in agent configuration

**Example:**
```yaml
models:
  - id: qwen-qwen3-coder-next-fp8
  - id: codellama-7b
```

### Full Global Configuration Example

```yaml
modelProviders:
  - name: local
    apiBase: http://localhost:11434/v1
    apiKey: dummy
    models:
      - id: qwen-qwen3-coder-next-fp8
      - id: codellama-7b

  - name: openai
    apiBase: https://api.openai.com/v1
    apiKey: sk-...
    models:
      - id: gpt-4
      - id: gpt-4o
```

## Complete Configuration Examples

### Minimal Agent Configuration

```yaml
agents:
 简单的agent:
    harness_command:
      - echo
      - hello
    image: busybox
```

### Full-Featured Agent Configuration

```yaml
default_model: qwen-qwen3-coder-next-fp8

agents:
  opencode-agent:
    harness_command:
      - opencode
    containerfile: ./Containerfile
    default_model: qwen-qwen3-coder-next-fp8
    sync_approach: automerge
    autodelete_agent_branch: true
    git_workspace:
      base_directory: /tmp/opencode-workspace
    mounts:
      - source: /home/user/.config/opencode
        destination: /home/agent/.config/opencode
      - source: /home/user/.local/state/opencode
        destination: /home/agent/.local/state/opencode
      - source: /home/user/.local/share/opencode
        destination: /home/agent/.local/share/opencode
    files_to_copy:
      - source: ~/.gitconfig
        destination: /home/agent/.gitconfig
```

## Validation Rules

Djinni validates configuration files on load. Key validation rules:

1. Agent must specify either `image` or `containerfile` (not both)
2. `harness_command` is required for each agent
3. `harness_command` cannot be an empty array
4. Agent names cannot be empty
5. `sync_approach` must be one of: `none`, `gitpatch`, `automerge`
6. Model provider name cannot be empty
7. Model provider must have at least one model defined

## File Locations

- **Local config**: `.djinni.yml` in project root (or specify with `--config` flag)
- **Global config**: `~/.config/djinni/config.yaml` (or `$DJINNI_CONFIG_DIR/config.yaml`)

## Environment Variables

- `DJINNI_CONFIG_DIR`: Override the global config directory (default: `~/.config`)
