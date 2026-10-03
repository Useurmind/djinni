# Security and Isolation

Djinni provides multiple layers of security isolation for AI agents running in containers. This document describes the isolation model and implementation details.

## Container-Based Isolation

### Container Runtime

Djinni uses **Podman** as the container runtime (see `pkg/docker/client.go:21`). Podman provides:
- Rootless container execution support
- Built-in user namespace mapping
-SELinux integration without requiring systemd

All container operations use Podman explicitly (see `pkg/docker/overlay.go:99,127,139`).

### Volume Mounts with SELinux Labels

All volume mounts use SELinux labels for isolation:

```go
// Read-only mounts
"-v", "source:destination:Z,ro,U"

// Read-write mounts
"-v", "source:destination:Z,U"
```

The flags mean:
- `:Z` — Private shared SELinux label (automatically relabeled)
- `:ro` — Read-only access (only for explicit read-only mounts)
- `:U` — User namespace mapping (SELinux user mapping)

Mounts are configured in `pkg/docker/client.go:139-147`.

### Read-Only Root Filesystem

Djinni enables read-only root filesystem by default:

```yaml
agents:
  secure-agent:
    harness_command: [opencode]
    containerfile: ./Containerfile
    forceReadOnlyRootOff: false
    tmpfsMounts:
      - destination: /tmp
      - destination: /cache
        size: "512m"
```

控制通过 `forceReadOnlyRootOff: true` 禁用。实现见 `pkg/docker/client.go:125-127`。

**Read-only root prevents:**
- Container escape attacks modifying system files
- Unauthorized software installation within the container
- Malware persistence through filesystem modifications

### Tmpfs Mounts

Tmpfs (RAM-backed) mounts provide writable storage for read-only containers:

```yaml
tmpfsMounts:
  - destination: /tmp
  - destination: /cache
    size: "512m"
```

**Benefits:**
- Data automatically cleared on container exit
- RAM-backed performance for temporary files
- No host filesystem exposure
- Size limits prevent memory exhaustion (see `pkg/docker/client.go:129-137`)

### Overlay-Based Writable Paths

For agents needing persistent writable storage, Djinni uses **overlayfs** with upper/lower/work directory structure:

```yaml
agents:
  secure-agent:
    harness_command: [opencode]
    containerfile: ./Containerfile
    writablePaths:
      - name: home
        destination: /home/agent
```

**How it works:**
- **Lower directory** — Read-only content from container image (populated at startup)
- **Upper directory** — Write layer (task-specific, isolated per execution)
- **Work directory** — Overlayfs working files
- Mount path: `{baseDir}/repo/agent/writablePaths/{name}/mnt`

**Benefits:**
- Write operations isolated to upper directory
- Content from image preserved in lower directory
- Task-specific isolation via upper directory naming
- Cleanup on task completion (see `pkg/docker/overlay.go:120-144`)

**Implementation details:**
- Overlay structure created at `{baseDir}/{repo}/{agent}/writablePaths/{name}/`
- Lower directory populated by copying from container image (see `pkg/docker/overlay.go:60-118`)
- Uses `podman unshare` for namespace-aware file operations
- Temporary container created to extract image content (see `TempContainerName` constant)

### File Copy Mechanism

Critical files are copied into the container via a temporary mount:

```yaml
files_to_copy:
  - source: ~/.gitconfig
    destination: /home/agent/.gitconfig
```

**Process:**
1. Temp mount created at `{baseDir}/{repo}/{agent}/copyMounts/{task}/`
2. Source files copied to temp mount
3. Container entrypoint copies files to destinations (see `pkg/docker/client.go:167-175`)
4. Temp mount cleaned up after execution

**Benefits:**
- Static snapshot at container startup
- No live symlink to host files
- Independent file ownership and permissions
- Prevents live modification from within container

### Git Workspace Isolation

Git operations use isolated workspace directories:

```yaml
agents:
  git-agent:
    harness_command: [python, -m, agent.harness]
    containerfile: ./Containerfile
    git_workspace:
      base_directory: /var/tmp/git-agent
```

**Details:**
- Default: `/var/tmp/djinni` (configurable via `base_directory` in agent config)
- Configurable per-agent via `base_directory`
- Isolated from host filesystem
- Typically mounted as tmpfs or overlay path

### Internal Network and Proxy Isolation

Djinni supports **internal network isolation** with optional **HTTP proxy control**:

```yaml
network:
  internal: true
  proxy:
    enabled: true
    allowList:
      - github.com
      - google.com
```

**Components:**
- **Internal Network**: Creates isolated bridge network `djinni-ai-{agentName}`
- **Proxy Container**: Runs Squid proxy on port 3128
- **Proxy Network**: Connects proxy to both internal and bridge networks

**Traffic Flow:**
```
Agent Container → Internal Network → Squid Proxy → Bridge Network → Internet
```

**Security Benefits:**
- Agent cannot access internet directly
- All outbound traffic must pass through Squid proxy
- ACL-based access control via `allowList`
- CONNECT tunneling support for HTTPS traffic

**Squid Configuration:**
- ACL definitions must be ordered before http_access rules
- SSL ports ACL for port 443 (HTTPS CONNECT)
- Safe ports ACL for ports 80 (HTTP) and 443 (HTTPS)
- Explicit allow rules for configured domains
- Catch-all deny for unauthorized access

**Proxy Environment Variables:**
Agent containers automatically receive:
- `HTTP_PROXY=http://proxy-container-name:3128`
- `HTTPS_PROXY=http://proxy-container-name:3128`
- `NO_PROXY=localhost,127.0.0.1`

**Implementation Details:**
- Network and proxy setup handled by `pkg/container/network.go`
- Proxy container lifecycle managed by `pkg/container/proxy.go`
- Proxy info persisted to `{baseDir}/proxyInfo/{repo}/{agent}/proxy.json`
- Cleanup handled by `pkg/container/client.go` cleanup methods

**Requirements:**
- Agent name must only contain alphanumeric characters, hyphens, and underscores
- Proxy image must be fully-qualified (e.g., `docker.io/ubuntu/squid:latest`)

## Security Model Summary

Djinni's security model focuses on:

| Layer | Implementation | Purpose |
|-------|---------------|---------|
| Container runtime | Podman (rootless support) | Secure container execution |
| File system isolation | SELinux `:Z,U` labels | Prevent cross-contamination |
| Mount isolation | Granular read/write mounts | Limit file access scope |
| Root filesystem | Read-only by default | Prevent filesystem modifications |
| Writable paths | Overlayfs (upper/lower/work) | Isolated writable storage |
| Network isolation | Bridge network (hardcoded) | Restrict container communications |
| Internal network | Isolated bridge network | Prevent direct internet access |
| Proxy isolation | Squid ACL-based proxy | Control outbound traffic |
| File copying | Temporary mount + copy | Static file snapshots |
| Namespace isolation | User namespace mapping | Isolate from host UIDs/GIDs |
| Temporary storage | Tmpfs mounts | Secure in-memory storage |
| Git workspace | Isolated base directory | Container-internal git ops |

This multi-layered approach provides defense-in-depth against container escape and lateral movement attacks.

## Security Best Practices

### 1. Minimal Container Images

Build containers with only required dependencies:

```dockerfile
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*
```

### 2. Read-Only Root Filesystem

Enable read-only root (default):

```yaml
agents:
  secure-agent:
    forceReadOnlyRootOff: false
    tmpfsMounts:
      - destination: /tmp
```

### 3. Use Overlay for Writable Paths

For agents needing write access:

```yaml
agents:
  worker-agent:
    harness_command: [python, -m, agent.harness]
    containerfile: ./Containerfile
    writablePaths:
      - name: data
        destination: /app/data
```

### 4. Copy Sensitive Files

Use `files_to_copy` for secrets/config instead of mounts:

```yaml
agents:
  secure-agent:
    files_to_copy:
      - source: ~/.ssh/id_rsa
        destination: /home/agent/.ssh/id_rsa
```

### 5. Network Restrictions

Djinni enforces bridge networking. Avoid requiring host network access.

### 6. Proxy Isolation

Enable proxy isolation for agents that should not have direct internet access:

```yaml
agents:
  secure-agent:
    harness_command: [opencode]
    containerfile: ./Containerfile
    network:
      internal: true
      proxy:
        enabled: true
        allowList:
          - github.com
          - gitlab.com
          - example.com
```

**Benefits:**
- Agent cannot access internet directly
- All outbound traffic routed through Squid proxy
- ACL-based access control via allowlist
- HTTPS CONNECT tunneling supported

### 7. Resource Limits

Configure CPU/memory limits at Podman level when running Djinni:

```bash
podman run --memory=512m --cpus=1.0 ...
```

Note: Current implementation doesn't expose resource limits in agent config.
