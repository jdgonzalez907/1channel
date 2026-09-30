package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contactshttp "github.com/jdgonzalez907/1channel/internal/modules/contacts/infra/http"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/app"
	convhttp "github.com/jdgonzalez907/1channel/internal/modules/conversations/infra/http"
	convmeta "github.com/jdgonzalez907/1channel/internal/modules/conversations/infra/meta"
	usersapp "github.com/jdgonzalez907/1channel/internal/modules/users/app"
	usershttp "github.com/jdgonzalez907/1channel/internal/modules/users/infra/http"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb/sqlc"
)

func newTestRouter() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	userWriteHandler := usershttp.NewUserWriteHandler(&usersapp.MockCreateUser{})
	conversationWriteHandler := convhttp.NewConversationWriteHandler(
		&app.MockSendAgentMessage{},
		&app.MockAgentReadConversation{},
		&app.MockResolveConversation{},
	)
	contactWebhookHandler := newTestContactWebhookHandler()

	return newRouter(logger, dependencies{
		userWriteHandler:         userWriteHandler,
		userReadHandler:          usershttp.NewUserReadHandler(sqlc.New(nil)),
		conversationWriteHandler: conversationWriteHandler,
		contactWebhookHandler:    contactWebhookHandler,
		metaWebhookHandler:       newTestMetaWebhookHandler(),
		conversationReadHandler:  convhttp.NewConversationReadHandler(sqlc.New(nil)),
		contactReadHandler:       contactshttp.NewContactReadHandler(sqlc.New(nil)),
	})
}

func newTestContactWebhookHandler() *convhttp.ContactWebhookHandler {
	return convhttp.NewContactWebhookHandler(
		&app.MockReceiveContactMessage{},
		&app.MockReceiveContactMessageEdit{},
		&app.MockReceiveContactMessageDelete{},
		&app.MockReceiveContactMessageRead{},
	)
}

func newTestMetaWebhookHandler() *convmeta.WebhookHandler {
	return convmeta.NewWebhookHandler(
		&app.MockReceiveContactMessage{},
		&app.MockReceiveContactMessageEdit{},
		convmeta.Config{},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func TestRouter_Healthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	newTestRouter().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRouter_UnknownRouteReturnsProblemJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/does-not-exist", nil)
	rec := httptest.NewRecorder()

	newTestRouter().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, httperror.ContentType, rec.Header().Get("Content-Type"))
}

func TestRouter_MethodNotAllowedReturnsProblemJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/v1/users", nil)
	rec := httptest.NewRecorder()

	newTestRouter().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, httperror.ContentType, rec.Header().Get("Content-Type"))
}

func TestRouter_AuthenticatedRouteRequiresBearer(t *testing.T) {
	conversationID := uuid.NewV7()
	req := httptest.NewRequest(http.MethodPost, "/v1/conversations/"+conversationID.String()+"/messages", nil)
	rec := httptest.NewRecorder()

	newTestRouter().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRouter_ContactWebhookIsPublic(t *testing.T) {
	receive := &app.MockReceiveContactMessage{}
	receive.On("Execute", mock.Anything, mock.Anything).Return(nil).Once()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := newRouter(logger, dependencies{
		userWriteHandler: usershttp.NewUserWriteHandler(&usersapp.MockCreateUser{}),
		conversationWriteHandler: convhttp.NewConversationWriteHandler(
			&app.MockSendAgentMessage{},
			&app.MockAgentReadConversation{},
			&app.MockResolveConversation{},
		),
		contactWebhookHandler: convhttp.NewContactWebhookHandler(
			receive,
			&app.MockReceiveContactMessageEdit{},
			&app.MockReceiveContactMessageDelete{},
			&app.MockReceiveContactMessageRead{},
		),
		metaWebhookHandler:      newTestMetaWebhookHandler(),
		userReadHandler:         usershttp.NewUserReadHandler(sqlc.New(nil)),
		conversationReadHandler: convhttp.NewConversationReadHandler(sqlc.New(nil)),
		contactReadHandler:      contactshttp.NewContactReadHandler(sqlc.New(nil)),
	})

	body := `{"event":"message.received","external_contact_id":"54911","external_message_id":"wamid.1","text":"hola"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/1channel", strings.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	receive.AssertExpectations(t)
}

func TestRouter_ReadRoutesRequireBearer(t *testing.T) {
	paths := []string{
		"/v1/conversations?status=open",
		"/v1/conversations/" + uuid.NewV7().String(),
		"/v1/contacts/" + uuid.NewV7().String(),
		"/v1/users/" + uuid.NewV7().String(),
	}

	router := newTestRouter()
	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code, path)
		assert.NotEqual(t, http.StatusNotFound, rec.Code, path)
	}
}

func TestRouter_CreateUserIsPublic(t *testing.T) {
	create := &usersapp.MockCreateUser{}
	create.On("Execute", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := newRouter(logger, dependencies{
		userWriteHandler: usershttp.NewUserWriteHandler(create),
		conversationWriteHandler: convhttp.NewConversationWriteHandler(
			&app.MockSendAgentMessage{},
			&app.MockAgentReadConversation{},
			&app.MockResolveConversation{},
		),
		userReadHandler:         usershttp.NewUserReadHandler(sqlc.New(nil)),
		conversationReadHandler: convhttp.NewConversationReadHandler(sqlc.New(nil)),
		contactReadHandler:      contactshttp.NewContactReadHandler(sqlc.New(nil)),
		metaWebhookHandler:      newTestMetaWebhookHandler(),
		userLookup:              func(context.Context, uuid.UUID) error { return nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/users", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusUnauthorized, rec.Code)
	create.AssertExpectations(t)
}
