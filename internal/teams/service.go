package teams

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"task-service/internal/users"
)

const (
	maxEmailLength = 254
)

var (
	ErrAccessDenied = errors.New("access denied")
	ErrInvalidEmail = errors.New("invalid email")
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidName  = errors.New("invalid name")
)

type Service struct {
	teamsRepo *Repository
	usersRepo *users.Repository
}

func NewService(teamsRepo *Repository, usersRepo *users.Repository) *Service {
	return &Service{
		teamsRepo: teamsRepo,
		usersRepo: usersRepo,
	}
}

func (s *Service) Create(ctx context.Context, req CreateTeamRequest, userID int64) (*CreateTeamResponse, error) {
	name := strings.TrimSpace(req.Name)

	if name == "" {
		return nil, ErrInvalidName
	}

	team := Team{
		Name:      name,
		CreatedBy: userID,
	}

	id, err := s.teamsRepo.Create(ctx, team)

	if err != nil {
		return nil, fmt.Errorf("create team: %w", err)
	}

	return &CreateTeamResponse{
		ID:   id,
		Name: name,
	}, nil
}

func (s *Service) List(ctx context.Context, userID int64) ([]TeamResponse, error) {
	userTeams, err := s.teamsRepo.ListByUserID(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}

	res := make([]TeamResponse, 0, len(userTeams))

	for _, team := range userTeams {
		res = append(res, TeamResponse{
			ID:   team.ID,
			Name: team.Name,
			Role: string(team.Role),
		})
	}

	return res, nil
}

func (s *Service) Invite(ctx context.Context, req InviteUserRequest, teamID, inviterID int64) (*InviteUserResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	role, err := s.teamsRepo.GetUserRole(ctx, teamID, inviterID)

	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return nil, ErrAccessDenied
		}

		return nil, fmt.Errorf("get user role: %w", err)
	}

	if role != RoleOwner && role != RoleAdmin {
		return nil, ErrAccessDenied
	}

	user, err := s.usersRepo.FindByEmail(ctx, email)

	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("find user by email: %w", err)
	}

	if user.Status != users.StatusActive {
		return nil, ErrUserNotFound
	}

	if err := s.teamsRepo.AddMember(ctx, teamID, user.ID, RoleMember); err != nil {
		if errors.Is(err, ErrUserAlreadyInTeam) {
			return nil, ErrUserAlreadyInTeam
		}

		return nil, fmt.Errorf("add team member: %w", err)
	}

	return &InviteUserResponse{
		TeamID: teamID,
		UserID: user.ID,
		Role:   string(RoleMember),
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