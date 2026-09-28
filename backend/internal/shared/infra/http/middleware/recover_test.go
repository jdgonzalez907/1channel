package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
)

func TestRecover(t *testing.T) {
	// Arrange
	var logged bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logged, nil))
	handler := Recover(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	// Act
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, httperror.ContentType, rec.Header().Get("Content-Type"))

	var problem httperror.Problem
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &problem))
	assert.Equal(t, http.StatusInternalServerError, problem.Status)
	assert.Equal(t, "Internal Server Error", problem.Title)

	assert.Contains(t, logged.String(), "panic recovered")
	assert.Contains(t, logged.String(), "boom")
	assert.Contains(t, logged.String(), "/panic")
}

func TestRecover_PropagatesAbortHandler(t *testing.T) {
	// Arrange
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := Recover(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	// Act + Assert
	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/abort", nil))
	})
}

func TestRecover_NoPanic(t *testing.T) {
	// Arrange
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := Recover(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	// Act
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ok", nil))

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
}
