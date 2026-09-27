package http

import (
	"errors"
	"net/http"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/app"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/jdgonzalez907/1channel/internal/modules/users"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httputil"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/middleware"
)

type ConversationHandler struct {
	sendAgentMessage      app.SendAgentMessage
	agentReadConversation app.AgentReadConversation
	resolveConversation   app.ResolveConversation
}

func NewConversationHandler(
	sendAgentMessage app.SendAgentMessage,
	agentReadConversation app.AgentReadConversation,
	resolveConversation app.ResolveConversation,
) *ConversationHandler {
	return &ConversationHandler{
		sendAgentMessage:      sendAgentMessage,
		agentReadConversation: agentReadConversation,
		resolveConversation:   resolveConversation,
	}
}

func (h *ConversationHandler) Register(r chi.Router) {
	r.Post("/conversations/{id}/messages", h.handleSendAgentMessage)
	r.Patch("/conversations/{id}/messages", h.handleMarkMessagesAsRead)
	r.Patch("/conversations/{id}", h.handleResolveConversation)
}

func (h *ConversationHandler) handleSendAgentMessage(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := conversationIDFrom(r)
	if !ok {
		httperror.BadRequest(w, r, "invalid conversation id")
		return
	}

	agentID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		httperror.Unauthorized(w, r, "unauthorized")
		return
	}

	var req SendAgentMessageRequest
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httperror.BadRequest(w, r, "invalid request body")
		return
	}

	messageID := uuid.NewV7()
	err := h.sendAgentMessage.Execute(r.Context(), app.SendAgentMessageInput{
		ConversationID: conversationID,
		MessageID:      messageID,
		AgentID:        agentID,
		Text:           req.Text,
		SentAt:         time.Now().UTC(),
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, MessageResponse{ID: messageID.String()})
}

func (h *ConversationHandler) handleMarkMessagesAsRead(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := conversationIDFrom(r)
	if !ok {
		httperror.BadRequest(w, r, "invalid conversation id")
		return
	}

	agentID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		httperror.Unauthorized(w, r, "unauthorized")
		return
	}

	var req MarkMessagesReadRequest
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httperror.BadRequest(w, r, "invalid request body")
		return
	}

	if req.Status != domain.MessageStatusRead.String() {
		httperror.Unprocessable(w, r, "status must be read")
		return
	}

	err := h.agentReadConversation.Execute(r.Context(), app.AgentReadConversationInput{
		ConversationID: conversationID,
		AgentID:        agentID,
		ReadAt:         time.Now().UTC(),
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConversationHandler) handleResolveConversation(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := conversationIDFrom(r)
	if !ok {
		httperror.BadRequest(w, r, "invalid conversation id")
		return
	}

	agentID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		httperror.Unauthorized(w, r, "unauthorized")
		return
	}

	var req ResolveConversationRequest
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httperror.BadRequest(w, r, "invalid request body")
		return
	}

	if req.Status != domain.ConversationStatusResolved.String() {
		httperror.Unprocessable(w, r, "status must be resolved")
		return
	}

	err := h.resolveConversation.Execute(r.Context(), app.ResolveConversationInput{
		ConversationID: conversationID,
		AgentID:        agentID,
		ResolvedAt:     time.Now().UTC(),
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConversationHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, users.ErrUserNotFound):
		httperror.Unauthorized(w, r, err.Error())
	case errors.Is(err, domain.ErrConversationAgentNotOwner),
		errors.Is(err, domain.ErrConversationContactNotOwner):
		httperror.Forbidden(w, r, err.Error())
	case errors.Is(err, domain.ErrConversationNotFound),
		errors.Is(err, domain.ErrMessageNotFound):
		httperror.NotFound(w, r, err.Error())
	case errors.Is(err, domain.ErrConversationFinished),
		errors.Is(err, domain.ErrConversationNotAcceptingMessages),
		errors.Is(err, domain.ErrConversationDuplicateMessage),
		errors.Is(err, domain.ErrMessageAlreadyDeleted),
		errors.Is(err, domain.ErrMessageFailed):
		httperror.Conflict(w, r, err.Error())
	case errors.Is(err, domain.ErrMessageInvalidID),
		errors.Is(err, domain.ErrMessageStatusInvalid),
		errors.Is(err, domain.ErrMessageTypeInvalid),
		errors.Is(err, domain.ErrMessageEmptyText),
		errors.Is(err, domain.ErrMessageTextTooLong),
		errors.Is(err, domain.ErrMessageExternalIDInvalid),
		errors.Is(err, domain.ErrMessageExternalIDAlreadySet),
		errors.Is(err, domain.ErrMessageInvalidOwner),
		errors.Is(err, domain.ErrMessageNotText):
		httperror.Unprocessable(w, r, err.Error())
	default:
		httperror.Generic(w, r, err)
	}
}

func conversationIDFrom(r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil(), false
	}

	return id, true
}
