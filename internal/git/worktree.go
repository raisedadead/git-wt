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
// Uses --relative-paths for portability (Git 2.36+)
func CreateWorktreeWithBase(projectRoot, branchName, baseBranch string) (string, error) {
	// Flatten branch name for directory (e.g., feature/auth -> feature-auth)
	dirName := FlattenBranchName(branchName)
	worktreePath := filepath.Join(projectRoot, dirName)

	// Create worktree with new branch, optionally from a base branch
	// Use --relative-paths so the repo can be moved without breaking paths
	args := []string{"worktree", "add", "--relative-paths", worktreePath, "-b", branchName}
	if baseBranch != "" {
		args = append(args, baseBranch)
	}

	if _, err := RunInDir(projectRoot, args...); err != nil {
		return "", fmt.Errorf("failed to create worktree: %w", err)
	}

	return worktreePath, nil
}

// CreateWorktreeFromRemote creates a worktree tracking a remote branch
// If local branch exists, checks it out; otherwise creates new branch tracking remote
func CreateWorktreeFromRemote(projectRoot, branchName, remote string) (string, error) {
	// Flatten branch name for directory (e.g., feature/auth -> feature-auth)
	dirName := FlattenBranchName(branchName)
	worktreePath := filepath.Join(projectRoot, dirName)

	var args []string
	if LocalBranchExists(projectRoot, branchName) {
		// Local branch exists - just check it out
		args = []string{"worktree", "add", "--relative-paths", worktreePath, branchName}
	} else {
		// Local branch doesn't exist - create new branch tracking remote
		remoteBranch := remote + "/" + branchName
		args = []string{"worktree", "add", "--relative-paths", worktreePath, "-b", branchName, "--track", remoteBranch}
	}

	if _, err := RunInDir(projectRoot, args...); err != nil {
		errStr := err.Error()
		// Check for stale worktree error
		if strings.Contains(errStr, "already registered worktree") {
			return "", fmt.Errorf("stale worktree entry exists. Run 'git worktree prune' first, then retry")
		}
		return "", fmt.Errorf("failed to create worktree: %w", err)
	}

	return worktreePath, nil
}

// CreateWorktreeFromBranch creates a worktree from an existing branch
// The directory name is flattened (slashes become dashes)
// Uses --relative-paths for portability (Git 2.36+)
func CreateWorktreeFromBranch(projectRoot, branchName string) (string, error) {
	// Flatten branch name for directory (e.g., feature/auth -> feature-auth)
	dirName := FlattenBranchName(branchName)
	worktreePath := filepath.Join(projectRoot, dirName)

	// Create worktree from existing branch
	// Use --relative-paths so the repo can be moved without breaking paths
	if _, err := RunInDir(projectRoot, "worktree", "add", "--relative-paths", worktreePath, branchName); err != nil {
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

// RepairWorktrees repairs worktree paths after a repository has been moved
func RepairWorktrees(projectRoot string) (string, error) {
	output, err := RunInDir(projectRoot, "worktree", "repair")
	if err != nil {
		return "", fmt.Errorf("failed to repair worktrees: %w", err)
	}
	return output, nil
}

// IsBranchMerged checks if a branch has been merged into the target branch
// Uses git merge-base --is-ancestor to check if branch is an ancestor of target
func IsBranchMerged(projectRoot, branch, targetBranch string) bool {
	_, err := RunInDir(projectRoot, "merge-base", "--is-ancestor", branch, targetBranch)
	return err == nil
}

// GetDefaultBranchName returns the default branch (main or master)
func GetDefaultBranchName(projectRoot string) string {
	// Check if main exists
	if _, err := RunInDir(projectRoot, "rev-parse", "--verify", "refs/heads/main"); err == nil {
		return "main"
	}
	// Fall back to master
	if _, err := RunInDir(projectRoot, "rev-parse", "--verify", "refs/heads/master"); err == nil {
		return "master"
	}
	// Default to main
	return "main"
}

// HasBranchUpstream checks if a branch has an upstream tracking branch configured
func HasBranchUpstream(projectRoot, branch string) bool {
	_, err := RunInDir(projectRoot, "config", "--get", "branch."+branch+".remote")
	return err == nil
}

// GetCommitsAhead returns the number of commits the branch has ahead of the target branch
// Returns 0 if there's an error or if the branch has no commits ahead
func GetCommitsAhead(projectRoot, branch, targetBranch string) (int, error) {
	output, err := RunInDir(projectRoot, "rev-list", "--count", targetBranch+".."+branch)
	if err != nil {
		return 0, err
	}

	var count int
	_, err = fmt.Sscanf(strings.TrimSpace(output), "%d", &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetCommitsBehind returns the number of commits the branch is behind the target branch
func GetCommitsBehind(projectRoot, branch, targetBranch string) (int, error) {
	output, err := RunInDir(projectRoot, "rev-list", "--count", branch+".."+targetBranch)
	if err != nil {
		return 0, err
	}

	var count int
	_, err = fmt.Sscanf(strings.TrimSpace(output), "%d", &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// IsTrulyMerged checks if a branch has been truly merged into the target branch
// A branch is truly merged if:
// 1. It has no commits ahead of the target branch (all its work is in target)
// 2. It is BEHIND the target branch (target has moved on since the branch was created)
// 3. It is an ancestor of the target branch (merge-base check)
// This avoids false positives for newly created branches that are at the same point as main
func IsTrulyMerged(projectRoot, branch, targetBranch string) bool {
	// First check: does the branch have any commits ahead of target?
	commitsAhead, err := GetCommitsAhead(projectRoot, branch, targetBranch)
	if err != nil {
		return false
	}

	// If branch has commits ahead, it's not merged
	if commitsAhead > 0 {
		return false
	}

	// Second check: is the branch behind target?
	// If branch is at the same commit as target, it's not "merged" - it just hasn't diverged
	commitsBehind, err := GetCommitsBehind(projectRoot, branch, targetBranch)
	if err != nil {
		return false
	}

	// Branch must be behind target to be considered "merged"
	// (0 ahead, 0 behind = at same commit = not merged, just fresh)
	if commitsBehind == 0 {
		return false
	}

	// Third check: is the branch an ancestor of target? (standard merge check)
	// This confirms the branch's commits are actually in the target
	return IsBranchMerged(projectRoot, branch, targetBranch)
}
