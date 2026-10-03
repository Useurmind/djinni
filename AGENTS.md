# Djinni AGENTS.md

## Build & verification

```bash
make build test vet lint deadcode
```

A task is complete only when all four succeed.

**Do not** modify `.golangci.yml` to suppress issues—fix them in the code.

---

## Architecture notes

- **Main entrypoint**: `main.go` → `cmd.Execute()` → root Cobra command with subcommands (`start`, `prepare`, `attach`, `clean`)
- **Agent execution**: `pkg/ai/agent.go:Execute()` reads git changes via `GetChangedFilesWithDiffs()`, generates commit messages via LLM
- **Container runtime**: Uses `podman` exclusively (see `pkg/container/client.go:NewClient()`)
- **Config file**: `.djinni.yml` defines agents with `harness_command`, `image`/`containerfile`, and mounts
- **Network isolation**: Optional internal network with Squid proxy for agent internet access control

### Package Structure

| Package | Purpose |
|---------|---------|
| `pkg/container` | Podman client, network setup, proxy management |
| `pkg/config` | Configuration types and validation |
| `pkg/ai` | Agent execution and LLM integration |
| `pkg/git` | Git operations (see note below) |
| `pkg/ui` | User interaction and prompts |
| `pkg/log` | Logging infrastructure |
| `pkg/utils` | Utility functions |

**Note on pkg/git**: The git package contains multiple files for git operations: `commit.go`, `diff.go`, `push.go`, `status.go`, `sync.go`, `workspace.go`.

### Network and Proxy Lifecycle

1. **Network Setup** (`pkg/container/network.go`):
   - Creates internal bridge network `djinni-ai-{agentName}` (hyphens replaced with underscores)
   - Handles network cleanup and removal

2. **Proxy Setup** (`pkg/container/proxy.go`):
   - Generates Squid configuration with ACL rules
   - Starts proxy container connected to internal network
   - Connects proxy to bridge network for internet access

3. **Environment Configuration** (`pkg/container/client.go`):
   - Sets `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` environment variables
   - Proxy info persisted to `{baseDir}/proxyInfo/{repo}/{agent}/proxy.json`

4. **Cleanup** (`cmd/clean.go`):
   - Stops and removes proxy containers
   - Removes internal networks

---

## Key conventions

1. **Git tools**: `GetChangedFilesWithDiffs()` in `pkg/git/diff.go` returns diffs for all changed files; returns "No changes detected." if repo is clean
2. **Test package**: `testify/assert` for assertions; use `require.NoError(t, err, "descriptive message")` for errors
3. **Test scope**: Skip tests that only verify struct field access—assume that works
4. **Mount paths in Podman**: Source → destination; use `:Z,ro,U` (read-only) or `:Z,U` (read-write) SELinux labels
5. **Code Documentation**: Add documentation to structs, struct fields (especially in the config package) and funcs (especially public funcs)
6. **Code structure**: 
    - Dont make them too long/short, strive to keep functions on single abstraction levels and extract code to new funcs if the abstraction level does not match the current function.
    - prefer handing the variables from func to func instead of defining global variables, create conceptually sound 'classes' or container structs if signatures become to large
---

## Dev loop

1. Changes in `./...`
2. `make build` creates `./bin/djinni`
3. Run config via `./bin/djinni` (reads `.djinni.yml`)
4. Agents execute in containers via `pkg/container` client
5. Network/proxy setup/teardown handled automatically

---

## Dependencies

- Go 1.26.0
- `github.com/stretchr/testify` v1.10.0 (test assertions)
- `github.com/tmc/langchaingo` v0.1.14 (LLM integration)
- `github.com/spf13/cobra` v1.8.1 (CLI)
- `github.com/uber-go/zap` (logging)
