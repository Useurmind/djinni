We want to offer a flexible workflow to get the changes from the agent back to the users repo

Always:
- The agent develops on a feature branch 
- The changes are automatically commited to that branch and synced back to the user repo

Now the user has several options for applying the changes

- no action, leave it up to the user how to proceed
- automerge the branch to the users branch (optionally with auto delete of the agent branch)
- apply a patch of the changes the agent did to the workspace of the user (optionally with auto delete of the agent branch)


# Push Branch 

Initial development and sync are as follows

- Create task branch in agent repo
- Commit after container exits (with automatic commit message from ai agent)
- Push branch to user repo

# Applying the changes

Once the container has exited and the agent branch was pushed to the user branch, several approaches are possible to apply the changes from the agent branch.

## Automerge Approach

First option is automerge, to automatically merge the agent task branch into the user branch.
This should be configurable in the agent config in the repo.
If `sync_approach` is set to `automerge`, the agents task branch will automatically be merged onto the users branch.

## Apply Patch Approach

This approach is as follows:

- Create a git patch from the commit(s) on the agent branch
- Apply the patch to workspace of the user (no commit, just apply to files in filesystem)

This approach can be configured via `sync_approach: gitpatch`

## Autodelete agent branch

After the sync of the changes has taken place either via automerge or via git patch, the user has the option to automatically delete the agent branch.
Can be configured via `autodelete_agent_branch`.

## Interactivity

The user should be able to choose all of this interactively if not configured in the agent config.
If the flags are false, there should be a prompt that ask the user to decide.

Details:

- If `sync_approach` is not set, ask the user for the sync approach `none`, `gitpatch`, `automerge`
- If `autodelete_agent_branch` is false and the `sync_approach` is not `none` ask the user if the agent branch should be deleted

Autodelete of agent branch should be done after the git patch was applied or the automerge was performed.