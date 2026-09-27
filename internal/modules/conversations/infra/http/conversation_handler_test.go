package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/app"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/jdgonzalez907/1channel/internal/modules/users"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/middleware"
)

func newConversationRouter(h *ConversationHandler) chi.Router {
	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(middleware.Auth)
		h.Register(r)
	})

	return router
}

func doConversationRequest(router chi.Router, method, path, body string, agentID *uuid.UUID) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, path, reader)
	if agentID != nil {
		req.Header.Set("Authorization", "Bearer "+agentID.String())
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestConversationHandler_SendMessage(t *testing.T) {
	agentID := uuid.NewV7()
	conversationID := uuid.NewV7()
	sendErr := errors.New("send failure")

	tests := []struct {
		name       string
		path       string
		body       string
		agentID    *uuid.UUID
		setup      func(t *testing.T, m *app.MockSendAgentMessage)
		wantStatus int
	}{
		{
			name:    "sends message",
			path:    "/conversations/" + conversationID.String() + "/messages",
			body:    `{"text":"hola"}`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockSendAgentMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.MatchedBy(func(in app.SendAgentMessageInput) bool {
					return in.ConversationID == conversationID && in.AgentID == agentID && in.Text == "hola"
				})).Return(nil).Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid conversation id",
			path:       "/conversations/not-a-uuid/messages",
			body:       `{"text":"hola"}`,
			agentID:    &agentID,
			setup:      func(t *testing.T, m *app.MockSendAgentMessage) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing bearer",
			path:       "/conversations/" + conversationID.String() + "/messages",
			body:       `{"text":"hola"}`,
			setup:      func(t *testing.T, m *app.MockSendAgentMessage) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "malformed body",
			path:    "/conversations/" + conversationID.String() + "/messages",
			body:    `{"text":`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockSendAgentMessage) {
				t.Helper()
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "conversation not found",
			path:    "/conversations/" + conversationID.String() + "/messages",
			body:    `{"text":"hola"}`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockSendAgentMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).
					Return(errors.Join(app.ErrSendingAgentMessage, domain.ErrConversationNotFound)).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "agent not owner",
			path:    "/conversations/" + conversationID.String() + "/messages",
			body:    `{"text":"hola"}`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockSendAgentMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).
					Return(errors.Join(app.ErrSendingAgentMessage, domain.ErrConversationAgentNotOwner)).Once()
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:    "conversation not accepting messages",
			path:    "/conversations/" + conversationID.String() + "/messages",
			body:    `{"text":"hola"}`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockSendAgentMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).
					Return(errors.Join(app.ErrSendingAgentMessage, domain.ErrConversationNotAcceptingMessages)).Once()
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:    "invalid text",
			path:    "/conversations/" + conversationID.String() + "/messages",
			body:    `{"text":""}`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockSendAgentMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).
					Return(errors.Join(app.ErrSendingAgentMessage, domain.ErrMessageEmptyText)).Once()
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "unknown user",
			path:    "/conversations/" + conversationID.String() + "/messages",
			body:    `{"text":"hola"}`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockSendAgentMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).
					Return(errors.Join(app.ErrSendingAgentMessage, users.ErrUserNotFound)).Once()
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "use case error returns 500",
			path:    "/conversations/" + conversationID.String() + "/messages",
			body:    `{"text":"hola"}`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockSendAgentMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(sendErr).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:    "request timed out",
			path:    "/conversations/" + conversationID.String() + "/messages",
			body:    `{"text":"hola"}`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockSendAgentMessage) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).
					Return(errors.Join(app.ErrSendingAgentMessage, context.DeadlineExceeded)).Once()
			},
			wantStatus: http.StatusGatewayTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			send := &app.MockSendAgentMessage{}
			tt.setup(t, send)

			router := newConversationRouter(NewConversationHandler(send, &app.MockAgentReadConversation{}, &app.MockResolveConversation{}))

			// Act
			rec := doConversationRequest(router, http.MethodPost, tt.path, tt.body, tt.agentID)

			// Assert
			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantStatus == http.StatusCreated {
				var resp MessageResponse
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.NotEmpty(t, resp.ID)
			}

			send.AssertExpectations(t)
		})
	}
}

func TestConversationHandler_UpdateMessages(t *testing.T) {
	agentID := uuid.NewV7()
	conversationID := uuid.NewV7()

	tests := []struct {
		name       string
		body       string
		agentID    *uuid.UUID
		setup      func(t *testing.T, m *app.MockAgentReadConversation)
		wantStatus int
	}{
		{
			name:    "marks messages as read",
			body:    `{"status":"read"}`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockAgentReadConversation) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.MatchedBy(func(in app.AgentReadConversationInput) bool {
					return in.ConversationID == conversationID && in.AgentID == agentID
				})).Return(nil).Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "status not allowed",
			body:       `{"status":"deleted"}`,
			agentID:    &agentID,
			setup:      func(t *testing.T, m *app.MockAgentReadConversation) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "agent not owner",
			body:    `{"status":"read"}`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockAgentReadConversation) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).
					Return(errors.Join(app.ErrAgentReadingConversation, domain.ErrConversationAgentNotOwner)).Once()
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			read := &app.MockAgentReadConversation{}
			tt.setup(t, read)

			router := newConversationRouter(NewConversationHandler(&app.MockSendAgentMessage{}, read, &app.MockResolveConversation{}))

			// Act
			rec := doConversationRequest(router, http.MethodPatch, "/conversations/"+conversationID.String()+"/messages", tt.body, tt.agentID)

			// Assert
			assert.Equal(t, tt.wantStatus, rec.Code)
			read.AssertExpectations(t)
		})
	}
}

func TestConversationHandler_UpdateConversation(t *testing.T) {
	agentID := uuid.NewV7()
	conversationID := uuid.NewV7()

	tests := []struct {
		name       string
		body       string
		agentID    *uuid.UUID
		setup      func(t *testing.T, m *app.MockResolveConversation)
		wantStatus int
	}{
		{
			name:    "resolves conversation",
			body:    `{"status":"resolved"}`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockResolveConversation) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.MatchedBy(func(in app.ResolveConversationInput) bool {
					return in.ConversationID == conversationID && in.AgentID == agentID
				})).Return(nil).Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "agent cannot expire",
			body:       `{"status":"expired"}`,
			agentID:    &agentID,
			setup:      func(t *testing.T, m *app.MockResolveConversation) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "agent not owner",
			body:    `{"status":"resolved"}`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockResolveConversation) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).
					Return(errors.Join(app.ErrResolvingConversation, domain.ErrConversationAgentNotOwner)).Once()
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:    "conversation not found",
			body:    `{"status":"resolved"}`,
			agentID: &agentID,
			setup: func(t *testing.T, m *app.MockResolveConversation) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).
					Return(errors.Join(app.ErrResolvingConversation, domain.ErrConversationNotFound)).Once()
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			resolve := &app.MockResolveConversation{}
			tt.setup(t, resolve)

			router := newConversationRouter(NewConversationHandler(&app.MockSendAgentMessage{}, &app.MockAgentReadConversation{}, resolve))

			// Act
			rec := doConversationRequest(router, http.MethodPatch, "/conversations/"+conversationID.String(), tt.body, tt.agentID)

			// Assert
			assert.Equal(t, tt.wantStatus, rec.Code)
			resolve.AssertExpectations(t)
		})
	}
}
