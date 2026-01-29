package git

import (
	"testing"
)

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "my-project", false},
		{"valid with numbers", "project123", false},
		{"empty", "", true},
		{"path traversal", "../secret", true},
		{"absolute path", "/etc/passwd", true},
		{"backslash", "foo\\bar", true},
		{"current dir", ".", true},
		{"parent dir", "..", true},
		{"reserved git", ".git", true},
		{"reserved bare", ".bare", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "feature/auth", false},
		{"valid with numbers", "issue-123", false},
		{"empty", "", true},
		{"double dot", "foo..bar", true},
		{"tilde", "foo~bar", true},
		{"caret", "foo^bar", true},
		{"colon", "foo:bar", true},
		{"space", "foo bar", true},
		{"question mark", "foo?bar", true},
		{"asterisk", "foo*bar", true},
		{"bracket", "foo[bar", true},
		{"starts with dot", ".foo", true},
		{"ends with dot", "foo.", true},
		{"starts with slash", "/foo", true},
		{"ends with slash", "foo/", true},
		{"consecutive slashes", "foo//bar", true},
		{"ends with .lock", "foo.lock", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBranchName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBranchName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestFlattenBranchName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"feature/auth", "feature-auth"},
		{"feature/deep/nested", "feature-deep-nested"},
		{"simple", "simple"},
		{"already-flat", "already-flat"},
		{"mix/of-styles", "mix-of-styles"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FlattenBranchName(tt.input)
			if result != tt.expected {
				t.Errorf("FlattenBranchName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
