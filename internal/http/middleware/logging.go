package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type Middleware func(next http.Handler) http.Handler

func LoggingMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			writer := NewResponseWriter(w)

			next.ServeHTTP(writer, r)

			duration := time.Since(start)

			requestID, _ := RequestIDFromContext(r.Context())

			logger.Info(
				"request completed",
				"request_id", requestID,
				"metod", r.Method,
				"url", r.URL.Path,
				"status", writer.statusCode,
				"duration", duration.String(),
			)
		})
	}
}
