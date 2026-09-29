package http

import (
	"net/http"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/app"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httputil"
)

type ContactWebhookHandler struct {
	receiveContactMessage       app.ReceiveContactMessage
	receiveContactMessageEdit   app.ReceiveContactMessageEdit
	receiveContactMessageDelete app.ReceiveContactMessageDelete
	receiveContactMessageRead   app.ReceiveContactMessageRead
}

func NewContactWebhookHandler(
	receiveContactMessage app.ReceiveContactMessage,
	receiveContactMessageEdit app.ReceiveContactMessageEdit,
	receiveContactMessageDelete app.ReceiveContactMessageDelete,
	receiveContactMessageRead app.ReceiveContactMessageRead,
) *ContactWebhookHandler {
	return &ContactWebhookHandler{
		receiveContactMessage:       receiveContactMessage,
		receiveContactMessageEdit:   receiveContactMessageEdit,
		receiveContactMessageDelete: receiveContactMessageDelete,
		receiveContactMessageRead:   receiveContactMessageRead,
	}
}

func (h *ContactWebhookHandler) Register(r chi.Router) {
	r.Post("/webhooks/1channel", h.handleContactWebhook)
}

func (h *ContactWebhookHandler) handleContactWebhook(w http.ResponseWriter, r *http.Request) {
	var req ContactWebhookRequest
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httperror.BadRequest(w, r, "invalid request body")
		return
	}

	at := time.Now().UTC()

	var err error

	switch req.Event {
	case "message.received":
		if req.ExternalContactID == "" || req.ExternalMessageID == "" || req.Text == nil {
			httperror.Unprocessable(w, r, "invalid contact webhook request")
			return
		}
		err = h.receiveContactMessage.Execute(r.Context(), app.ReceiveContactMessageInput{
			ConversationID:    uuid.NewV7(),
			MessageID:         uuid.NewV7(),
			ExternalMessageID: &req.ExternalMessageID,
			ExternalContactID: req.ExternalContactID,
			DisplayName:       req.DisplayName,
			Text:              *req.Text,
			ReceivedAt:        at,
		})
	case "message.edited":
		if req.ExternalContactID == "" || req.ExternalMessageID == "" || req.Text == nil {
			httperror.Unprocessable(w, r, "invalid contact webhook request")
			return
		}
		err = h.receiveContactMessageEdit.Execute(r.Context(), app.ReceiveContactMessageEditInput{
			ExternalMessageID: req.ExternalMessageID,
			ExternalContactID: req.ExternalContactID,
			NewText:           *req.Text,
			EditedAt:          at,
		})
	case "message.deleted":
		if req.ExternalContactID == "" || req.ExternalMessageID == "" {
			httperror.Unprocessable(w, r, "invalid contact webhook request")
			return
		}
		err = h.receiveContactMessageDelete.Execute(r.Context(), app.ReceiveContactMessageDeleteInput{
			ExternalMessageID: req.ExternalMessageID,
			ExternalContactID: req.ExternalContactID,
			DeletedAt:         at,
		})
	case "message.read":
		if req.ExternalContactID == "" || req.ExternalMessageID == "" {
			httperror.Unprocessable(w, r, "invalid contact webhook request")
			return
		}
		err = h.receiveContactMessageRead.Execute(r.Context(), app.ReceiveContactMessageReadInput{
			ExternalMessageID: req.ExternalMessageID,
			ExternalContactID: req.ExternalContactID,
			ReadAt:            at,
		})
	default:
		w.WriteHeader(http.StatusOK)
		return
	}

	if err != nil {
		writeConversationError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
