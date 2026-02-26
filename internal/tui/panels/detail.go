package panels

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/ui"
)

var (
	dtTabActive   = lipgloss.NewStyle().Bold(true).Foreground(ui.AdaptiveTabActive).Underline(true).Padding(0, 1)
	dtTabInactive = lipgloss.NewStyle().Foreground(ui.AdaptiveTabInactive).Padding(0, 1)
	dtTabBarLine  = lipgloss.NewStyle().Foreground(ui.AdaptiveTabBar)
	dtDiffAdd     = lipgloss.NewStyle().Foreground(ui.AdaptiveDiffAdd)
	dtDiffRemove  = lipgloss.NewStyle().Foreground(ui.AdaptiveDiffRemove)
	dtDiffHunk    = lipgloss.NewStyle().Foreground(ui.AdaptiveDiffHunk)
	dtDiffMeta    = lipgloss.NewStyle().Bold(true).Foreground(ui.AdaptiveText)
	dtLogHash     = lipgloss.NewStyle().Foreground(ui.AdaptiveWarning)
	dtLogGraph    = lipgloss.NewStyle().Foreground(ui.AdaptivePrimary)
	dtInfoLabel   = lipgloss.NewStyle().Bold(true).Foreground(ui.AdaptiveSubtle).Width(10)
	dtInfoValue   = lipgloss.NewStyle().Foreground(ui.AdaptiveText)
	dtPathValue   = lipgloss.NewStyle().Foreground(ui.AdaptiveSubtle)
	dtSection     = lipgloss.NewStyle().Bold(true).Foreground(ui.AdaptivePrimary)
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
}

func NewDetailPanel() DetailPanel {
	vp := viewport.New(0, 0)
	vp.SetContent("")

	return DetailPanel{
		Viewport: vp,
		Tab:      TabInfo,
	}
}

func (dp *DetailPanel) SetSize(width, height int) {
	dp.Width = width
	dp.Height = height
	// Subtract 2 for tab bar + separator line
	vpHeight := height - 2
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

	// Shorten path: show last 2 components
	shortPath := dp.Path
	if parts := strings.Split(dp.Path, "/"); len(parts) > 2 {
		shortPath = "\u2026/" + strings.Join(parts[len(parts)-2:], "/")
	}

	fmt.Fprintf(&b, "  %s%s\n", dtInfoLabel.Render("Branch"), dtInfoValue.Render(dp.Branch))
	fmt.Fprintf(&b, "  %s%s\n", dtInfoLabel.Render("Path"), dtPathValue.Render(shortPath))
	fmt.Fprintf(&b, "  %s%s\n", dtInfoLabel.Render("Status"), dtInfoValue.Render(dp.Status))

	sepWidth := dp.Width - 4
	if sepWidth < 1 {
		sepWidth = 1
	}
	sep := dtTabBarLine.Render(strings.Repeat("\u2500", sepWidth))

	if dp.Commits != "" {
		b.WriteString("\n")
		fmt.Fprintf(&b, "  %s\n", dtSection.Render("Recent Commits"))
		fmt.Fprintf(&b, "  %s\n", sep)
		for _, line := range strings.Split(dp.Commits, "\n") {
			if line != "" {
				fmt.Fprintf(&b, "  %s\n", line)
			}
		}
	}

	if dp.Files != "" {
		b.WriteString("\n")
		fmt.Fprintf(&b, "  %s\n", dtSection.Render("Modified Files"))
		fmt.Fprintf(&b, "  %s\n", sep)
		for _, line := range strings.Split(dp.Files, "\n") {
			if line != "" {
				fmt.Fprintf(&b, "  %s\n", line)
			}
		}
	} else if dp.Status == "clean" {
		b.WriteString("\n")
		fmt.Fprintf(&b, "  %s\n", dtSection.Render("Modified Files"))
		fmt.Fprintf(&b, "  %s\n", sep)
		b.WriteString("  (no uncommitted changes)\n")
	}

	return b.String()
}

func (dp *DetailPanel) renderDiff() string {
	if dp.Diff == "" {
		return "\n  No diff available (working tree clean)"
	}

	var b strings.Builder
	for _, line := range strings.Split(dp.Diff, "\n") {
		var styled string
		switch {
		case strings.HasPrefix(line, "+++"), strings.HasPrefix(line, "---"),
			strings.HasPrefix(line, "diff "), strings.HasPrefix(line, "index "):
			styled = dtDiffMeta.Render(line)
		case strings.HasPrefix(line, "@@"):
			styled = dtDiffHunk.Render(line)
		case strings.HasPrefix(line, "+"):
			styled = dtDiffAdd.Render(line)
		case strings.HasPrefix(line, "-"):
			styled = dtDiffRemove.Render(line)
		default:
			styled = line
		}
		fmt.Fprintf(&b, "  %s\n", styled)
	}
	return b.String()
}

func isGraphChar(r rune) bool {
	return r == '*' || r == '|' || r == '/' || r == '\\' || r == '_' || r == ' '
}

func isHexHash(s string) bool {
	if len(s) < 7 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

func (dp *DetailPanel) renderLog() string {
	if dp.Log == "" {
		return "\n  Loading log..."
	}

	var b strings.Builder
	for _, line := range strings.Split(dp.Log, "\n") {
		if line == "" {
			b.WriteString("\n")
			continue
		}

		graphEnd := 0
		for i, r := range line {
			if !isGraphChar(r) {
				graphEnd = i
				break
			}
			if i == len(line)-1 {
				graphEnd = len(line)
			}
		}

		graphPart := line[:graphEnd]
		rest := line[graphEnd:]

		var rendered string
		if graphPart != "" {
			rendered = dtLogGraph.Render(graphPart)
		}

		rest = strings.TrimLeftFunc(rest, unicode.IsSpace)
		if rest != "" {
			spaceIdx := strings.IndexByte(rest, ' ')
			var firstWord, remainder string
			if spaceIdx > 0 {
				firstWord = rest[:spaceIdx]
				remainder = rest[spaceIdx:]
			} else {
				firstWord = rest
			}

			if isHexHash(firstWord) {
				rendered += dtLogHash.Render(firstWord) + remainder
			} else {
				rendered += rest
			}
		}

		fmt.Fprintf(&b, "  %s\n", rendered)
	}
	return b.String()
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
	infoTab := dtTabInactive.Render("Info")
	diffTab := dtTabInactive.Render("Diff")
	logTab := dtTabInactive.Render("Log")

	switch dp.Tab {
	case TabInfo:
		infoTab = dtTabActive.Render("Info")
	case TabDiff:
		diffTab = dtTabActive.Render("Diff")
	case TabLog:
		logTab = dtTabActive.Render("Log")
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Top, infoTab, diffTab, logTab)
	sepWidth := dp.Width
	if sepWidth < 1 {
		sepWidth = 1
	}
	separator := dtTabBarLine.Render(strings.Repeat("\u2500", sepWidth))
	return tabs + "\n" + separator
}
