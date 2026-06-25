package main

import (
	"context"
	"log"
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
}
