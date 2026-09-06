package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecovery(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.Handler
		wantStatus int
	}{
		{
			name: "panic becomes 500 instead of crashing server",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("boom")
			}),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "normal handler passes through untouched",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			}),
			wantStatus: http.StatusTeapot,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			Recovery()(tt.handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, recorder.Code)
			}
		})
	}
}
