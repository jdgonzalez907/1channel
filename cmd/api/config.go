package main

import (
	"log/slog"
	"net"
	"net/url"
	"os"
	"strings"
)

func newLogger() *slog.Logger {
	level := slog.LevelInfo

	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

func httpPort() string {
	if port := os.Getenv("HTTP_PORT"); port != "" {
		return port
	}

	return "8080"
}

func postgresDSN() string {
	if dsn := os.Getenv("POSTGRES_URL"); dsn != "" {
		return dsn
	}

	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(os.Getenv("POSTGRES_USERNAME"), os.Getenv("POSTGRES_PASSWORD")),
		Host:     net.JoinHostPort(os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT")),
		Path:     os.Getenv("POSTGRES_DATABASE"),
		RawQuery: "sslmode=disable",
	}

	return u.String()
}
