package app

import (
	"database/sql"
	"net/http"
	"task-service/internal/auth"
	"task-service/internal/config"
	httpserver "task-service/internal/http-server"
	"task-service/internal/middleware"
	"task-service/internal/tasks"
	"task-service/internal/teams"
	"task-service/internal/users"
)

type App struct {
	Router http.Handler
	DB     *sql.DB
}

func New(cfg *config.Config, db *sql.DB) *App {
	router := httpserver.NewRouter()

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTTTL)
	authMiddleware := middleware.NewAuth(jwtManager)

	auth.RegisterRoutes(router, buildAuth(db, jwtManager))
	teams.RegisterRoutes(router, buildTeams(db), authMiddleware)
	tasks.RegisterRoutes(router, buildTasks(db), authMiddleware)

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

func buildTeams(db *sql.DB) *teams.Handler {
	teamsRepo := teams.NewRepository(db)
	usersRepo := users.NewRepository(db)

	teamsService := teams.NewService(teamsRepo, usersRepo)

	return teams.NewHandler(teamsService)
}

func buildTasks(db *sql.DB) *tasks.Handler {
	tasksRepo := tasks.NewRepository(db)
	teamsRepo := teams.NewRepository(db)

	tasksService := tasks.NewService(tasksRepo, teamsRepo)

	return tasks.NewHandler(tasksService)
}