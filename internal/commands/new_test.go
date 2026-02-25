package commands

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
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

// TestNewCommandHasSwitchFlag verifies that --switch/-s flag is registered on the add command
func TestNewCommandHasSwitchFlag(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"add"})
	if err != nil {
		t.Fatalf("add command not found: %v", err)
	}

	flag := cmd.Flags().Lookup("switch")
	if flag == nil {
		t.Fatal("--switch flag not found on add command")
	}
	if flag.Shorthand != "s" {
		t.Errorf("--switch shorthand = %q, want %q", flag.Shorthand, "s")
	}
	if flag.DefValue != "false" {
		t.Errorf("--switch default = %q, want %q", flag.DefValue, "false")
	}
	if flag.Value.Type() != "bool" {
		t.Errorf("--switch type = %q, want %q", flag.Value.Type(), "bool")
	}
}

// TestNewCommandSwitchFlagDescription verifies the --switch flag has a helpful description
func TestNewCommandSwitchFlagDescription(t *testing.T) {
	cmd, _, _ := rootCmd.Find([]string{"add"})
	flag := cmd.Flags().Lookup("switch")
	if flag == nil {
		t.Fatal("--switch flag not found on add command")
	}
	if !strings.Contains(strings.ToLower(flag.Usage), "switch") {
		t.Errorf("--switch usage %q should mention 'switch'", flag.Usage)
	}
}

// TestNewCommandSwitchFlagDefaultFalse verifies switchFlag package var defaults to false
func TestNewCommandSwitchFlagDefaultFalse(t *testing.T) {
	// Reset by looking up the flag and checking its stored value
	cmd, _, _ := rootCmd.Find([]string{"add"})
	flag := cmd.Flags().Lookup("switch")
	if flag == nil {
		t.Fatal("--switch flag not found on add command")
	}
	// When no args parsed, default should be false
	val, err := cmd.Flags().GetBool("switch")
	if err != nil {
		t.Fatalf("failed to get --switch value: %v", err)
	}
	if val != false {
		t.Errorf("--switch default value = %v, want false", val)
	}
}

// TestNewCommandSwitchFlagNotPersistent verifies --switch is local to add, not inherited
func TestNewCommandSwitchFlagNotPersistent(t *testing.T) {
	cmd, _, _ := rootCmd.Find([]string{"add"})
	// Should be a local flag, not persistent
	var found bool
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "switch" {
			found = true
		}
	})
	if !found {
		t.Error("--switch should be a local flag on the add command")
	}

	// Should NOT be on the root command
	rootFlag := rootCmd.Flags().Lookup("switch")
	if rootFlag != nil {
		t.Error("--switch should not be on the root command")
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
