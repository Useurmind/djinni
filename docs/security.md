# Security and Isolation

Djinni provides multiple layers of security isolation for AI agents running in containers. This document describes the isolation model and how it compares to other approaches.

## Container-Based Isolation

### Non-Root User

Each agent runs as a non-root user (`agent`) inside the container. The container image creates a dedicated user:

```dockerfile
RUN useradd -m -s /bin/bash agent
USER agent
WORKDIR /home/agent
```

This prevents:
- Container escape attacks that rely on root privileges
- Unauthorized system modifications within the container
- Access to sensitive host files that require root

### Separate User Namespace

Djinni uses separate user namespaces to isolate the agent from the host system:

- The `agent` user inside the container has different UID/GID than the host user
- Permissions are mapped to prevent cross-contamination
- File operations are constrained to the container's filesystem

### Volume Mounts with SELinux Labels

All volume mounts use SELinux labels for additional isolation:

```go
// Read-only mounts
"-v", "source:destination:Z,RO,U"

// Read-write mounts
"-v", "source:destination:Z,U"
```

The `:Z` flag indicates private shared SELinux label, and `:U` flag maps the mount to the user namespace.

### Read-Only Mounts

Sensitive configuration can be mounted as read-only:

```yaml
mounts:
  - source: ~/.kube/config
    destination: /home/agent/.kube/config
    readOnly: true
```

This prevents:
- Malicious agents from tampering with configuration
- Accidental modification of critical files
- Ransomware-style encryption of configuration

### Read-Only Root Filesystem

Djinni enables read-only root filesystem by default for maximum security:

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

**Benefits of tmpfs:**
- Data automatically cleared on container exit
- RAM-backed performance for temporary files
- No host filesystem exposure
- Size limits prevent memory exhaustion



### File Copy Mechanism

Critical files (like `.gitconfig`, SSH keys) are copied into the container rather than mounted:

(The gitconfig is automatically copied to the agent container with this mechanism, no need to configure it here)

```yaml
files_to_copy:
  - source: ~/.gitconfig
    destination: /home/agent/.gitconfig
```

This approach provides:
- Static snapshot of files at container startup
- No live symlink to host files (cannot be modified from within container)
- Independent file ownership and permissions

## Comparison with VS Code Agent Isolation

### VS Code Remote Containers

VS Code agents typically run in containers with:

| Aspect | VS Code Remote | Djinni |
|--------|---------------|--------|
| User | Root (default) or dev user | Non-root `agent` user |
| SELinux labeling | Optional | Mandatory (`:Z,U` flags) |
| File system access | Mounts entire workspace | Granular mounts + file copy |
| Default permissions | Full workspace read/write | Limited scope |
| Network isolation | Host network | Bridge network only |

### Key Security Advantages

1. **Default non-root**: Djinni always uses non-root user; VS Code often uses root
2. **SELinux enforcement**: Djinni applies `:Z,U` labels by default
3. **Granular file access**: Djinni uses file copy for sensitive data
4. **Network isolation**: Djinni uses bridge network; VS Code uses host network

## Running TUI Agents Inside Containers

TUI (Text User Interface) agents can run inside Djinni containers with enhanced security:

### Supported TUI Frameworks

- **opencode**: Run directly via `harness_command: [opencode]`
- **Custom TUI applications**: Build them into the container image
- **SSH-based TUI**: Use SSH inside container for terminal access

### Container Configuration

```dockerfile
FROM ubuntu:24.04

RUN apt-get update && apt-get install -y \
    git \
    curl \
    make \
    build-essential \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -m -s /bin/bash agent
USER agent
WORKDIR /home/agent

CMD ["opencode"]
```

### Agent Configuration

```yaml
agents:
  tui-agent:
    harness_command:
      - opencode
    containerfile: ./Containerfile
    mounts:
      - source: ~/.config/opencode
        destination: /home/agent/.config/opencode
        readOnly: true
```

### Benefits of TUI in Container

1. **No GUI dependencies**: TUI works without X11/Wayland
2. **Reduced attack surface**: No display server to exploit
3. **Simplified isolation**: No GPU or display permissions needed
4. **Deterministic environment**: TUI behaves identically across hosts

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

Enable read-only root filesystem in agent configuration:

```yaml
agents:
  secure-agent:
    forceReadOnlyRootOff: false
    tmpfsMounts:
      - destination: /tmp
```

### 3. Network Restrictions

Djinni uses bridge networking to restrict container access to the host:

- Containers cannot access host services directly
- No direct container-to-container communication
- Outbound network access limited to bridge

### 4. Resource Limits

Configure CPU and memory limits in production:

```yaml
agents:
  restricted-agent:
    harness_command: [opencode]
    image: opencode:latest
    # Add host_config for memory limits
```

## Security Model Summary

Djinni's security model focuses on:

| Layer | Implementation | Purpose |
|-------|---------------|---------|
| User isolation | Non-root `agent` user | Prevent privilege escalation |
| Namespace isolation | Separate user namespace | Isolate from host |
| File system isolation | SELinux `:Z,U` labels | Prevent cross-contamination |
| Mount isolation | Granular read/write mounts | Limit file access scope |
| Network isolation | Bridge network | Restrict container comms |
| File copying | Static file copy | Prevent live symlink attacks |
| Root filesystem isolation | Read-only root | Prevent filesystem modifications |
| Temporary storage | Tmpfs mounts | Secure writable temporary storage |

This multi-layered approach provides defense-in-depth against container escape and lateral movement attacks.
