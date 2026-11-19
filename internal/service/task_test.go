package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Dima-Vasilenko/task-manager/internal/domain"
)

type FakeTaskRepository struct {
	tasks []domain.Task

	getAllCalled  bool
	getByIDCalled bool
	createCalled  bool
	updateCalled  bool
	deleteCalled  bool

	err error
}

func (f *FakeTaskRepository) GetAll(ctx context.Context) ([]domain.Task, error) {
	f.getAllCalled = true
	return f.tasks, f.err
}

func (f *FakeTaskRepository) GetTaskByID(ctx context.Context, id int64) (domain.Task, error) {
	f.getByIDCalled = true
	return domain.Task{}, nil
}

func (f *FakeTaskRepository) CreateTask(ctx context.Context, req domain.CreateTaskParams) (domain.Task, error) {
	f.createCalled = true
	return domain.Task{}, f.err
}

func (f *FakeTaskRepository) UpdateTaskByID(ctx context.Context, id int64, req domain.UpdateTaskParams) (domain.Task, error) {
	f.updateCalled = true
	return domain.Task{}, f.err
}

func (f *FakeTaskRepository) DeleteTaskByID(ctx context.Context, id int64) error {
	f.deleteCalled = true
	return f.err
}

func TestTaskService_GetAll(t *testing.T) {
	ctx := context.Background()

	rep := &FakeTaskRepository{
		tasks: []domain.Task{
			{
				ID:          1,
				Title:       "new",
				Description: "new",
				Importance:  4,
			},
		},
	}
	serv := NewTaskService(rep)

	tasks, err := serv.GetAll(ctx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("expected len: %v, get: %v", 1, len(tasks))
	}

	if tasks[0].ID != 1 {
		t.Errorf("expected ID: %v, get: %v", 1, tasks[0].ID)
	}
}

func TestTaskService_Create(t *testing.T) {
	tests := []struct {
		name    string
		params  domain.CreateTaskParams
		wantErr bool
	}{
		{
			name: "valid task",
			params: domain.CreateTaskParams{
				Title:       "New title",
				Description: "New description",
				Importance:  5,
			},
			wantErr: false,
		},
		{
			name: "invalid title",
			params: domain.CreateTaskParams{
				Title:       "  ",
				Description: "New description",
				Importance:  5,
			},
			wantErr: true,
		},
		{
			name: "iverfllow importance",
			params: domain.CreateTaskParams{
				Title:       "New title",
				Description: "New description",
				Importance:  12,
			},
			wantErr: true,
		},
		{
			name: "negative importance",
			params: domain.CreateTaskParams{
				Title:       "New title",
				Description: "New description",
				Importance:  -1,
			},
			wantErr: true,
		},
		{
			name: "zero importance",
			params: domain.CreateTaskParams{
				Title:       "New title",
				Description: "New description",
				Importance:  0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			rep := &FakeTaskRepository{}
			serv := NewTaskService(rep)

			_, err := serv.Create(ctx, tt.params)

			if tt.wantErr {
				if !errors.Is(err, ErrIncorrectTask) {
					t.Errorf("expected err: %v, but not get: %v", ErrIncorrectTask, err)
				}
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTaskService_Update(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		importance  int
		completed   bool
		wantErr     bool
	}{
		{
			name:        "valid task",
			title:       "New title",
			description: "new description",
			importance:  3,
			completed:   true,
			wantErr:     false,
		},
		{
			name:        "empty title",
			title:       "   ",
			description: "new description",
			importance:  3,
			completed:   true,
			wantErr:     true,
		},
		{
			name:        "overflow importance",
			title:       "New title",
			description: "new description",
			importance:  11,
			completed:   true,
			wantErr:     true,
		},
		{
			name:        "negative importance",
			title:       "New title",
			description: "new description",
			importance:  -1,
			completed:   true,
			wantErr:     true,
		},
		{
			name:        "zero importance",
			title:       "New title",
			description: "new description",
			importance:  0,
			completed:   true,
			wantErr:     true,
		},
		{
			name:    "empty request",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			rep := &FakeTaskRepository{}
			serv := NewTaskService(rep)
			params := domain.UpdateTaskParams{Title: &tt.title, Description: &tt.description, Importance: &tt.importance, Completed: &tt.completed}

			_, err := serv.Update(ctx, 1, params)

			if tt.wantErr {
				if !errors.Is(err, ErrIncorrectTask) {
					t.Errorf("expected err: %v, but not get: %v", ErrIncorrectTask, err)
				}
			}

			if tt.wantErr && rep.updateCalled {
				t.Errorf("shouldn't be called repo")
			}

			if !tt.wantErr && !rep.updateCalled {
				t.Errorf("expected: %v, get %v", true, rep.updateCalled)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTaskService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "valid task",
			id:      1,
			wantErr: false,
		},
		{
			name:    "zero id",
			id:      0,
			wantErr: true,
		},
		{
			name:    "negative id",
			id:      -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			rep := &FakeTaskRepository{}
			serv := NewTaskService(rep)

			_, err := serv.GetByID(ctx, tt.id)

			if tt.wantErr {
				if !errors.Is(err, ErrIncorrectId) {
					t.Errorf("expected err: %v, but not get: %v", ErrIncorrectId, err)
				}
			}

			if tt.wantErr && rep.getByIDCalled {
				t.Errorf("shouldn't be called repo")
			}

			if !tt.wantErr && !rep.getByIDCalled {
				t.Errorf("expected: %v, get %v", true, rep.getByIDCalled)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTaskService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "valid task",
			id:      1,
			wantErr: false,
		},
		{
			name:    "zero id",
			id:      0,
			wantErr: true,
		},
		{
			name:    "negative id",
			id:      -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			rep := &FakeTaskRepository{}
			serv := NewTaskService(rep)

			err := serv.Delete(ctx, tt.id)

			if tt.wantErr {
				if !errors.Is(err, ErrIncorrectId) {
					t.Errorf("expected err: %v, but not get: %v", ErrIncorrectId, err)
				}
			}

			if tt.wantErr && rep.deleteCalled {
				t.Errorf("shouldn't be called repo")
			}

			if !tt.wantErr && !rep.deleteCalled {
				t.Errorf("expected: %v, get %v", true, rep.deleteCalled)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
