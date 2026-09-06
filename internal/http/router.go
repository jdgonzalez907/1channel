package http

import (
	nethttp "net/http"

	"github.com/jdgonzalez907/1channel/internal/http/middleware"
)

type Router struct {
	mux         *nethttp.ServeMux
	middlewares []middleware.Middleware
}

func NewRouter() *Router {
	return &Router{mux: nethttp.NewServeMux()}
}

func (r *Router) Handle(pattern string, handler nethttp.HandlerFunc) {
	r.mux.HandleFunc(pattern, handler)
}

func (r *Router) Use(mws ...middleware.Middleware) {
	r.middlewares = append(r.middlewares, mws...)
}

func (r *Router) ServeHTTP(w nethttp.ResponseWriter, req *nethttp.Request) {
	middleware.Chain(r.mux, r.middlewares...).ServeHTTP(w, req)
}
