package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"task-service/internal/users"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.service.Register(r.Context(), req)

	if err != nil {
		var validErr ValidationError

		if errors.As(err, &validErr) {
			http.Error(w, validErr.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, users.ErrUserAlreadyExists) {
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

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.service.Login(r.Context(), req)

	if err != nil {
		var validErr ValidationError

		if errors.As(err, &validErr) {
			http.Error(w, validErr.Error(), http.StatusUnauthorized)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}