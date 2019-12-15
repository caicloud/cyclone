package exec

import (
	"testing"
)

// TestExecutil tests executil.
func TestExecutil(t *testing.T) {
	_, err := RunInDir(".", "the-comand-is-not-exist")
	if err == nil {
		t.Error("Expected error to occur but it was nil")
	}
}
