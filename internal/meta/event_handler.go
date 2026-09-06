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

type EventHandler struct {
	cfg config.Configuration
}

func NewEventHandler(cfg config.Configuration) *EventHandler {
	return &EventHandler{cfg: cfg}
}

func (h *EventHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	slog.Info("meta webhook received")
	slog.Debug("meta webhook payload",
		"payload", string(bodyBytes),
	)

	w.WriteHeader(http.StatusOK)
}
