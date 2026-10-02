package group

import (
	entity "github.com/josofm/liliana/internal/entity/group"
	user "github.com/josofm/liliana/internal/entity/user"
	repository "github.com/josofm/liliana/internal/repository/group"
	userRepo "github.com/josofm/liliana/internal/repository/user"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInvitationPermissions(t *testing.T) {
	users := userRepo.NewInMemoryRepo()
	for _, name := range []string{"Owner", "Invited", "Outside", "Member"} {
		require.NoError(t, users.Create(&user.User{Name: name}))
	}
	r := repository.NewInMemoryRepo()
	s := NewService(r, users)
	g, err := s.Create(1, "Group", "")
	require.NoError(t, err)
	require.NoError(t, r.AddMember(g.ID, 4))
	for _, actor := range []int64{0, 2, 3, 4} {
		_, err := s.CreateInvitation(actor, g.ID, 2)
		require.ErrorIs(t, err, ErrForbidden)
	}
	_, err = s.CreateInvitation(1, g.ID, 0)
	require.ErrorIs(t, err, ErrInvalidPlayerID)
	_, err = s.CreateInvitation(1, g.ID, 999)
	require.ErrorIs(t, err, userRepo.ErrNotFound)
	_, err = s.CreateInvitation(1, g.ID, 4)
	require.ErrorIs(t, err, repository.ErrAlreadyMember)
	i, err := s.CreateInvitation(1, g.ID, 2)
	require.NoError(t, err)
	require.Equal(t, int64(1), i.InvitedByPlayerID)
	for _, actor := range []int64{0, 1, 3, 4} {
		for _, status := range []entity.InvitationStatus{entity.InvitationAccepted, entity.InvitationDeclined} {
			_, err := s.RespondInvitation(actor, i.ID, status)
			require.ErrorIs(t, err, ErrForbidden)
		}
	}
	for _, actor := range []int64{0, 2, 3, 4} {
		_, err := s.RespondInvitation(actor, i.ID, entity.InvitationCancelled)
		require.ErrorIs(t, err, ErrForbidden)
	}
	for _, actor := range []int64{2, 3, 4} {
		_, err := s.GetGroupInvitations(actor, g.ID)
		require.ErrorIs(t, err, ErrForbidden)
	}
	inbox, err := s.GetPlayerInvitations(3)
	require.NoError(t, err)
	require.Empty(t, inbox)
	inbox, err = s.GetPlayerInvitations(2)
	require.NoError(t, err)
	require.Len(t, inbox, 1)
	_, err = s.GetByID(2, g.ID)
	require.ErrorIs(t, err, ErrForbidden)
	accepted, err := s.RespondInvitation(2, i.ID, entity.InvitationAccepted)
	require.NoError(t, err)
	require.Equal(t, entity.InvitationAccepted, accepted.Status)
	_, err = s.GetByID(2, g.ID)
	require.NoError(t, err)
	_, err = s.CreateInvitation(1, g.ID, 2)
	require.ErrorIs(t, err, repository.ErrAlreadyMember)
}
