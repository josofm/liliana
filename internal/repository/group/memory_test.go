package group

import (
	entity "github.com/josofm/liliana/internal/entity/group"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestMemoryMembershipAndIsolation(t *testing.T) {
	r := NewInMemoryRepo()
	g := &entity.Group{Name: "Group", OwnerID: 1}
	require.NoError(t, r.Create(g))
	other := &entity.Group{Name: "Other", OwnerID: 2}
	require.NoError(t, r.Create(other))
	require.NoError(t, r.AddMember(g.ID, 2))
	groups, err := r.GetByPlayerID(2)
	require.NoError(t, err)
	require.Len(t, groups, 2)
	g.Name = "Mutated"
	stored, err := r.GetByID(g.ID)
	require.NoError(t, err)
	require.Equal(t, "Group", stored.Name)
	stored.OwnerID = 3
	require.NoError(t, r.Update(stored))
	require.Equal(t, int64(1), stored.OwnerID)
	require.ErrorIs(t, r.RemoveMember(g.ID, 1), ErrOwnerCannotLeave)
	require.ErrorIs(t, r.AddMember(999, 1), ErrNotFound)
	require.ErrorIs(t, r.RemoveMember(g.ID, 999), ErrMemberNotFound)
}

func TestMemoryConcurrentMembership(t *testing.T) {
	r := NewInMemoryRepo()
	g := &entity.Group{Name: "Group", OwnerID: 1}
	require.NoError(t, r.Create(g))
	var wg sync.WaitGroup
	errors := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errors <- r.AddMember(g.ID, 2) }()
	}
	wg.Wait()
	close(errors)
	added := 0
	for err := range errors {
		if err == nil {
			added++
		} else {
			require.ErrorIs(t, err, ErrAlreadyMember)
		}
	}
	require.Equal(t, 1, added)
	members, err := r.GetMembers(g.ID)
	require.NoError(t, err)
	require.Len(t, members, 2)
}
