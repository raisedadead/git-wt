package commands

import (
	"strings"
	"testing"
)

// TestMenuLabels verifies that the workflow menu options contain the expected template hints
func TestMenuLabels(t *testing.T) {
	tests := []struct {
		name          string
		label         string
		shouldHave    string
		shouldNotHave string
	}{
		{
			name:       "feature has template hint",
			label:      WorkflowMenuLabels["feature"],
			shouldHave: "(feat/...)",
		},
		{
			name:       "bugfix has template hint",
			label:      WorkflowMenuLabels["bugfix"],
			shouldHave: "(fix/...)",
		},
		{
			name:          "pr-review has no template hint",
			label:         WorkflowMenuLabels["pr-review"],
			shouldNotHave: "(...)",
		},
		{
			name:          "branch has no template hint",
			label:         WorkflowMenuLabels["branch"],
			shouldNotHave: "(...)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldHave != "" && !strings.Contains(tt.label, tt.shouldHave) {
				t.Errorf("menu label %q should contain %q", tt.label, tt.shouldHave)
			}
			if tt.shouldNotHave != "" && strings.Contains(tt.label, tt.shouldNotHave) {
				t.Errorf("menu label %q should not contain %q", tt.label, tt.shouldNotHave)
			}
		})
	}
}

// TestMenuLabelsComplete ensures all expected workflows have labels
func TestMenuLabelsComplete(t *testing.T) {
	expectedWorkflows := []string{"feature", "bugfix", "pr-review", "branch"}

	for _, workflow := range expectedWorkflows {
		t.Run(workflow, func(t *testing.T) {
			label, ok := WorkflowMenuLabels[workflow]
			if !ok {
				t.Errorf("WorkflowMenuLabels missing key %q", workflow)
			}
			if label == "" {
				t.Errorf("WorkflowMenuLabels[%q] is empty", workflow)
			}
		})
	}
}
