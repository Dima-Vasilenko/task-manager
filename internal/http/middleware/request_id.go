package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func GenerateRequestID() (string, error) {
	id := make([]byte, 16)

	if _, err := rand.Read(id); err != nil {
		return "", err
	}

	return hex.EncodeToString(id), nil
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	res, ok := ctx.Value(requestIDKey).(string)
	return res, ok
}

func RequestIDMiddleware(generator func() (string, error)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID, err := generator()
			if err != nil {
				http.Error(w, "error in generating request_id", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r = r.WithContext(ctx)

			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r)
		})
	}
}
