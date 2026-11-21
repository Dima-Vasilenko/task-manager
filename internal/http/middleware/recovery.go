package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func RecoveryMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrappedWriter := NewResponseWriter(w)
			defer func() {
				if rec := recover(); rec != nil {

					logger.Error(
						"panic occured",
						"panic", rec,
						"method", r.Method,
						"url", r.URL.Path,
						"stack trace", string(debug.Stack()),
					)

					if !wrappedWriter.hasWroten {
						http.Error(w, "Internal server error", http.StatusInternalServerError)
						return
					}
				}
			}()

			next.ServeHTTP(wrappedWriter, r)
		})
	}
}
