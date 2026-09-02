package config

import (
	"net/http"
)

func NewServer(cfg *Config) *http.ServeMux {
	mux := http.NewServeMux()

	return mux
}
