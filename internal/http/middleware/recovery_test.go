package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecovery(t *testing.T) {
	tests := []struct {
		title     string
		handler   http.Handler
		expStatus int
		expBody   string
		expLogged string
	}{
		{
			title: "success - normal handler passes through untouched",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			}),
			expStatus: http.StatusTeapot,
			expBody:   "",
			expLogged: "",
		},
		{
			title: "failure - panic becomes 500 instead of crashing server",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("boom")
			}),
			expStatus: http.StatusInternalServerError,
			expBody:   "Internal Server Error\n",
			expLogged: "panic recovered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			logs := captureSlogLogs(t)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)

			// Act
			Recovery()(tt.handler).ServeHTTP(recorder, request)

			// Assert
			assert.Equal(t, tt.expStatus, recorder.Code)
			assert.Equal(t, tt.expBody, recorder.Body.String())
			if tt.expLogged != "" {
				assert.Contains(t, logs.String(), tt.expLogged)
			} else {
				assert.Empty(t, logs.String())
			}
		})
	}
}
