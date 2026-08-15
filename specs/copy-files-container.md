We need to copy some files to the agent container.
The files must be writable an available before the entrypoint starts executing commands.
They should be copied via `podman cp` to the container.
The entrypoint must wait until the files were all copied. This is done by waiting for a marker file /home/agent/.copydone.
Once that file is present the entrypoint script may continue.

no, mounts will not work
instead do the following:
- start a gorouting that waits until the container has started
- start the container
- first command in the entrypoint should be to wait for a marker file in the home directory of the agent
- go routine sees that container started and starts to copy
- go routine copy all defined files (currently only ~/.gitconfig to their places in the container, e.g. /home/agent/.gitconfig)
- go routine creates the marker file .copydone in the directory /home/agent
- entrypoint continues

the container should be started in the main thread, and copying is done in the goroutine
