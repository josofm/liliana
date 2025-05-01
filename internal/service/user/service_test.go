//go:build unit

package user_test

import (
	"testing"

	entityUser "github.com/josofm/liliana/internal/entity/user"
	repo "github.com/josofm/liliana/internal/repository/user"
	"github.com/josofm/liliana/internal/service/user"
	"github.com/stretchr/testify/assert"
)

func TestUserService(t *testing.T) {
	repository := repo.NewInMemoryRepo()
	service := user.NewService(repository)

	t.Run("should create and retrieve user", func(t *testing.T) {
		u := &entityUser.User{
			Name:     "Liliana",
			Email:    "liliana@mtg.com",
			Password: "necromancer",
		}

		err := service.Create(u)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), u.ID)

		all, err := service.GetAll()
		assert.NoError(t, err)
		assert.Len(t, all, 1)
		assert.Equal(t, "Liliana", all[0].Name)

		found, err := service.GetByID(u.ID)
		assert.NoError(t, err)
		assert.Equal(t, u.Email, found.Email)
	})

	t.Run("should delete user", func(t *testing.T) {
		err := service.Delete(1)
		assert.NoError(t, err)

		_, err = service.GetByID(1)
		assert.Error(t, err)
	})
}
