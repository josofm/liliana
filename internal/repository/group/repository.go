package group

import (
	"errors"
	entity "github.com/josofm/liliana/internal/entity/group"
)

var (
	ErrNotFound         = errors.New("group not found")
	ErrMemberNotFound   = errors.New("member not found")
	ErrAlreadyMember    = errors.New("player is already a member")
	ErrOwnerCannotLeave = errors.New("owner cannot leave the group")
)

type Repository interface {
	// Create atomically persists the group and its owner membership.
	Create(*entity.Group) error
	GetByID(int64) (*entity.Group, error)
	GetByPlayerID(int64) ([]*entity.Group, error)
	Update(*entity.Group) error
	GetMembers(int64) ([]entity.Member, error)
	// AddMember is persistence infrastructure for the future invitation flow.
	AddMember(int64, int64) error
	RemoveMember(int64, int64) error
}
