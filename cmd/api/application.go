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
}

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
	Importance  *int    `json:"importance"`
}

type application struct {
	tasks  map[int]domain.Task
	nextID int
	mu     sync.RWMutex
}

func (a *application) addTask(taskReq CreateTaskRequest) (task domain.Task) {
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

func (a *application) updateTask(id int, taskReq UpdateTaskRequest) (domain.Task, bool) {
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

func (a *application) deleteTask(id int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, ok := a.tasks[id]
	if !ok {
		return false
	}

	delete(a.tasks, id)
	return true
}

func (a *application) writeJSON(w http.ResponseWriter, status int, data any) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(js))
	return nil
}

func (a *application) allHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func (a *application) taskHandler(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	tasks := maps.Clone(a.tasks)
	a.mu.RUnlock()

	err := a.writeJSON(w, http.StatusOK, tasks)
	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (a *application) createTaskHandler(w http.ResponseWriter, r *http.Request) {
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

func (a *application) taskByIDHandler(w http.ResponseWriter, r *http.Request) {

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

func (a *application) updateTaskHandler(w http.ResponseWriter, r *http.Request) {

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
		http.Error(w, "Bad request", http.StatusBadRequest)
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

func (a *application) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {

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

func (a *application) healthHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status": "ok",
	}

	err := a.writeJSON(w, http.StatusOK, data)
	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (a *application) route() *http.ServeMux {

	mux := http.NewServeMux()

	mux.HandleFunc("/", a.allHandler)
	mux.HandleFunc("GET /tasks", a.taskHandler)
	mux.HandleFunc("POST /tasks", a.createTaskHandler)
	mux.HandleFunc("GET /tasks/{id}", a.taskByIDHandler)
	mux.HandleFunc("PATCH /tasks/{id}", a.updateTaskHandler)
	mux.HandleFunc("DELETE /tasks/{id}", a.deleteTaskHandler)
	mux.HandleFunc("GET /health", a.healthHandler)

	return mux
}
