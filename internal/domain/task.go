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
