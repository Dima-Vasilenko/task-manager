package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Dima-Vasilenko/task-manager/internal/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func setUpTestDB(t *testing.T) *sql.DB {
	t.Helper()

	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Файл .env не найден, используем системные переменные")
	}

	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		t.Skip("test database is not set")
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("error %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("filed to close database %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("no connection sets %v", err)
	}

	_, errClear := db.ExecContext(ctx, "TRUNCATE TABLE app.tasks RESTART IDENTITY CASCADE")
	if errClear != nil {
		t.Errorf("can't clear table %v", err)
	}

	return db
}

func TestCreateTask(t *testing.T) {
	db := setUpTestDB(t)
	repo := NewTaskRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := domain.CreateTaskParams{Title: "New task", Description: "New description", Importance: 5}

	task, err := repo.CreateTask(ctx, req)
	if err != nil {
		t.Fatalf("error in proccess of creating task: %v", err)
	}

	if task.ID <= 0 {
		t.Error("incorrect id of created task")
	}

	if task.Title != req.Title {
		t.Error("incorrect title")
	}

	if task.Description != req.Description {
		t.Error("incorrect description")
	}

	if task.Importance != req.Importance {
		t.Error("incorrect importance")
	}
}

func TestGetTaskByID(t *testing.T) {
	db := setUpTestDB(t)
	repo := NewTaskRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := domain.CreateTaskParams{Title: "New task", Description: "New description", Importance: 5}

	task, err := repo.CreateTask(ctx, req)
	if err != nil {
		t.Fatalf("error in proccess of creating task: %v", err)
	}

	getTask, err := repo.GetTaskByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("error in proccess of getting task by id: %v", err)
	}

	if getTask != task {
		t.Error("geted task is not equal to created")
	}
}

func TestGetTaskByIncorrectID(t *testing.T) {
	db := setUpTestDB(t)
	repo := NewTaskRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, errGet := repo.GetTaskByID(ctx, 99999)
	if !errors.Is(errGet, ErrTaskNotFound) {
		t.Errorf("task with such id shouldn't be in table")
	}
}

func TestGetAllTasks(t *testing.T) {
	db := setUpTestDB(t)
	repo := NewTaskRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := domain.CreateTaskParams{Title: "New task", Description: "New description", Importance: 5}
	req2 := domain.CreateTaskParams{Title: "New task2", Description: "New description2", Importance: 6}

	task, err := repo.CreateTask(ctx, req)
	if err != nil {
		t.Fatalf("error in proccess of creating task: %v", err)
	}

	task2, err := repo.CreateTask(ctx, req2)
	if err != nil {
		t.Fatalf("error in proccess of creating task: %v", err)
	}

	getTask, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("error in getting all tasks: %v", err)
	}

	if len(getTask) != 2 {
		t.Fatalf("get: %v, expected: %v ", len(getTask), 2)
	}

	if getTask[0] != task || getTask[1] != task2 {
		t.Error("incorrects tasks in list")
	}
}

func TestGetEmptyTasks(t *testing.T) {
	db := setUpTestDB(t)
	repo := NewTaskRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	getTask, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("error in getting all tasks: %v", err)
	}

	if getTask == nil {
		t.Error("expected empry collection, not nil")
	}

	if len(getTask) != 0 {
		t.Errorf("get: %v, expected: %v ", len(getTask), 0)
	}
}

func TestUpdateTask(t *testing.T) {
	db := setUpTestDB(t)
	repo := NewTaskRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	completence := true

	req := domain.CreateTaskParams{Title: "New task", Description: "New description", Importance: 5}
	reqUp := domain.UpdateTaskParams{Completed: &completence}

	task, err := repo.CreateTask(ctx, req)
	if err != nil {
		t.Errorf("failde to create task: %v", err)
	}

	upgradedTask, err := repo.UpdateTaskByID(ctx, task.ID, reqUp)
	if err != nil {
		t.Fatalf("error in upgreding task: %v", err)
	}

	if upgradedTask.Completed != true {
		t.Errorf("get: %v, expect: %v", upgradedTask.Completed, completence)
	}

	if task.Title != upgradedTask.Title {
		t.Errorf("Title shouldn't change")
	}

	if task.Description != upgradedTask.Description {
		t.Errorf("Description shouldn't change")
	}

	if task.Importance != upgradedTask.Importance {
		t.Errorf("Importance shouldn't change")
	}

	if task.ID != upgradedTask.ID {
		t.Errorf("ID shouldn't change")
	}
}

func TestUpdateCompletedFalse(t *testing.T) {
	db := setUpTestDB(t)
	repo := NewTaskRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	completence1 := true
	completence2 := false

	req := domain.CreateTaskParams{Title: "New task", Description: "New description", Importance: 5}
	reqUp := domain.UpdateTaskParams{Completed: &completence1}
	reqUp2 := domain.UpdateTaskParams{Completed: &completence2}

	task, err := repo.CreateTask(ctx, req)
	if err != nil {
		t.Errorf("failde to create task: %v", err)
	}

	upgradedTask, err := repo.UpdateTaskByID(ctx, task.ID, reqUp)
	if err != nil {
		t.Fatalf("error in upgreding task: %v", err)
	}

	if upgradedTask.Completed != true {
		t.Errorf("get: %v, expect: %v", upgradedTask.Completed, completence1)
	}

	upgradedTask2, err := repo.UpdateTaskByID(ctx, task.ID, reqUp2)
	if err != nil {
		t.Fatalf("error in upgreding task: %v", err)
	}

	if upgradedTask2.Completed != false {
		t.Errorf("get: %v, expect: %v", upgradedTask2.Completed, completence2)
	}
}

func TestUpdateMultipleFields(t *testing.T) {
	db := setUpTestDB(t)
	repo := NewTaskRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	title := "Changed title"
	description := "Changed description"
	importance := 10
	completence := true

	req := domain.CreateTaskParams{Title: "New task", Description: "New description", Importance: 5}
	reqUp := domain.UpdateTaskParams{Title: &title, Description: &description, Importance: &importance, Completed: &completence}

	task, err := repo.CreateTask(ctx, req)
	if err != nil {
		t.Errorf("failde to create task: %v", err)
	}

	upgradedTask, err := repo.UpdateTaskByID(ctx, task.ID, reqUp)
	if err != nil {
		t.Fatalf("error in upgreding task: %v", err)
	}

	if upgradedTask.Completed != true {
		t.Errorf("get: %v, expect: %v", upgradedTask.Completed, completence)
	}

	if upgradedTask.Title != title {
		t.Errorf("Title should change")
	}

	if upgradedTask.Description != description {
		t.Errorf("Description should change")
	}

	if upgradedTask.Importance != importance {
		t.Errorf("Importance should change")
	}

	if task.ID != upgradedTask.ID {
		t.Errorf("ID shouldn't change")
	}
}

func TestUpdateInvalidIDTask(t *testing.T) {
	db := setUpTestDB(t)
	repo := NewTaskRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	completence := true

	reqUp := domain.UpdateTaskParams{Completed: &completence}

	_, err := repo.UpdateTaskByID(ctx, 9999, reqUp)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("task with such id shouldn't be in table")
	}
}

func TestDeleteTask(t *testing.T) {
	db := setUpTestDB(t)
	repo := NewTaskRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := domain.CreateTaskParams{Title: "New task", Description: "New description", Importance: 5}

	task, err := repo.CreateTask(ctx, req)
	if err != nil {
		t.Errorf("failde to create task: %v", err)
	}

	errDel := repo.DeleteTaskByID(ctx, task.ID)
	if errors.Is(errDel, ErrNothingToDelete) {
		t.Errorf("err in deliting: %v", errDel)
	}

	if errDel != nil {
		t.Fatalf("error in delliting task: %v", errDel)
	}

	err2 := repo.DeleteTaskByID(ctx, task.ID)
	if !errors.Is(err2, ErrNothingToDelete) {
		t.Errorf("expected such error: %v,  with deliting", err2)
	}
}
