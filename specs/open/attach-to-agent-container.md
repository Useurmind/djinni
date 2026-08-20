I want to be able to attach to an agent container via a command in the djinni cli.

    djinni attach-agent <agentName> --cmd bash

The default command should be bash.

This commands just execs into the container that was started with the `djinni start-agent <agentName>` command.
The user than has an interactive shell in the container to start tools or just write bash commands.