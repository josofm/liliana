package group

import (
	"errors"
	entity "github.com/josofm/liliana/internal/entity/group"
)

var (
	ErrNotFound                = errors.New("group not found")
	ErrMemberNotFound          = errors.New("member not found")
	ErrAlreadyMember           = errors.New("player is already a member")
	ErrOwnerCannotLeave        = errors.New("owner cannot leave the group")
	ErrInvitationNotFound      = errors.New("invitation not found")
	ErrInvitationPending       = errors.New("player already has a pending invitation")
	ErrInvitationClosed        = errors.New("invitation is no longer pending")
	ErrInvalidInvitationStatus = errors.New("invalid invitation response")
)

type Repository interface {
	// Create atomically persists the group and its owner membership.
	Create(*entity.Group) error
	GetByID(int64) (*entity.Group, error)
	GetByPlayerID(int64) ([]*entity.Group, error)
	Update(*entity.Group) error
	GetMembers(int64) ([]entity.Member, error)
	// AddMember is an internal persistence operation, not a public join flow.
	AddMember(int64, int64) error
	RemoveMember(int64, int64) error
	CreateInvitation(groupID, playerID, invitedByID int64) (*entity.Invitation, error)
	GetInvitation(int64) (*entity.Invitation, error)
	GetGroupInvitations(int64) ([]entity.Invitation, error)
	GetPlayerInvitations(int64) ([]entity.Invitation, error)
	// RespondInvitation checks pending status and atomically adds membership on accept.
	RespondInvitation(int64, entity.InvitationStatus) (*entity.Invitation, error)
}
