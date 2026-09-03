Workspaces with results from agents should never be lost accidentially.

After the agent container exits there are several choices to make and actions to take:

- generate commit message (ai model call)
- delete workspace (user choice)

These two if implemented wrongly can lead to loosing the changes implemented by the agent.

## Situation

If the ai model call fails and the sync of the branc to the users repo is not performed and the workspace is deleted, the implemented changes will be lost. 

## Goal

### No workspace deletion on error

Independent of the result of the ai commit message call or any other errors, the workspace content should never be lost.
If the ai call or any other preliminary steps fail, the workspace should NOT be deleted.
Workspace deletion should be the last step at the end of a run.

### Default AI commit message

In addition if the ai commit message call fails the app should fall back to a default commit message "no ai available for commit message" and not automerge the branch. Instead it should only sync the branch to the user repo.