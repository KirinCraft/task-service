package main

import (
	"log"
	"task-service/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	env := config.GetEnv()
	envFile := ".env." + env

	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Not found %s, using system variables", envFile)
	}

	_ = config.MustLoad()
}
