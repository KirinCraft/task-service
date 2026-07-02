package reports

type TeamStatsResponse struct {
	Items []TeamStatsItem `json:"items"`
}

type TeamStatsItem struct {
	TeamID             int64  `json:"team_id"`
	TeamName           string `json:"team_name"`
	MembersCount       int64  `json:"members_count"`
	DoneTasksLast7Days int64  `json:"done_tasks_last_7_days"`
}

type TopCreatorsResponse struct {
	Items []TopCreatorItem `json:"items"`
}

type TopCreatorItem struct {
	TeamID     int64  `json:"team_id"`
	TeamName   string `json:"team_name"`
	UserID     int64  `json:"user_id"`
	Username   string `json:"username"`
	TasksCount int64  `json:"tasks_count"`
	RankInTeam int64  `json:"rank_in_team"`
}

type InvalidAssigneesResponse struct {
	Items []InvalidAssigneeItem `json:"items"`
}

type InvalidAssigneeItem struct {
	TaskID     int64  `json:"task_id"`
	Title      string `json:"title"`
	TeamID     int64  `json:"team_id"`
	TeamName   string `json:"team_name"`
	AssigneeID int64  `json:"assignee_id"`
}