package git

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Worktree represents a git worktree
type Worktree struct {
	Path   string
	Branch string
	Commit string
}

// CreateWorktree creates a new worktree with a new branch
// The directory name is flattened (slashes become dashes)
func CreateWorktree(projectRoot, branchName string) (string, error) {
	return CreateWorktreeWithBase(projectRoot, branchName, "")
}

// CreateWorktreeWithBase creates a new worktree with a new branch from a specific base
// The directory name is flattened (slashes become dashes)
func CreateWorktreeWithBase(projectRoot, branchName, baseBranch string) (string, error) {
	// Flatten branch name for directory (e.g., feature/auth -> feature-auth)
	dirName := FlattenBranchName(branchName)
	worktreePath := filepath.Join(projectRoot, dirName)

	// Create worktree with new branch, optionally from a base branch
	args := []string{"worktree", "add", worktreePath, "-b", branchName}
	if baseBranch != "" {
		args = append(args, baseBranch)
	}

	if _, err := RunInDir(projectRoot, args...); err != nil {
		return "", fmt.Errorf("failed to create worktree: %w", err)
	}

	return worktreePath, nil
}

// CreateWorktreeFromBranch creates a worktree from an existing branch
// The directory name is flattened (slashes become dashes)
func CreateWorktreeFromBranch(projectRoot, branchName string) (string, error) {
	// Flatten branch name for directory (e.g., feature/auth -> feature-auth)
	dirName := FlattenBranchName(branchName)
	worktreePath := filepath.Join(projectRoot, dirName)

	// Create worktree from existing branch
	if _, err := RunInDir(projectRoot, "worktree", "add", worktreePath, branchName); err != nil {
		return "", fmt.Errorf("failed to create worktree: %w", err)
	}

	return worktreePath, nil
}

// ListWorktrees lists all worktrees in the project
func ListWorktrees(projectRoot string) ([]Worktree, error) {
	output, err := RunInDir(projectRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	return parseWorktreeList(output), nil
}

// parseWorktreeList parses the porcelain output of git worktree list
func parseWorktreeList(output string) []Worktree {
	var worktrees []Worktree
	var current Worktree

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)

		if line == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			current.Path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "HEAD ") {
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		} else if strings.HasPrefix(line, "branch ") {
			branch := strings.TrimPrefix(line, "branch ")
			// Extract branch name from refs/heads/... (preserves slashes in names like feature/auth)
			current.Branch = strings.TrimPrefix(branch, "refs/heads/")
		}
	}

	// Don't forget the last one
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees
}

// RemoveWorktree removes a worktree
func RemoveWorktree(projectRoot, worktreePath string) error {
	if _, err := RunInDir(projectRoot, "worktree", "remove", worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}
	return nil
}

// RemoveWorktreeForce forcefully removes a worktree
func RemoveWorktreeForce(projectRoot, worktreePath string) error {
	if _, err := RunInDir(projectRoot, "worktree", "remove", "--force", worktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}
	return nil
}

// DeleteBranch deletes a local branch
func DeleteBranch(projectRoot, branchName string) error {
	if _, err := RunInDir(projectRoot, "branch", "-D", branchName); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}
	return nil
}

// PruneWorktrees removes stale worktree entries
func PruneWorktrees(projectRoot string) error {
	if _, err := RunInDir(projectRoot, "worktree", "prune"); err != nil {
		return fmt.Errorf("failed to prune worktrees: %w", err)
	}
	return nil
}

// GetWorktreeStatus returns the status of a worktree (clean, modified files count)
func GetWorktreeStatus(worktreePath string) (string, error) {
	output, err := RunInDir(worktreePath, "status", "--porcelain")
	if err != nil {
		return "unknown", nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if output == "" || len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return "clean", nil
	}

	return fmt.Sprintf("%d modified", len(lines)), nil
}
