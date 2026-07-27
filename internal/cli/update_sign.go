package cli

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// SignChecksumsFile reads an ed25519 private-key seed (hex-encoded, 32
// bytes) from the MDD_RELEASE_SIGNING_KEY environment variable, signs the
// file at path, and returns the base64-encoded signature. Invoked only by
// .github/workflows/release.yml, via the hidden "mdd __sign-checksums"
// subcommand, during the release build — never by end users, and never by
// the update client itself (which only ever verifies, via
// verifyChecksumsSignature in update_verify.go).
func SignChecksumsFile(path string) (string, error) {
	seedHex := strings.TrimSpace(os.Getenv("MDD_RELEASE_SIGNING_KEY"))
	seed, err := hex.DecodeString(seedHex)
	if err != nil || len(seed) != ed25519.SeedSize {
		return "", fmt.Errorf("MDD_RELEASE_SIGNING_KEY must be a %d-byte hex-encoded ed25519 seed", ed25519.SeedSize)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	sig := ed25519.Sign(ed25519.NewKeyFromSeed(seed), data)
	return base64.StdEncoding.EncodeToString(sig), nil
}
