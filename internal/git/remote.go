package git

import (
	"strings"
)

// RemoteBranch represents a branch on a remote
type RemoteBranch struct {
	Remote string
	Branch string
}

// parseRemoteBranches parses output of git branch -r --list
func parseRemoteBranches(output string) []RemoteBranch {
	var branches []RemoteBranch
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip HEAD pointers like "origin/HEAD -> origin/main"
		if strings.Contains(line, "->") {
			continue
		}
		// Parse "origin/feature/auth" into Remote="origin", Branch="feature/auth"
		parts := strings.SplitN(line, "/", 2)
		if len(parts) == 2 {
			branches = append(branches, RemoteBranch{
				Remote: parts[0],
				Branch: parts[1],
			})
		}
	}
	return branches
}

// FindRemoteBranches checks all remotes for a branch name
// Returns list of remotes that have this branch
func FindRemoteBranches(projectRoot, branchName string) ([]RemoteBranch, error) {
	// git branch -r --list "*/<branchName>"
	output, err := RunInDir(projectRoot, "branch", "-r", "--list", "*/"+branchName)
	if err != nil {
		return nil, err
	}
	return parseRemoteBranches(output), nil
}

// FetchAllRemotes fetches from all configured remotes
func FetchAllRemotes(projectRoot string) error {
	return RunWithProgress(projectRoot, "fetch", "--all", "--prune")
}

// LocalBranchExists checks if a local branch exists
func LocalBranchExists(projectRoot, branchName string) bool {
	_, err := RunInDir(projectRoot, "rev-parse", "--verify", "refs/heads/"+branchName)
	return err == nil
}
