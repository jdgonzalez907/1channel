package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/app"
	convhttp "github.com/jdgonzalez907/1channel/internal/modules/conversations/infra/http"
	usersapp "github.com/jdgonzalez907/1channel/internal/modules/users/app"
	usershttp "github.com/jdgonzalez907/1channel/internal/modules/users/infra/http"
)

func newTestRouter() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	usersHandler := usershttp.NewUserHandler(&usersapp.MockCreateUser{})
	conversationsHandler := convhttp.NewConversationHandler(
		&app.MockSendAgentMessage{},
		&app.MockAgentReadConversation{},
		&app.MockResolveConversation{},
	)

	return newRouter(logger, dependencies{
		usersHandler:         usersHandler,
		conversationsHandler: conversationsHandler,
	})
}

func TestRouter_Healthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	newTestRouter().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRouter_AuthenticatedRouteRequiresBearer(t *testing.T) {
	conversationID := uuid.NewV7()
	req := httptest.NewRequest(http.MethodPost, "/v1/conversations/"+conversationID.String()+"/messages", nil)
	rec := httptest.NewRecorder()

	newTestRouter().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRouter_CreateUserIsPublic(t *testing.T) {
	create := &usersapp.MockCreateUser{}
	create.On("Execute", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := newRouter(logger, dependencies{
		usersHandler: usershttp.NewUserHandler(create),
		conversationsHandler: convhttp.NewConversationHandler(
			&app.MockSendAgentMessage{},
			&app.MockAgentReadConversation{},
			&app.MockResolveConversation{},
		),
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/users", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusUnauthorized, rec.Code)
	create.AssertExpectations(t)
}
