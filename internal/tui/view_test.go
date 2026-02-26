package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/tui/panels"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// testModel creates a model pre-loaded with fake worktree data,
// bypassing Init() which would make real git calls.
func testModel(width, height int) model {
	m := newModel()
	m.loading = false
	m.projectRoot = "/Users/test/DEV/myproject"
	m.defaultBranch = "main"
	m.currentPath = "/Users/test/DEV/myproject/main"

	// Simulate window size
	sized, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	m = sized.(model)

	// Set up fake worktrees
	worktrees := []git.Worktree{
		{Branch: "feat-TUI", Path: "/Users/test/DEV/myproject/feat-TUI"},
		{Branch: "fix/integrations", Path: "/Users/test/DEV/myproject/fix-integrations"},
		{Branch: "main", Path: "/Users/test/DEV/myproject/main"},
	}
	m.setWorktrees(worktrees)
	m.header.WorktreeCount = len(worktrees)
	m.header.ProjectRoot = m.projectRoot
	m.header.DefaultBranch = m.defaultBranch

	// Simulate status loaded for each
	m.worktrees.UpdateItemStatus(worktrees[0].Path, "16 modified", 3, 0)
	m.worktrees.UpdateItemStatus(worktrees[1].Path, "5 modified", 0, 2)
	m.worktrees.UpdateItemStatus(worktrees[2].Path, "clean", 0, 0)

	// Load detail for selected item
	if item := m.worktrees.SelectedItem(); item != nil {
		m.detail.SetInfo(item.Wt.Branch, item.Wt.Path, "16 modified",
			"346576e feat: more aliases\nba57942 feat: add more aliases",
			"M internal/tui/model.go\nM internal/tui/styles.go")
	}

	return m
}

func TestViewLayout(t *testing.T) {
	m := testModel(120, 40)
	view := m.View()

	// Strip ANSI for content checks
	plain := ansi.Strip(view)
	lines := strings.Split(plain, "\n")

	t.Logf("View has %d lines (expected ~40 for height=40)", len(lines))
	t.Logf("\n--- RAW VIEW (stripped ANSI) ---\n%s\n--- END ---", plain)

	// Header should contain project info
	if !strings.Contains(plain, "wt") {
		t.Error("header missing 'wt' project name")
	}
	if !strings.Contains(plain, "main") {
		t.Error("header missing default branch 'main'")
	}
	if !strings.Contains(plain, "3 worktrees") {
		t.Error("header missing worktree count")
	}

	// Border titles should be embedded
	if !strings.Contains(plain, "Worktrees") {
		t.Error("missing 'Worktrees' border title")
	}
	if !strings.Contains(plain, "Detail") {
		t.Error("missing 'Detail' border title")
	}

	// Worktree items should be visible
	if !strings.Contains(plain, "feat-TUI") {
		t.Error("missing worktree 'feat-TUI'")
	}
	if !strings.Contains(plain, "fix/integrations") {
		t.Error("missing worktree 'fix/integrations'")
	}

	// Cursor bar should be on the selected item (first item)
	// Unicode left bar: ▎ (U+258E)
	if !strings.Contains(view, "\u258e") {
		t.Error("missing cursor bar character on selected item")
	}

	// Status indicators should be present
	if !strings.Contains(plain, "16 modified") {
		t.Error("missing status '16 modified'")
	}

	// Detail panel should show info
	if !strings.Contains(plain, "Branch") {
		t.Error("detail panel missing 'Branch' label")
	}
}

func TestViewCursorMovement(t *testing.T) {
	m := testModel(120, 40)

	// Initial: cursor on index 0 (feat-TUI)
	view1 := m.View()
	plain1 := ansi.Strip(view1)

	// Move cursor down
	moved, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = moved.(model)

	view2 := m.View()
	plain2 := ansi.Strip(view2)

	// Both views should have the cursor bar, but the bar should have moved
	if !strings.Contains(view1, "\u258e") {
		t.Error("view1 missing cursor bar")
	}
	if !strings.Contains(view2, "\u258e") {
		t.Error("view2 missing cursor bar")
	}

	// Views should be different (cursor moved)
	if plain1 == plain2 {
		t.Error("view did not change after cursor movement")
	}

	t.Logf("\n--- AFTER j (cursor down) ---\n%s\n--- END ---", plain2)
}

func TestViewPanelResize(t *testing.T) {
	m := testModel(120, 40)

	initialLeft := m.leftWidth

	// Resize left panel wider
	resized, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("L")})
	m = resized.(model)

	if m.leftWidth <= initialLeft {
		t.Errorf("expected leftWidth to increase: was %d, now %d", initialLeft, m.leftWidth)
	}

	view := m.View()
	plain := ansi.Strip(view)
	t.Logf("\n--- AFTER L (resize right) ---\n%s\n--- END ---", plain)
}

func TestViewFocusSwitch(t *testing.T) {
	m := testModel(120, 40)

	// Initial: list focused
	if m.focused != panelList {
		t.Fatal("expected initial focus on list")
	}

	// Tab to detail
	tabbed, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = tabbed.(model)

	if m.focused != panelDetail {
		t.Fatal("expected focus on detail after tab")
	}

	view := m.View()
	plain := ansi.Strip(view)
	t.Logf("\n--- AFTER Tab (detail focused) ---\n%s\n--- END ---", plain)
}

func TestViewVimNavigation(t *testing.T) {
	m := testModel(120, 40)

	// l → detail panel
	moved, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m = moved.(model)
	if m.focused != panelDetail {
		t.Fatal("expected focus on detail after 'l'")
	}

	// h → back to list
	moved, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m = moved.(model)
	if m.focused != panelList {
		t.Fatal("expected focus on list after 'h'")
	}

	// l → detail, then ] to cycle tab
	moved, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m = moved.(model)

	if m.detail.Tab != panels.TabInfo {
		t.Fatal("expected initial tab to be Info")
	}

	moved, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")})
	m = moved.(model)
	if m.detail.Tab != panels.TabDiff {
		t.Fatalf("expected Diff tab after ], got %d", m.detail.Tab)
	}

	moved, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("[")})
	m = moved.(model)
	if m.detail.Tab != panels.TabInfo {
		t.Fatalf("expected Info tab after [, got %d", m.detail.Tab)
	}
}

func TestViewOverlays(t *testing.T) {
	m := testModel(120, 40)

	// Open help overlay
	helped, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = helped.(model)

	if !m.helpOverlay.Active {
		t.Fatal("help overlay should be active")
	}

	view := m.View()
	plain := ansi.Strip(view)

	if !strings.Contains(plain, "Keybindings") {
		t.Error("help overlay should show 'Keybindings'")
	}

	t.Logf("\n--- HELP OVERLAY ---\n%s\n--- END ---", plain)
}

func TestViewSmallTerminal(t *testing.T) {
	m := testModel(40, 15)
	view := m.View()
	plain := ansi.Strip(view)
	lines := strings.Split(plain, "\n")

	t.Logf("Small terminal (%dx%d): %d lines", 40, 15, len(lines))
	t.Logf("\n--- SMALL TERMINAL ---\n%s\n--- END ---", plain)

	// Should still render without panic
	if len(plain) == 0 {
		t.Error("empty view on small terminal")
	}
}

func TestViewWorktreeListRendering(t *testing.T) {
	// Test the worktree delegate directly
	wl := panels.NewWorktreeList()
	wl.SetSize(40, 20)
	wl.SetFocused(true)

	items := []panels.WorktreeItem{
		{
			Wt:        git.Worktree{Branch: "feat-TUI", Path: "/tmp/feat-TUI"},
			Status:    "16 modified",
			Dirty:     true,
			Current:   true,
			Ahead:     3,
			Behind:    0,
			IsDefault: false,
		},
		{
			Wt:        git.Worktree{Branch: "main", Path: "/tmp/main"},
			Status:    "clean",
			Dirty:     false,
			Current:   false,
			Ahead:     0,
			Behind:    0,
			IsDefault: true,
		},
	}
	wl.SetItems(items)

	view := wl.View()
	plain := ansi.Strip(view)

	t.Logf("\n--- WORKTREE LIST ---\n%s\n--- END ---", plain)

	// First item should have cursor bar (selected by default)
	if !strings.Contains(view, "\u258e") {
		t.Error("missing cursor bar on first item")
	}

	// Should show status
	if !strings.Contains(plain, "16 modified") {
		t.Error("missing status text")
	}

	// Should show current marker text
	if !strings.Contains(plain, "(current)") {
		t.Error("missing (current) marker on line 2")
	}

	// Default branch should have dot icon
	if !strings.Contains(view, "\u25cf") {
		t.Error("missing default branch icon")
	}
}

// TestViewDump is a helper test that dumps the full rendered view.
// Run with: go test -run TestViewDump -v ./internal/tui/
func TestViewDump(t *testing.T) {
	sizes := []struct{ w, h int }{
		{120, 40},
		{80, 24},
		{160, 50},
	}

	for _, sz := range sizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			m := testModel(sz.w, sz.h)
			view := m.View()
			plain := ansi.Strip(view)
			t.Logf("\n=== %dx%d ===\n%s\n=== END ===", sz.w, sz.h, plain)
		})
	}
}

// --- Golden file snapshot tests ---
// Run with -update to regenerate: go test -run TestGolden -update ./internal/tui/

func TestGoldenDefault(t *testing.T) {
	m := testModel(120, 40)
	view := m.View()
	golden.RequireEqual(t, []byte(ansi.Strip(view)))
}

func TestGoldenSmall(t *testing.T) {
	m := testModel(80, 24)
	view := m.View()
	golden.RequireEqual(t, []byte(ansi.Strip(view)))
}

func TestGoldenWide(t *testing.T) {
	m := testModel(160, 50)
	view := m.View()
	golden.RequireEqual(t, []byte(ansi.Strip(view)))
}

func TestGoldenCursorDown(t *testing.T) {
	m := testModel(120, 40)
	moved, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = moved.(model)
	view := m.View()
	golden.RequireEqual(t, []byte(ansi.Strip(view)))
}

func TestGoldenDetailFocused(t *testing.T) {
	m := testModel(120, 40)
	tabbed, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = tabbed.(model)
	view := m.View()
	golden.RequireEqual(t, []byte(ansi.Strip(view)))
}

func TestGoldenHelpOverlay(t *testing.T) {
	m := testModel(120, 40)
	helped, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = helped.(model)
	view := m.View()
	golden.RequireEqual(t, []byte(ansi.Strip(view)))
}

func TestGoldenResized(t *testing.T) {
	m := testModel(120, 40)
	// Grow left panel twice
	resized, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("L")})
	m = resized.(model)
	resized, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("L")})
	m = resized.(model)
	view := m.View()
	golden.RequireEqual(t, []byte(ansi.Strip(view)))
}
