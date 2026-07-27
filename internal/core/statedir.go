package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// stateDir is the base directory markdown-do uses for cache/undo state
// that shouldn't live inside the project directory it's operating on.
var stateDir string

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not determine home directory (%v); storing mdd state under the current directory\n", err)
		homeDir = "."
	}
	stateDir = filepath.Join(homeDir, ".config", "markdowndo")
}

// hashPath returns a stable filename-safe key for an absolute path, so
// cache/undo files for different projects don't collide.
func hashPath(p string) string {
	sum := sha256.Sum256([]byte(p))
	return hex.EncodeToString(sum[:])
}

// SetStateDirForTesting overrides the base directory used for cache/undo
// state and returns a function that restores the previous value. Intended
// for tests only, so package-level tests never read or write the real
// ~/.config/markdowndo on the machine running `go test`.
func SetStateDirForTesting(dir string) (restore func()) {
	orig := stateDir
	stateDir = dir
	return func() { stateDir = orig }
}
