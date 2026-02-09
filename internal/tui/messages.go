package tui

import "github.com/raisedadead/wt/internal/git"

// projectLoadedMsg is sent when the project root and default branch are resolved.
type projectLoadedMsg struct {
	projectRoot   string
	defaultBranch string
	err           error
}

// worktreesLoadedMsg is sent when the worktree list finishes loading.
type worktreesLoadedMsg struct {
	worktrees []git.Worktree
	err       error
}

// statusLoadedMsg is sent when a worktree's status finishes loading.
type statusLoadedMsg struct {
	path   string
	status string
	err    error
}

// detailLoadedMsg is sent when detail info for a worktree is loaded.
type detailLoadedMsg struct {
	path    string
	commits string
	status  string
	files   string
	err     error
}

// diffLoadedMsg is sent when diff content for a worktree is loaded.
type diffLoadedMsg struct {
	path string
	diff string
}

// logLoadedMsg is sent when log content for a worktree is loaded.
type logLoadedMsg struct {
	path string
	log  string
}

// worktreeCreatedMsg is sent after a new worktree is created.
type worktreeCreatedMsg struct {
	path   string
	branch string
	err    error
}

// worktreeDeletedMsg is sent after a worktree is deleted.
type worktreeDeletedMsg struct {
	branch string
	err    error
}

// pruneMsg is sent after pruning stale worktrees.
type pruneMsg struct {
	err error
}

// fetchMsg is sent after fetching from remotes.
type fetchMsg struct {
	err error
}

// flashMsg is a temporary message displayed in the footer.
type flashMsg struct {
	text    string
	isError bool
}

// clearFlashMsg clears the flash message.
type clearFlashMsg struct{}
