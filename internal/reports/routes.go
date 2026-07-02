package reports

import (
	"task-service/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler, authMiddleware *middleware.AuthMiddleware) {
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Handle)

		r.Get("/api/v1/reports/team-stats", h.TeamStats)
		r.Get("/api/v1/reports/top-creators", h.TopCreators)
		r.Get("/api/v1/reports/invalid-assignees", h.InvalidAssignees)
	})
}
