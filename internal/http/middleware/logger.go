package middleware

import (
	"log/slog"
	"net/http"
	"os"
)

func New(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	return logger
}

func Chain(handler http.Handler, middleware ...Middleware) http.Handler {
	if len(middleware) == 0 {
		return handler
	}

	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}
