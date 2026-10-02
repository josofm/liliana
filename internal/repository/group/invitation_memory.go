package group

import (
	entity "github.com/josofm/liliana/internal/entity/group"
	"sort"
	"time"
)

// invitationCopy is called while the repository lock is held.
func (r *inMemoryRepo) invitationCopy(i entity.Invitation) entity.Invitation {
	i.GroupName = r.groups[i.GroupID].Name
	if i.RespondedAt != nil {
		at := *i.RespondedAt
		i.RespondedAt = &at
	}
	return i
}

func (r *inMemoryRepo) CreateInvitation(groupID, playerID, invitedByID int64) (*entity.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.groups[groupID]; !ok {
		return nil, ErrNotFound
	}
	if _, ok := r.members[groupID][playerID]; ok {
		return nil, ErrAlreadyMember
	}
	for _, i := range r.invitations {
		if i.GroupID == groupID && i.InvitedPlayerID == playerID && i.Status == entity.InvitationPending {
			return nil, ErrInvitationPending
		}
	}
	i := entity.Invitation{ID: r.nextInvitationID, GroupID: groupID, InvitedPlayerID: playerID, InvitedByPlayerID: invitedByID, Status: entity.InvitationPending, CreatedAt: time.Now().UTC()}
	r.nextInvitationID++
	r.invitations[i.ID] = i
	i = r.invitationCopy(i)
	return &i, nil
}

func (r *inMemoryRepo) GetInvitation(id int64) (*entity.Invitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	i, ok := r.invitations[id]
	if !ok {
		return nil, ErrInvitationNotFound
	}
	i = r.invitationCopy(i)
	return &i, nil
}

func (r *inMemoryRepo) pendingInvitations(groupID, playerID int64) ([]entity.Invitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]entity.Invitation, 0)
	for _, i := range r.invitations {
		if i.Status == entity.InvitationPending && (groupID == 0 || i.GroupID == groupID) && (playerID == 0 || i.InvitedPlayerID == playerID) {
			result = append(result, r.invitationCopy(i))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}
func (r *inMemoryRepo) GetGroupInvitations(id int64) ([]entity.Invitation, error) {
	return r.pendingInvitations(id, 0)
}
func (r *inMemoryRepo) GetPlayerInvitations(id int64) ([]entity.Invitation, error) {
	return r.pendingInvitations(0, id)
}

func (r *inMemoryRepo) RespondInvitation(id int64, status entity.InvitationStatus) (*entity.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if status != entity.InvitationAccepted && status != entity.InvitationDeclined && status != entity.InvitationCancelled {
		return nil, ErrInvalidInvitationStatus
	}
	i, ok := r.invitations[id]
	if !ok {
		return nil, ErrInvitationNotFound
	}
	if i.Status != entity.InvitationPending {
		return nil, ErrInvitationClosed
	}
	now := time.Now().UTC()
	if status == entity.InvitationAccepted {
		if _, ok := r.members[i.GroupID][i.InvitedPlayerID]; ok {
			return nil, ErrAlreadyMember
		}
		r.members[i.GroupID][i.InvitedPlayerID] = entity.Member{GroupID: i.GroupID, PlayerID: i.InvitedPlayerID, Role: entity.RoleMember, JoinedAt: now}
	}
	i.Status = status
	i.RespondedAt = &now
	r.invitations[id] = i
	i = r.invitationCopy(i)
	return &i, nil
}
