package app

import (
	"database/sql"
	"net/http"
	"task-service/internal/auth"
	"task-service/internal/config"
	httpserver "task-service/internal/http-server"
	"task-service/internal/users"
)

type App struct {
	Router http.Handler
	DB     *sql.DB
}

func New(cfg *config.Config, db *sql.DB) *App {
	router := httpserver.NewRouter()

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTTTL)

	auth.RegisterRoutes(router, buildAuth(db, jwtManager))

	return &App{
		Router: router,
		DB:     db,
	}
}

func buildAuth(db *sql.DB, jwtManager *auth.JWTManager) *auth.Handler {
	usersRepo := users.NewRepository(db)
	authService := auth.NewService(usersRepo, jwtManager)
	authHandler := auth.NewHandler(authService)

	return authHandler
}