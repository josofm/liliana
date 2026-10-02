package group

import (
	entity "github.com/josofm/liliana/internal/entity/group"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

// Shared contract runs against memory and real PostgreSQL repositories.
func testInvitations(t *testing.T, r Repository) {
	t.Helper()
	g := &entity.Group{Name: "Invitations", OwnerID: 1}
	require.NoError(t, r.Create(g))
	_, err := r.CreateInvitation(g.ID, 1, 1)
	require.ErrorIs(t, err, ErrAlreadyMember)
	i, err := r.CreateInvitation(g.ID, 2, 1)
	require.NoError(t, err)
	require.Equal(t, entity.InvitationPending, i.Status)
	require.Nil(t, i.RespondedAt)
	require.Equal(t, g.Name, i.GroupName)
	_, err = r.CreateInvitation(g.ID, 2, 1)
	require.ErrorIs(t, err, ErrInvitationPending)
	inbox, err := r.GetPlayerInvitations(2)
	require.NoError(t, err)
	require.Len(t, inbox, 1)
	_, err = r.RespondInvitation(i.ID, entity.InvitationPending)
	require.ErrorIs(t, err, ErrInvalidInvitationStatus)
	declined, err := r.RespondInvitation(i.ID, entity.InvitationDeclined)
	require.NoError(t, err)
	require.NotNil(t, declined.RespondedAt)
	_, err = r.RespondInvitation(i.ID, entity.InvitationAccepted)
	require.ErrorIs(t, err, ErrInvitationClosed)
	members, err := r.GetMembers(g.ID)
	require.NoError(t, err)
	require.Len(t, members, 1)
	i, err = r.CreateInvitation(g.ID, 2, 1)
	require.NoError(t, err)
	cancelled, err := r.RespondInvitation(i.ID, entity.InvitationCancelled)
	require.NoError(t, err)
	require.Equal(t, entity.InvitationCancelled, cancelled.Status)
	i, err = r.CreateInvitation(g.ID, 2, 1)
	require.NoError(t, err)
	accepted, err := r.RespondInvitation(i.ID, entity.InvitationAccepted)
	require.NoError(t, err)
	require.Equal(t, entity.InvitationAccepted, accepted.Status)
	_, err = r.RespondInvitation(i.ID, entity.InvitationAccepted)
	require.ErrorIs(t, err, ErrInvitationClosed)
	members, err = r.GetMembers(g.ID)
	require.NoError(t, err)
	require.Len(t, members, 2)
	require.Equal(t, entity.RoleMember, members[1].Role)
	inbox, err = r.GetPlayerInvitations(2)
	require.NoError(t, err)
	require.Empty(t, inbox)
	pending, err := r.GetGroupInvitations(g.ID)
	require.NoError(t, err)
	require.Empty(t, pending)
	require.NoError(t, r.RemoveMember(g.ID, 2))
	_, err = r.CreateInvitation(g.ID, 2, 1)
	require.NoError(t, err)
	_, err = r.GetInvitation(999999)
	require.ErrorIs(t, err, ErrInvitationNotFound)
}

func testConcurrentInvitations(t *testing.T, r Repository) {
	t.Helper()
	g := &entity.Group{Name: "Concurrent", OwnerID: 1}
	require.NoError(t, r.Create(g))
	var wg sync.WaitGroup
	results := make(chan error, 12)
	for n := 0; n < 12; n++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := r.CreateInvitation(g.ID, 2, 1); results <- err }()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else {
			require.ErrorIs(t, err, ErrInvitationPending)
		}
	}
	require.Equal(t, 1, successes)
	invitations, err := r.GetGroupInvitations(g.ID)
	require.NoError(t, err)
	require.Len(t, invitations, 1)
	results = make(chan error, 12)
	for n := 0; n < 12; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			status := []entity.InvitationStatus{entity.InvitationAccepted, entity.InvitationDeclined, entity.InvitationCancelled}[n%3]
			_, err := r.RespondInvitation(invitations[0].ID, status)
			results <- err
		}(n)
	}
	wg.Wait()
	close(results)
	successes = 0
	for err := range results {
		if err == nil {
			successes++
		} else {
			require.ErrorIs(t, err, ErrInvitationClosed)
		}
	}
	require.Equal(t, 1, successes)
	i, err := r.GetInvitation(invitations[0].ID)
	require.NoError(t, err)
	require.NotNil(t, i.RespondedAt)
	members, err := r.GetMembers(g.ID)
	require.NoError(t, err)
	if i.Status == entity.InvitationAccepted {
		require.Len(t, members, 2)
	} else {
		require.Len(t, members, 1)
	}
}

func TestMemoryInvitations(t *testing.T)           { testInvitations(t, NewInMemoryRepo()) }
func TestMemoryConcurrentInvitations(t *testing.T) { testConcurrentInvitations(t, NewInMemoryRepo()) }
