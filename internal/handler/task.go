package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Dima-Vasilenko/task-manager/internal/domain"
	"github.com/Dima-Vasilenko/task-manager/internal/service"
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

type TaskService interface {
	GetAll(ctx context.Context) ([]domain.Task, error)
	GetByID(ctx context.Context, id int64) (domain.Task, error)
	Create(ctx context.Context, req domain.CreateTaskParams) (domain.Task, error)
	Update(ctx context.Context, id int64, req domain.UpdateTaskParams) (domain.Task, error)
	Delete(ctx context.Context, id int64) error
}

type TaskHandler struct {
	service TaskService
}

func NewTaskHandler(serv TaskService) *TaskHandler {
	return &TaskHandler{
		service: serv,
	}
}

func (h *TaskHandler) writeJSON(w http.ResponseWriter, status int, data any) error {
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

func (h *TaskHandler) AllHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func (h *TaskHandler) TaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tasks, err := h.service.GetAll(ctx)
	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	errWrite := h.writeJSON(w, http.StatusOK, tasks)
	if errWrite != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) TaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	task, errTake := h.service.GetByID(ctx, id)
	if errors.Is(errTake, service.ErrIncorrectId) {
		http.Error(w, "id not found", http.StatusBadRequest)
		return
	}

	if errors.Is(errTake, service.ErrTaskNotFound) {
		http.Error(w, "id not found", http.StatusNotFound)
		return
	}

	if errTake != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	errW := h.writeJSON(w, http.StatusOK, task)
	if errW != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var taskReq CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&taskReq); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	createReq := domain.CreateTaskParams{
		Title:       taskReq.Title,
		Description: taskReq.Description,
		Importance:  taskReq.Importance,
	}

	task, errCreate := h.service.Create(ctx, createReq)
	if errors.Is(errCreate, service.ErrIncorrectTask) {
		http.Error(w, "incorrect request", http.StatusBadRequest)
		return
	}

	if errCreate != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	err := h.writeJSON(w, http.StatusCreated, task)
	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	var taskReq UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&taskReq); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	updateReq := domain.UpdateTaskParams{
		Title:       taskReq.Title,
		Description: taskReq.Description,
		Importance:  taskReq.Importance,
		Completed:   taskReq.Completed,
	}

	task, errUpdate := h.service.Update(ctx, id, updateReq)
	if errors.Is(errUpdate, service.ErrIncorrectTask) {
		http.Error(w, "incorrect request", http.StatusBadRequest)
		return
	}

	if errors.Is(errUpdate, service.ErrIncorrectId) {
		http.Error(w, "incorrect request", http.StatusBadRequest)
		return
	}

	if errors.Is(errUpdate, service.ErrTaskNotFound) {
		http.Error(w, "id not found", http.StatusNotFound)
		return
	}

	if errUpdate != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	errW := h.writeJSON(w, http.StatusOK, task)
	if errW != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	errDel := h.service.Delete(ctx, id)
	if errors.Is(errDel, service.ErrIncorrectId) {
		http.Error(w, "id not found", http.StatusBadRequest)
		return
	}

	if errors.Is(errDel, service.ErrTaskNotFound) {
		http.Error(w, "id not found", http.StatusNotFound)
		return
	}

	if errDel != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status": "ok",
	}

	err := h.writeJSON(w, http.StatusOK, data)
	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
}
