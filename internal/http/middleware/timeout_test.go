package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeout(t *testing.T) {
	tests := []struct {
		title     string
		timeout   time.Duration
		slow      bool
		expStatus int
		expBody   string
	}{
		{
			title:     "success - fast handler answers normally",
			timeout:   time.Second,
			slow:      false,
			expStatus: http.StatusOK,
			expBody:   "ok",
		},
		{
			title:     "failure - blocked handler gets 503 timeout",
			timeout:   20 * time.Millisecond,
			slow:      true,
			expStatus: http.StatusServiceUnavailable,
			expBody:   timeoutResponseMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			release := make(chan struct{})
			t.Cleanup(func() { close(release) })
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.slow {
					<-release
					return
				}
				w.Write([]byte("ok"))
			})
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)

			// Act
			Timeout(tt.timeout)(handler).ServeHTTP(recorder, request)

			// Assert
			assert.Equal(t, tt.expStatus, recorder.Code)
			assert.Equal(t, tt.expBody, recorder.Body.String())
		})
	}
}
