package cli

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// latestReleaseTag fetches the tag name of the latest GitHub release for
// this repo. Release tags are bare semver (e.g. "3.2.0", no "v" prefix).
func latestReleaseTag() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/i-am-fran/markdown-do/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", errors.New("GitHub API response had no tag_name")
	}
	return release.TagName, nil
}

// updateBaseURL returns the base URL under which release assets live.
// MDD_UPDATE_BASE_URL overrides it for tests only; production always uses
// the real GitHub download URL since the env var is unset.
func updateBaseURL() string {
	if v := os.Getenv("MDD_UPDATE_BASE_URL"); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return "https://github.com/i-am-fran/markdown-do/releases/download"
}

// assetNameForPlatform maps a GOOS/GOARCH pair to the release asset name
// produced by the Makefile's build-all target.
func assetNameForPlatform(goos, goarch string) (string, error) {
	supported := map[string]bool{
		"darwin/amd64":  true,
		"darwin/arm64":  true,
		"linux/amd64":   true,
		"linux/arm64":   true,
		"windows/amd64": true,
	}
	key := goos + "/" + goarch
	if !supported[key] {
		return "", fmt.Errorf("no prebuilt mdd binary for %s — build from source (see README) or open an issue at https://github.com/i-am-fran/markdown-do/issues", key)
	}
	name := fmt.Sprintf("mdd-%s-%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}
	return name, nil
}

// parseChecksums finds wantName's sha256 digest in a checksums.txt in
// `sha256sum` output format ("<hex>  <filename>" or "<hex> *<filename>").
func parseChecksums(data []byte, wantName string) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}
		digest := fields[0]
		name := strings.TrimPrefix(fields[1], "*")
		if name == wantName {
			return strings.ToLower(digest), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no checksum found for %s", wantName)
}

// downloadFile GETs url and writes the response body into a new file
// created in dir, returning its path. dir should be the directory of the
// file that will ultimately be replaced, so the later swap-rename is
// same-filesystem (and therefore atomic).
func downloadFile(url, dir, pattern string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", url, resp.Status)
	}

	out, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		os.Remove(out.Name())
		return "", err
	}
	return out.Name(), nil
}

// downloadBytes GETs url and returns the response body.
func downloadBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// sha256HexOfFile returns the lowercase hex sha256 digest of the file at path.
func sha256HexOfFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// swapBinaryUnix replaces currentPath with newPath's contents via chmod
// 0755 + os.Rename. This is safe even when currentPath is the running
// process's own executable: Unix repoints the directory entry and the
// running process keeps its already-open old inode.
func swapBinaryUnix(currentPath, newPath string) error {
	if err := os.Chmod(newPath, 0o755); err != nil {
		return err
	}
	return os.Rename(newPath, currentPath)
}

// swapBinaryWindowsStyle implements the rename-old/move-new dance required
// because Windows won't let you overwrite or delete a running .exe, but
// will let you rename it. It renames currentPath -> currentPath+".old"
// (freeing the name), then newPath -> currentPath. If the second rename
// fails, it best-effort restores the original file's name so a failed
// update doesn't leave mdd missing. Returns the surviving ".old" path on
// success (empty string on failure) so the caller can tell the user it's
// safe to delete.
func swapBinaryWindowsStyle(currentPath, newPath string) (string, error) {
	oldPath := currentPath + ".old"
	if err := os.Rename(currentPath, oldPath); err != nil {
		return "", err
	}
	if err := os.Rename(newPath, currentPath); err != nil {
		if rollbackErr := os.Rename(oldPath, currentPath); rollbackErr != nil {
			return "", fmt.Errorf("%w (rollback also failed: %v — the original binary is at %s)", err, rollbackErr, oldPath)
		}
		return "", err
	}
	return oldPath, nil
}

// swapBinary picks the correct swap strategy for the current OS.
func swapBinary(currentPath, newPath string) (string, error) {
	if runtime.GOOS == "windows" {
		return swapBinaryWindowsStyle(currentPath, newPath)
	}
	return "", swapBinaryUnix(currentPath, newPath)
}

// performUpdate downloads, checksum-verifies, and swaps in tag's release
// asset for the current platform, replacing the binary at execPath.
func performUpdate(tag, execPath string) (string, error) {
	asset, err := assetNameForPlatform(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", err
	}
	base := updateBaseURL() + "/" + tag

	binPath, err := downloadFile(base+"/"+asset, filepath.Dir(execPath), "mdd-update-*")
	if err != nil {
		return "", fmt.Errorf("downloading %s: %w", asset, err)
	}
	defer os.Remove(binPath) // no-op once swapBinary has moved it into place

	sums, err := downloadBytes(base + "/checksums.txt")
	if err != nil {
		return "", fmt.Errorf("downloading checksums.txt: %w", err)
	}
	want, err := parseChecksums(sums, asset)
	if err != nil {
		return "", fmt.Errorf("release %s looks incomplete: %w", tag, err)
	}
	got, err := sha256HexOfFile(binPath)
	if err != nil {
		return "", err
	}
	if !strings.EqualFold(got, want) {
		return "", fmt.Errorf("checksum mismatch for %s: expected %s, got %s — download may be corrupted, try again", asset, want, got)
	}

	return swapBinary(execPath, binPath)
}

// UpdateBinary downloads the latest release binary for the current
// platform, verifies its checksum, and swaps it in for the running mdd
// executable.
func UpdateBinary() error {
	tag, err := latestReleaseTag()
	if err != nil {
		return fmt.Errorf("could not determine the latest release: %w", err)
	}
	if tag == Version {
		fmt.Println(green(fmt.Sprintf("Already up to date (%s).", Version)))
		return nil
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine the running binary's path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(execPath); err == nil {
		execPath = resolved
	}

	oldPath, err := performUpdate(tag, execPath)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("update failed: %w — %s isn't writable by your user; retry with sudo, or move mdd to a directory you own (e.g. ~/bin) and reinstall from https://github.com/i-am-fran/markdown-do/releases", err, filepath.Dir(execPath))
		}
		return fmt.Errorf("update failed: %w", err)
	}

	msg := fmt.Sprintf("Updated mdd on disk at %s to %s.", execPath, tag)
	if oldPath != "" {
		msg += fmt.Sprintf(" You can delete %s.", oldPath)
	}
	msg += fmt.Sprintf(" This running process will keep reporting %s until it exits — open a new shell and run \"mdd version\" to confirm.", Version)
	fmt.Println(green(msg))
	return nil
}
