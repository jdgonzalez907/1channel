package httperror

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	Write(rec, req, Problem{
		Title:  "Unprocessable Entity",
		Status: http.StatusUnprocessableEntity,
		Detail: "message text is empty",
	})

	assert.Equal(t, ContentType, rec.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var got Problem
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "Unprocessable Entity", got.Title)
	assert.Equal(t, http.StatusUnprocessableEntity, got.Status)
	assert.Equal(t, "message text is empty", got.Detail)
	assert.Empty(t, got.Instance)
}

func TestWrite_SetsInstanceFromRequestID(t *testing.T) {
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		NotFound(w, r, "boom")
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	var got Problem
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.NotEmpty(t, got.Instance)
}

func TestGeneric(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "deadline exceeded returns 504",
			err:        context.DeadlineExceeded,
			wantStatus: http.StatusGatewayTimeout,
		},
		{
			name:       "unknown error returns 500",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			Generic(rec, req, tt.err)

			assert.Equal(t, ContentType, rec.Header().Get("Content-Type"))
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
