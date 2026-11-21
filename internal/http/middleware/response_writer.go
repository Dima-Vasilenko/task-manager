package middleware

import (
	"net/http"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	hasWroten  bool
}

func (r *responseWriter) WriteHeader(status int) {
	if r.hasWroten {
		return
	}

	r.hasWroten = true

	r.statusCode = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseWriter) Write(data []byte) (int, error) {
	if !r.hasWroten {
		r.WriteHeader(http.StatusOK)
	}

	return r.ResponseWriter.Write(data)
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}
