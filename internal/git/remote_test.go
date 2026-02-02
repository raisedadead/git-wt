package git

import (
	"testing"
)

func TestParseRemoteBranches(t *testing.T) {
	// Simulates output of: git branch -r --list "*/<branch>"
	tests := []struct {
		name     string
		output   string
		expected []RemoteBranch
	}{
		{
			name:     "no matches",
			output:   "",
			expected: []RemoteBranch{},
		},
		{
			name:   "single remote",
			output: "  origin/feature/auth\n",
			expected: []RemoteBranch{
				{Remote: "origin", Branch: "feature/auth"},
			},
		},
		{
			name:   "multiple remotes",
			output: "  origin/feature/auth\n  upstream/feature/auth\n",
			expected: []RemoteBranch{
				{Remote: "origin", Branch: "feature/auth"},
				{Remote: "upstream", Branch: "feature/auth"},
			},
		},
		{
			name:   "with HEAD pointer ignored",
			output: "  origin/HEAD -> origin/main\n  origin/feature/auth\n",
			expected: []RemoteBranch{
				{Remote: "origin", Branch: "feature/auth"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRemoteBranches(tt.output)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d branches, got %d", len(tt.expected), len(result))
			}
			for i, rb := range result {
				if rb.Remote != tt.expected[i].Remote || rb.Branch != tt.expected[i].Branch {
					t.Errorf("at index %d: expected %+v, got %+v", i, tt.expected[i], rb)
				}
			}
		})
	}
}

func TestLocalBranchExists_NonexistentRepo(t *testing.T) {
	// Should return false for nonexistent directory
	exists := LocalBranchExists("/nonexistent", "main")
	if exists {
		t.Error("expected false for nonexistent repo")
	}
}
