package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"task-service/internal/users"

	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLength = 6
	maxPasswordLength = 72
	maxEmailLength    = 254
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

type Service struct {
	usersRepo   *users.Repository
	jwtManager *JWTManager
}

func NewService(usersRep *users.Repository, jwtManager *JWTManager) *Service {
	return &Service{
		usersRepo:   usersRep,
		jwtManager: jwtManager,
	}
}

func (s *Service) Register(ctx context.Context, r RegisterRequest) (*RegisterResponse, error) {
	name := strings.TrimSpace(r.Name)
	email := strings.ToLower(strings.TrimSpace(r.Email))
	password := r.Password

	if name == "" {
		return nil, ValidationError{
			Message: "invalid name",
		}
	}
	if !isValidEmail(email) {
		return nil, ValidationError{
			Message: "invalid email",
		}
	}
	if !isValidPassword(password) {
		return nil, ValidationError{
			Message: "invalid password",
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := users.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(passwordHash),
	}

	id, err := s.usersRepo.Create(ctx, user)

	if err != nil {

		if errors.Is(err, users.ErrUserAlreadyExists) {
			return nil, users.ErrUserAlreadyExists
		}

		return nil, fmt.Errorf("create user: %w", err)
	}

	return &RegisterResponse{
		ID:    id,
		Name:  name,
		Email: email,
	}, nil
}

func (s *Service) Login(ctx context.Context, r LoginRequest) (*LoginResponse, error) {
	email := strings.ToLower(strings.TrimSpace(r.Email))
	password := r.Password

	if !isValidEmail(email) {
		return nil, ValidationError{
			Message: "invalid email",
		}
	}
	if !isValidPassword(password) {
		return nil, ValidationError{
			Message: "invalid password",
		}
	}

	user, err := s.usersRepo.FindByEmail(ctx, email)

	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return nil, ValidationError{
				Message: "invalid email or password",
			}
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	if user.Status != users.StatusActive {
		return nil, ValidationError{
			Message: "invalid email or password",
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ValidationError{
			Message: "invalid email or password",
		}
	}

	token, err := s.jwtManager.Generate(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &LoginResponse{
		Token: token,
	}, nil
}

func isValidEmail(e string) bool {
	if e == "" {
		return false
	}

	if len(e) > maxEmailLength {
		return false
	}

	e = strings.ToLower(strings.TrimSpace(e))

	address, err := mail.ParseAddress(e)

	if err != nil {
		return false
	}

	return address.Address == e
}

func isValidPassword(p string) bool {
	if minPasswordLength > len(p) || len(p) > maxPasswordLength {
		return false
	}

	return !strings.ContainsAny(p, " \t\n\r")
}