package http

import (
	"errors"
	"net/http"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/app"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httputil"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/middleware"
)

type ConversationWriteHandler struct {
	sendAgentMessage      app.SendAgentMessage
	agentReadConversation app.AgentReadConversation
	resolveConversation   app.ResolveConversation
}

func NewConversationWriteHandler(
	sendAgentMessage app.SendAgentMessage,
	agentReadConversation app.AgentReadConversation,
	resolveConversation app.ResolveConversation,
) *ConversationWriteHandler {
	return &ConversationWriteHandler{
		sendAgentMessage:      sendAgentMessage,
		agentReadConversation: agentReadConversation,
		resolveConversation:   resolveConversation,
	}
}

func (h *ConversationWriteHandler) Register(r chi.Router) {
	r.Post("/conversations/{id}/messages", h.handleSendAgentMessage)
	r.Patch("/conversations/{id}/messages", h.handleMarkMessagesAsRead)
	r.Patch("/conversations/{id}", h.handleResolveConversation)
}

func (h *ConversationWriteHandler) handleSendAgentMessage(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := httputil.RequirePathUUID(w, r, "id", "invalid conversation id")
	if !ok {
		return
	}

	agentID, ok := middleware.RequireUserID(w, r)
	if !ok {
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

func (h *ConversationWriteHandler) handleMarkMessagesAsRead(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := httputil.RequirePathUUID(w, r, "id", "invalid conversation id")
	if !ok {
		return
	}

	agentID, ok := middleware.RequireUserID(w, r)
	if !ok {
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

func (h *ConversationWriteHandler) handleResolveConversation(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := httputil.RequirePathUUID(w, r, "id", "invalid conversation id")
	if !ok {
		return
	}

	agentID, ok := middleware.RequireUserID(w, r)
	if !ok {
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

func (h *ConversationWriteHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
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
