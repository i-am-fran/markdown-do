package cli

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// updatePublicKeyHex is the ed25519 public key (32 bytes, hex-encoded) used
// to verify the signature over each release's checksums.txt before
// trusting any checksum in it. It's the public half of a keypair generated
// for this purpose; the private half is kept only as the
// MDD_RELEASE_SIGNING_KEY GitHub Actions secret used by
// .github/workflows/release.yml to sign each release (see
// update_sign.go) — it never appears in this repo. If the signing key is
// ever rotated, this constant must be updated to match the new public half.
const updatePublicKeyHex = "219d8de8eab642f4eaa887c524c848d92ef69d6d4c1cef28030cfa2db3de04d7"

// productionUpdatePublicKey decodes the embedded public key.
func productionUpdatePublicKey() (ed25519.PublicKey, error) {
	key, err := hex.DecodeString(updatePublicKeyHex)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return nil, errors.New("invalid embedded update public key")
	}
	return ed25519.PublicKey(key), nil
}

// verifyChecksumsSignature verifies sigB64 (base64-encoded) is a valid
// ed25519 signature of checksums by the current public key (the production
// embedded key, or its test-only override — see update_testhooks.go). A
// non-nil error here means the release may have been tampered with (a
// compromised release pipeline/account, or a MITM able to forge both the
// binary and its checksums.txt); the caller must not trust the checksum.
func verifyChecksumsSignature(checksums, sigB64 []byte) error {
	pub, err := currentUpdatePublicKey()
	if err != nil {
		return err
	}

	sig, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(sigB64)))
	if err != nil {
		return fmt.Errorf("malformed checksums.txt.sig: %w", err)
	}

	if !ed25519.Verify(pub, checksums, sig) {
		return errors.New("checksums.txt signature verification failed — the release may have been tampered with; aborting update")
	}
	return nil
}
