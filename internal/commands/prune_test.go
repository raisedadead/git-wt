package commands

import (
	"testing"
)

func TestPruneConfirmationTitle(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{"singular", 1, "Remove 1 stale worktree?"},
		{"plural two", 2, "Remove 2 stale worktrees?"},
		{"plural three", 3, "Remove 3 stale worktrees?"},
		{"plural many", 10, "Remove 10 stale worktrees?"},
		{"zero", 0, "Remove 0 stale worktrees?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pruneConfirmationTitle(tt.count)
			if result != tt.expected {
				t.Errorf("pruneConfirmationTitle(%d) = %q, want %q", tt.count, result, tt.expected)
			}
		})
	}
}

func TestPruneConfirmationDescription(t *testing.T) {
	tests := []struct {
		name     string
		branches []string
		expected string
	}{
		{
			"single branch",
			[]string{"feature/old"},
			"Branches: feature/old",
		},
		{
			"two branches",
			[]string{"feature/old", "fix/done"},
			"Branches: feature/old, fix/done",
		},
		{
			"three branches",
			[]string{"feature/old", "fix/done", "temp/test"},
			"Branches: feature/old, fix/done, temp/test",
		},
		{
			"five branches exact",
			[]string{"a", "b", "c", "d", "e"},
			"Branches: a, b, c, d, e",
		},
		{
			"six branches truncated",
			[]string{"a", "b", "c", "d", "e", "f"},
			"Branches: a, b, c, d, e and 1 more...",
		},
		{
			"ten branches truncated",
			[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
			"Branches: a, b, c, d, e and 5 more...",
		},
		{
			"empty",
			[]string{},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pruneConfirmationDescription(tt.branches)
			if result != tt.expected {
				t.Errorf("pruneConfirmationDescription(%v) = %q, want %q", tt.branches, result, tt.expected)
			}
		})
	}
}
