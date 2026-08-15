We use an ai agent to generate a commit message for the changes implemented by the agent container.

The generation of the message works as follows:
- determine changed files with git status --porcelain
- if no files changed the commit and pushback is not performed
- if any files changed we continue
- read out the file patches of all changed files as determined above
- integrate the file paths and status together with the file patches into the prompt
- the list of files and patches in the prompt should look likes this

    - <filepath> <status>
      <patch1>
      <patch2>
    - <filepath2> <status2>
      <patch3>
      <patch4>

More info:
- no files should be preloaded from disk, only patches should be provided
- The agent should be allowed to read files inside the container repository and have a tool for that.