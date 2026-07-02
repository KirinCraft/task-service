package main

import (
	"context"
	"log"
	"net/http"
	application "task-service/internal/app"
	"task-service/internal/config"
	database "task-service/internal/db"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	env := config.GetEnv()
	envFile := ".env." + env

	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Not found %s, using system variables", envFile)
	}

	cfg := config.MustLoad()
	ctx := context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("redis unavailable: %v", err)
	}

	db, err := database.NewMySQL(ctx, cfg.MySQLDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatal(err)
	}

	app := application.New(cfg, db, redisClient)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: app.Router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}