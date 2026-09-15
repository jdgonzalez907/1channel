package meta

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/jdgonzalez907/1channel/internal/config"
)

const (
	maxRequestBodyBytes = 4 << 20
	signatureHeader     = "X-Hub-Signature-256"
)

type MessageEventHandler struct {
	cfg config.Configuration
}

func NewMessageEventHandler(cfg config.Configuration) *MessageEventHandler {
	return &MessageEventHandler{cfg: cfg}
}

func (h *MessageEventHandler) Handle(w http.ResponseWriter, r *http.Request) {
	signature := r.Header.Get(signatureHeader)
	if signature == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if !VerifySignature(bodyBytes, h.cfg.MetaSecret(), signature) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	slog.Info("message event received")
	w.WriteHeader(http.StatusOK)
}
