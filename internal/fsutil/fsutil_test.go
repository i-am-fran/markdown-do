package fsutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAtomicWriteFileWritesContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")

	if err := AtomicWriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("expected content %q, got %q", "hello", got)
	}
}

func TestAtomicWriteFileLeavesNoTempFileOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	if err := AtomicWriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "out.txt" {
		t.Errorf("expected only out.txt left in dir, got %v", entries)
	}
}

func TestAtomicWriteFileSetsPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")

	if err := AtomicWriteFile(path, []byte("hello"), 0600); err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected permissions 0600, got %v", info.Mode().Perm())
	}
}

func TestAtomicWriteFileOverwritesExistingContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")
	if err := os.WriteFile(path, []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if err := AtomicWriteFile(path, []byte("new"), 0644); err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("expected content %q, got %q", "new", got)
	}
}

func TestAcquireLockExclusivity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "TODO.md")

	lock, err := AcquireLock(path)
	if err != nil {
		t.Fatalf("first AcquireLock failed: %v", err)
	}
	defer func() { _ = lock.Unlock() }()

	if _, err := os.Stat(path + ".lock"); err != nil {
		t.Fatalf("expected lock file to exist: %v", err)
	}

	start := time.Now()
	_, err = AcquireLock(path)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected second AcquireLock to fail while the first is held")
	}
	if elapsed < maxWait {
		t.Errorf("expected AcquireLock to wait out maxWait (%v) before giving up, only waited %v", maxWait, elapsed)
	}
}

func TestAcquireLockSucceedsAfterUnlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "TODO.md")

	lock, err := AcquireLock(path)
	if err != nil {
		t.Fatalf("first AcquireLock failed: %v", err)
	}
	if err := lock.Unlock(); err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	second, err := AcquireLock(path)
	if err != nil {
		t.Fatalf("expected AcquireLock to succeed after Unlock, got: %v", err)
	}
	_ = second.Unlock()
}

func TestAcquireLockStealsAbandonedLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "TODO.md")
	lockPath := path + ".lock"

	if err := os.WriteFile(lockPath, []byte("12345\n"), 0o600); err != nil {
		t.Fatalf("failed to seed a stale lock file: %v", err)
	}
	old := time.Now().Add(-staleAfter - time.Second)
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatalf("Chtimes failed: %v", err)
	}

	start := time.Now()
	lock, err := AcquireLock(path)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("expected AcquireLock to steal the abandoned lock, got: %v", err)
	}
	defer func() { _ = lock.Unlock() }()
	if elapsed >= maxWait {
		t.Errorf("expected a stale lock to be stolen well under maxWait (%v), took %v", maxWait, elapsed)
	}
}
