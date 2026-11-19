package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Dima-Vasilenko/task-manager/internal/domain"
	"github.com/Dima-Vasilenko/task-manager/internal/repository"
)

var (
	ErrIncorrectTask = errors.New("incorrect task args")
	ErrIncorrectId   = errors.New("incorrect id")
	ErrTaskNotFound  = errors.New("task not found")
)

type TaskRepository interface {
	GetAll(ctx context.Context) ([]domain.Task, error)
	CreateTask(ctx context.Context, req domain.CreateTaskParams) (domain.Task, error)
	GetTaskByID(ctx context.Context, id int64) (domain.Task, error)
	UpdateTaskByID(ctx context.Context, id int64, req domain.UpdateTaskParams) (domain.Task, error)
	DeleteTaskByID(ctx context.Context, id int64) error
}

type TaskService struct {
	repo TaskRepository
}

func NewTaskService(rep TaskRepository) *TaskService {
	return &TaskService{
		repo: rep,
	}
}

func (s *TaskService) GetAll(ctx context.Context) ([]domain.Task, error) {
	return s.repo.GetAll(ctx)
}

func (s *TaskService) GetByID(ctx context.Context, id int64) (domain.Task, error) {
	if id <= 0 {
		return domain.Task{}, ErrIncorrectId
	}

	task, err := s.repo.GetTaskByID(ctx, id)
	if errors.Is(err, repository.ErrTaskNotFound) {
		return domain.Task{}, ErrTaskNotFound
	}

	return task, err
}

func (s *TaskService) Create(ctx context.Context, req domain.CreateTaskParams) (domain.Task, error) {
	if strings.TrimSpace(req.Title) == "" {
		return domain.Task{}, ErrIncorrectTask
	}

	if req.Importance < 1 || req.Importance > 10 {
		return domain.Task{}, ErrIncorrectTask
	}

	return s.repo.CreateTask(ctx, req)
}

func (s *TaskService) Update(ctx context.Context, id int64, req domain.UpdateTaskParams) (domain.Task, error) {
	if id <= 0 {
		return domain.Task{}, ErrIncorrectId
	}

	if req.Title == nil && req.Description == nil && req.Importance == nil && req.Completed == nil {
		return domain.Task{}, fmt.Errorf("no field to update %w", ErrIncorrectTask)
	}

	if req.Title != nil && strings.TrimSpace(*req.Title) == "" {
		return domain.Task{}, fmt.Errorf("incorrect title %w", ErrIncorrectTask)
	}

	if req.Importance != nil && (*req.Importance < 1 || *req.Importance > 10) {
		return domain.Task{}, fmt.Errorf("incorrect importance %w", ErrIncorrectTask)
	}

	task, err := s.repo.UpdateTaskByID(ctx, id, req)
	if errors.Is(err, repository.ErrTaskNotFound) {
		return domain.Task{}, ErrTaskNotFound
	}

	return task, err
}

func (s *TaskService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrIncorrectId
	}

	err := s.repo.DeleteTaskByID(ctx, id)
	if errors.Is(err, repository.ErrNothingToDelete) {
		return ErrTaskNotFound
	}

	return err
}
