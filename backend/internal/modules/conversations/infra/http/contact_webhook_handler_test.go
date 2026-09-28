package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/app"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

type webhookMocks struct {
	receive *app.MockReceiveContactMessage
	edit    *app.MockReceiveContactMessageEdit
	delete  *app.MockReceiveContactMessageDelete
	read    *app.MockReceiveContactMessageRead
}

func newWebhookMocks() *webhookMocks {
	return &webhookMocks{
		receive: &app.MockReceiveContactMessage{},
		edit:    &app.MockReceiveContactMessageEdit{},
		delete:  &app.MockReceiveContactMessageDelete{},
		read:    &app.MockReceiveContactMessageRead{},
	}
}

func newContactWebhookRouter(m *webhookMocks) chi.Router {
	router := chi.NewRouter()
	NewContactWebhookHandler(m.receive, m.edit, m.delete, m.read).Register(router)

	return router
}

func doWebhookRequest(router chi.Router, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/webhooks/1channel", strings.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestContactWebhookHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(t *testing.T, m *webhookMocks)
		wantStatus int
	}{
		{
			name: "message.received calls receive use case",
			body: `{"event":"message.received","external_contact_id":"54911","external_message_id":"wamid.1","text":"hola"}`,
			setup: func(t *testing.T, m *webhookMocks) {
				t.Helper()
				m.receive.On("Execute", mock.Anything, mock.MatchedBy(func(in app.ReceiveContactMessageInput) bool {
					return in.ConversationID != uuid.Nil() &&
						in.MessageID != uuid.Nil() &&
						in.ExternalContactID == "54911" &&
						in.ExternalMessageID != nil && *in.ExternalMessageID == "wamid.1" &&
						in.Text == "hola" &&
						in.ReceivedAt.Location() == time.UTC
				})).Return(nil).Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "message.received without external message id is rejected",
			body:       `{"event":"message.received","external_contact_id":"54911","text":"hola"}`,
			setup:      func(t *testing.T, m *webhookMocks) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "message.edited calls edit use case",
			body: `{"event":"message.edited","external_contact_id":"54911","external_message_id":"wamid.1","text":"hola"}`,
			setup: func(t *testing.T, m *webhookMocks) {
				t.Helper()
				m.edit.On("Execute", mock.Anything, mock.MatchedBy(func(in app.ReceiveContactMessageEditInput) bool {
					return in.ExternalContactID == "54911" &&
						in.ExternalMessageID == "wamid.1" &&
						in.NewText == "hola" &&
						in.EditedAt.Location() == time.UTC
				})).Return(nil).Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "message.deleted calls delete use case",
			body: `{"event":"message.deleted","external_contact_id":"54911","external_message_id":"wamid.1"}`,
			setup: func(t *testing.T, m *webhookMocks) {
				t.Helper()
				m.delete.On("Execute", mock.Anything, mock.MatchedBy(func(in app.ReceiveContactMessageDeleteInput) bool {
					return in.ExternalContactID == "54911" &&
						in.ExternalMessageID == "wamid.1" &&
						in.DeletedAt.Location() == time.UTC
				})).Return(nil).Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "message.read calls read use case",
			body: `{"event":"message.read","external_contact_id":"54911","external_message_id":"wamid.1"}`,
			setup: func(t *testing.T, m *webhookMocks) {
				t.Helper()
				m.read.On("Execute", mock.Anything, mock.MatchedBy(func(in app.ReceiveContactMessageReadInput) bool {
					return in.ExternalContactID == "54911" &&
						in.ExternalMessageID == "wamid.1" &&
						in.ReadAt.Location() == time.UTC
				})).Return(nil).Once()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "unknown event is ignored",
			body:       `{"event":"message.unknown","external_contact_id":"54911","external_message_id":"wamid.1"}`,
			setup:      func(t *testing.T, m *webhookMocks) {},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing required field is rejected",
			body:       `{"event":"message.received","external_contact_id":"54911"}`,
			setup:      func(t *testing.T, m *webhookMocks) {},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "malformed body is rejected",
			body:       `{"event":`,
			setup:      func(t *testing.T, m *webhookMocks) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "domain not found maps to 404",
			body: `{"event":"message.deleted","external_contact_id":"54911","external_message_id":"wamid.1"}`,
			setup: func(t *testing.T, m *webhookMocks) {
				t.Helper()
				m.delete.On("Execute", mock.Anything, mock.Anything).
					Return(errors.Join(app.ErrReceivingContactMessageDelete, domain.ErrMessageNotFound)).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "domain ownership error maps to 403",
			body: `{"event":"message.read","external_contact_id":"54911","external_message_id":"wamid.1"}`,
			setup: func(t *testing.T, m *webhookMocks) {
				t.Helper()
				m.read.On("Execute", mock.Anything, mock.Anything).
					Return(errors.Join(app.ErrReceivingContactMessageRead, domain.ErrConversationContactNotOwner)).Once()
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			m := newWebhookMocks()
			tt.setup(t, m)
			router := newContactWebhookRouter(m)

			// Act
			rec := doWebhookRequest(router, tt.body)

			// Assert
			assert.Equal(t, tt.wantStatus, rec.Code)

			m.receive.AssertExpectations(t)
			m.edit.AssertExpectations(t)
			m.delete.AssertExpectations(t)
			m.read.AssertExpectations(t)
		})
	}
}
