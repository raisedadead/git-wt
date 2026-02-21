package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func TestNewTable(t *testing.T) {
	tbl := NewTable().
		Headers("Name", "Status").
		Row("main", "active").
		Row("feature", "clean")

	output := tbl.String()

	if output == "" {
		t.Fatal("expected non-empty table output")
	}

	// Rounded border uses ╭ ╮ ╰ ╯ characters
	if !strings.Contains(output, "╭") {
		t.Error("expected rounded border (╭), got:\n" + output)
	}
	if !strings.Contains(output, "╰") {
		t.Error("expected rounded border (╰), got:\n" + output)
	}

	// Headers and data should be present
	if !strings.Contains(output, "Name") {
		t.Error("expected header 'Name' in output")
	}
	if !strings.Contains(output, "main") {
		t.Error("expected row data 'main' in output")
	}
	if !strings.Contains(output, "feature") {
		t.Error("expected row data 'feature' in output")
	}
}

func TestNewStyledTable(t *testing.T) {
	called := make(map[int]bool)

	tbl := NewStyledTable(func(row, col int) lipgloss.Style {
		called[row] = true
		if row == table.HeaderRow {
			return lipgloss.NewStyle().Bold(true).Padding(0, 1)
		}
		return lipgloss.NewStyle().Padding(0, 1)
	}).
		Headers("A", "B").
		Row("x", "y")

	output := tbl.String()

	if output == "" {
		t.Fatal("expected non-empty table output")
	}

	// StyleFunc should have been called for header row
	if !called[table.HeaderRow] {
		t.Error("expected StyleFunc to be called for header row")
	}

	// Should still have rounded borders
	if !strings.Contains(output, "╭") {
		t.Error("expected rounded border in styled table")
	}
}

func TestNewTableAutoSizes(t *testing.T) {
	short := NewTable().
		Headers("A", "B").
		Row("x", "y").
		String()

	long := NewTable().
		Headers("A", "B").
		Row("a very long cell value here", "y").
		String()

	// Longer content should produce wider table
	shortWidth := len(strings.Split(short, "\n")[0])
	longWidth := len(strings.Split(long, "\n")[0])

	if longWidth <= shortWidth {
		t.Errorf("expected auto-sized table to be wider with longer content: short=%d, long=%d", shortWidth, longWidth)
	}
}
