package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"task-service/internal/teams"

	"github.com/redis/go-redis/v9"
)

const (
	teamTasksCacheTTL = 5 * time.Minute

	limitDefault  = 20
	limitMax      = 100
	offsetDefault = 0
)

var (
	ErrInvalidTitle    = errors.New("invalid title")
	ErrInvalidTeam     = errors.New("invalid team")
	ErrAccessDenied    = errors.New("access denied")
	ErrInvalidStatus   = errors.New("invalid status")
	ErrInvalidAssignee = errors.New("invalid assignee")
)

type Service struct {
	tasksRepo *Repository
	teamsRepo *teams.Repository
	redis     *redis.Client
}

func NewService(tasksRepo *Repository, teamsRepo *teams.Repository, redisClient *redis.Client) *Service {
	return &Service{
		tasksRepo: tasksRepo,
		teamsRepo: teamsRepo,
		redis:     redisClient,
	}
}

func (s *Service) Create(ctx context.Context, req CreateTaskRequest, userID int64) (*TaskResponse, error) {
	title := strings.TrimSpace(req.Title)

	if title == "" {
		return nil, ErrInvalidTitle
	}

	if req.TeamID <= 0 {
		return nil, ErrInvalidTeam
	}

	_, err := s.teamsRepo.GetUserRole(ctx, req.TeamID, userID)

	if err != nil {
		if errors.Is(err, teams.ErrMemberNotFound) {
			return nil, ErrAccessDenied
		}

		return nil, fmt.Errorf("get user role: %w", err)
	}

	task := Task{
		TeamID:      req.TeamID,
		Title:       title,
		Description: strings.TrimSpace(req.Description),
		Status:      StatusTodo,
		AssigneeID:  req.AssigneeID,
		CreatedBy:   userID,
	}

	taskID, err := s.tasksRepo.Create(ctx, task)

	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	s.invalidateTeamTasksCache(ctx, req.TeamID)

	return &TaskResponse{
		ID:          taskID,
		TeamID:      task.TeamID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		AssigneeID:  task.AssigneeID,
		CreatedBy:   task.CreatedBy,
	}, nil
}

func (s *Service) List(ctx context.Context, req ListTasksRequest, userID int64) (*ListTasksResponse, error) {
	if req.TeamID <= 0 {
		return nil, ErrInvalidTeam
	}

	_, err := s.teamsRepo.GetUserRole(ctx, req.TeamID, userID)

	if err != nil {
		if errors.Is(err, teams.ErrMemberNotFound) {
			return nil, ErrAccessDenied
		}

		return nil, fmt.Errorf("get user role: %w", err)
	}

	filter := TaskFilter{
		TeamID:     req.TeamID,
		AssigneeID: req.AssigneeID,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	if filter.Limit <= 0 {
		filter.Limit = limitDefault
	}

	if filter.Limit > limitMax {
		filter.Limit = limitMax
	}

	if filter.Offset < 0 {
		filter.Offset = offsetDefault
	}

	if req.Status != "" {
		status := TaskStatus(req.Status)

		if !isValidTaskStatus(status) {
			return nil, ErrInvalidStatus
		}

		filter.Status = &status
	}

	items, err := s.listTasksCached(ctx, filter)

	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	res := &ListTasksResponse{
		Items: make([]TaskResponse, 0, len(items)),
	}

	for _, item := range items {
		res.Items = append(res.Items, taskToResponse(item))
	}

	return res, nil
}

func (s *Service) Update(ctx context.Context, taskID int64, req UpdateTaskRequest, userID int64) (*TaskResponse, error) {
	if taskID <= 0 {
		return nil, ErrTaskNotFound
	}

	if req.ClearAssignee && req.AssigneeID != nil {
		return nil, ErrInvalidAssignee
	}

	task, err := s.tasksRepo.GetByID(ctx, taskID)

	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}

		return nil, fmt.Errorf("get task: %w", err)
	}

	_, err = s.teamsRepo.GetUserRole(ctx, task.TeamID, userID)

	if err != nil {
		if errors.Is(err, teams.ErrMemberNotFound) {
			return nil, ErrAccessDenied
		}

		return nil, fmt.Errorf("get user role: %w", err)
	}

	update := UpdateTask{
		AssigneeID:    req.AssigneeID,
		ClearAssignee: req.ClearAssignee,
	}

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)

		if title == "" {
			return nil, ErrInvalidTitle
		}

		update.Title = &title
	}

	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		update.Description = &description
	}

	if req.Status != nil {
		status := TaskStatus(*req.Status)

		if !isValidTaskStatus(status) {
			return nil, ErrInvalidStatus
		}

		update.Status = &status
	}

	updatedTask, err := s.tasksRepo.Update(ctx, taskID, userID, update)

	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}

		return nil, fmt.Errorf("update task: %w", err)
	}

	s.invalidateTeamTasksCache(ctx, updatedTask.TeamID)

	return &TaskResponse{
		ID:          updatedTask.ID,
		TeamID:      updatedTask.TeamID,
		Title:       updatedTask.Title,
		Description: updatedTask.Description,
		Status:      string(updatedTask.Status),
		AssigneeID:  updatedTask.AssigneeID,
		CreatedBy:   updatedTask.CreatedBy,
	}, nil
}

func (s *Service) History(ctx context.Context, taskID int64, userID int64) ([]TaskHistoryResponse, error) {
	if taskID <= 0 {
		return nil, ErrTaskNotFound
	}

	task, err := s.tasksRepo.GetByID(ctx, taskID)

	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}

		return nil, fmt.Errorf("get task: %w", err)
	}

	_, err = s.teamsRepo.GetUserRole(ctx, task.TeamID, userID)

	if err != nil {
		if errors.Is(err, teams.ErrMemberNotFound) {
			return nil, ErrAccessDenied
		}

		return nil, fmt.Errorf("get user role: %w", err)
	}

	items, err := s.tasksRepo.History(ctx, taskID)

	if err != nil {
		return nil, fmt.Errorf("get task history: %w", err)
	}

	res := make([]TaskHistoryResponse, 0, len(items))

	for _, item := range items {
		res = append(res, TaskHistoryResponse{
			ID:        item.ID,
			TaskID:    item.TaskID,
			ChangedBy: item.ChangedBy,
			Action:    string(item.Action),
			FieldName: item.FieldName,
			OldValue:  item.OldValue,
			NewValue:  item.NewValue,
			CreatedAt: item.CreatedAt.Format(time.RFC3339),
		})
	}

	return res, nil
}

func isValidTaskStatus(status TaskStatus) bool {
	return status == StatusTodo || status == StatusInProgress || status == StatusDone
}

func taskToResponse(task Task) TaskResponse {
	return TaskResponse{
		ID:          task.ID,
		TeamID:      task.TeamID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		AssigneeID:  task.AssigneeID,
		CreatedBy:   task.CreatedBy,
	}
}

func (s *Service) invalidateTeamTasksCache(ctx context.Context, teamID int64) {
	if s.redis == nil || teamID <= 0 {
		return
	}

	if err := s.redis.Del(ctx, teamTasksCacheKey(teamID)).Err(); err != nil {
		log.Printf("redis del error: %v", err)
	}
}

func (s *Service) listTasksCached(ctx context.Context, filter TaskFilter) ([]Task, error) {
	if !isCacheableTaskList(filter) {
		return s.tasksRepo.List(ctx, filter)
	}

	cacheKey := teamTasksCacheKey(filter.TeamID)

	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()

		switch {
		case err == nil:
			var items []Task
			if err := json.Unmarshal([]byte(cached), &items); err == nil {
				return items, nil
			}
			log.Printf("redis unmarshal error: %v", err)
		case errors.Is(err, redis.Nil):
			//cache miss
		default:
			log.Printf("redis get error: %v", err)
		}
	}

	items, err := s.tasksRepo.List(ctx, filter)

	if err != nil {
		return nil, err
	}

	if s.redis != nil {
		data, err := json.Marshal(items)

		if err != nil {
			log.Printf("redis marshal error: %v", err)
			return items, nil
		}

		if err := s.redis.Set(ctx, cacheKey, data, teamTasksCacheTTL).Err(); err != nil {
			log.Printf("redis set error: %v", err)
		}
	}

	return items, nil
}

func isCacheableTaskList(filter TaskFilter) bool {
	return filter.TeamID > 0 &&
		filter.AssigneeID == nil &&
		filter.Status == nil &&
		filter.Limit == limitDefault &&
		filter.Offset == offsetDefault
}

func teamTasksCacheKey(teamID int64) string {
	return fmt.Sprintf("tasks:team:%d", teamID)
}