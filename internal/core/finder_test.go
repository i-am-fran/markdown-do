package core

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// writeFile creates path (and any parent directories) with the given
// content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
}

func relPaths(files []FoundFile) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = filepath.ToSlash(f.RelativePath)
	}
	sort.Strings(out)
	return out
}

func TestFindTodoFilesNonRecursiveOnlyTopLevel(t *testing.T) {
	dir := chdirTemp(t)
	writeFile(t, filepath.Join(dir, "TODO.md"), "# TODO\n")
	writeFile(t, filepath.Join(dir, "sub", "TODO.md"), "# TODO\n")

	files, err := FindTodoFiles(dir, false)
	if err != nil {
		t.Fatalf("FindTodoFiles failed: %v", err)
	}
	if got := relPaths(files); len(got) != 1 || got[0] != "TODO.md" {
		t.Errorf("expected only top-level TODO.md, got %v", got)
	}
}

func TestFindTodoFilesRecursiveFindsNested(t *testing.T) {
	dir := chdirTemp(t)
	writeFile(t, filepath.Join(dir, "TODO.md"), "# TODO\n")
	writeFile(t, filepath.Join(dir, "sub", "TODO.md"), "# TODO\n")
	writeFile(t, filepath.Join(dir, "sub", "nested", "TODO.md"), "# TODO\n")

	files, err := FindTodoFiles(dir, true)
	if err != nil {
		t.Fatalf("FindTodoFiles failed: %v", err)
	}
	want := []string{"TODO.md", "sub/TODO.md", "sub/nested/TODO.md"}
	sort.Strings(want)
	if got := relPaths(files); !equalStringSlices(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestFindTodoFilesExcludesNodeModules(t *testing.T) {
	dir := chdirTemp(t)
	writeFile(t, filepath.Join(dir, "TODO.md"), "# TODO\n")
	writeFile(t, filepath.Join(dir, "node_modules", "pkg", "TODO.md"), "# TODO\n")

	files, err := FindTodoFiles(dir, true)
	if err != nil {
		t.Fatalf("FindTodoFiles failed: %v", err)
	}
	if got := relPaths(files); len(got) != 1 || got[0] != "TODO.md" {
		t.Errorf("expected node_modules/ excluded, got %v", got)
	}
}

func TestFindTodoFilesExcludesDotGit(t *testing.T) {
	dir := chdirTemp(t)
	writeFile(t, filepath.Join(dir, "TODO.md"), "# TODO\n")
	writeFile(t, filepath.Join(dir, ".git", "TODO.md"), "# TODO\n")

	files, err := FindTodoFiles(dir, true)
	if err != nil {
		t.Fatalf("FindTodoFiles failed: %v", err)
	}
	if got := relPaths(files); len(got) != 1 || got[0] != "TODO.md" {
		t.Errorf("expected .git/ excluded, got %v", got)
	}
}

func TestFindTodoFilesMatchesCaseVariants(t *testing.T) {
	dir := chdirTemp(t)
	writeFile(t, filepath.Join(dir, "TODO.md"), "# TODO\n")
	writeFile(t, filepath.Join(dir, "todo-list.md"), "# TODO\n")

	files, err := FindTodoFiles(dir, false)
	if err != nil {
		t.Fatalf("FindTodoFiles failed: %v", err)
	}
	want := []string{"TODO.md", "todo-list.md"}
	sort.Strings(want)
	if got := relPaths(files); !equalStringSlices(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestFindDefaultTodoFilePrefersExactTODOmd(t *testing.T) {
	dir := chdirTemp(t)
	writeFile(t, filepath.Join(dir, "todo-other.md"), "# TODO\n")
	writeFile(t, filepath.Join(dir, "TODO.md"), "# TODO\n")

	path, err := FindDefaultTodoFile(dir)
	if err != nil {
		t.Fatalf("FindDefaultTodoFile failed: %v", err)
	}
	if want := filepath.Join(dir, "TODO.md"); path != want {
		t.Errorf("expected %q, got %q", want, path)
	}
}

func TestFindDefaultTodoFileFallsBackToOtherVariant(t *testing.T) {
	dir := chdirTemp(t)
	writeFile(t, filepath.Join(dir, "todo-project.md"), "# TODO\n")

	path, err := FindDefaultTodoFile(dir)
	if err != nil {
		t.Fatalf("FindDefaultTodoFile failed: %v", err)
	}
	if want := filepath.Join(dir, "todo-project.md"); path != want {
		t.Errorf("expected fallback to %q, got %q", want, path)
	}
}

func TestFindDefaultTodoFileReturnsDefaultPathWhenNoneExist(t *testing.T) {
	dir := chdirTemp(t)

	path, err := FindDefaultTodoFile(dir)
	if err != nil {
		t.Fatalf("FindDefaultTodoFile failed: %v", err)
	}
	want := filepath.Join(dir, "TODO.md")
	if path != want {
		t.Errorf("expected default path %q, got %q", want, path)
	}
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected %q to not exist yet", path)
	}
}

func TestFindTodoFilesInSubdirsGroupsByDirectory(t *testing.T) {
	dir := chdirTemp(t)
	writeFile(t, filepath.Join(dir, "TODO.md"), "# TODO\n")
	writeFile(t, filepath.Join(dir, "sub", "TODO.md"), "# TODO\n")

	byDir, err := FindTodoFilesInSubdirs(dir)
	if err != nil {
		t.Fatalf("FindTodoFilesInSubdirs failed: %v", err)
	}

	rootFiles, ok := byDir["(root)"]
	if !ok || len(rootFiles) != 1 {
		t.Errorf("expected 1 file under \"(root)\", got %v", byDir["(root)"])
	}
	subFiles, ok := byDir[filepath.FromSlash("sub")]
	if !ok || len(subFiles) != 1 {
		t.Errorf("expected 1 file under \"sub\", got %v", byDir[filepath.FromSlash("sub")])
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
