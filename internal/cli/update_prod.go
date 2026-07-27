//go:build !testhooks

package cli

import "crypto/ed25519"

// updateBaseURL returns the base URL under which release assets live. This
// is the production implementation: the URL is fixed, with no runtime
// override — see update_testhooks.go (compiled only under `-tags
// testhooks`, i.e. only by `go test`/CI, never by `make build`) for the
// test-only hook this file is swapped out for under that tag. Gating the
// override behind a build tag (rather than a runtime env var check) means
// it's compiled out of every shipped binary entirely, not just disabled at
// runtime.
func updateBaseURL() string {
	return "https://github.com/i-am-fran/markdown-do/releases/download"
}

// currentUpdatePublicKey returns the production embedded public key. See
// update_testhooks.go for the test-only override of this hook.
func currentUpdatePublicKey() (ed25519.PublicKey, error) {
	return productionUpdatePublicKey()
}
