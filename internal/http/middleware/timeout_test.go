package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		handler  http.Handler
		wantCode int
		wantBody string
	}{
		{
			name:    "slow handler gets 503 timeout",
			timeout: 20 * time.Millisecond,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(200 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			}),
			wantCode: http.StatusServiceUnavailable,
			wantBody: timeoutResponseMessage,
		},
		{
			name:    "fast handler answers normally",
			timeout: time.Second,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("ok"))
			}),
			wantCode: http.StatusOK,
			wantBody: "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			Timeout(tt.timeout)(tt.handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

			if recorder.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d", tt.wantCode, recorder.Code)
			}
			if recorder.Body.String() != tt.wantBody {
				t.Fatalf("expected body %q, got %q", tt.wantBody, recorder.Body.String())
			}
		})
	}
}
