package domain

type Task struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	Importance  int    `json:"importance"`
}

type CreateTaskParams struct {
	Title       string
	Description string
	Importance  int
}

type UpdateTaskParams struct {
	Title       *string
	Description *string
	Completed   *bool
	Importance  *int
}

// func CreateTasks() map[int]Task {
// 	tasks := make(map[int]Task, 10)
// 	var i int64
// 	for i = 1; i <= 5; i++ {
// 		tasks[i] = Task{
// 			ID:          i,
// 			Title:       "task" + strconv.Itoa(i),
// 			Description: "about " + strconv.Itoa(i) + " task",
// 			Completed:   false,
// 			Importance:  10 - i,
// 		}
// 	}
// 	return tasks
// }
