package commands

import (
	"slices"
	"testing"
)

// TestDefaultFormKeyMapIncludesEsc verifies that ESC key is included in the quit binding
// so users can cancel interactive forms with either Ctrl+C or ESC
func TestDefaultFormKeyMapIncludesEsc(t *testing.T) {
	keyMap := DefaultFormKeyMap()
	quitKeys := keyMap.Quit.Keys()

	if !slices.Contains(quitKeys, "esc") {
		t.Errorf("DefaultFormKeyMap Quit binding should include 'esc', got: %v", quitKeys)
	}
	if !slices.Contains(quitKeys, "ctrl+c") {
		t.Errorf("DefaultFormKeyMap Quit binding should include 'ctrl+c', got: %v", quitKeys)
	}
}
