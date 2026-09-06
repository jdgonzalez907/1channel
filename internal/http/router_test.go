package http

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/jdgonzalez907/1channel/internal/http/middleware"
)

func TestRouterHandlesRegisteredPatterns(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"matching method and path", nethttp.MethodGet, "/probe", nethttp.StatusOK},
		{"path match with wrong method returns 405", nethttp.MethodPost, "/probe", nethttp.StatusMethodNotAllowed},
		{"unregistered path returns 404", nethttp.MethodGet, "/absent", nethttp.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			router.Handle("GET /probe", func(w nethttp.ResponseWriter, r *nethttp.Request) {
				w.WriteHeader(nethttp.StatusOK)
			})

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(tt.method, tt.path, nil))

			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, recorder.Code)
			}
		})
	}
}

func TestRouterAppliesMiddlewaresInOrder(t *testing.T) {
	marker := func(name string) middleware.Middleware {
		return func(next nethttp.Handler) nethttp.Handler {
			return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
				w.Header().Add("X-Middleware", name)
				next.ServeHTTP(w, r)
			})
		}
	}

	router := NewRouter()
	router.Use(marker("first"), marker("second"))
	router.Handle("GET /probe", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.WriteHeader(nethttp.StatusTeapot)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(nethttp.MethodGet, "/probe", nil))

	got := recorder.Header().Values("X-Middleware")
	if len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("expected middlewares [first second], got %v", got)
	}
	if recorder.Code != nethttp.StatusTeapot {
		t.Fatalf("expected 418, got %d", recorder.Code)
	}
}
