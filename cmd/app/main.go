package main

import (
	"context"
	"log"
	"net/http"
	application "task-service/internal/app"
	"task-service/internal/config"
	database "task-service/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	env := config.GetEnv()
	envFile := ".env." + env

	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Not found %s, using system variables", envFile)
	}

	cfg := config.MustLoad()
	ctx := context.Background()

	db, err := database.NewMySQL(ctx, cfg.MySQLDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatal(err)
	}

	app := application.New(cfg, db)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: app.Router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}