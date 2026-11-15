package domain

import (
	"strconv"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	Importance  int    `json:"importance"`
}

func CreateTasks() map[int]Task {
	tasks := make(map[int]Task, 10)
	for i := 1; i <= 5; i++ {
		tasks[i] = Task{
			ID:          i,
			Title:       "task" + strconv.Itoa(i),
			Description: "about " + strconv.Itoa(i) + " task",
			Completed:   false,
			Importance:  10 - i,
		}
	}
	return tasks
}
