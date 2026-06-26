package teams

import "time"

type Team struct {
	ID        int64
	Name      string
	CreatedBy int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MemberRole string

const (
	RoleOwner  MemberRole = "owner"
	RoleAdmin  MemberRole = "admin"
	RoleMember MemberRole = "member"
)