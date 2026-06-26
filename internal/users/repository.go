package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

const (
	mysqlDuplicateEntry = 1062
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, user User) (int64, error) {
	query := `
	INSERT INTO users (name, email, password_hash)
	VALUES (?, ?, ?)
	`
	res, err := r.db.ExecContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.PasswordHash,
	)

	if err != nil {
		var mysqlErr *mysql.MySQLError

		if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntry {
			return 0, ErrUserAlreadyExists
		}

		return 0, fmt.Errorf("insert user: %w", err)
	}

	id, err := res.LastInsertId()

	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	return id, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
	SELECT id, name, email, password_hash, status, created_at, updated_at, deleted_at
	FROM users
	WHERE email = ?
	`
	var user User

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("query user by email: %w", err)
	}

	return &user, nil
}