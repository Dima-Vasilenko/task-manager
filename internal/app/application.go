package app

import (
	"database/sql"
	"net/http"

	"github.com/Dima-Vasilenko/task-manager/internal/handler"
	"github.com/Dima-Vasilenko/task-manager/internal/repository"
	"github.com/Dima-Vasilenko/task-manager/internal/service"
)

type Application struct {
	handler *handler.TaskHandler
}

func NewApplication(db *sql.DB) *Application {

	repo := repository.NewTaskRepository(db)
	taskService := service.NewTaskService(repo)

	return &Application{
		handler: handler.NewTaskHandler(taskService),
	}
}

func (a *Application) Route() *http.ServeMux {

	mux := http.NewServeMux()

	mux.HandleFunc("/", a.handler.AllHandler)
	mux.HandleFunc("GET /tasks", a.handler.TaskHandler)
	mux.HandleFunc("POST /tasks", a.handler.CreateTaskHandler)
	mux.HandleFunc("GET /tasks/{id}", a.handler.TaskByIDHandler)
	mux.HandleFunc("PATCH /tasks/{id}", a.handler.UpdateTaskHandler)
	mux.HandleFunc("DELETE /tasks/{id}", a.handler.DeleteTaskHandler)
	mux.HandleFunc("GET /health", a.handler.HealthHandler)

	return mux
}
