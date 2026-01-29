package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/raisedadead/git-wt/internal/git"
	"github.com/raisedadead/git-wt/internal/ui"
	"github.com/spf13/cobra"
)

var (
	listJSONOutput bool
	pathOutput     bool
)

// ListData represents the JSON output for the list command
type ListData struct {
	Worktrees []worktreeInfo `json:"worktrees"`
	Count     int            `json:"count"`
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all worktrees",
	RunE:    runList,
}

func init() {
	listCmd.Flags().BoolVar(&listJSONOutput, "json", false, "Output as JSON (legacy, use global --json)")
	listCmd.Flags().BoolVar(&pathOutput, "path", false, "Output paths only")
	rootCmd.AddCommand(listCmd)
}

type worktreeInfo struct {
	Branch string `json:"branch"`
	Path   string `json:"path"`
	Status string `json:"status"`
}

func runList(cmd *cobra.Command, args []string) error {
	// Find project root
	projectRoot, err := git.GetProjectRoot(".")
	if err != nil {
		return fmt.Errorf("not in a git-wt project: %w", err)
	}

	worktrees, err := git.ListWorktrees(projectRoot)
	if err != nil {
		return err
	}

	// Build info with status (skip .bare directory)
	var infos []worktreeInfo
	for _, wt := range worktrees {
		// Skip the bare repository itself
		if strings.HasSuffix(wt.Path, "/.bare") || wt.Branch == "" {
			continue
		}
		status, _ := git.GetWorktreeStatus(wt.Path)
		infos = append(infos, worktreeInfo{
			Branch: wt.Branch,
			Path:   wt.Path,
			Status: status,
		})
	}

	// Output based on flags - check global --json first, then legacy list --json
	if IsJSONOutput() {
		data := ListData{
			Worktrees: infos,
			Count:     len(infos),
		}
		return ui.OutputJSON(os.Stdout, "list", data, nil)
	}

	// Legacy --json flag for backward compatibility
	if listJSONOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(infos)
	}

	if pathOutput {
		for _, info := range infos {
			fmt.Println(info.Path)
		}
		return nil
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, ui.BoldStyle.Render("BRANCH\tSTATUS\tPATH"))

	for _, info := range infos {
		statusStyle := ui.SuccessStyle
		if info.Status != "clean" {
			statusStyle = ui.SubtleStyle
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n",
			info.Branch,
			statusStyle.Render(info.Status),
			ui.SubtleStyle.Render(shortenPath(info.Path)),
		)
	}

	return w.Flush()
}

func shortenPath(path string) string {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
