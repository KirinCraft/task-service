package tasks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, task Task) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}

	defer tx.Rollback()

	res, err := tx.ExecContext(
		ctx,
		`
		INSERT INTO tasks (
			team_id,
			title,
			description,
			status,
			assignee_id,
			created_by
		)
		VALUES (?, ?, ?, ?, ?, ?)
		`,
		task.TeamID,
		task.Title,
		task.Description,
		task.Status,
		task.AssigneeID,
		task.CreatedBy,
	)

	if err != nil {
		return 0, fmt.Errorf("insert task: %w", err)
	}

	taskID, err := res.LastInsertId()

	if err != nil {
		return 0, fmt.Errorf("get task id: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		`
		INSERT INTO task_history (
			task_id,
			changed_by,
			action
		)
		VALUES (?, ?, ?)
		`,
		taskID,
		task.CreatedBy,
		ActionCreated,
	)

	if err != nil {
		return 0, fmt.Errorf("insert task history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return taskID, nil
}

func (r *Repository) List(ctx context.Context, filter TaskFilter) ([]Task, error) {
	query := `
		SELECT
			id,
			team_id,
			title,
			description,
			status,
			assignee_id,
			created_by,
			created_at,
			updated_at,
			done_at
		FROM tasks
		WHERE team_id = ?
	`

	args := []any{filter.TeamID}

	if filter.Status != nil {
		query += " AND status = ?"
		args = append(args, *filter.Status)
	}

	if filter.AssigneeID != nil {
		query += " AND assignee_id = ?"
		args = append(args, *filter.AssigneeID)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	defer rows.Close()

	tasks := make([]Task, 0)

	for rows.Next() {
		task, err := scanTask(rows)

		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tasks, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (Task, error) {
	row := r.db.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			team_id,
			title,
			description,
			status,
			assignee_id,
			created_by,
			created_at,
			updated_at,
			done_at
		FROM tasks
		WHERE id = ?
		`,
		id,
	)

	task, err := scanTask(row)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, ErrTaskNotFound
		}

		return Task{}, fmt.Errorf("get task by id: %w", err)
	}

	return task, nil
}

func (r *Repository) Update(ctx context.Context, id, changedBy int64, update UpdateTask) (Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return Task{}, fmt.Errorf("begin tx: %w", err)
	}

	defer tx.Rollback()

	oldTask, err := getTaskByIDTx(ctx, tx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, ErrTaskNotFound
		}

		return Task{}, fmt.Errorf("get old task: %w", err)
	}

	setParts := make([]string, 0)
	args := make([]any, 0)
	history := make([]TaskHistory, 0)

	if update.Title != nil && *update.Title != oldTask.Title {
		setParts = append(setParts, "title = ?")
		args = append(args, *update.Title)
		history = append(history, newHistory(id, changedBy, "title", oldTask.Title, *update.Title))
	}

	if update.Description != nil && *update.Description != oldTask.Description {
		setParts = append(setParts, "description = ?")
		args = append(args, *update.Description)
		history = append(history, newHistory(id, changedBy, "description", oldTask.Description, *update.Description))
	}

	if update.Status != nil && *update.Status != oldTask.Status {
		setParts = append(setParts, "status = ?")
		args = append(args, *update.Status)

		if *update.Status == StatusDone {
			setParts = append(setParts, "done_at = CURRENT_TIMESTAMP")
		} else {
			setParts = append(setParts, "done_at = NULL")
		}

		history = append(history, newHistory(id, changedBy, "status", string(oldTask.Status), string(*update.Status)))
	}

	if update.ClearAssignee {
		if oldTask.AssigneeID != nil {
			setParts = append(setParts, "assignee_id = NULL")
			history = append(history, newHistory(
				id,
				changedBy,
				"assignee_id",
				int64PtrToString(oldTask.AssigneeID),
				"",
			))
		}
	} else if update.AssigneeID != nil && !sameInt64Ptr(update.AssigneeID, oldTask.AssigneeID) {
		setParts = append(setParts, "assignee_id = ?")
		args = append(args, *update.AssigneeID)
		history = append(history, newHistory(
			id,
			changedBy,
			"assignee_id",
			int64PtrToString(oldTask.AssigneeID),
			strconv.FormatInt(*update.AssigneeID, 10),
		))
	}

	if len(setParts) == 0 {
		return oldTask, nil
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")

	query := fmt.Sprintf(
		"UPDATE tasks SET %s WHERE id = ?",
		strings.Join(setParts, ", "),
	)

	args = append(args, id)

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return Task{}, fmt.Errorf("update task: %w", err)
	}

	for _, item := range history {
		_, err := tx.ExecContext(
			ctx,
			`
			INSERT INTO task_history (
				task_id,
				changed_by,
				action,
				field_name,
				old_value,
				new_value
			)
			VALUES (?, ?, ?, ?, ?, ?)
			`,
			item.TaskID,
			item.ChangedBy,
			item.Action,
			item.FieldName,
			item.OldValue,
			item.NewValue,
		)

		if err != nil {
			return Task{}, fmt.Errorf("insert task history: %w", err)
		}
	}

	updatedTask, err := getTaskByIDTx(ctx, tx, id)

	if err != nil {
		return Task{}, fmt.Errorf("get updated task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Task{}, fmt.Errorf("commit tx: %w", err)
	}

	return updatedTask, nil
}

func (r *Repository) History(ctx context.Context, taskID int64) ([]TaskHistory, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`
		SELECT
			id,
			task_id,
			changed_by,
			action,
			field_name,
			old_value,
			new_value,
			created_at
		FROM task_history
		WHERE task_id = ?
		ORDER BY created_at ASC
		`,
		taskID,
	)

	if err != nil {
		return nil, fmt.Errorf("get task history: %w", err)
	}

	defer rows.Close()

	history := make([]TaskHistory, 0)

	for rows.Next() {
		var item TaskHistory

		if err := rows.Scan(
			&item.ID,
			&item.TaskID,
			&item.ChangedBy,
			&item.Action,
			&item.FieldName,
			&item.OldValue,
			&item.NewValue,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task history: %w", err)
		}

		history = append(history, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return history, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(s scanner) (Task, error) {
	var task Task
	var assigneeID sql.NullInt64
	var doneAt sql.NullTime

	err := s.Scan(
		&task.ID,
		&task.TeamID,
		&task.Title,
		&task.Description,
		&task.Status,
		&assigneeID,
		&task.CreatedBy,
		&task.CreatedAt,
		&task.UpdatedAt,
		&doneAt,
	)

	if err != nil {
		return Task{}, err
	}

	if assigneeID.Valid {
		task.AssigneeID = &assigneeID.Int64
	}

	if doneAt.Valid {
		task.DoneAt = &doneAt.Time
	}

	return task, nil
}

func getTaskByIDTx(ctx context.Context, tx *sql.Tx, id int64) (Task, error) {
	row := tx.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			team_id,
			title,
			description,
			status,
			assignee_id,
			created_by,
			created_at,
			updated_at,
			done_at
		FROM tasks
		WHERE id = ?
		`,
		id,
	)

	return scanTask(row)
}

func newHistory(taskID, changedBy int64, fieldName, oldValue, newValue string) TaskHistory {
	return TaskHistory{
		TaskID:    taskID,
		ChangedBy: changedBy,
		Action:    ActionUpdated,
		FieldName: &fieldName,
		OldValue:  &oldValue,
		NewValue:  &newValue,
	}
}

func sameInt64Ptr(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return *a == *b
}

func int64PtrToString(v *int64) string {
	if v == nil {
		return ""
	}

	return strconv.FormatInt(*v, 10)
}