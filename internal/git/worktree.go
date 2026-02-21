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
	IsBare bool
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

// parseWorktreeList parses the porcelain output of git worktree list,
// filtering out bare repo entries
func parseWorktreeList(output string) []Worktree {
	all := parseWorktreeListAll(output)
	var worktrees []Worktree
	for _, wt := range all {
		if !wt.IsBare {
			worktrees = append(worktrees, wt)
		}
	}
	return worktrees
}

// parseWorktreeListAll parses the porcelain output of git worktree list
// without filtering, including bare repo entries
func parseWorktreeListAll(output string) []Worktree {
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
			current.Branch = strings.TrimPrefix(branch, "refs/heads/")
		} else if line == "detached" {
			if current.Commit != "" {
				shortCommit := current.Commit
				if len(shortCommit) > 7 {
					shortCommit = shortCommit[:7]
				}
				current.Branch = "HEAD detached at " + shortCommit
			} else {
				current.Branch = "HEAD detached"
			}
		} else if line == "bare" {
			current.IsBare = true
		}
	}

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

// IsTrulyMerged checks if a branch has been truly merged into the target branch.
// A branch is truly merged if it had unique commits that are now in the target.
// Branches that never diverged from the target are NOT considered merged —
// they are just stale (main moved ahead while the branch sat idle).
func IsTrulyMerged(projectRoot, branch, targetBranch string) bool {
	// Quick check: branch must be an ancestor of target
	if !IsBranchMerged(projectRoot, branch, targetBranch) {
		return false
	}

	// Branch must have no unmerged commits
	commitsAhead, err := GetCommitsAhead(projectRoot, branch, targetBranch)
	if err != nil || commitsAhead > 0 {
		return false
	}

	// If branch is at the same commit as target, it hasn't diverged — not merged
	commitsBehind, err := GetCommitsBehind(projectRoot, branch, targetBranch)
	if err != nil || commitsBehind == 0 {
		return false
	}

	// Key distinction: is the branch tip on the target's first-parent line?
	// If yes, the branch never diverged — it's just a stale pointer on the
	// main line (e.g., branched off main, never committed, main moved on).
	// If no, the branch tip was brought in via a merge commit's second parent,
	// meaning it was genuinely merged.
	branchTip, err := RunInDir(projectRoot, "rev-parse", branch)
	if err != nil {
		return false
	}
	branchTip = strings.TrimSpace(branchTip)

	// Only scan commits between branch and target (not the entire history)
	output, err := RunInDir(projectRoot, "log", "--first-parent", "--format=%H", branchTip+".."+targetBranch)
	if err != nil {
		return false
	}

	// Also check if the branch tip itself is the first parent of any commit in the range.
	// If it appears as a first-parent ancestor, the branch never diverged.
	// We check by looking at the parent of the oldest commit in our range.
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 0 && lines[0] != "" {
		// Get the parents of the oldest first-parent commit in range
		oldest := lines[len(lines)-1]
		parentOutput, err := RunInDir(projectRoot, "rev-parse", oldest+"^")
		if err == nil && strings.TrimSpace(parentOutput) == branchTip {
			// Branch tip is the first parent of the next commit — it's on the
			// first-parent line. Never diverged, just stale.
			return false
		}
	}

	// Branch tip is NOT on the first-parent line but IS an ancestor of target.
	// It was brought in via a merge commit — genuinely merged.
	return true
}
