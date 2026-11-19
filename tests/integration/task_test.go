package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dima-Vasilenko/task-manager/internal/app"
	"github.com/Dima-Vasilenko/task-manager/internal/domain"
	"github.com/Dima-Vasilenko/task-manager/internal/handler"
)

func TestCreateTask(t *testing.T) {
	db := setUpTestDB(t)
	app := app.NewApplication(db)

	route := app.Route()
	task := CreateTestTask(t, route)
	if task.ID <= 0 {
		t.Errorf("Incorrect task ID:  %v", task.ID)
	}

	if task.Title != "First Task" {
		t.Errorf("Title = %v, ожидали %v", task.Title, "First Task")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var storedTask domain.Task
	err := db.QueryRowContext(ctx, "SELECT id, title, description, importance, completed FROM app.tasks WHERE id = $1", task.ID).Scan(
		&storedTask.ID, &storedTask.Title, &storedTask.Description, &storedTask.Importance, &storedTask.Completed)
	if err != nil {
		t.Fatalf("error in getting task from db: %v", err)
	}

	if storedTask != task {
		t.Error("task from db not equal to task from response")
	}
}

func TestGetTaskByID(t *testing.T) {
	db := setUpTestDB(t)
	app := app.NewApplication(db)

	route := app.Route()
	createdTask := CreateTestTask(t, route)

	url := fmt.Sprintf("/tasks/%d", createdTask.ID)

	req := httptest.NewRequest(http.MethodGet, url, nil)
	res := httptest.NewRecorder()
	route.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status %v, required %v", res.Code, http.StatusOK)
	}

	var task domain.Task
	if err := json.NewDecoder(res.Body).Decode(&task); err != nil {
		t.Fatalf("не удалось распарсить ответ: %v", err)
	}
	if task != createdTask {
		t.Errorf("get task not equal to created task")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var storedTask domain.Task
	err := db.QueryRowContext(ctx, "SELECT id, title, description, importance, completed FROM app.tasks WHERE id = $1", createdTask.ID).Scan(
		&storedTask.ID, &storedTask.Title, &storedTask.Description, &storedTask.Importance, &storedTask.Completed)
	if err != nil {
		t.Fatalf("error in getting task from db: %v", err)
	}

	if storedTask != task {
		t.Error("task from db not equal to task from response")
	}
}

func TestUpdateTask(t *testing.T) {
	db := setUpTestDB(t)
	app := app.NewApplication(db)

	route := app.Route()
	createdTask := CreateTestTask(t, route)

	title := "New title"
	importance := 9
	taskReq := handler.UpdateTaskRequest{Title: &title, Importance: &importance}
	body, _ := json.Marshal(taskReq)

	url := fmt.Sprintf("/tasks/%d", createdTask.ID)

	req := httptest.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
	res := httptest.NewRecorder()

	route.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status %v, required %v", res.Code, http.StatusOK)
	}

	var task domain.Task
	if err := json.NewDecoder(res.Body).Decode(&task); err != nil {
		t.Fatalf("не удалось распарсить ответ: %v", err)
	}

	if task.Title != title {
		t.Errorf("title = %v, ожидали %v", task.Title, title)
	}

	if task.Importance != importance {
		t.Errorf("importance = %v, ожидали %v", task.Importance, importance)
	}

	if task.ID != createdTask.ID {
		t.Errorf("incorrect task id: %v, expected %v", task.ID, createdTask.ID)
	}

	if task.Description != "10 byn" {
		t.Errorf("expected: %v, recive: %v", "10 byn", task.Description)
	}

	if task.Completed != false {
		t.Error("completed shouldn't changed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var storedTask domain.Task
	err := db.QueryRowContext(ctx, "SELECT id, title, description, importance, completed FROM app.tasks WHERE id = $1", createdTask.ID).Scan(
		&storedTask.ID, &storedTask.Title, &storedTask.Description, &storedTask.Importance, &storedTask.Completed)
	if err != nil {
		t.Fatalf("error in getting task from db: %v", err)
	}

	if storedTask != task {
		t.Error("task from db not equal to task from response")
	}
}

func TestDeleteTask(t *testing.T) {
	db := setUpTestDB(t)
	app := app.NewApplication(db)

	route := app.Route()
	createdTask := CreateTestTask(t, route)

	url := fmt.Sprintf("/tasks/%d", createdTask.ID)

	req := httptest.NewRequest(http.MethodDelete, url, nil)
	res := httptest.NewRecorder()

	route.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status %v, required %v", res.Code, http.StatusNoContent)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var storedTask domain.Task
	err := db.QueryRowContext(ctx, "SELECT id, title, description, importance, completed FROM app.tasks WHERE id = $1", createdTask.ID).Scan(
		&storedTask.ID, &storedTask.Title, &storedTask.Description, &storedTask.Importance, &storedTask.Completed)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("task was not deleted from bd")
	}
}
