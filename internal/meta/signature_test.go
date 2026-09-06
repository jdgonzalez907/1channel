package meta

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifySignature(t *testing.T) {
	const secret = "app-secret"

	body := []byte(`{"object":"whatsapp_business_account"}`)
	valid := func() string {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		return signaturePrefix + hex.EncodeToString(mac.Sum(nil))
	}

	tests := []struct {
		name      string
		body      []byte
		secret    string
		signature string
		want      bool
	}{
		{
			name:      "valid signature passes",
			body:      body,
			secret:    secret,
			signature: valid(),
			want:      true,
		},
		{
			name:      "tampered body fails",
			body:      []byte(`{"object":"other"}`),
			secret:    secret,
			signature: valid(),
			want:      false,
		},
		{
			name:      "wrong secret fails",
			body:      body,
			secret:    "other-secret",
			signature: valid(),
			want:      false,
		},
		{
			name:      "missing sha256 prefix fails",
			body:      body,
			secret:    secret,
			signature: hex.EncodeToString([]byte("deadbeef")),
			want:      false,
		},
		{
			name:      "non-hex payload after prefix fails",
			body:      body,
			secret:    secret,
			signature: signaturePrefix + "not-hex!!",
			want:      false,
		},
		{
			name:      "empty signature fails",
			body:      body,
			secret:    secret,
			signature: "",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifySignature(tt.body, tt.secret, tt.signature); got != tt.want {
				t.Fatalf("VerifySignature() = %v, want %v", got, tt.want)
			}
		})
	}
}
