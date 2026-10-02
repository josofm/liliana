package group

import "time"

type InvitationStatus string

const (
	InvitationPending   InvitationStatus = "pending"
	InvitationAccepted  InvitationStatus = "accepted"
	InvitationDeclined  InvitationStatus = "declined"
	InvitationCancelled InvitationStatus = "cancelled"
)

type Invitation struct {
	ID                int64            `json:"id"`
	GroupID           int64            `json:"group_id"`
	GroupName         string           `json:"group_name"`
	InvitedPlayerID   int64            `json:"invited_player_id"`
	InvitedByPlayerID int64            `json:"invited_by_player_id"`
	Status            InvitationStatus `json:"status"`
	CreatedAt         time.Time        `json:"created_at"`
	RespondedAt       *time.Time       `json:"responded_at"`
}
