package http

import (
	nethttp "net/http"
)

type Router struct {
	mux *nethttp.ServeMux
}

func NewRouter() *Router {
	return &Router{mux: nethttp.NewServeMux()}
}

func (r *Router) Handle(pattern string, handler nethttp.HandlerFunc) {
	r.mux.HandleFunc(pattern, handler)
}

func (r *Router) ServeHTTP(w nethttp.ResponseWriter, req *nethttp.Request) {
	r.mux.ServeHTTP(w, req)
}
