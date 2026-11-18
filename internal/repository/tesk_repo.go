package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Dima-Vasilenko/task-manager/internal/domain"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (t *TaskRepository) GetAll(ctx context.Context) ([]domain.Task, error) {
	res := make([]domain.Task, 0)
	query := "SELECT id, title, description, importance, completed FROM app.tasks ORDER BY id"
	rows, err := t.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		task := domain.Task{}
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Importance, &task.Completed); err != nil {
			return nil, err
		}
		res = append(res, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (t *TaskRepository) CreateTask(ctx context.Context, cr domain.CreateTaskParams) (domain.Task, error) {
	task := domain.Task{}
	query := "INSERT INTO app.tasks (title, description, importance) VALUES ($1, $2, $3) RETURNING id, title, description, importance, completed"
	res := t.db.QueryRowContext(ctx, query, cr.Title, cr.Description, cr.Importance)
	if err := res.Scan(&task.ID, &task.Title, &task.Description, &task.Importance, &task.Completed); err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

func (t *TaskRepository) GetTaskByID(ctx context.Context, id int64) (domain.Task, error) {
	task := domain.Task{}
	query := "SELECT id, title, description, importance, completed FROM app.tasks WHERE id = $1"
	res := t.db.QueryRowContext(ctx, query, id)
	err := res.Scan(&task.ID, &task.Title, &task.Description, &task.Importance, &task.Completed)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, ErrTaskNotFound
	}
	if err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

func (t *TaskRepository) UpdateTaskByID(ctx context.Context, id int64, upTask domain.UpdateTaskParams) (domain.Task, error) {

	var task domain.Task
	query := "UPDATE app.tasks SET title = COALESCE($1, title), description = COALESCE($2, description), importance = COALESCE($3, importance), completed = COALESCE($4, completed), updated_at = now() WHERE id = $5 RETURNING id, title, description, importance, completed"
	res := t.db.QueryRowContext(ctx, query, upTask.Title, upTask.Description, upTask.Importance, upTask.Completed, id)
	err := res.Scan(&task.ID, &task.Title, &task.Description, &task.Importance, &task.Completed)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, ErrTaskNotFound
	}
	if err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

func (t *TaskRepository) DeleteTaskByID(ctx context.Context, id int64) (bool, error) {
	query := "DELETE FROM app.tasks WHERE id = $1"
	res, err := t.db.ExecContext(ctx, query, id)
	if err != nil {
		return false, err
	}

	amount, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return amount > 0, nil
}
