package group

import (
	"errors"
	entity "github.com/josofm/liliana/internal/entity/group"
)

var ErrInvalidPlayerID = errors.New("invited_player_id must be greater than zero")

func (s *Service) requireOwner(actorID, groupID int64) error {
	g, err := s.GetByID(actorID, groupID)
	if err != nil {
		return err
	}
	if g.OwnerID != actorID {
		return ErrForbidden
	}
	return nil
}
func (s *Service) CreateInvitation(actorID, groupID, playerID int64) (*entity.Invitation, error) {
	if err := s.requireOwner(actorID, groupID); err != nil {
		return nil, err
	}
	if playerID <= 0 {
		return nil, ErrInvalidPlayerID
	}
	if _, err := s.users.GetByID(playerID); err != nil {
		return nil, err
	}
	return s.repo.CreateInvitation(groupID, playerID, actorID)
}
func (s *Service) GetGroupInvitations(actorID, groupID int64) ([]entity.Invitation, error) {
	if err := s.requireOwner(actorID, groupID); err != nil {
		return nil, err
	}
	return s.repo.GetGroupInvitations(groupID)
}
func (s *Service) GetPlayerInvitations(actorID int64) ([]entity.Invitation, error) {
	if actorID <= 0 {
		return nil, ErrForbidden
	}
	return s.repo.GetPlayerInvitations(actorID)
}
func (s *Service) RespondInvitation(actorID, id int64, status entity.InvitationStatus) (*entity.Invitation, error) {
	if actorID <= 0 {
		return nil, ErrForbidden
	}
	i, err := s.repo.GetInvitation(id)
	if err != nil {
		return nil, err
	}
	switch status {
	case entity.InvitationAccepted, entity.InvitationDeclined:
		if i.InvitedPlayerID != actorID {
			return nil, ErrForbidden
		}
	case entity.InvitationCancelled:
		if err := s.requireOwner(actorID, i.GroupID); err != nil {
			return nil, err
		}
	default:
		return nil, ErrForbidden
	}
	return s.repo.RespondInvitation(id, status)
}
