package cli

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestShowHelpOutputUnchanged pins ShowHelp's exact output against a golden
// file captured before its internal implementation was refactored from one
// large positionally-interpolated Printf into one heading+body print per
// section. The refactor must be text-preserving.
func TestShowHelpOutputUnchanged(t *testing.T) {
	golden, err := os.ReadFile(filepath.Join("testdata", "help_golden.txt"))
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w
	ShowHelp()
	w.Close()
	os.Stdout = origStdout

	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}

	if string(got) != string(golden) {
		t.Errorf("ShowHelp() output changed.\n--- got ---\n%s\n--- want ---\n%s", got, golden)
	}
}
