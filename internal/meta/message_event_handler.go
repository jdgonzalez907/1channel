package meta

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jdgonzalez907/1channel/internal/config"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations"
)

const (
	maxRequestBodyBytes = 4 << 20
	signatureHeader     = "X-Hub-Signature-256"
)

type webhookPayload struct {
	Entry []entry `json:"entry"`
}

type entry struct {
	Changes []change `json:"changes"`
}

type change struct {
	Value value `json:"value"`
}

type value struct {
	Messages []message `json:"messages"`
}

type message struct {
	ID         string    `json:"id"`
	From       string    `json:"from"`
	FromUserID string    `json:"from_user_id"`
	Timestamp  string    `json:"timestamp"`
	Type       string    `json:"type"`
	Text       *textBody `json:"text"`
}

type textBody struct {
	Body string `json:"body"`
}

type MessageEventHandler struct {
	cfg              config.Configuration
	conversationsAPI conversations.ConversationsAPI
}

func NewMessageEventHandler(cfg config.Configuration, conversationsAPI conversations.ConversationsAPI) *MessageEventHandler {
	return &MessageEventHandler{cfg: cfg, conversationsAPI: conversationsAPI}
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

	var payload webhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		slog.Error("failed to parse webhook payload", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	for _, e := range payload.Entry {
		for _, c := range e.Changes {
			for _, msg := range c.Value.Messages {
				if msg.Type != "text" {
					slog.Info("skipping non-text message", "type", msg.Type, "id", msg.ID)
					continue
				}

				if err := h.processMessage(r.Context(), msg); err != nil {
					slog.Error("failed to process message", "error", err, "message_id", msg.ID)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MessageEventHandler) processMessage(ctx context.Context, msg message) error {
	timestamp, err := strconv.ParseInt(msg.Timestamp, 10, 64)
	if err != nil {
		return err
	}

	return h.conversationsAPI.ReceiveContactMessage(ctx, conversations.ReceiveContactMessageInput{
		ExternalMessageID: msg.ID,
		ExternalContactID: msg.FromUserID,
		Text:              msg.Text.Body,
		ReceivedAt:        time.Unix(timestamp, 0).UTC(),
	})
}
