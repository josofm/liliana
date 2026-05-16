//go:build integration

package deck

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	deckEntity "github.com/josofm/liliana/internal/entity/deck"
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

	ensurePostgresDeckSchema(t, db)
	return NewPostgresRepo(db)
}

func ensurePostgresDeckSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS decks (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			color TEXT NOT NULL,
			commander TEXT NOT NULL,
			owner_id BIGINT NOT NULL,
			source_link TEXT NOT NULL DEFAULT ''
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`TRUNCATE TABLE decks RESTART IDENTITY`)
	require.NoError(t, err)
}

func TestNewPostgresRepo(t *testing.T) {
	repo := setupPostgresRepo(t)
	assert.NotNil(t, repo)
}

func TestPostgresRepo_Create(t *testing.T) {
	repo := setupPostgresRepo(t)

	deck := &deckEntity.Deck{
		Name:       "Test Deck",
		Color:      "WUBRG",
		Commander:  "Atraxa, Praetors' Voice",
		OwnerID:    1,
		SourceLink: "https://archidekt.com/decks/123456",
	}

	err := repo.Create(deck)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deck.ID)
}

func TestPostgresRepo_GetAll(t *testing.T) {
	repo := setupPostgresRepo(t)

	deck1 := &deckEntity.Deck{Name: "Deck 1", Color: "WU", Commander: "Azorius", OwnerID: 1}
	deck2 := &deckEntity.Deck{Name: "Deck 2", Color: "BR", Commander: "Rakdos", OwnerID: 2}

	require.NoError(t, repo.Create(deck1))
	require.NoError(t, repo.Create(deck2))

	decks, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, decks, 2)
}

func TestPostgresRepo_GetByID(t *testing.T) {
	repo := setupPostgresRepo(t)

	deck := &deckEntity.Deck{Name: "Test Deck", Color: "WUBRG", Commander: "Atraxa", OwnerID: 1}
	require.NoError(t, repo.Create(deck))

	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, deck.Name, found.Name)
	assert.Equal(t, deck.Color, found.Color)
	assert.Equal(t, deck.Commander, found.Commander)

	notFound, err := repo.GetByID(999)
	assert.Error(t, err)
	assert.Nil(t, notFound)
	assert.Equal(t, "deck not found", err.Error())
}

func TestPostgresRepo_Update(t *testing.T) {
	repo := setupPostgresRepo(t)

	deck := &deckEntity.Deck{Name: "Original Deck", Color: "WU", Commander: "Azorius", OwnerID: 1}
	require.NoError(t, repo.Create(deck))

	updatedDeck := &deckEntity.Deck{Name: "Updated Deck", Color: "BR", Commander: "Rakdos", OwnerID: 2}
	err := repo.Update(1, updatedDeck)
	assert.NoError(t, err)

	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Deck", found.Name)
	assert.Equal(t, "BR", found.Color)
	assert.Equal(t, "Rakdos", found.Commander)

	err = repo.Update(999, updatedDeck)
	assert.Error(t, err)
	assert.Equal(t, "deck not found", err.Error())
}

func TestPostgresRepo_Delete(t *testing.T) {
	repo := setupPostgresRepo(t)

	deck := &deckEntity.Deck{Name: "Test Deck", Color: "WUBRG", Commander: "Atraxa", OwnerID: 1}
	require.NoError(t, repo.Create(deck))

	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, found)

	err = repo.Delete(1)
	assert.NoError(t, err)

	found, err = repo.GetByID(1)
	assert.Error(t, err)
	assert.Nil(t, found)
}
