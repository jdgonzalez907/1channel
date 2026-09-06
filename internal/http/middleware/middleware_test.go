package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChainAppliesMiddlewaresInOnionOrder(t *testing.T) {
	tracker := func(got *[]string, name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				*got = append(*got, "before:"+name)
				next.ServeHTTP(w, r)
				*got = append(*got, "after:"+name)
			})
		}
	}

	tests := []struct {
		name  string
		build func(got *[]string) []Middleware
		want  []string
	}{
		{
			name:  "no middlewares passes through",
			build: func(*[]string) []Middleware { return nil },
			want:  []string{"handler"},
		},
		{
			name:  "single middleware wraps handler",
			build: func(g *[]string) []Middleware { return []Middleware{tracker(g, "a")} },
			want:  []string{"before:a", "handler", "after:a"},
		},
		{
			name:  "two middlewares nest in declaration order",
			build: func(g *[]string) []Middleware { return []Middleware{tracker(g, "a"), tracker(g, "b")} },
			want:  []string{"before:a", "before:b", "handler", "after:b", "after:a"},
		},
		{
			name: "three middlewares keep onion order",
			build: func(g *[]string) []Middleware {
				return []Middleware{tracker(g, "a"), tracker(g, "b"), tracker(g, "c")}
			},
			want: []string{"before:a", "before:b", "before:c", "handler", "after:c", "after:b", "after:a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = append(got, "handler")
			})

			Chain(handler, tt.build(&got)...).ServeHTTP(
				httptest.NewRecorder(),
				httptest.NewRequest(http.MethodGet, "/", nil),
			)

			if len(got) != len(tt.want) {
				t.Fatalf("expected %d entries %v, got %v", len(tt.want), tt.want, got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("position %d: expected %q, got %v", i, tt.want[i], got)
				}
			}
		})
	}
}
