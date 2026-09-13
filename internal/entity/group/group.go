package group

import "time"

type Group struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     int64     `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Role string

const (
	RoleOwner  Role = "owner"
	RoleMember Role = "member"
)

// PlayerID references the existing User entity.
type Member struct {
	GroupID  int64     `json:"group_id"`
	PlayerID int64     `json:"player_id"`
	Role     Role      `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}
