package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	main "github.com/Dima-Vasilenko/task-manager/cmd/api"
	"github.com/Dima-Vasilenko/task-manager/internal/domain"
)

func TestGetingAllTasks(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("status %v, required %v", res.Code, http.StatusOK)
	}

	var body map[string]domain.Task
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("не удалось распарсить ответ: %v", err)
	}
	if body["1"].ID != 1 {
		t.Errorf("id = %v, ожидали %v", body["1"].ID, 1)
	}
}

func TestGetTaskByID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	req.SetPathValue("id", "1")
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("status %v, required %v", res.Code, http.StatusOK)
	}

	var body domain.Task
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("не удалось распарсить ответ: %v", err)
	}
	if body.Importance != 9 {
		t.Errorf("id = %v, ожидали %v", body.Importance, 9)
	}
}

func TestGetTaskByInvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks/abc", nil)
	req.SetPathValue("id", "abc")
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("status %v, required %v", res.Code, http.StatusBadRequest)
	}
}

func TestGetTaskByNotFoundID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks/10", nil)
	req.SetPathValue("id", "10")
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Errorf("status %v, required %v", res.Code, http.StatusNotFound)
	}
}

func TestCreateTask(t *testing.T) {
	payload := main.CreateTaskRequest{Title: "Buy milk", Description: "10 byn", Importance: 4}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Errorf("status %v, required %v", res.Code, http.StatusCreated)
	}

	var task domain.Task
	if err := json.NewDecoder(res.Body).Decode(&task); err != nil {
		t.Fatalf("не удалось распарсить ответ: %v", err)
	}
	if task.ID != 6 {
		t.Errorf("id = %v, ожидали %v", task.ID, 6)
	}
}

func TestCreateTaskWithInvalidTitle(t *testing.T) {
	payload := main.CreateTaskRequest{Title: "   ", Description: "10 byn", Importance: 4}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("status %v, required %v", res.Code, http.StatusBadRequest)
	}
}

func TestCreateTaskWithInvalidImportance(t *testing.T) {
	payload := main.CreateTaskRequest{Title: "AAA", Description: "10 byn", Importance: -1}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("status %v, required %v", res.Code, http.StatusBadRequest)
	}
}

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		name          string
		importance    int
		title         string
		completed     bool
		wantCode      int
		wantCompleted bool
	}{
		{
			"valid request",
			4,
			"New title",
			false,
			http.StatusOK,
			false,
		},
		{
			"invalid title",
			4,
			"   ",
			false,
			http.StatusBadRequest,
			false,
		},
		{
			"negative importance",
			-1,
			"New title",
			false,
			http.StatusBadRequest,
			false,
		},
		{
			"overallow importance",
			11,
			"New title",
			false,
			http.StatusBadRequest,
			false,
		},
		{
			"changing completence",
			4,
			"New title",
			true,
			http.StatusOK,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskReq := main.UpdateTaskRequest{Title: &tt.title, Importance: &tt.importance, Completed: &tt.completed}
			body, _ := json.Marshal(taskReq)
			req := httptest.NewRequest(http.MethodPatch, "/tasks/1", bytes.NewBuffer(body))
			req.SetPathValue("id", "1")
			res := httptest.NewRecorder()

			app := main.NewApplication()
			app.Route().ServeHTTP(res, req)

			if res.Code != tt.wantCode {
				t.Errorf("status %v, required %v", res.Code, tt.wantCode)
			}

			if res.Code == http.StatusBadRequest {
				return
			}
			var task domain.Task
			if err := json.NewDecoder(res.Body).Decode(&task); err != nil {
				t.Fatalf("не удалось распарсить ответ: %v", err)
			}
			if task.Completed != tt.wantCompleted {
				t.Errorf("id = %v, ожидали %v", task.Importance, 4)
			}
		})
	}
}

func TestUpdateTaskWithInvalidID(t *testing.T) {
	payload := main.CreateTaskRequest{Title: "New title", Description: "10 byn", Importance: 4}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/tasks/abc", bytes.NewBuffer(body))
	req.SetPathValue("id", "abc")
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("status %v, required %v", res.Code, http.StatusBadRequest)
	}
}

func TestUpdateTaskWithNotFoundID(t *testing.T) {
	payload := main.CreateTaskRequest{Title: "New title", Description: "10 byn", Importance: 4}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/tasks/10", bytes.NewBuffer(body))
	req.SetPathValue("id", "10")
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Errorf("status %v, required %v", res.Code, http.StatusNotFound)
	}
}

func TestUpdateTaskWithInvalidTitle(t *testing.T) {
	payload := main.CreateTaskRequest{Title: "  ", Description: "10 byn", Importance: 4}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/tasks/1", bytes.NewBuffer(body))
	req.SetPathValue("id", "1")
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("status %v, required %v", res.Code, http.StatusBadRequest)
	}
}

func TestUpdateTaskWithEmptyRequest(t *testing.T) {
	payload := main.CreateTaskRequest{}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPatch, "/tasks/1", bytes.NewBuffer(body))
	req.SetPathValue("id", "1")
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("status %v, required %v", res.Code, http.StatusBadRequest)
	}
}

func TestDeleteTask(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	req.SetPathValue("id", "1")
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Errorf("status %v, required %v", res.Code, http.StatusNoContent)
	}

	req = httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	req.SetPathValue("id", "1")
	res = httptest.NewRecorder()

	app.Route().ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Errorf("status %v, required %v", res.Code, http.StatusNotFound)
	}
}

func TestDeleteTaskWithInvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/tasks/abc", nil)
	req.SetPathValue("id", "abc")
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("status %v, required %v", res.Code, http.StatusBadRequest)
	}
}

func TestDeleteTaskWithNotFoundID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/tasks/11", nil)
	req.SetPathValue("id", "11")
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Errorf("status %v, required %v", res.Code, http.StatusNotFound)
	}
}

func TestHealtHandlerWithIncorrectMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	res := httptest.NewRecorder()

	app := main.NewApplication()
	app.Route().ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Errorf("status %v, required %v", res.Code, http.StatusMethodNotAllowed)
	}
}

func TestRaceAdding(t *testing.T) {

	var wg sync.WaitGroup

	app := main.NewApplication()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			payload := main.CreateTaskRequest{Title: "New task", Description: "10 byn", Importance: 4}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
			res := httptest.NewRecorder()

			app.Route().ServeHTTP(res, req)
		}()
	}

	wg.Wait()

	if len(app.GetTasks()) != 105 {
		t.Errorf("amount of tasks: %v, expected: %v", len(app.GetTasks()), 105)
	}
}
