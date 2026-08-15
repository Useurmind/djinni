We want to offer the user several ways to sync back the changes the ai agent made in the container

    1. push the branch from the agent repo to the users repo (optionally automerge into users branch)
    2. create a patch from the agent repo and apply it the users repo

# Push Branch Approach

This is already implemented.

- Create task branch in agent repo
- Commit after container exits (with automatic commit message from ai agent)
- Push branch to user repo

No changes required, except for the option to automatically merge the agent task branch into the user branch.
This should be configurable in the agent config in the repo.
If `automerge_agent_branch` is set to true, the agents task branch will automatically be merged onto the users branch.

This approach can be configured via

    sync_approach: "branch_sync"

## Apply Patch Approach

This is the new approach:

- Create task branch in agent repo
- After container exits create a git patch from the changes in the branch of the agent repo
- Apply the patch to the users repo

This approach can be configured via

    sync_approach: "git_patch"