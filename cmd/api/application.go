package main

import (
	"encoding/json"
	"maps"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Dima-Vasilenko/task-manager/internal/domain"
)

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Importance  int    `json:"importance"`
	Completed   bool   `json:"completed"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
	Importance  *int    `json:"importance"`
}

type Application struct {
	tasks  map[int]domain.Task
	nextID int
	mu     sync.RWMutex
}

func (a *Application) GetTasks() map[int]domain.Task {
	return a.tasks
}

func NewApplication() *Application {
	t := domain.CreateTasks()
	return &Application{
		tasks:  t,
		nextID: len(t) + 1,
	}
}

func (a *Application) addTask(taskReq CreateTaskRequest) (task domain.Task) {
	a.mu.Lock()
	defer a.mu.Unlock()

	task.Title = taskReq.Title
	task.Description = taskReq.Description
	task.Importance = taskReq.Importance
	task.ID = a.nextID
	a.nextID++
	a.tasks[task.ID] = task
	return
}

func (a *Application) updateTask(id int, taskReq UpdateTaskRequest) (domain.Task, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	task, ok := a.tasks[id]
	if !ok {
		return domain.Task{}, false
	}

	if taskReq.Title != nil {
		task.Title = *taskReq.Title
	}
	if taskReq.Description != nil {
		task.Description = *taskReq.Description
	}
	if taskReq.Completed != nil {
		task.Completed = *taskReq.Completed
	}
	if taskReq.Importance != nil {
		task.Importance = *taskReq.Importance
	}

	a.tasks[task.ID] = task
	return task, true
}

func (a *Application) deleteTask(id int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, ok := a.tasks[id]
	if !ok {
		return false
	}

	delete(a.tasks, id)
	return true
}

func (a *Application) writeJSON(w http.ResponseWriter, status int, data any) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')
	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(status)
	w.Write([]byte(js))
	return nil
}

func (a *Application) AllHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func (a *Application) TaskHandler(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	tasks := maps.Clone(a.tasks)
	a.mu.RUnlock()

	err := a.writeJSON(w, http.StatusOK, tasks)
	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (a *Application) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var taskReq CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&taskReq); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(taskReq.Title) == "" || taskReq.Importance < 1 || taskReq.Importance > 10 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	task := a.addTask(taskReq)

	err := a.writeJSON(w, http.StatusCreated, task)
	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (a *Application) TaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	a.mu.RLock()
	task, ok := a.tasks[id]
	a.mu.RUnlock()
	if !ok {
		http.Error(w, "id not found", http.StatusNotFound)
		return
	}

	errW := a.writeJSON(w, http.StatusOK, task)
	if errW != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (a *Application) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	var taskReq UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&taskReq); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if taskReq.Title == nil && taskReq.Description == nil && taskReq.Importance == nil && taskReq.Completed == nil {
		http.Error(w, "Bad request ", http.StatusBadRequest)
		return
	}

	if taskReq.Title != nil && strings.TrimSpace(*taskReq.Title) == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if taskReq.Importance != nil && (*taskReq.Importance < 1 || *taskReq.Importance > 10) {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	task, ok := a.updateTask(id, taskReq)
	if !ok {
		http.Error(w, "id not found", http.StatusNotFound)
		return
	}

	errW := a.writeJSON(w, http.StatusOK, task)
	if errW != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (a *Application) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	if ok := a.deleteTask(id); !ok {
		http.Error(w, "id not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *Application) HealthHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status": "ok",
	}

	err := a.writeJSON(w, http.StatusOK, data)
	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (a *Application) Route() *http.ServeMux {

	mux := http.NewServeMux()

	mux.HandleFunc("/", a.AllHandler)
	mux.HandleFunc("GET /tasks", a.TaskHandler)
	mux.HandleFunc("POST /tasks", a.CreateTaskHandler)
	mux.HandleFunc("GET /tasks/{id}", a.TaskByIDHandler)
	mux.HandleFunc("PATCH /tasks/{id}", a.UpdateTaskHandler)
	mux.HandleFunc("DELETE /tasks/{id}", a.DeleteTaskHandler)
	mux.HandleFunc("GET /health", a.HealthHandler)

	return mux
}
