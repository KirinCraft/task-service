package main

import (
	"context"
	"log"
	"net/http"
	"task-service/internal/config"
	database "task-service/internal/db"
	httpserver "task-service/internal/http-server"

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

	r := httpserver.NewRouter()

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}