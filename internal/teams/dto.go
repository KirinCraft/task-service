package teams

type CreateTeamRequest struct {
	Name string `json:"name"`
}

type CreateTeamResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type TeamResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type UserTeam struct {
	ID   int64
	Name string
	Role MemberRole
}

type InviteUserRequest struct {
	Email string `json:"email"`
}

type InviteUserResponse struct {
	TeamID int64  `json:"team_id"`
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}