package meta

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return signaturePrefix + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature(t *testing.T) {
	const secret = "app-secret"

	body := []byte(`{"object":"whatsapp_business_account"}`)
	tampered := []byte(`{"object":"other"}`)

	tests := []struct {
		title     string
		body      []byte
		secret    string
		signature string
		expValid  bool
	}{
		{
			title:     "success - valid signature passes",
			body:      body,
			secret:    secret,
			signature: sign(secret, body),
			expValid:  true,
		},
		{
			title:     "failure - tampered body fails",
			body:      tampered,
			secret:    secret,
			signature: sign(secret, body),
			expValid:  false,
		},
		{
			title:     "failure - wrong secret fails",
			body:      body,
			secret:    "other-secret",
			signature: sign(secret, body),
			expValid:  false,
		},
		{
			title:     "failure - missing sha256 prefix fails",
			body:      body,
			secret:    secret,
			signature: hex.EncodeToString([]byte("deadbeef")),
			expValid:  false,
		},
		{
			title:     "failure - non-hex payload after prefix fails",
			body:      body,
			secret:    secret,
			signature: signaturePrefix + "not-hex!!",
			expValid:  false,
		},
		{
			title:     "failure - empty signature fails",
			body:      body,
			secret:    secret,
			signature: "",
			expValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange

			// Act
			got := VerifySignature(tt.body, tt.secret, tt.signature)

			// Assert
			assert.Equal(t, tt.expValid, got)
		})
	}
}
