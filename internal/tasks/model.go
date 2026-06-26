package tasks

import "time"

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

const (
	ActionCreated HistoryAction = "created"
	ActionUpdated HistoryAction = "updated"
)

type Task struct {
	ID          int64
	TeamID      int64
	Title       string
	Description string
	Status      TaskStatus
	AssigneeID  *int64
	CreatedBy   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DoneAt      *time.Time
}

type TaskStatus string

type TaskHistory struct {
	ID        int64
	TaskID    int64
	ChangedBy int64
	Action    HistoryAction
	FieldName *string
	OldValue  *string
	NewValue  *string
	CreatedAt time.Time
}

type HistoryAction string

type UpdateTask struct {
	Title         *string
	Description   *string
	Status        *TaskStatus
	AssigneeID    *int64
	ClearAssignee bool
}

type TaskFilter struct {
	TeamID     int64
	Status     *TaskStatus
	AssigneeID *int64
	Limit      int
	Offset     int
}