package meta

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const signaturePrefix = "sha256="

func VerifySignature(body []byte, secret string, signature string) bool {
	receivedHex, found := strings.CutPrefix(signature, signaturePrefix)
	if !found {
		return false
	}

	receivedBytes, err := hex.DecodeString(receivedHex)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	return hmac.Equal(mac.Sum(nil), receivedBytes)
}
