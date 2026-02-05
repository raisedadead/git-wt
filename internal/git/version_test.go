package git

import (
	"testing"
)

func TestParseGitVersion(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMajor   int
		wantMinor   int
		expectError bool
	}{
		{
			name:      "standard format",
			input:     "git version 2.39.0",
			wantMajor: 2,
			wantMinor: 39,
		},
		{
			name:      "with platform suffix",
			input:     "git version 2.36.1 (Apple Git-133)",
			wantMajor: 2,
			wantMinor: 36,
		},
		{
			name:      "windows format",
			input:     "git version 2.40.0.windows.1",
			wantMajor: 2,
			wantMinor: 40,
		},
		{
			name:      "old version",
			input:     "git version 1.8.5",
			wantMajor: 1,
			wantMinor: 8,
		},
		{
			name:        "invalid format",
			input:       "not git output",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, err := parseGitVersion(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if major != tt.wantMajor {
				t.Errorf("major version = %d, want %d", major, tt.wantMajor)
			}
			if minor != tt.wantMinor {
				t.Errorf("minor version = %d, want %d", minor, tt.wantMinor)
			}
		})
	}
}

func TestCheckGitVersion(t *testing.T) {
	// This test runs against the actual installed git
	// It should pass since we require git 2.20+ and most systems have newer
	err := CheckGitVersion(2, 20)
	if err != nil {
		t.Errorf("expected git 2.20+ to be available: %v", err)
	}

	// Test with impossibly high version requirement
	err = CheckGitVersion(99, 0)
	if err == nil {
		t.Error("expected error for git version 99.0 requirement")
	}
}
