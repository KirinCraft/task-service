package tasks

type CreateTaskRequest struct {
	TeamID      int64  `json:"team_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AssigneeID  *int64 `json:"assignee_id"`
}

type ListTasksRequest struct {
	TeamID     int64
	Status     string
	AssigneeID *int64
	Limit      int
	Offset     int
}

type ListTasksResponse struct {
	Items []TaskResponse `json:"items"`
}

type TaskResponse struct {
	ID          int64  `json:"id"`
	TeamID      int64  `json:"team_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	AssigneeID  *int64 `json:"assignee_id"`
	CreatedBy   int64  `json:"created_by"`
}

type UpdateTaskRequest struct {
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	Status        *string `json:"status"`
	AssigneeID    *int64  `json:"assignee_id"`
	ClearAssignee bool    `json:"clear_assignee"`
}

type TaskHistoryResponse struct {
	ID        int64   `json:"id"`
	TaskID    int64   `json:"task_id"`
	ChangedBy int64   `json:"changed_by"`
	Action    string  `json:"action"`
	FieldName *string `json:"field_name"`
	OldValue  *string `json:"old_value"`
	NewValue  *string `json:"new_value"`
	CreatedAt string  `json:"created_at"`
}