package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/ui"
)

// DetailTab represents which tab is active in the detail pane.
type DetailTab int

const (
	TabInfo DetailTab = iota
	TabDiff
	TabLog
)

// DetailPanel shows info/diff/log for the selected worktree.
type DetailPanel struct {
	Viewport viewport.Model
	Focused  bool
	Tab      DetailTab
	Width    int
	Height   int

	// Content for each tab
	Branch  string
	Path    string
	Status  string
	Commits string
	Files   string
	Diff    string
	Log     string

	tabStyles struct {
		active   lipgloss.Style
		inactive lipgloss.Style
		label    lipgloss.Style
		value    lipgloss.Style
		section  lipgloss.Style
	}
}

func NewDetailPanel() DetailPanel {
	vp := viewport.New(0, 0)
	vp.SetContent("")

	dp := DetailPanel{
		Viewport: vp,
		Tab:      TabInfo,
	}
	dp.tabStyles.active = lipgloss.NewStyle().Bold(true).Foreground(ui.AdaptivePrimary).Underline(true).Padding(0, 1)
	dp.tabStyles.inactive = lipgloss.NewStyle().Foreground(ui.AdaptiveSubtle).Padding(0, 1)
	dp.tabStyles.label = lipgloss.NewStyle().Bold(true).Foreground(ui.AdaptiveSubtle)
	dp.tabStyles.value = lipgloss.NewStyle().Foreground(ui.AdaptiveText)
	dp.tabStyles.section = lipgloss.NewStyle().Foreground(ui.AdaptiveSubtle).Bold(true)

	return dp
}

func (dp *DetailPanel) SetSize(width, height int) {
	dp.Width = width
	dp.Height = height
	// Subtract 1 for tab bar
	vpHeight := height - 1
	if vpHeight < 0 {
		vpHeight = 0
	}
	dp.Viewport.Width = width
	dp.Viewport.Height = vpHeight
	dp.updateContent()
}

func (dp *DetailPanel) SetFocused(focused bool) {
	dp.Focused = focused
}

func (dp *DetailPanel) SetTab(tab DetailTab) {
	dp.Tab = tab
	dp.updateContent()
}

func (dp *DetailPanel) SetInfo(branch, path, status, commits, files string) {
	dp.Branch = branch
	dp.Path = path
	dp.Status = status
	dp.Commits = commits
	dp.Files = files
	if dp.Tab == TabInfo {
		dp.updateContent()
	}
}

func (dp *DetailPanel) SetDiff(diff string) {
	dp.Diff = diff
	if dp.Tab == TabDiff {
		dp.updateContent()
	}
}

func (dp *DetailPanel) SetLog(log string) {
	dp.Log = log
	if dp.Tab == TabLog {
		dp.updateContent()
	}
}

func (dp *DetailPanel) Clear() {
	dp.Branch = ""
	dp.Path = ""
	dp.Status = ""
	dp.Commits = ""
	dp.Files = ""
	dp.Diff = ""
	dp.Log = ""
	dp.Viewport.SetContent("")
}

func (dp *DetailPanel) updateContent() {
	var content string
	switch dp.Tab {
	case TabInfo:
		content = dp.renderInfo()
	case TabDiff:
		content = dp.renderDiff()
	case TabLog:
		content = dp.renderLog()
	}
	dp.Viewport.SetContent(content)
	dp.Viewport.GotoTop()
}

func (dp *DetailPanel) renderInfo() string {
	if dp.Branch == "" {
		return "\n  No worktree selected"
	}

	var b strings.Builder

	fmt.Fprintf(&b, "  %s  %s\n",
		dp.tabStyles.label.Render("Branch:"),
		dp.tabStyles.value.Render(dp.Branch))
	fmt.Fprintf(&b, "  %s    %s\n",
		dp.tabStyles.label.Render("Path:"),
		dp.tabStyles.value.Render(dp.Path))
	fmt.Fprintf(&b, "  %s  %s\n",
		dp.tabStyles.label.Render("Status:"),
		dp.tabStyles.value.Render(dp.Status))

	if dp.Commits != "" {
		b.WriteString("\n")
		fmt.Fprintf(&b, "  %s\n", dp.tabStyles.section.Render("── Recent commits ──────────────"))
		for _, line := range strings.Split(dp.Commits, "\n") {
			if line != "" {
				b.WriteString("  " + line + "\n")
			}
		}
	}

	if dp.Files != "" {
		b.WriteString("\n")
		fmt.Fprintf(&b, "  %s\n", dp.tabStyles.section.Render("── Modified files ──────────────"))
		for _, line := range strings.Split(dp.Files, "\n") {
			if line != "" {
				b.WriteString("  " + line + "\n")
			}
		}
	} else if dp.Status == "clean" {
		b.WriteString("\n")
		fmt.Fprintf(&b, "  %s\n", dp.tabStyles.section.Render("── Modified files ──────────────"))
		b.WriteString("  (no uncommitted changes)\n")
	}

	return b.String()
}

func (dp *DetailPanel) renderDiff() string {
	if dp.Diff == "" {
		return "\n  No diff available (working tree clean)"
	}
	return dp.Diff
}

func (dp *DetailPanel) renderLog() string {
	if dp.Log == "" {
		return "\n  Loading log..."
	}
	return dp.Log
}

func (dp *DetailPanel) Update(msg tea.Msg) (DetailPanel, tea.Cmd) {
	var cmd tea.Cmd
	dp.Viewport, cmd = dp.Viewport.Update(msg)
	return *dp, cmd
}

func (dp *DetailPanel) View() string {
	tabs := dp.renderTabs()
	return tabs + "\n" + dp.Viewport.View()
}

func (dp *DetailPanel) renderTabs() string {
	infoTab := dp.tabStyles.inactive.Render("Info")
	diffTab := dp.tabStyles.inactive.Render("Diff")
	logTab := dp.tabStyles.inactive.Render("Log")

	switch dp.Tab {
	case TabInfo:
		infoTab = dp.tabStyles.active.Render("Info")
	case TabDiff:
		diffTab = dp.tabStyles.active.Render("Diff")
	case TabLog:
		logTab = dp.tabStyles.active.Render("Log")
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, infoTab, diffTab, logTab)
}
