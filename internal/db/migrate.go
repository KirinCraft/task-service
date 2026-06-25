package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const migrationsPath = "migrations"

func RunMigrations(db *sql.DB) error {
	sourceURL := "file://" + migrationsPath

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("create mysql migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"mysql",
		driver,
	)

	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
