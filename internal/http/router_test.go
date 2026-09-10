package http

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/jdgonzalez907/1channel/internal/http/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	// Act
	router := NewRouter()

	// Assert
	require.NotNil(t, router)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(nethttp.MethodGet, "/anything", nil))
	assert.Equal(t, nethttp.StatusNotFound, recorder.Code)
}

func TestHandle(t *testing.T) {
	// Arrange
	router := NewRouter()

	// Act
	router.Handle("GET /probe", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.WriteHeader(nethttp.StatusTeapot)
	})

	// Assert
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(nethttp.MethodGet, "/probe", nil))
	assert.Equal(t, nethttp.StatusTeapot, recorder.Code)
}

func TestUse(t *testing.T) {
	marker := func(name string) middleware.Middleware {
		return func(next nethttp.Handler) nethttp.Handler {
			return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
				w.Header().Add("X-Middleware", name)
				next.ServeHTTP(w, r)
			})
		}
	}

	tests := []struct {
		title     string
		use       [][]middleware.Middleware
		expected  []string
		expResult int
	}{
		{
			title:     "success - middlewares run in declaration order",
			use:       [][]middleware.Middleware{{marker("first"), marker("second")}},
			expected:  []string{"first", "second"},
			expResult: nethttp.StatusTeapot,
		},
		{
			title:     "success - repeated use calls accumulate in order",
			use:       [][]middleware.Middleware{{marker("first")}, {marker("second")}},
			expected:  []string{"first", "second"},
			expResult: nethttp.StatusTeapot,
		},
		{
			title:     "success - no middlewares leaves response untouched",
			use:       nil,
			expected:  nil,
			expResult: nethttp.StatusTeapot,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			router := NewRouter()
			router.Handle("GET /probe", func(w nethttp.ResponseWriter, r *nethttp.Request) {
				w.WriteHeader(nethttp.StatusTeapot)
			})

			// Act
			for _, group := range tt.use {
				router.Use(group...)
			}

			// Assert
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(nethttp.MethodGet, "/probe", nil))
			assert.Equal(t, tt.expected, recorder.Header().Values("X-Middleware"))
			assert.Equal(t, tt.expResult, recorder.Code)
		})
	}
}

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		title     string
		method    string
		path      string
		expResult int
	}{
		{
			title:     "success - matching method and path reaches handler",
			method:    nethttp.MethodGet,
			path:      "/probe",
			expResult: nethttp.StatusOK,
		},
		{
			title:     "failure - path match with wrong method returns 405",
			method:    nethttp.MethodPost,
			path:      "/probe",
			expResult: nethttp.StatusMethodNotAllowed,
		},
		{
			title:     "failure - unregistered path returns 404",
			method:    nethttp.MethodGet,
			path:      "/absent",
			expResult: nethttp.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// Arrange
			router := NewRouter()
			router.Handle("GET /probe", func(w nethttp.ResponseWriter, r *nethttp.Request) {
				w.WriteHeader(nethttp.StatusOK)
			})
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.path, nil)

			// Act
			router.ServeHTTP(recorder, request)

			// Assert
			assert.Equal(t, tt.expResult, recorder.Code)
		})
	}
}
