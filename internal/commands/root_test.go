package commands

import (
	"testing"
)

// TestSilentExitError verifies that SilentExit implements the error interface
// and returns an empty string
func TestSilentExitError(t *testing.T) {
	e := SilentExit{Code: 130}
	if e.Error() != "" {
		t.Errorf("SilentExit.Error() = %q, want empty string", e.Error())
	}
}
