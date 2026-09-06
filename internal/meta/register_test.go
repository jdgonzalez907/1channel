package meta

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	httprouter "github.com/jdgonzalez907/1channel/internal/http"
)

func TestRegisterRoutes(t *testing.T) {
	const secret = "app-secret"

	newRouter := func() *httprouter.Router {
		router := httprouter.NewRouter()
		RegisterRoutes(router, fakeConfig{
			metaSecret:       secret,
			oneChannelSecret: "verify-token",
		})
		return router
	}

	validBody := []byte(`{"object":"whatsapp_business_account"}`)

	tests := []struct {
		name       string
		method     string
		target     string
		header     func() (string, string)
		wantStatus int
		wantBody   string
	}{
		{
			name:       "GET webhook with valid token echoes challenge",
			method:     http.MethodGet,
			target:     "/meta/webhook?hub.mode=subscribe&hub.verify_token=verify-token&hub.challenge=c1",
			wantStatus: http.StatusOK,
			wantBody:   "c1",
		},
		{
			name:       "POST webhook with valid signature accepted",
			method:     http.MethodPost,
			target:     "/meta/webhook",
			header:     func() (string, string) { return signatureHeader, sign(secret, validBody) },
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST webhook without signature rejected",
			method:     http.MethodPost,
			target:     "/meta/webhook",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown path is 404",
			method:     http.MethodGet,
			target:     "/meta/hash",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "wrong method on webhook path is 405",
			method:     http.MethodDelete,
			target:     "/meta/webhook",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.target, bytes.NewReader(validBody))
			if tt.header != nil {
				key, value := tt.header()
				request.Header.Set(key, value)
			}

			newRouter().ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, recorder.Code)
			}
			if tt.wantBody != "" && recorder.Body.String() != tt.wantBody {
				t.Fatalf("expected body %q, got %q", tt.wantBody, recorder.Body.String())
			}
		})
	}
}
