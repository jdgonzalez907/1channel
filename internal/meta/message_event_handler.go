package meta

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/jdgonzalez907/1channel/internal/nats"
)

const (
	maxRequestBodyBytes = 4 << 20
	signatureHeader     = "X-Hub-Signature-256"
)

type MessageEventHandler struct {
	cfg    config.Configuration
	stream *nats.MessageEventReceivedStream
}

func NewMessageEventHandler(cfg config.Configuration, stream *nats.MessageEventReceivedStream) *MessageEventHandler {
	return &MessageEventHandler{cfg: cfg, stream: stream}
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

	if err := h.stream.Publish(r.Context(), bodyBytes); err != nil {
		slog.Error("publishing message event to nats",
			"subject", nats.MessageEventReceivedSubject,
			"error", err,
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("message event received", "subject", nats.MessageEventReceivedSubject)
	w.WriteHeader(http.StatusOK)
}
