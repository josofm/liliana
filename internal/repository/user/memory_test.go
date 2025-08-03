package user

import (
	"testing"

	userEntity "github.com/josofm/liliana/internal/entity/user"
	"github.com/stretchr/testify/assert"
)

func TestNewInMemoryRepo(t *testing.T) {
	repo := NewInMemoryRepo()
	assert.NotNil(t, repo)
}

func TestInMemoryRepo_Create(t *testing.T) {
	repo := NewInMemoryRepo()
	
	user := &userEntity.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	
	err := repo.Create(user)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
}

func TestInMemoryRepo_GetAll(t *testing.T) {
	repo := NewInMemoryRepo()
	
	// Create test users
	user1 := &userEntity.User{Name: "User 1", Email: "user1@example.com", Password: "pass1"}
	user2 := &userEntity.User{Name: "User 2", Email: "user2@example.com", Password: "pass2"}
	
	repo.Create(user1)
	repo.Create(user2)
	
	users, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestInMemoryRepo_GetByID(t *testing.T) {
	repo := NewInMemoryRepo()
	
	user := &userEntity.User{Name: "Test User", Email: "test@example.com", Password: "password"}
	repo.Create(user)
	
	// Test successful retrieval
	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, user.Name, found.Name)
	assert.Equal(t, user.Email, found.Email)
	
	// Test not found
	notFound, err := repo.GetByID(999)
	assert.Error(t, err)
	assert.Nil(t, notFound)
	assert.Equal(t, "user not found", err.Error())
}

func TestInMemoryRepo_Update(t *testing.T) {
	repo := NewInMemoryRepo()
	
	// Create user
	user := &userEntity.User{Name: "Original Name", Email: "original@example.com", Password: "pass"}
	repo.Create(user)
	
	// Update user
	updatedUser := &userEntity.User{Name: "Updated Name", Email: "updated@example.com", Password: "newpass"}
	err := repo.Update(1, updatedUser)
	assert.NoError(t, err)
	
	// Verify update
	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, "updated@example.com", found.Email)
	
	// Test update non-existent user
	err = repo.Update(999, updatedUser)
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestInMemoryRepo_Delete(t *testing.T) {
	repo := NewInMemoryRepo()
	
	// Create user
	user := &userEntity.User{Name: "Test User", Email: "test@example.com", Password: "password"}
	repo.Create(user)
	
	// Verify user exists
	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	
	// Delete user
	err = repo.Delete(1)
	assert.NoError(t, err)
	
	// Verify user is deleted
	found, err = repo.GetByID(1)
	assert.Error(t, err)
	assert.Nil(t, found)
} 