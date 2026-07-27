//go:build testhooks

package cli

import (
	"crypto/ed25519"
	"strings"
)

var updateBaseURLOverride string

// updateBaseURL honors SetUpdateBaseURLForTesting when set, falling back to
// the real GitHub download URL otherwise. Only compiled under the
// `testhooks` build tag (never in `make build`'s production binary) — see
// update_prod.go for the production implementation this file replaces.
func updateBaseURL() string {
	if updateBaseURLOverride != "" {
		return updateBaseURLOverride
	}
	return "https://github.com/i-am-fran/markdown-do/releases/download"
}

// SetUpdateBaseURLForTesting overrides the base URL used for downloading
// update assets and returns a function that restores the previous value.
// Intended for tests only; compiled in only under the `testhooks` tag, so
// it can't be reached by a production binary at all.
func SetUpdateBaseURLForTesting(url string) (restore func()) {
	orig := updateBaseURLOverride
	updateBaseURLOverride = strings.TrimSuffix(url, "/")
	return func() { updateBaseURLOverride = orig }
}

var updatePublicKeyOverride ed25519.PublicKey

// currentUpdatePublicKey honors SetUpdatePublicKeyForTesting when set,
// falling back to the production embedded key otherwise.
func currentUpdatePublicKey() (ed25519.PublicKey, error) {
	if updatePublicKeyOverride != nil {
		return updatePublicKeyOverride, nil
	}
	return productionUpdatePublicKey()
}

// SetUpdatePublicKeyForTesting overrides the public key used to verify a
// release's checksums.txt signature and returns a function that restores
// the previous value. Intended for tests only; compiled in only under the
// `testhooks` tag.
func SetUpdatePublicKeyForTesting(pub ed25519.PublicKey) (restore func()) {
	orig := updatePublicKeyOverride
	updatePublicKeyOverride = pub
	return func() { updatePublicKeyOverride = orig }
}
