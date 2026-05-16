//go:build integration

package user

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	userEntity "github.com/josofm/liliana/internal/entity/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultPostgresTestDSN = "postgres://liliana:liliana@localhost:5433/liliana_test?sslmode=disable"

func setupPostgresRepo(t *testing.T) Repository {
	t.Helper()

	dsn := os.Getenv("LILIANA_POSTGRES_TEST_DSN")
	if dsn == "" {
		dsn = defaultPostgresTestDSN
	}

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Minute)

	if err := db.Ping(); err != nil {
		t.Skipf("postgres test database unavailable: %v", err)
	}

	ensurePostgresUserSchema(t, db)
	return NewPostgresRepo(db)
}

func ensurePostgresUserSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`TRUNCATE TABLE users RESTART IDENTITY`)
	require.NoError(t, err)
}

func TestNewPostgresRepo(t *testing.T) {
	repo := setupPostgresRepo(t)
	assert.NotNil(t, repo)
}

func TestPostgresRepo_Create(t *testing.T) {
	repo := setupPostgresRepo(t)

	user := &userEntity.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := repo.Create(user)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
}

func TestPostgresRepo_GetAll(t *testing.T) {
	repo := setupPostgresRepo(t)

	user1 := &userEntity.User{Name: "User 1", Email: "user1@example.com", Password: "pass1"}
	user2 := &userEntity.User{Name: "User 2", Email: "user2@example.com", Password: "pass2"}

	require.NoError(t, repo.Create(user1))
	require.NoError(t, repo.Create(user2))

	users, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestPostgresRepo_GetByID(t *testing.T) {
	repo := setupPostgresRepo(t)

	user := &userEntity.User{Name: "Test User", Email: "test@example.com", Password: "password"}
	require.NoError(t, repo.Create(user))

	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, user.Name, found.Name)
	assert.Equal(t, user.Email, found.Email)

	notFound, err := repo.GetByID(999)
	assert.Error(t, err)
	assert.Nil(t, notFound)
	assert.Equal(t, "user not found", err.Error())
}

func TestPostgresRepo_GetByEmail(t *testing.T) {
	repo := setupPostgresRepo(t)

	user := &userEntity.User{Name: "Test User", Email: "test@example.com", Password: "password"}
	require.NoError(t, repo.Create(user))

	found, err := repo.GetByEmail("test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user.Name, found.Name)
	assert.Equal(t, user.Email, found.Email)

	notFound, err := repo.GetByEmail("missing@example.com")
	assert.Error(t, err)
	assert.Nil(t, notFound)
	assert.Equal(t, "user not found", err.Error())
}

func TestPostgresRepo_Update(t *testing.T) {
	repo := setupPostgresRepo(t)

	user := &userEntity.User{Name: "Original Name", Email: "original@example.com", Password: "pass"}
	require.NoError(t, repo.Create(user))

	updatedUser := &userEntity.User{Name: "Updated Name", Email: "updated@example.com", Password: "newpass"}
	err := repo.Update(1, updatedUser)
	assert.NoError(t, err)

	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, "updated@example.com", found.Email)

	err = repo.Update(999, updatedUser)
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestPostgresRepo_Delete(t *testing.T) {
	repo := setupPostgresRepo(t)

	user := &userEntity.User{Name: "Test User", Email: "test@example.com", Password: "password"}
	require.NoError(t, repo.Create(user))

	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, found)

	err = repo.Delete(1)
	assert.NoError(t, err)

	found, err = repo.GetByID(1)
	assert.Error(t, err)
	assert.Nil(t, found)
}
