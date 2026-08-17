For proper functioning of containers some paths must be writable and should contain the content from the source container image.

# Example

This should be configured as follows

agents:
  local-agent:
    writablePaths:
      - name: home
        destination: /home/agent

# Implementation

## Filesystem

The implementation should be made with overlayfs. We create an overlayfs where the lowerdir is a copy of the folder in the image.
The upperdir will be empty on container start, and cleaned after container exit.
That enables us to only once create the lowerdir, which might be costly and reuse it across multiple agent runs.

The folder structure should be like this

- /tmp/djinni/<repoName>/<agentName>/writablePaths/<writablePathName>
  | - /lower
  | - /upper/<taskName>
  | - /work/<taskName>

The lower folder is created  on prepare-agent from the docker image folder via `docker cp`.
The upper and work folder are created and mounted with overlayfs during start of an agent.

## Unshare

To avoid sudo/root permissions during start, we want to use `podman unshare` to manage the overlay mount inside the podman namespace. 
Inside that namespace we will create the overlayfs mount.
Then finally the mounted overlayfs work dir can be mounted into the pod at the destination, as other mounts too.

Example bash script

  set -e
  podman unshare mount -t overlay overlay -o lowerdir=/path/to/lowerdir,upperdir=/path/to/upperdir,workdir=/path/to/workdir /path/to/podmannamespace/mountdir
  podman run -i -v /path/to/podmannamespace/mountdir:/path/to/container/mountdir <image>
  podman unshare umount /path/to/podmannamespace/mountdir

## Code

The code should simply call the corresponding commands in order and ensuring that cleanup is done at the end.
Create functions for podman unshare mount and unshare unmount actions.
Call the mount action, run the container as we are used to with additional unshare mount, then after the container exits call the unmount function. 