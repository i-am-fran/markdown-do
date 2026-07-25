package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCompleteCommandPrefix(t *testing.T) {
	got := Complete([]string{"ar"})
	want := []string{"archive"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Complete([ar]) = %v, want %v", got, want)
	}
}

func TestCompleteCommandPrefixEmpty(t *testing.T) {
	got := Complete([]string{""})
	if len(got) != len(allCommands) {
		t.Errorf("Complete([\"\"]) returned %d candidates, want %d (all commands)", len(got), len(allCommands))
	}
}

func TestCompleteConfigSubcommand(t *testing.T) {
	got := Complete([]string{"config", "s"})
	want := []string{"set"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Complete([config s]) = %v, want %v", got, want)
	}
}

func TestCompleteSectionAliasAndHeaders(t *testing.T) {
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	defer os.Chdir(origWd)

	content := "# TODO\n\n## Bugs\n\n- [ ] fix it\n\n## Budget\n\n- [ ] save money\n"
	if err := os.WriteFile(filepath.Join(dir, "TODO.md"), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	got := Complete([]string{"task", "text", "@Bu"})
	want := []string{"@Budget", "@Bugs"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Complete([task text @Bu]) = %v, want %v", got, want)
	}
}

func TestCompleteSectionAliasBuiltin(t *testing.T) {
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	defer os.Chdir(origWd)

	got := Complete([]string{"@f"})
	want := []string{"@ff"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Complete([@f]) = %v, want %v", got, want)
	}
}

func TestCompleteUnknownPositionReturnsNil(t *testing.T) {
	got := Complete([]string{"complete", "1", ""})
	if got != nil {
		t.Errorf("Complete([complete 1 \"\"]) = %v, want nil", got)
	}
}
