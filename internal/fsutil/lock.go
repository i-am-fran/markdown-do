package fsutil

import (
	"fmt"
	"os"
	"time"
)

const (
	// staleAfter is how old an unheld-looking lock file's mtime must be
	// before a new acquirer treats it as abandoned (e.g. left behind by a
	// process that crashed instead of calling Unlock) and steals it.
	staleAfter = 10 * time.Second
	// retryInterval is how often AcquireLock retries after contention.
	retryInterval = 25 * time.Millisecond
	// maxWait bounds how long AcquireLock will wait for a lock that's
	// genuinely held (i.e. not stale) by another process before giving up.
	maxWait = 2 * time.Second
)

// Lock is a held advisory sidecar-file lock. Callers must call Unlock
// (typically via defer) to release it.
type Lock struct {
	path string
}

// AcquireLock takes an advisory lock on path+".lock" using O_CREATE|O_EXCL,
// so it's exclusive across cooperating processes on the same machine (it
// does not lock against processes that ignore the convention, nor across
// networked filesystems without atomic O_EXCL semantics). It retries
// briefly on contention and steals a lock that looks abandoned (mtime
// older than staleAfter) rather than waiting on it indefinitely.
func AcquireLock(path string) (*Lock, error) {
	lockPath := path + ".lock"
	deadline := time.Now().Add(maxWait)

	for {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			_, _ = fmt.Fprintf(f, "%d\n", os.Getpid())
			_ = f.Close()
			return &Lock{path: lockPath}, nil
		}
		if !os.IsExist(err) {
			return nil, err
		}

		if info, statErr := os.Stat(lockPath); statErr == nil && time.Since(info.ModTime()) > staleAfter {
			_ = os.Remove(lockPath) // best-effort steal; the next loop iteration retries the create
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for lock on %s (held by another mdd process?)", path)
		}
		time.Sleep(retryInterval)
	}
}

// Unlock releases the lock. Safe to call on a nil *Lock.
func (l *Lock) Unlock() error {
	if l == nil {
		return nil
	}
	return os.Remove(l.path)
}
