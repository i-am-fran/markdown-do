// Package fsutil provides small filesystem helpers (atomic writes, advisory
// locking) shared by every package that persists state to disk (TODO files,
// config, cache).
package fsutil

import (
	"os"
	"path/filepath"
)

// AtomicWriteFile writes data to path by first writing to a temp file in
// the same directory, then renaming it into place. Same-directory
// placement keeps the final rename on the same filesystem, which is what
// makes it atomic on both POSIX and Windows: readers never observe a
// partially-written file, and a crash mid-write leaves path's previous
// contents untouched instead of truncated/corrupted.
func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-"+filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	cleanup = false
	return nil
}
