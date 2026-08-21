# Djinni AGENTS.md

## Build & verification

```bash
make build test vet lint deadcode
```

A task is complete only when all four succeed.

**Do not** modify `.golangci.yml` to suppress issues—fix them in the code.

---

## Architecture notes

- **Main entrypoint**: `main.go` → `cmd.Execute()` → root Cobra command
- **Agent execution**: `pkg/ai/agent.go:Execute()` reads git changes, generates commit messages via LLM
- **Container runtime**: Uses `podman` exclusively (see `pkg/docker/client.go:NewClient()`)
- **Config file**: `.djinni.yml` defines agents with `harness_command`, `image`/`containerfile`, and mounts

---

## Key conventions

1. **Git tools**: `GitChangedFilesTool` in `pkg/ai/tools.go` returns diffs for all changed files; returns "No changes detected." if repo is clean
2. **Test package**: `testify/assert` for assertions; use `require.NoError(t, err, "descriptive message")` for errors
3. **Test scope**: Skip tests that only verify struct field access—assume that works
4. **Mount paths in Podman**: Source → destination; use `:Z` (rw) or `:Zro` (ro) SELinux labels
5. **Code Documentation**: Add documentation to structs, struct fields (especially in the config package) and funcs (especially public funcs)
6. **Code structure**: 
    - Dont make them too long/short, strive to keep functions on single abstraction levels and extract code to new funcs if the abstraction level does not match the current function.
    - prefer handing the variables from func to func instead of defining global variables, create conceptually sound 'classes' or container structs if signatures become to large
---

## Dev loop

1. Changes in `./...`
2. `make build` creates `./bin/djinni`
3. Run config via `./bin/djinni` (reads `.djinni.yml`)
4. Agents execute in containers via `pkg/docker` client

---

## Dependencies

- Go 1.25.0
- `github.com/stretchr/testify` v1.10.0 (test assertions)
- `github.com/tmc/langchaingo` v0.1.14 (LLM integration)
- `github.com/spf13/cobra` v1.8.1 (CLI)
