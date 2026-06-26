package teams

import (
	"task-service/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler, authMiddleware *middleware.AuthMiddleware) {
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Handle)

		r.Post("/api/v1/teams", h.Create)
		r.Get("/api/v1/teams", h.List)
		r.Post("/api/v1/teams/{id}/invite", h.Invite)
	})
}