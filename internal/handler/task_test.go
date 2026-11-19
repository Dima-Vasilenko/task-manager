package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/Dima-Vasilenko/task-manager/internal/domain"
)

type FakeTaskService struct {
	tasks []domain.Task
	task  domain.Task
	err   error

	getAllCalled  bool
	getByIDCalled bool
	createCalled  bool
	updateCalled  bool
	deleteCalled  bool
}

func (f *FakeTaskService) GetAll(ctx context.Context) ([]domain.Task, error) {
	f.getAllCalled = true
	return f.tasks, f.err
}

func (f *FakeTaskService) GetByID(ctx context.Context, id int64) (domain.Task, error) {
	f.getByIDCalled = true
	index := slices.IndexFunc(f.tasks, func(t domain.Task) bool {
		return t.ID == id
	})
	if index == -1 {
		return domain.Task{}, fmt.Errorf("not such id %v in rep", id)
	}
	return f.tasks[index], f.err
}

func (f *FakeTaskService) Create(ctx context.Context, req domain.CreateTaskParams) (domain.Task, error) {
	task := domain.Task{
		ID:          int64(len(f.tasks)) + 1,
		Title:       req.Title,
		Description: req.Description,
		Importance:  req.Importance,
	}
	f.tasks = append(f.tasks, task)
	f.createCalled = true
	return f.tasks[len(f.tasks)-1], f.err
}

func (f *FakeTaskService) Update(ctx context.Context, id int64, req domain.UpdateTaskParams) (domain.Task, error) {
	f.updateCalled = true
	index := slices.IndexFunc(f.tasks, func(t domain.Task) bool {
		return t.ID == id
	})
	if index == -1 {
		return domain.Task{}, fmt.Errorf("not such id %v in rep", id)
	}
	task := &f.tasks[index]
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Importance != nil {
		task.Importance = *req.Importance
	}
	if req.Completed != nil {
		task.Completed = *req.Completed
	}
	return *task, f.err
}

func (f *FakeTaskService) Delete(ctx context.Context, id int64) error {
	f.deleteCalled = true
	index := slices.IndexFunc(f.tasks, func(t domain.Task) bool {
		return t.ID == id
	})
	if index == -1 {
		return fmt.Errorf("not such id %v in rep", id)
	}
	f.tasks = append(f.tasks[:index], f.tasks[index+1:]...)
	return f.err
}

func CreateService() *FakeTaskService {
	serv := FakeTaskService{
		tasks: func() []domain.Task {
			var i int64 = 1
			var res []domain.Task
			for ; i <= 10; i++ {
				res = append(res, domain.Task{
					ID:          i,
					Title:       fmt.Sprintf("Task %v", i),
					Description: fmt.Sprintf("Description %v", i),
					Importance:  4,
				})
			}
			return res
		}(),
	}

	serv.task = serv.tasks[0]
	return &serv
}

func TestTaskHandler_GetAll(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	res := httptest.NewRecorder()

	serv := CreateService()
	handl := NewTaskHandler(serv)

	handl.TaskHandler(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("status %v, required %v", res.Code, http.StatusOK)
	}

	var body []domain.Task
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("не удалось распарсить ответ: %v", err)
	}

	if len(body) != 10 {
		t.Errorf("len = %v, ожидали %v", len(body), 10)
	}

	if body[0].ID != 1 {
		t.Errorf("id = %v, ожидали %v", body[0].ID, 1)
	}

	if !serv.getAllCalled {
		t.Errorf("service wasn bee called")
	}
}

func TestTaskHandler_GetByID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/tasks/6", nil)
	req.SetPathValue("id", "6")
	res := httptest.NewRecorder()

	serv := CreateService()
	handl := NewTaskHandler(serv)

	handl.TaskByIDHandler(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("status %v, required %v", res.Code, http.StatusOK)
	}

	fmt.Println(res.Body)

	var body domain.Task
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("не удалось распарсить ответ: %v", err)
	}
	if body.ID != 6 {
		t.Errorf("id = %v, ожидали %v", body.ID, 6)
	}

	if !serv.getByIDCalled {
		t.Errorf("service wasn bee called")
	}
}

func TestTaskHandler_Create(t *testing.T) {
	payload := CreateTaskRequest{Title: "Buy milk", Description: "10 byn", Importance: 4}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
	res := httptest.NewRecorder()

	serv := CreateService()
	handl := NewTaskHandler(serv)

	handl.CreateTaskHandler(res, req)

	if res.Code != http.StatusCreated {
		t.Errorf("status %v, required %v", res.Code, http.StatusCreated)
	}

	var task domain.Task
	if err := json.NewDecoder(res.Body).Decode(&task); err != nil {
		t.Fatalf("не удалось распарсить ответ: %v", err)
	}
	if task.ID != 11 {
		t.Errorf("Title = %v, ожидали %v", task.ID, 11)
	}

	if task.Title != "Buy milk" {
		t.Errorf("Title = %v, ожидали %v", task.Title, "Buy milk")
	}

	if !serv.createCalled {
		t.Errorf("service wasn bee called")
	}
}

func TestTaskHandler_CreateInvalidJson(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer([]byte("lojddksjgojsd;lfgjl;sjdg")))
	res := httptest.NewRecorder()

	serv := CreateService()
	handl := NewTaskHandler(serv)

	handl.CreateTaskHandler(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("status %v, required %v", res.Code, http.StatusBadRequest)
	}
}

func TestTaskHandler_Update(t *testing.T) {
	title := "New title"
	importance := 9
	taskReq := UpdateTaskRequest{Title: &title, Importance: &importance}
	body, _ := json.Marshal(taskReq)
	req := httptest.NewRequest(http.MethodPatch, "/tasks/1", bytes.NewBuffer(body))
	req.SetPathValue("id", "1")
	res := httptest.NewRecorder()

	serv := CreateService()
	handl := NewTaskHandler(serv)

	handl.UpdateTaskHandler(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("status %v, required %v", res.Code, http.StatusOK)
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

	if task.ID != 1 {
		t.Errorf("id = %v, ожидали %v", task.ID, 1)
	}

	if !serv.updateCalled {
		t.Errorf("service wasn bee called")
	}
}

func TestTaskHandler_UpdateInvalidID(t *testing.T) {
	title := "New title"
	importance := 9
	taskReq := UpdateTaskRequest{Title: &title, Importance: &importance}
	body, _ := json.Marshal(taskReq)
	req := httptest.NewRequest(http.MethodPatch, "/tasks/1", bytes.NewBuffer(body))
	req.SetPathValue("id", "abc")
	res := httptest.NewRecorder()

	serv := CreateService()
	handl := NewTaskHandler(serv)

	handl.UpdateTaskHandler(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("status %v, required %v", res.Code, http.StatusBadRequest)
	}
}

func TestTaskHandler_Delete(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	req.SetPathValue("id", "1")
	res := httptest.NewRecorder()

	serv := CreateService()
	handl := NewTaskHandler(serv)

	handl.DeleteTaskHandler(res, req)

	if res.Code != http.StatusNoContent {
		t.Errorf("status %v, required %v", res.Code, http.StatusNoContent)
	}

	if !serv.deleteCalled {
		t.Errorf("service wasn bee called")
	}
}
