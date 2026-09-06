package meta

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/jdgonzalez907/1channel/internal/config"
)

// fakeConfig satisfies config.Configuration for the methods the meta handlers
// actually use. Unused methods would panic if ever called, which is exactly
// what we want: a test that starts depending on more config is a signal.
type fakeConfig struct {
	config.Configuration

	metaSecret       string
	oneChannelSecret string
}

func (f fakeConfig) MetaSecret() string { return f.metaSecret }

func (f fakeConfig) OneChannelSecret() string { return f.oneChannelSecret }

// sign produces a Meta-style signature header value ("sha256=<hex>").
// Test-only helper: production code must never be able to forge signatures.
func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return signaturePrefix + hex.EncodeToString(mac.Sum(nil))
}
