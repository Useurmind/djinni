After the agent container exits all changes in the workspace repo should be commited to the branch.
After commit, the branch should be pushed from the workspace git repo to the original repository.
Use the available ai agent to generate an automatic commit message for the changed files.

Detailed rundown:
- container exits
- AI agent looks at the changed files and creates a commit message
- all files are added and committed using the generated commit message
- workspace repo feature branch is pushed to original repo on the machine