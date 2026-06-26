package users

import "time"

type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

const (
	StatusActive  Status = "active"
	StatusDeleted Status = "deleted"
)

type Status string