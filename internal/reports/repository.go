package reports

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) TeamStats(ctx context.Context) ([]TeamStatsItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			t.id AS team_id,
			t.name AS team_name,
			COUNT(DISTINCT tm.user_id) AS members_count,
			COUNT(DISTINCT CASE
				WHEN task.status = 'done'
				AND task.done_at >= CURRENT_TIMESTAMP - INTERVAL 7 DAY
				THEN task.id
			END) AS done_tasks_last_7_days
		FROM teams t
		LEFT JOIN team_members tm ON tm.team_id = t.id
		LEFT JOIN tasks task ON task.team_id = t.id
		GROUP BY t.id, t.name
		ORDER BY t.name
	`)

	if err != nil {
		return nil, fmt.Errorf("get team stats: %w", err)
	}

	defer rows.Close()

	items := make([]TeamStatsItem, 0)

	for rows.Next() {
		var item TeamStatsItem

		if err := rows.Scan(
			&item.TeamID,
			&item.TeamName,
			&item.MembersCount,
			&item.DoneTasksLast7Days,
		); err != nil {
			return nil, fmt.Errorf("scan team stats: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows team stats: %w", err)
	}

	return items, nil
}

func (r *Repository) TopCreators(ctx context.Context) ([]TopCreatorItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			team_id,
			team_name,
			user_id,
			username,
			tasks_count,
			rank_in_team
		FROM (
			SELECT
				t.id AS team_id,
				t.name AS team_name,
				u.id AS user_id,
				u.name AS username,
				COUNT(task.id) AS tasks_count,
				ROW_NUMBER() OVER (
					PARTITION BY t.id
					ORDER BY COUNT(task.id) DESC, u.id
				) AS rank_in_team
			FROM teams t
			JOIN tasks task ON task.team_id = t.id
			JOIN users u ON u.id = task.created_by
			WHERE task.created_at >= CURRENT_TIMESTAMP - INTERVAL 1 MONTH
			GROUP BY t.id, t.name, u.id, u.name
		) ranked
		WHERE rank_in_team <= 3
		ORDER BY team_name, rank_in_team
	`)

	if err != nil {
		return nil, fmt.Errorf("get top creators: %w", err)
	}

	defer rows.Close()

	items := make([]TopCreatorItem, 0)

	for rows.Next() {
		var item TopCreatorItem

		if err := rows.Scan(
			&item.TeamID,
			&item.TeamName,
			&item.UserID,
			&item.Username,
			&item.TasksCount,
			&item.RankInTeam,
		); err != nil {
			return nil, fmt.Errorf("scan top creators: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows top creators: %w", err)
	}

	return items, nil
}

func (r *Repository) InvalidAssignees(ctx context.Context) ([]InvalidAssigneeItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			task.id AS task_id,
			task.title,
			task.team_id,
			t.name AS team_name,
			task.assignee_id
		FROM tasks task
		JOIN teams t ON t.id = task.team_id
		LEFT JOIN team_members tm
			ON tm.team_id = task.team_id
		   AND tm.user_id = task.assignee_id
		WHERE task.assignee_id IS NOT NULL
		AND tm.user_id IS NULL
		ORDER BY task.id
	`)

	if err != nil {
		return nil, fmt.Errorf("get invalid assignees: %w", err)
	}

	defer rows.Close()

	items := make([]InvalidAssigneeItem, 0)

	for rows.Next() {
		var item InvalidAssigneeItem

		if err := rows.Scan(
			&item.TaskID,
			&item.Title,
			&item.TeamID,
			&item.TeamName,
			&item.AssigneeID,
		); err != nil {
			return nil, fmt.Errorf("scan invalid assignees: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows invalid assignees: %w", err)
	}

	return items, nil
}