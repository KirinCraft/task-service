package teams

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
	ErrMemberNotFound    = errors.New("member not found")
	ErrUserAlreadyInTeam = errors.New("user already in team")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, team Team) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}

	defer tx.Rollback()

	res, err := tx.ExecContext(
		ctx,
		`
		INSERT INTO teams (name, created_by)
		VALUES (?, ?)
		`,
		team.Name,
		team.CreatedBy,
	)

	if err != nil {
		return 0, fmt.Errorf("insert team: %w", err)
	}

	teamID, err := res.LastInsertId()

	if err != nil {
		return 0, fmt.Errorf("get team id: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		`
		INSERT INTO team_members (team_id, user_id, role)
		VALUES (?, ?, ?)
		`,
		teamID,
		team.CreatedBy,
		RoleOwner,
	)
	if err != nil {
		return 0, fmt.Errorf("insert team owner: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return teamID, nil
}

func (r *Repository) ListByUserID(ctx context.Context, userID int64) ([]UserTeam, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`
		SELECT t.id, t.name, tm.role
		FROM team_members tm
		JOIN teams t ON t.id = tm.team_id
		WHERE tm.user_id = ?
		ORDER BY t.created_at DESC
		`,
		userID,
	)

	if err != nil {
		return nil, fmt.Errorf("query teams by user id: %w", err)
	}

	defer rows.Close()

	teams := make([]UserTeam, 0)

	for rows.Next() {

		var team UserTeam

		err := rows.Scan(
			&team.ID,
			&team.Name,
			&team.Role,
		)

		if err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}

		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate teams: %w", err)
	}

	return teams, nil
}

func (r *Repository) GetUserRole(ctx context.Context, teamID, userID int64) (MemberRole, error) {
	var role MemberRole

	err := r.db.QueryRowContext(
		ctx,
		`
		SELECT role
		FROM team_members
		WHERE team_id = ? AND user_id = ?
		`,
		teamID,
		userID,
	).Scan(&role)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrMemberNotFound
		}

		return "", fmt.Errorf("get user role: %w", err)
	}

	return role, nil
}

func (r *Repository) AddMember(ctx context.Context, teamID, userID int64, role MemberRole) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		INSERT INTO team_members (team_id, user_id, role)
		VALUES (?, ?, ?)
		`,
		teamID,
		userID,
		role,
	)

	if err != nil {
		var mysqlErr *mysql.MySQLError

		if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntry {
			return ErrUserAlreadyInTeam
		}
		return fmt.Errorf("insert team member: %w", err)
	}

	return nil
}