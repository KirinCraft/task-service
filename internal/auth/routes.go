package auth

import (
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Post("/api/v1/register", h.Register)
	r.Post("/api/v1/login", h.Login)
}