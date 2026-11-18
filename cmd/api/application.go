package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Dima-Vasilenko/task-manager/internal/domain"
	"github.com/Dima-Vasilenko/task-manager/internal/repository"
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

type Application struct {
	tasks *repository.TaskRepository
}

func NewApplication(db *sql.DB) *Application {
	return &Application{
		tasks: repository.NewTaskRepository(db),
	}
}

func (a *Application) addTask(ctx context.Context, taskReq CreateTaskRequest) (domain.Task, error) {
	task, err := a.tasks.CreateTask(ctx, domain.CreateTaskParams{
		Title:       taskReq.Title,
		Description: taskReq.Description,
		Importance:  taskReq.Importance,
	})
	if err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

func (a *Application) updateTask(ctx context.Context, id int64, taskReq UpdateTaskRequest) (domain.Task, error) {
	task, err := a.tasks.UpdateTaskByID(ctx, id, domain.UpdateTaskParams{
		Title:       taskReq.Title,
		Description: taskReq.Description,
		Importance:  taskReq.Importance,
		Completed:   taskReq.Completed,
	})
	if err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

func (a *Application) deleteTask(ctx context.Context, id int64) (bool, error) {
	ok, err := a.tasks.DeleteTaskByID(ctx, id)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (a *Application) writeJSON(w http.ResponseWriter, status int, data any) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, errWrite := w.Write([]byte(js))
	return errWrite
}

func (a *Application) AllHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func (a *Application) TaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	tasks, err := a.tasks.GetAll(ctx)
	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
	errWrite := a.writeJSON(w, http.StatusOK, tasks)
	if errWrite != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (a *Application) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var taskReq CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&taskReq); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(taskReq.Title) == "" || taskReq.Importance < 1 || taskReq.Importance > 10 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	task, errCreate := a.addTask(ctx, taskReq)
	if errCreate != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	err := a.writeJSON(w, http.StatusCreated, task)
	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (a *Application) TaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	task, errTake := a.tasks.GetTaskByID(ctx, id)
	if errors.Is(errTake, repository.ErrTaskNotFound) {
		http.Error(w, "id not found", http.StatusNotFound)
		return
	}

	if errTake != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	errW := a.writeJSON(w, http.StatusOK, task)
	if errW != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (a *Application) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
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

	task, errUpdate := a.updateTask(ctx, id, taskReq)
	if errors.Is(errUpdate, repository.ErrTaskNotFound) {
		http.Error(w, "id not found", http.StatusNotFound)
		return
	}

	if errUpdate != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	errW := a.writeJSON(w, http.StatusOK, task)
	if errW != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (a *Application) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	ok, err := a.deleteTask(ctx, id)
	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	if !ok {
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
