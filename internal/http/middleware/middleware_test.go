package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIDGeneration(t *testing.T) {
	id, err := GenerateRequestID()

	if err != nil {
		t.Errorf("not expect error: %v", err)
	}

	if len(id) != 32 {
		t.Errorf("expect len : %v, get : %v", 32, len(id))
	}
}

func TestIDGenerationTwice(t *testing.T) {
	id, err := GenerateRequestID()

	if err != nil {
		t.Errorf("not expect error: %v", err)
	}

	id2, err := GenerateRequestID()

	if err != nil {
		t.Errorf("not expect error: %v", err)
	}

	if len(id) != 32 {
		t.Errorf("expect len : %v, get : %v", 32, len(id))
	}

	if id == id2 {
		t.Errorf("id: %v, shouldn't be equal", id)
	}
}

func TestRequestIDToContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	res := httptest.NewRecorder()

	var generetaedID string

	handler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		generetaedID, _ = RequestIDFromContext(req.Context())
		res.Header().Add("ID-from-middleware", generetaedID)
	})

	wrappedHandler := RequestIDMiddleware(GenerateRequestID)(handler)

	wrappedHandler.ServeHTTP(res, req)

	if res.Header().Get("X-Request-ID") == "" {
		t.Errorf("expect value request_id in header!")
	}

	if generetaedID == "" || len(generetaedID) != 32 {
		t.Errorf("incorrect request_id, expect len %v, len %v", 32, len(generetaedID))
	}
}

func TestRequestIDToContextTwice(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	res := httptest.NewRecorder()

	var generetaedID string

	handler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		generetaedID, _ = RequestIDFromContext(req.Context())
		res.Header().Add("ID-from-middleware", generetaedID)
	})

	wrappedHandler := RequestIDMiddleware(GenerateRequestID)(handler)

	wrappedHandler.ServeHTTP(res, req)
	generetaedID1 := generetaedID
	wrappedHandler.ServeHTTP(res, req)

	if generetaedID1 == generetaedID {
		t.Errorf("expect unequls values: %v and %v", generetaedID1, generetaedID)
	}
}

func fakeGenerator() (string, error) {
	return "", fmt.Errorf("error in generating request_id")
}

func TestInvalidGeneration(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	res := httptest.NewRecorder()

	var generetaedID string
	var called bool

	handler := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		called = true
		generetaedID, _ = RequestIDFromContext(req.Context())
		res.Header().Add("ID-from-middleware", generetaedID)
	})
	wrappedHandler := RequestIDMiddleware(fakeGenerator)(handler)

	wrappedHandler.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Errorf("get: %v, expect: %v", res.Code, http.StatusInternalServerError)
	}

	if called {
		t.Errorf("expect that handler doesn't called")
	}
}

func TestSimpleRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	recorder := httptest.NewRecorder()
	res := NewResponseWriter(recorder)

	var okRequest bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okRequest = true
		w.Write([]byte("Hello"))
	})

	wrappedHandler := RecoveryMiddleware(New(slog.LevelInfo))(handler)
	wrappedHandler.ServeHTTP(res, req)

	if res.statusCode != http.StatusOK {
		t.Errorf("get value: %v, expect: %v", res.statusCode, http.StatusOK)
	}

	if !okRequest {
		t.Errorf("get value: %v, expect: %v", okRequest, true)
	}

	if recorder.Body.String() != "Hello" {
		t.Errorf("get value: %v, expect: %v", recorder.Body.String(), "Hello")
	}
}

func TestPanicRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	recorder := httptest.NewRecorder()
	res := NewResponseWriter(recorder)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("panic for test")
	})

	wrappedHandler := RecoveryMiddleware(New(slog.LevelInfo))(handler)
	wrappedHandler.ServeHTTP(res, req)

	if res.statusCode != http.StatusInternalServerError {
		t.Errorf("get value: %v, expect: %v", res.statusCode, http.StatusInternalServerError)
	}
}

func TestSimpleRequestAfterPanicRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req2 := httptest.NewRequest(http.MethodGet, "/by", nil)
	recorder := httptest.NewRecorder()
	res := NewResponseWriter(recorder)

	logger := New(slog.LevelInfo)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/hello" {
			panic("panic for test")
		}
	})

	wrappedHandler := RecoveryMiddleware(logger)(handler)
	wrappedHandler.ServeHTTP(res, req)

	if res.statusCode != http.StatusInternalServerError {
		t.Errorf("get value: %v, expect: %v", res.statusCode, http.StatusInternalServerError)
	}

	recorder = httptest.NewRecorder()
	res = NewResponseWriter(recorder)

	wrappedHandler.ServeHTTP(res, req2)

	if res.statusCode != http.StatusOK {
		t.Errorf("get value: %v, expect: %v", res.statusCode, http.StatusOK)
	}
}

func TestPanicRequestAfetrHeadersSet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	recorder := httptest.NewRecorder()
	res := NewResponseWriter(recorder)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		panic("panic for test")
	})

	wrappedHandler := RecoveryMiddleware(New(slog.LevelInfo))(handler)
	wrappedHandler.ServeHTTP(res, req)

	if res.statusCode != http.StatusOK {
		t.Errorf("get value: %v, expect: %v", res.statusCode, http.StatusOK)
	}
}

func TestNoOneMiddlewareComposition(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	res := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	wrappedHandler := Chain(handler)
	wrappedHandler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("get value: %v, expect: %v", res.Code, http.StatusOK)
	}
}

func TestOneMiddlewareComposition(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	res := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	var checker []string

	middleA := Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			checker = append(checker, "A: before")
			next.ServeHTTP(w, r)
			checker = append(checker, "A: after")
		})
	})

	wrappedHandler := Chain(handler, middleA)
	wrappedHandler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("get value: %v, expect: %v", res.Code, http.StatusOK)
	}

	if checker[0] != "A: before" {
		t.Errorf("checker value: %v, expect: %v", checker[0], "A: before")
	}

	if checker[1] != "A: after" {
		t.Errorf("checker value: %v, expect: %v", checker[1], "A: after")
	}
}

func TestTwoMiddlewareComposition(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	res := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	var checker []string

	middleA := Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			checker = append(checker, "A: before")
			next.ServeHTTP(w, r)
			checker = append(checker, "A: after")
		})
	})

	middleB := Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			checker = append(checker, "B: before")
			next.ServeHTTP(w, r)
			checker = append(checker, "B: after")
		})
	})

	wrappedHandler := Chain(handler, middleA, middleB)
	wrappedHandler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("get value: %v, expect: %v", res.Code, http.StatusOK)
	}

	if checker[0] != "A: before" {
		t.Errorf("checker value: %v, expect: %v", checker[0], "A: before")
	}

	if checker[1] != "B: before" {
		t.Errorf("checker value: %v, expect: %v", checker[1], "B: before")
	}

	if checker[2] != "B: after" {
		t.Errorf("checker value: %v, expect: %v", checker[2], "B: after")
	}

	if checker[3] != "A: after" {
		t.Errorf("checker value: %v, expect: %v", checker[3], "A: after")
	}
}
