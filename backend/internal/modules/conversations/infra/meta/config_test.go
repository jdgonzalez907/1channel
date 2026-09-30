package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigFromEnv(t *testing.T) {
	t.Run("reads all values", func(t *testing.T) {
		t.Setenv(envPageID, "page-1")
		t.Setenv(envPageAccessToken, "token-1")
		t.Setenv(envAppSecret, "secret-1")
		t.Setenv(envVerifyToken, "verify-1")

		assert.Equal(t, Config{
			PageID:          "page-1",
			PageAccessToken: "token-1",
			AppSecret:       "secret-1",
			VerifyToken:     "verify-1",
		}, NewConfigFromEnv())
	})

	t.Run("absent values are empty", func(t *testing.T) {
		t.Setenv(envPageID, "")
		t.Setenv(envPageAccessToken, "")
		t.Setenv(envAppSecret, "")
		t.Setenv(envVerifyToken, "")

		assert.Equal(t, Config{}, NewConfigFromEnv())
	})
}
