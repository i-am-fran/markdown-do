package core

import (
	"os"
	"path/filepath"
	"testing"
)

// chdirTemp chdirs into a fresh temp directory for the duration of the
// test, restoring the original working directory after.
func chdirTemp(t *testing.T) string {
	t.Helper()
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
	t.Cleanup(func() { os.Chdir(origWd) })
	return dir
}

func TestCacheFilePathDoesNotLiveInCwd(t *testing.T) {
	dir := chdirTemp(t)

	path, err := GetCacheFilePath()
	if err != nil {
		t.Fatalf("GetCacheFilePath failed: %v", err)
	}
	if filepath.Dir(path) == dir {
		t.Errorf("expected cache file outside the project directory, got %q", path)
	}
}

func TestSaveLoadClearCacheRoundTrip(t *testing.T) {
	chdirTemp(t)

	if cache, err := LoadCache(); err != nil || cache != nil {
		t.Fatalf("expected no cache before any save, got %+v, err %v", cache, err)
	}

	cache := &Cache{
		Tasks: map[int]TaskCache{
			1: {GlobalID: 1, FilePath: "TODO.md", LocalID: 1, TaskText: "Buy milk"},
		},
	}
	if err := SaveCache(cache); err != nil {
		t.Fatalf("SaveCache failed: %v", err)
	}

	loaded, err := LoadCache()
	if err != nil {
		t.Fatalf("LoadCache failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected a cache to be loaded")
	}
	if got := loaded.Tasks[1].TaskText; got != "Buy milk" {
		t.Errorf("expected cached task text %q, got %q", "Buy milk", got)
	}

	if err := ClearCache(); err != nil {
		t.Fatalf("ClearCache failed: %v", err)
	}
	if cache, err := LoadCache(); err != nil || cache != nil {
		t.Fatalf("expected no cache after ClearCache, got %+v, err %v", cache, err)
	}
}

func TestCacheIsolatedPerDirectory(t *testing.T) {
	chdirTemp(t)

	if err := SaveCache(&Cache{Tasks: map[int]TaskCache{1: {TaskText: "here"}}}); err != nil {
		t.Fatalf("SaveCache failed: %v", err)
	}

	chdirTemp(t)
	cache, err := LoadCache()
	if err != nil {
		t.Fatalf("LoadCache failed: %v", err)
	}
	if cache != nil {
		t.Errorf("expected a different directory to see no cache, got %+v", cache)
	}
}
