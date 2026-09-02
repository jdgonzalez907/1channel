package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jdgonzalez907/1channel/internal/config"
)

func NewMetaWebhooks(cfg *config.Config, mux *http.ServeMux) {
	mux.HandleFunc("POST /meta/hash", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil || len(bodyBytes) == 0 {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		mac := hmac.New(sha256.New, []byte(cfg.MetaSecret))
		mac.Write(bodyBytes)
		signatureHex := hex.EncodeToString(mac.Sum(nil))

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(signatureHex))
	})

	mux.HandleFunc("GET /meta/webhook", func(w http.ResponseWriter, r *http.Request) {
		const (
			subscribeMode  = "subscribe"
			hubMode        = "hub.mode"
			hubVerifyToken = "hub.verify_token"
			hubChallenge   = "hub.challenge"
		)

		query := r.URL.Query()

		mode := query.Get(hubMode)
		token := query.Get(hubVerifyToken)
		challenge := query.Get(hubChallenge)

		if mode == subscribeMode && token == cfg.OneChannelSecret {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(challenge))
			return
		}

		http.Error(w, "Forbidden", http.StatusForbidden)
	})

	mux.HandleFunc("POST /meta/webhook", func(w http.ResponseWriter, r *http.Request) {
		signatureHeader := r.Header.Get("X-Hub-Signature-256")

		if !strings.HasPrefix(signatureHeader, "sha256=") {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		expectedSignatureHex := strings.TrimPrefix(signatureHeader, "sha256=")
		actualSignatureBytes, err := hex.DecodeString(expectedSignatureHex)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil && len(bodyBytes) == 0 {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		mac := hmac.New(sha256.New, []byte(cfg.MetaSecret))
		mac.Write(bodyBytes)
		computedSignatureBytes := mac.Sum(nil)

		if !hmac.Equal(computedSignatureBytes, actualSignatureBytes) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		slog.Info("meta received message",
			"timestamp", time.Now().UTC().Format(time.RFC3339),
			"method", r.Method,
			"url", r.URL.Path,
			"payload", json.RawMessage(bodyBytes),
		)

		w.WriteHeader(http.StatusOK)
	})
}
