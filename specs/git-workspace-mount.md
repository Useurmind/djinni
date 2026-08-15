This is the description of the automatic workspace mount in the agent container.
The current working dir which is expceted to be a git repo should be cloned to a another directory.
In addition a new branch should be started from the users active branch in the cloned repo.

The base directory for the clone should be configured in the agent config. By default /tmp/djinni/<repoName>/<agentname>/.
The name of the folder to clone to should be the task name.
The task name should be specified via command line. The new branch should be feature/<taskname>.
Also in the clone repository the active branch of user should be checked out after cloning.

The permissions of the workspace will be adapted to the container via the :U option of the mount.

