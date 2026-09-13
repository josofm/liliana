package group

import (
	entity "github.com/josofm/liliana/internal/entity/group"
	user "github.com/josofm/liliana/internal/entity/user"
	repository "github.com/josofm/liliana/internal/repository/group"
	userRepo "github.com/josofm/liliana/internal/repository/user"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestGroupPermissionsAndMembership(t *testing.T) {
	users := userRepo.NewInMemoryRepo()
	for _, name := range []string{"Owner", "Member", "Outside"} {
		require.NoError(t, users.Create(&user.User{Name: name, Email: name + "@example.com"}))
	}
	repo := repository.NewInMemoryRepo()
	s := NewService(repo, users)
	g, err := s.Create(1, " Commander Night ", "Weekly games")
	require.NoError(t, err)
	require.Equal(t, "Commander Night", g.Name)
	members, err := s.GetMembers(1, g.ID)
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, entity.RoleOwner, members[0].Role)
	_, err = s.Create(1, "Second group", "")
	require.NoError(t, err)
	groups, err := s.GetByPlayerID(1)
	require.NoError(t, err)
	require.Len(t, groups, 2)
	require.NoError(t, repo.AddMember(g.ID, 2))
	require.ErrorIs(t, repo.AddMember(g.ID, 2), repository.ErrAlreadyMember)
	_, err = s.GetByID(2, g.ID)
	require.NoError(t, err)
	_, err = s.GetByID(3, g.ID)
	require.ErrorIs(t, err, ErrForbidden)
	groups, err = s.GetByPlayerID(3)
	require.NoError(t, err)
	require.Empty(t, groups)
	name := "Updated"
	_, err = s.Update(2, g.ID, &name, nil)
	require.ErrorIs(t, err, ErrForbidden)
	updated, err := s.Update(1, g.ID, &name, nil)
	require.NoError(t, err)
	require.Equal(t, "Weekly games", updated.Description)
	require.ErrorIs(t, s.RemoveMember(2, g.ID, 1), ErrForbidden)
	require.ErrorIs(t, s.RemoveMember(1, g.ID, 1), repository.ErrOwnerCannotLeave)
	require.NoError(t, s.RemoveMember(2, g.ID, 2))
	_, err = s.GetByID(2, g.ID)
	require.ErrorIs(t, err, ErrForbidden)
	require.NoError(t, repo.AddMember(g.ID, 2))
	require.NoError(t, s.RemoveMember(1, g.ID, 2))
	require.ErrorIs(t, s.RemoveMember(1, g.ID, 2), repository.ErrMemberNotFound)
}

func TestGroupValidation(t *testing.T) {
	users := userRepo.NewInMemoryRepo()
	require.NoError(t, users.Create(&user.User{Name: "Owner"}))
	s := NewService(repository.NewInMemoryRepo(), users)
	for _, name := range []string{"", " \t\n", strings.Repeat("a", 101)} {
		_, err := s.Create(1, name, "")
		require.ErrorIs(t, err, ErrInvalidName)
	}
	_, err := s.Create(999, "Group", "")
	require.Error(t, err)
	_, err = s.Create(0, "Group", "")
	require.ErrorIs(t, err, ErrForbidden)
	g, err := s.Create(1, strings.Repeat("ã", 100), "")
	require.NoError(t, err)
	blank := " "
	_, err = s.Update(1, g.ID, &blank, nil)
	require.ErrorIs(t, err, ErrInvalidName)
	stored, err := s.GetByID(1, g.ID)
	require.NoError(t, err)
	require.Equal(t, g.Name, stored.Name)
}
