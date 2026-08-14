Please plan the implementation of automatic workspace mount into the agent container.
The current working dir which is expceted to be a git repo should be cloned to a another directory.
The new repo should have permissions 777 so that when mounted into a container there are no problems with permissions.
In addition a new branch should be started from the main/master branch in the cloned repo.

The base directory for the clone should be configured in the agent config. By default /tmp/<agentname>.
The name of the folder to clone to should be the original repo folder name plus the task name.
The task name should be specified via command line. The new branch should be feature/<taskname>.

