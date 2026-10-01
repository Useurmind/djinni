package git

// GitClient defines the interface for git operations
// This provides an abstraction layer for testing and future extensibility
type GitClient interface {
	// GetChangedFilesWithDiffs returns formatted git status with diffs
	GetChangedFilesWithDiffs(repoPath string) (string, error)

	// GetDiffs returns detailed diff information for all changed files
	GetDiffs(repoPath string) ([]Diff, error)

	// CommitAll stages and commits all changes in the repository
	CommitAll(repoPath, message string) error

	// PushBranch pushes the specified branch to origin
	PushBranch(repoPath, branchName string) error

	// CreatePatch creates a git patch from HEAD~1 to HEAD
	CreatePatch(repoPath, patchDir string) error

	// ApplyPatch applies a patch to a repository
	ApplyPatch(repoPath, patchPath string) error

	// CloneToTemp clones a repository to a temporary directory
	CloneToTemp(sourceDir, baseDir, agentName, taskName string) (string, error)

	// CheckoutNewBranch creates a new branch
	CheckoutNewBranch(repoPath, taskName string) error

	// GetRepoName returns the repository name from a directory
	GetRepoName(dir string) (string, error)

	// DeleteBranch deletes a branch
	DeleteBranch(repoPath, branchName string) error

	// AddFiles stages all files in a repository
	AddFiles(repoPath string) error

	// IsRepositoryClean checks if the repository has no uncommitted changes
	IsRepositoryClean(repoPath string) (bool, error)
}

// GitClientImpl is the default implementation of GitClient
type GitClientImpl struct{}

// NewGitClient creates a new GitClient instance
func NewGitClient() GitClient {
	return &GitClientImpl{}
}

// GetChangedFilesWithDiffs returns formatted git status with diffs
func (c *GitClientImpl) GetChangedFilesWithDiffs(repoPath string) (string, error) {
	return GetChangedFilesWithDiffs(repoPath)
}

// GetDiffs returns detailed diff information for all changed files
func (c *GitClientImpl) GetDiffs(repoPath string) ([]Diff, error) {
	return GetDiffs(repoPath)
}

// CommitAll stages and commits all changes in the repository
func (c *GitClientImpl) CommitAll(repoPath, message string) error {
	return CommitAll(repoPath, message)
}

// PushBranch pushes the specified branch to origin
func (c *GitClientImpl) PushBranch(repoPath, branchName string) error {
	return PushBranch(repoPath, branchName)
}

// CreatePatch creates a git patch from HEAD~1 to HEAD
func (c *GitClientImpl) CreatePatch(repoPath, patchDir string) error {
	return CreatePatch(repoPath, patchDir)
}

// ApplyPatch applies a patch to a repository
func (c *GitClientImpl) ApplyPatch(repoPath, patchPath string) error {
	return ApplyPatch(repoPath, patchPath)
}

// CloneToTemp clones a repository to a temporary directory
func (c *GitClientImpl) CloneToTemp(sourceDir, baseDir, agentName, taskName string) (string, error) {
	return CloneToTemp(sourceDir, baseDir, agentName, taskName)
}

// CheckoutNewBranch creates a new branch
func (c *GitClientImpl) CheckoutNewBranch(repoPath, taskName string) error {
	return CheckoutNewBranch(repoPath, taskName)
}

// GetRepoName returns the repository name from a directory
func (c *GitClientImpl) GetRepoName(dir string) (string, error) {
	return GetRepoName(dir)
}

// DeleteBranch deletes a branch
func (c *GitClientImpl) DeleteBranch(repoPath, branchName string) error {
	return DeleteBranch(repoPath, branchName)
}

// AddFiles stages all files in a repository
func (c *GitClientImpl) AddFiles(repoPath string) error {
	return AddFiles(repoPath)
}

// IsRepositoryClean checks if the repository has no uncommitted changes
func (c *GitClientImpl) IsRepositoryClean(repoPath string) (bool, error) {
	return IsRepositoryClean(repoPath)
}
