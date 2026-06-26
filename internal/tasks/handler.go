package tasks

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"task-service/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.service.Create(r.Context(), req, userID)

	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, ErrInvalidTeam) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, ErrAccessDenied) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req ListTasksRequest

	teamIDStr := r.URL.Query().Get("team_id")

	if teamIDStr == "" {
		http.Error(w, "team_id is required", http.StatusBadRequest)
		return
	}

	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)

	if err != nil {
		http.Error(w, "invalid team_id", http.StatusBadRequest)
		return
	}

	req.TeamID = teamID
	req.Status = r.URL.Query().Get("status")

	if assigneeIDStr := r.URL.Query().Get("assignee_id"); assigneeIDStr != "" {
		assigneeID, err := strconv.ParseInt(assigneeIDStr, 10, 64)

		if err != nil {
			http.Error(w, "invalid assignee_id", http.StatusBadRequest)
			return
		}

		req.AssigneeID = &assigneeID
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)

		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}

		req.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)

		if err != nil {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}

		req.Offset = offset
	}

	res, err := h.service.List(r.Context(), req, userID)

	if err != nil {
		if errors.Is(err, ErrInvalidTeam) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, ErrInvalidStatus) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, ErrAccessDenied) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	var req UpdateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.service.Update(r.Context(), taskID, req, userID)

	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrInvalidTitle) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, ErrInvalidStatus) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, ErrAccessDenied) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if errors.Is(err, ErrInvalidAssignee) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	res, err := h.service.History(r.Context(), taskID, userID)

	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrAccessDenied) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
		return
	}
}