package meta

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/app"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httputil"
)

const (
	signatureHeader     = "X-Hub-Signature-256"
	signaturePrefix     = "sha256"
	signatureParam      = "="
	objectPage          = "page"
	hubModeParam        = "hub.mode"
	hubVerifyTokenParam = "hub.verify_token"
	hubChallengeParam   = "hub.challenge"
	hubModeSubscribe    = "subscribe"
)

type WebhookHandler struct {
	receive app.ReceiveContactMessage
	edit    app.ReceiveContactMessageEdit
	config  Config
	logger  *slog.Logger
}

func NewWebhookHandler(
	receive app.ReceiveContactMessage,
	edit app.ReceiveContactMessageEdit,
	config Config,
	logger *slog.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		receive: receive,
		edit:    edit,
		config:  config,
		logger:  logger,
	}
}

func (h *WebhookHandler) Register(r chi.Router) {
	r.Get("/webhooks/meta", h.handleVerification)
	r.Post("/webhooks/meta", h.handleNotification)
}

func (h *WebhookHandler) handleVerification(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if query.Get(hubModeParam) != hubModeSubscribe || query.Get(hubVerifyTokenParam) != h.config.VerifyToken {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(query.Get(hubChallengeParam)))
}

func (h *WebhookHandler) handleNotification(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, httputil.MaxBodyBytes))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !h.validSignature(body, r.Header.Get(signatureHeader)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var payload webhookPayload
	if err := json.Unmarshal(body, &payload); err != nil || payload.Object != objectPage {
		w.WriteHeader(http.StatusOK)
		return
	}

	if h.processEntries(r.Context(), payload.Entry) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) processEntries(ctx context.Context, entries []webhookEntry) (failed bool) {
	for _, entry := range entries {
		for _, event := range entry.Messaging {
			if err := h.processEvent(ctx, event); err != nil {
				h.logger.ErrorContext(ctx, "meta webhook: processing event failed", slog.String("error", err.Error()))
				failed = true
			}
		}
	}

	return failed
}

func (h *WebhookHandler) processEvent(ctx context.Context, event messagingEvent) error {
	switch {
	case event.Message != nil && !event.Message.IsEcho && event.Message.Text != nil:
		return h.receive.Execute(ctx, app.ReceiveContactMessageInput{
			ConversationID:    uuid.NewV7(),
			MessageID:         uuid.NewV7(),
			ExternalMessageID: &event.Message.Mid,
			ExternalContactID: event.Sender.ID,
			Text:              *event.Message.Text,
			ReceivedAt:        time.UnixMilli(event.Timestamp).UTC(),
		})
	case event.MessageEdit != nil:
		return h.edit.Execute(ctx, app.ReceiveContactMessageEditInput{
			ExternalMessageID: event.MessageEdit.Mid,
			ExternalContactID: event.Sender.ID,
			NewText:           event.MessageEdit.Text,
			EditedAt:          time.UnixMilli(event.Timestamp).UTC(),
		})
	default:
		return nil
	}
}

func (h *WebhookHandler) validSignature(body []byte, header string) bool {
	if header == "" {
		return false
	}

	parts := strings.SplitN(header, signatureParam, 2)
	if len(parts) != 2 || parts[0] != signaturePrefix {
		return false
	}

	signature, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(h.config.AppSecret))
	mac.Write(body)

	return hmac.Equal(mac.Sum(nil), signature)
}
