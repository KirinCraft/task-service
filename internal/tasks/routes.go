package tasks

import (
	"task-service/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler, authMiddleware *middleware.AuthMiddleware) {
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Handle)

		r.Post("/api/v1/tasks", h.Create)
		r.Get("/api/v1/tasks", h.List)
		r.Put("/api/v1/tasks/{id}", h.Update)
		r.Get("/api/v1/tasks/{id}/history", h.History)
	})
}