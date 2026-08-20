Currently we do asynchronous copy via podman exec and podman cp to copy files into the running container.
This leads to strange locking in wsl.

To alleviate the problem we want to change the copy process.
- create a tmp dir /tmp/djinni/<repo>/<agent>/copyMounts/<task> to copy the files to that we need
- bind mount the tmp dir readonly
- copy the files inside the container in the entrypoint from the bind mounted tmp dir to the destination

This makes everything simpler as no asynchronous logic is required anymore.