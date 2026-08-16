For proper functioning of containers some paths must be writable and should contain the content from the source container image.

# Example

This should be configured as follows

agents:
  local-agent:
    writablePaths:
      - name: home
        destination: /home/agent

# Implementation

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

To avoid sudo/root permissions during start, we want to use `unshare -rm` to create a new user mount namespace. 
Inside that namespace we will create the overlayfs mount and start the agent in his pod.
Then finally the mounted overlayfs work dir can be mounted into the pod at the destination, as other mounts too.

