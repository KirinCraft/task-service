package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Env     string
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret string
	JWTTTL    time.Duration
}

const (
	DefaultEnv     = "example"
	DefaultAppPort = "8080"
	DefaultJWTTTL  = 24 * time.Hour
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
		JWTSecret:  os.Getenv("JWT_SECRET"),
	}

	jwtTTL := os.Getenv("JWT_TTL")

	if jwtTTL == "" {
		cfg.JWTTTL = DefaultJWTTTL
	} else {
		duration, err := time.ParseDuration(jwtTTL)

		if err != nil {
			panic("JWT_TTL is invalid")
		}

		cfg.JWTTTL = duration
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
		{"JWT_SECRET", c.JWTSecret},
	}

	for _, field := range r {
		if field.value == "" {
			panic(fmt.Sprintf("%s is empty", field.name))
		}
	}

	if len(c.JWTSecret) < 32 {
		panic("JWT_SECRET must be at least 32 characters")
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
