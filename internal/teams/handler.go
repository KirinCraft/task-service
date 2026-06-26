package teams

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
		http.Error(w, "user id not found", http.StatusUnauthorized)
		return
	}

	var req CreateTeamRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.service.Create(r.Context(), req, userID)

	if err != nil {
		if errors.Is(err, ErrInvalidName) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, "user id not found", http.StatusUnauthorized)
		return
	}

	res, err := h.service.List(r.Context(), userID)

	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Invite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	inviterID, ok := middleware.UserIDFromContext(r.Context())

	if !ok {
		http.Error(w, "user id not found", http.StatusUnauthorized)
		return
	}

	teamID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil || teamID <= 0 {
		http.Error(w, "invalid team id", http.StatusBadRequest)
		return
	}

	var req InviteUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.service.Invite(r.Context(), req, teamID, inviterID)

	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if errors.Is(err, ErrInvalidEmail) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrUserAlreadyInTeam) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}