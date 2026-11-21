package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Dima-Vasilenko/task-manager/internal/domain"
	"github.com/Dima-Vasilenko/task-manager/internal/handler"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func setUpTestDB(t *testing.T) *sql.DB {
	t.Helper()

	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Файл .env не найден, используем системные переменные")
	}

	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		t.Skip("test database is not set")
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("error %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("filed to close database %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("no connection sets %v", err)
	}

	_, errClear := db.ExecContext(ctx, "TRUNCATE TABLE app.tasks RESTART IDENTITY CASCADE")
	if errClear != nil {
		t.Errorf("can't clear table %v", err)
	}

	return db
}

func CreateTestTask(t *testing.T, route http.Handler) domain.Task {
	t.Helper()
	payload := handler.CreateTaskRequest{Title: "First Task", Description: "10 byn", Importance: 4}
	body, mErr := json.Marshal(payload)
	if mErr != nil {
		t.Fatalf("error in marshalling: %v", mErr)
	}

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
	res := httptest.NewRecorder()

	route.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("status %v, required %v", res.Code, http.StatusCreated)
	}

	if res.Header().Get("X-Request-ID") == "" {
		t.Fatalf("resuest_id is empty: %v", res.Header().Get("X-Request-ID"))
	}

	var task domain.Task
	if err := json.NewDecoder(res.Body).Decode(&task); err != nil {
		t.Fatalf("не удалось распарсить ответ: %v", err)
	}
	return task
}
