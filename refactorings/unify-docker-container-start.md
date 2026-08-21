Currently we have two functions that start a Podman container in `pkg/docker/client.go`.
- `runContainerWithCommands`
- `runContainerDirect`

We only want one unified function `runContainer` that satifies all requirements.
Rename `runContainerWithCommands` to `runContainer`.
Drop `runContainerDirect` and use the unified version always.
Make sure proper defaults are applied when options are not actually used.