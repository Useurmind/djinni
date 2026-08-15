The user expects the agent to see everything he sees.
As we clone the repo and check out the users branch for the agent in the container, it can happen that changes are not
yet commited and therefore do not get into the container.

Therefore we want to ask the user if any changes should be commited before starting the agent.

Rundown:
- uncommited changes on repo
- user calls start-agent command
- user is presented a descriptive warning, output of git status and asked if changes should be commited
- if yes proceed with next steps, if no just start the container
- he is asked if the ai agent should commit the changes
- if yes: use same logic as for computing the commit message for the container repo
- if no: let the user type a commit message and commit
- start the container as before