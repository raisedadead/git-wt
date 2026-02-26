package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/raisedadead/wt/internal/git"
)

func loadProjectCmd() tea.Cmd {
	return func() tea.Msg {
		projectRoot, err := git.GetProjectRoot(".")
		if err != nil {
			return projectLoadedMsg{err: err}
		}
		defaultBranch := git.GetDefaultBranchName(projectRoot)
		return projectLoadedMsg{
			projectRoot:   projectRoot,
			defaultBranch: defaultBranch,
		}
	}
}

func loadWorktreesCmd(projectRoot string) tea.Cmd {
	return func() tea.Msg {
		worktrees, err := git.ListWorktrees(projectRoot)
		if err != nil {
			return worktreesLoadedMsg{err: err}
		}
		return worktreesLoadedMsg{worktrees: worktrees}
	}
}

func loadStatusCmd(wtPath string) tea.Cmd {
	return func() tea.Msg {
		status, err := git.GetWorktreeStatus(wtPath)

		var ahead, behind int
		if aheadStr, aErr := git.RunInDir(wtPath, "rev-list", "--count", "@{upstream}..HEAD"); aErr == nil {
			fmt.Sscanf(strings.TrimSpace(aheadStr), "%d", &ahead)
		}
		if behindStr, bErr := git.RunInDir(wtPath, "rev-list", "--count", "HEAD..@{upstream}"); bErr == nil {
			fmt.Sscanf(strings.TrimSpace(behindStr), "%d", &behind)
		}

		return statusLoadedMsg{
			path:   wtPath,
			status: status,
			err:    err,
			ahead:  ahead,
			behind: behind,
		}
	}
}

func loadDetailCmd(wtPath string) tea.Cmd {
	return func() tea.Msg {
		commits, _ := git.RunInDir(wtPath, "log", "--oneline", "-10")
		status, _ := git.GetWorktreeStatus(wtPath)
		files, _ := git.RunInDir(wtPath, "status", "--porcelain")
		return detailLoadedMsg{
			path:    wtPath,
			commits: commits,
			status:  status,
			files:   files,
		}
	}
}

func loadDiffCmd(wtPath string) tea.Cmd {
	return func() tea.Msg {
		diff, _ := git.RunInDir(wtPath, "diff")
		if diff == "" {
			diff, _ = git.RunInDir(wtPath, "diff", "--cached")
		}
		return diffLoadedMsg{
			path: wtPath,
			diff: diff,
		}
	}
}

func loadLogCmd(wtPath string) tea.Cmd {
	return func() tea.Msg {
		log, _ := git.RunInDir(wtPath, "log", "--oneline", "--graph", "-30")
		return logLoadedMsg{
			path: wtPath,
			log:  log,
		}
	}
}

func createWorktreeCmd(projectRoot, branchName string) tea.Cmd {
	return func() tea.Msg {
		path, err := git.CreateWorktree(projectRoot, branchName)
		return worktreeCreatedMsg{
			path:   path,
			branch: branchName,
			err:    err,
		}
	}
}

func deleteWorktreeCmd(projectRoot, wtPath, branch string, force bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if force {
			err = git.RemoveWorktreeForce(projectRoot, wtPath)
		} else {
			err = git.RemoveWorktree(projectRoot, wtPath)
		}
		if err != nil {
			return worktreeDeletedMsg{branch: branch, err: err}
		}
		_ = git.DeleteBranch(projectRoot, branch)
		return worktreeDeletedMsg{branch: branch}
	}
}

func pruneWorktreesCmd(projectRoot string) tea.Cmd {
	return func() tea.Msg {
		err := git.PruneWorktrees(projectRoot)
		return pruneMsg{err: err}
	}
}

func fetchCmd(projectRoot string) tea.Cmd {
	return func() tea.Msg {
		err := git.FetchAllRemotes(projectRoot)
		return fetchMsg{err: err}
	}
}

func clearFlashCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return clearFlashMsg{}
	})
}

func currentWorktreePath() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}
