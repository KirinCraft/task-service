package config

import (
	"fmt"
	"os"
)

type Config struct {
	Env     string
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

const (
	DefaultEnv     = "example"
	DefaultAppPort = "8080"
)

func GetEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = DefaultEnv
	}

	return env
}

func MustLoad() *Config {
	cfg := &Config{
		Env:        GetEnv(),
		AppPort:    os.Getenv("APP_PORT"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
	}

	if cfg.AppPort == "" {
		cfg.AppPort = DefaultAppPort
	}

	cfg.validate()

	return cfg
}

func (c *Config) validate() {
	type requiredField struct {
		name  string
		value string
	}

	r := []requiredField{
		{"DB_HOST", c.DBHost},
		{"DB_PORT", c.DBPort},
		{"DB_USER", c.DBUser},
		{"DB_NAME", c.DBName},
	}

	for _, field := range r {
		if field.value == "" {
			panic(fmt.Sprintf("%s is empty", field.name))
		}
	}
}

func (c *Config) MySQLDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}
