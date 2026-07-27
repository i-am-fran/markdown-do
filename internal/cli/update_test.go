package cli

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAssetNameForPlatform(t *testing.T) {
	cases := []struct {
		goos, goarch string
		want         string
	}{
		{"darwin", "amd64", "mdd-darwin-amd64"},
		{"darwin", "arm64", "mdd-darwin-arm64"},
		{"linux", "amd64", "mdd-linux-amd64"},
		{"linux", "arm64", "mdd-linux-arm64"},
		{"windows", "amd64", "mdd-windows-amd64.exe"},
	}
	for _, tc := range cases {
		got, err := assetNameForPlatform(tc.goos, tc.goarch)
		if err != nil {
			t.Errorf("assetNameForPlatform(%q, %q) unexpected error: %v", tc.goos, tc.goarch, err)
			continue
		}
		if got != tc.want {
			t.Errorf("assetNameForPlatform(%q, %q) = %q, want %q", tc.goos, tc.goarch, got, tc.want)
		}
	}

	unsupported := []struct{ goos, goarch string }{
		{"linux", "386"},
		{"windows", "arm64"},
	}
	for _, tc := range unsupported {
		_, err := assetNameForPlatform(tc.goos, tc.goarch)
		if err == nil {
			t.Errorf("assetNameForPlatform(%q, %q) expected error, got nil", tc.goos, tc.goarch)
			continue
		}
		if !strings.Contains(err.Error(), tc.goos+"/"+tc.goarch) {
			t.Errorf("assetNameForPlatform(%q, %q) error %q doesn't mention the platform", tc.goos, tc.goarch, err.Error())
		}
	}
}

func TestParseChecksums(t *testing.T) {
	data := []byte(
		"deadbeef00  mdd-darwin-amd64\n" +
			"cafebabe11 *mdd-linux-amd64\n" +
			"not a valid line\n" +
			"\n" +
			"ABCDEF0022  mdd-windows-amd64.exe\n",
	)

	got, err := parseChecksums(data, "mdd-darwin-amd64")
	if err != nil || got != "deadbeef00" {
		t.Errorf("parseChecksums darwin: got (%q, %v), want (%q, nil)", got, err, "deadbeef00")
	}

	got, err = parseChecksums(data, "mdd-linux-amd64")
	if err != nil || got != "cafebabe11" {
		t.Errorf("parseChecksums linux (binary-mode *prefix): got (%q, %v), want (%q, nil)", got, err, "cafebabe11")
	}

	got, err = parseChecksums(data, "mdd-windows-amd64.exe")
	if err != nil || got != "abcdef0022" {
		t.Errorf("parseChecksums windows (lowercased): got (%q, %v), want (%q, nil)", got, err, "abcdef0022")
	}

	if _, err := parseChecksums(data, "mdd-missing"); err == nil {
		t.Error("parseChecksums with no matching entry: expected error, got nil")
	}
}

// fakeReleaseServer serves /<tag>/<asset> with binContent, /<tag>/checksums.txt
// with a checksums.txt naming asset -> its real sha256 (unless
// overrideChecksum is non-empty, in which case that value is used instead,
// to simulate a corrupted/mismatched release), and /<tag>/checksums.txt.sig
// with a valid ed25519 signature of that checksums.txt content. It installs
// the matching public key via SetUpdatePublicKeyForTesting so
// verifyChecksumsSignature accepts it.
func fakeReleaseServer(t *testing.T, tag, asset string, binContent []byte, overrideChecksum string) *httptest.Server {
	t.Helper()
	sum := overrideChecksum
	if sum == "" {
		h := sha256.Sum256(binContent)
		sum = hex.EncodeToString(h[:])
	}
	checksums := fmt.Sprintf("%s  %s\n", sum, asset)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	t.Cleanup(SetUpdatePublicKeyForTesting(pub))
	sig := base64.StdEncoding.EncodeToString(ed25519.Sign(priv, []byte(checksums)))

	mux := http.NewServeMux()
	mux.HandleFunc("/"+tag+"/"+asset, func(w http.ResponseWriter, r *http.Request) {
		w.Write(binContent)
	})
	mux.HandleFunc("/"+tag+"/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, checksums)
	})
	mux.HandleFunc("/"+tag+"/checksums.txt.sig", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, sig)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestPerformUpdate_HappyPath(t *testing.T) {
	asset, err := assetNameForPlatform(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported platform for this test: %v", err)
	}
	tag := "9.9.9"
	binContent := []byte("fake mdd binary contents")
	srv := fakeReleaseServer(t, tag, asset, binContent, "")
	t.Cleanup(SetUpdateBaseURLForTesting(srv.URL))

	dir := t.TempDir()
	execPath := filepath.Join(dir, "mdd")
	if err := os.WriteFile(execPath, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("seeding execPath: %v", err)
	}

	oldPath, err := performUpdate(tag, execPath)
	if err != nil {
		t.Fatalf("performUpdate: %v", err)
	}
	if runtime.GOOS != "windows" && oldPath != "" {
		t.Errorf("expected no .old path on non-Windows, got %q", oldPath)
	}

	got, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("reading execPath after update: %v", err)
	}
	if string(got) != string(binContent) {
		t.Errorf("execPath contents = %q, want %q", got, binContent)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(execPath)
		if err != nil {
			t.Fatalf("stat execPath: %v", err)
		}
		if info.Mode().Perm() != 0o755 {
			t.Errorf("execPath mode = %v, want 0755", info.Mode().Perm())
		}
	}
}

func TestPerformUpdate_ChecksumMismatch(t *testing.T) {
	asset, err := assetNameForPlatform(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported platform for this test: %v", err)
	}
	tag := "9.9.9"
	srv := fakeReleaseServer(t, tag, asset, []byte("fake mdd binary contents"), "0000000000000000000000000000000000000000000000000000000000000000")
	t.Cleanup(SetUpdateBaseURLForTesting(srv.URL))

	dir := t.TempDir()
	execPath := filepath.Join(dir, "mdd")
	original := []byte("old binary")
	if err := os.WriteFile(execPath, original, 0o755); err != nil {
		t.Fatalf("seeding execPath: %v", err)
	}

	if _, err := performUpdate(tag, execPath); err == nil {
		t.Fatal("performUpdate with mismatched checksum: expected error, got nil")
	}

	got, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("reading execPath after failed update: %v", err)
	}
	if string(got) != string(original) {
		t.Errorf("execPath was modified despite checksum mismatch: got %q, want unchanged %q", got, original)
	}
}

func TestPerformUpdate_MissingAsset(t *testing.T) {
	asset, err := assetNameForPlatform(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported platform for this test: %v", err)
	}
	tag := "9.9.9"
	mux := http.NewServeMux()
	mux.HandleFunc("/"+tag+"/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "deadbeef  %s\n", asset)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	t.Cleanup(SetUpdateBaseURLForTesting(srv.URL))

	dir := t.TempDir()
	execPath := filepath.Join(dir, "mdd")
	if err := os.WriteFile(execPath, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("seeding execPath: %v", err)
	}

	if _, err := performUpdate(tag, execPath); err == nil {
		t.Fatal("performUpdate with a 404 asset: expected error, got nil")
	}
}

// TestPerformUpdate_RejectsInvalidSignature simulates a host that can serve
// a checksum matching a (possibly tampered) binary, but can't produce a
// valid signature for it — e.g. a compromised mirror or a MITM without the
// real signing key. The signature is real, just from an unrelated keypair
// that doesn't match the embedded/installed public key.
func TestPerformUpdate_RejectsInvalidSignature(t *testing.T) {
	asset, err := assetNameForPlatform(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("unsupported platform for this test: %v", err)
	}
	tag := "9.9.9"
	binContent := []byte("fake mdd binary contents")
	h := sha256.Sum256(binContent)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(h[:]), asset)

	// Install a legitimate public key for verification...
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	t.Cleanup(SetUpdatePublicKeyForTesting(pub))

	// ...but sign with a different, unrelated private key.
	_, attackerPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	badSig := base64.StdEncoding.EncodeToString(ed25519.Sign(attackerPriv, []byte(checksums)))

	mux := http.NewServeMux()
	mux.HandleFunc("/"+tag+"/"+asset, func(w http.ResponseWriter, r *http.Request) {
		w.Write(binContent)
	})
	mux.HandleFunc("/"+tag+"/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, checksums)
	})
	mux.HandleFunc("/"+tag+"/checksums.txt.sig", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, badSig)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	t.Cleanup(SetUpdateBaseURLForTesting(srv.URL))

	dir := t.TempDir()
	execPath := filepath.Join(dir, "mdd")
	original := []byte("old binary")
	if err := os.WriteFile(execPath, original, 0o755); err != nil {
		t.Fatalf("seeding execPath: %v", err)
	}

	if _, err := performUpdate(tag, execPath); err == nil {
		t.Fatal("performUpdate with a signature from an unrelated key: expected error, got nil")
	}

	got, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("reading execPath after rejected update: %v", err)
	}
	if string(got) != string(original) {
		t.Errorf("execPath was modified despite invalid signature: got %q, want unchanged %q", got, original)
	}
}

func TestVerifyChecksumsSignature_RoundTrip(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	t.Cleanup(SetUpdatePublicKeyForTesting(pub))

	data := []byte("some checksums.txt content\n")
	sig := base64.StdEncoding.EncodeToString(ed25519.Sign(priv, data))

	if err := verifyChecksumsSignature(data, []byte(sig)); err != nil {
		t.Errorf("expected a valid signature to verify, got: %v", err)
	}

	if err := verifyChecksumsSignature([]byte("tampered content\n"), []byte(sig)); err == nil {
		t.Error("expected verification to fail for tampered content, got nil")
	}
}

func TestSignChecksumsFile_ProducesVerifiableSignature(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	t.Setenv("MDD_RELEASE_SIGNING_KEY", hex.EncodeToString(priv.Seed()))

	path := filepath.Join(t.TempDir(), "checksums.txt")
	content := []byte("deadbeef  mdd-linux-amd64\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("seeding checksums.txt: %v", err)
	}

	sigB64, err := SignChecksumsFile(path)
	if err != nil {
		t.Fatalf("SignChecksumsFile failed: %v", err)
	}

	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		t.Fatalf("signature isn't valid base64: %v", err)
	}
	if !ed25519.Verify(pub, content, sig) {
		t.Error("expected the produced signature to verify against the matching public key")
	}
}

func TestSignChecksumsFile_RejectsInvalidKey(t *testing.T) {
	t.Setenv("MDD_RELEASE_SIGNING_KEY", "not-valid-hex")

	path := filepath.Join(t.TempDir(), "checksums.txt")
	if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
		t.Fatalf("seeding checksums.txt: %v", err)
	}

	if _, err := SignChecksumsFile(path); err == nil {
		t.Fatal("expected an error for an invalid signing key, got nil")
	}
}

func TestSwapBinaryUnixStrategy(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "current")
	newFile := filepath.Join(dir, "new")
	if err := os.WriteFile(current, []byte("old"), 0o644); err != nil {
		t.Fatalf("seeding current: %v", err)
	}
	if err := os.WriteFile(newFile, []byte("new"), 0o644); err != nil {
		t.Fatalf("seeding new: %v", err)
	}

	if err := swapBinaryUnix(current, newFile); err != nil {
		t.Fatalf("swapBinaryUnix: %v", err)
	}

	got, err := os.ReadFile(current)
	if err != nil {
		t.Fatalf("reading current after swap: %v", err)
	}
	if string(got) != "new" {
		t.Errorf("current contents = %q, want %q", got, "new")
	}
	info, err := os.Stat(current)
	if err != nil {
		t.Fatalf("stat current: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("current mode = %v, want 0755", info.Mode().Perm())
	}
	if _, err := os.Stat(newFile); !os.IsNotExist(err) {
		t.Errorf("expected newFile to be gone after rename, stat err = %v", err)
	}
}

func TestSwapBinaryWindowsStyleStrategy(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "mdd.exe")
	newFile := filepath.Join(dir, "new.exe")
	if err := os.WriteFile(current, []byte("old"), 0o644); err != nil {
		t.Fatalf("seeding current: %v", err)
	}
	if err := os.WriteFile(newFile, []byte("new"), 0o644); err != nil {
		t.Fatalf("seeding new: %v", err)
	}

	oldPath, err := swapBinaryWindowsStyle(current, newFile)
	if err != nil {
		t.Fatalf("swapBinaryWindowsStyle: %v", err)
	}
	if oldPath != current+".old" {
		t.Errorf("oldPath = %q, want %q", oldPath, current+".old")
	}

	gotOld, err := os.ReadFile(oldPath)
	if err != nil {
		t.Fatalf("reading .old file: %v", err)
	}
	if string(gotOld) != "old" {
		t.Errorf(".old contents = %q, want %q", gotOld, "old")
	}

	gotCurrent, err := os.ReadFile(current)
	if err != nil {
		t.Fatalf("reading current after swap: %v", err)
	}
	if string(gotCurrent) != "new" {
		t.Errorf("current contents = %q, want %q", gotCurrent, "new")
	}
}

func TestSwapBinaryWindowsStyleStrategy_RollsBackOnFailure(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "mdd.exe")
	missingNew := filepath.Join(dir, "does-not-exist.exe")
	original := []byte("old")
	if err := os.WriteFile(current, original, 0o644); err != nil {
		t.Fatalf("seeding current: %v", err)
	}

	if _, err := swapBinaryWindowsStyle(current, missingNew); err == nil {
		t.Fatal("swapBinaryWindowsStyle with a missing new binary: expected error, got nil")
	}

	got, err := os.ReadFile(current)
	if err != nil {
		t.Fatalf("reading current after failed swap (rollback should restore it): %v", err)
	}
	if string(got) != string(original) {
		t.Errorf("current contents after rollback = %q, want %q", got, original)
	}
	if _, err := os.Stat(current + ".old"); !os.IsNotExist(err) {
		t.Errorf("expected no leftover .old file after rollback, stat err = %v", err)
	}
}
