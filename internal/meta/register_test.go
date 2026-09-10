package meta

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jdgonzalez907/1channel/internal/config"
	httprouter "github.com/jdgonzalez907/1channel/internal/http"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRoutes(t *testing.T) {
	const secret = "app-secret"
	const verifyToken = "verify-token"

	validBody := []byte(`{"object":"whatsapp_business_account"}`)

	tests := []struct {
		title     string
		method    string
		target    string
		expStatus int
		expBody   string
	}{
		{
			title:     "success - get webhook with valid token echoes challenge",
			method:    http.MethodGet,
			target:    "/meta/webhook?hub.mode=subscribe&hub.verify_token=verify-token&hub.challenge=c1",
			expStatus: http.StatusOK,
			expBody:   "c1",
		},
		{
			title:     "failure - post webhook without signature is rejected",
			method:    http.MethodPost,
			target:    "/meta/webhook",
			expStatus: http.StatusBadRequest,
			expBody:   "Bad request\n",
		},
		{
			title:     "failure - unknown path is 404",
			method:    http.MethodGet,
			target:    "/meta/hash",
			expStatus: http.StatusNotFound,
		},
		{
			title:     "failure - wrong method on webhook path is 405",
			method:    http.MethodDelete,
			target:    "/meta/webhook",
			expStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			router := httprouter.NewRouter()
			RegisterRoutes(router, config.NewMockSecretsConfiguration(secret, verifyToken), nil)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.target, bytes.NewReader(validBody))

			// Act
			router.ServeHTTP(recorder, request)

			// Assert
			assert.Equal(t, tt.expStatus, recorder.Code)
			if tt.expBody != "" {
				assert.Equal(t, tt.expBody, recorder.Body.String())
			}
		})
	}
}
