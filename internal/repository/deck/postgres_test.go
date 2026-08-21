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

const defaultPostgresTestDSN = "postgres://liliana:liliana@localhost:5432/liliana_test?sslmode=disable"

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

	truncatePostgresDecks(t, db)
	return NewPostgresRepo(db)
}

func truncatePostgresDecks(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`TRUNCATE TABLE deck_cards, decks, cards RESTART IDENTITY`)
	require.NoError(t, err)
}

func TestNewPostgresRepo(t *testing.T) {
	repo := setupPostgresRepo(t)
	assert.NotNil(t, repo)
}

func TestPostgresRepo_Create(t *testing.T) {
	repo := setupPostgresRepo(t)

	deck := &deckEntity.Deck{
		Name:              "Test Deck",
		Color:             "WUBRG",
		Format:            "commander",
		Commander:         "Atraxa, Praetors' Voice",
		CommanderImageURI: "https://example.com/atraxa.jpg",
		OwnerID:           1,
		SourceLink:        "https://archidekt.com/decks/123456",
		Cards: []deckEntity.Card{{
			OracleID: "1b8d0a2b-79ab-4a2f-9d2f-0e6d5dc7c461", Name: "Aqueous Form", Quantity: 2,
			ManaCost: "{U}", TypeLine: "Enchantment — Aura", ColorIdentity: []string{"U"},
			CardFaces: []deckEntity.CardFace{{Name: "Front", ImageURI: "https://example.com/front.jpg"}, {Name: "Back", ImageURI: "https://example.com/back.jpg"}},
		}},
	}

	err := repo.Create(deck)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deck.ID)
	found, err := repo.GetByID(deck.ID)
	require.NoError(t, err)
	require.Len(t, found.Cards, 1)
	assert.Equal(t, "Aqueous Form", found.Cards[0].Name)
	assert.Equal(t, 2, found.Cards[0].Quantity)
	assert.Equal(t, "{U}", found.Cards[0].ManaCost)
	assert.Equal(t, deck.CommanderImageURI, found.CommanderImageURI)
	require.Len(t, found.Cards[0].CardFaces, 2)
	assert.Equal(t, "https://example.com/back.jpg", found.Cards[0].CardFaces[1].ImageURI)
}

func TestPostgresRepo_GetAll(t *testing.T) {
	repo := setupPostgresRepo(t)

	deck1 := &deckEntity.Deck{Name: "Deck 1", Color: "WU", Format: "commander", Commander: "Azorius", OwnerID: 1}
	deck2 := &deckEntity.Deck{Name: "Deck 2", Color: "BR", Format: "commander", Commander: "Rakdos", OwnerID: 2}

	require.NoError(t, repo.Create(deck1))
	require.NoError(t, repo.Create(deck2))

	decks, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, decks, 2)
}

func TestPostgresRepo_GetByID(t *testing.T) {
	repo := setupPostgresRepo(t)

	deck := &deckEntity.Deck{Name: "Test Deck", Color: "WUBRG", Format: "commander", Commander: "Atraxa", OwnerID: 1}
	require.NoError(t, repo.Create(deck))

	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, deck.Name, found.Name)
	assert.Equal(t, deck.Color, found.Color)
	assert.Equal(t, deck.Format, found.Format)
	assert.Equal(t, deck.Commander, found.Commander)

	notFound, err := repo.GetByID(999)
	assert.Error(t, err)
	assert.Nil(t, notFound)
	assert.Equal(t, "deck not found", err.Error())
}

func TestPostgresRepo_Update(t *testing.T) {
	repo := setupPostgresRepo(t)

	deck := &deckEntity.Deck{Name: "Original Deck", Color: "WU", Format: "commander", Commander: "Azorius", OwnerID: 1}
	require.NoError(t, repo.Create(deck))

	updatedDeck := &deckEntity.Deck{Name: "Updated Deck", Color: "BR", Format: "commander", Commander: "Rakdos", OwnerID: 2}
	err := repo.Update(1, updatedDeck)
	assert.NoError(t, err)

	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Deck", found.Name)
	assert.Equal(t, "BR", found.Color)
	assert.Equal(t, "commander", found.Format)
	assert.Equal(t, "Rakdos", found.Commander)

	err = repo.Update(999, updatedDeck)
	assert.Error(t, err)
	assert.Equal(t, "deck not found", err.Error())
}

func TestPostgresRepo_Delete(t *testing.T) {
	repo := setupPostgresRepo(t)

	deck := &deckEntity.Deck{Name: "Test Deck", Color: "WUBRG", Format: "commander", Commander: "Atraxa", OwnerID: 1}
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

func TestPostgresRepo_SharedCardRelationshipAndCascade(t *testing.T) {
	repo := setupPostgresRepo(t)
	postgres := repo.(*postgresRepo)
	card := deckEntity.Card{OracleID: "7a8a2d6f-8e24-4d87-8f42-42d4d7e535f5", Name: "Shared Card", Quantity: 1}
	first := &deckEntity.Deck{Name: "First", Color: "U", Format: "commander", Commander: "First Commander", OwnerID: 1, Cards: []deckEntity.Card{card}}
	second := &deckEntity.Deck{Name: "Second", Color: "U", Format: "commander", Commander: "Second Commander", OwnerID: 1, Cards: []deckEntity.Card{card}}
	require.NoError(t, repo.Create(first))
	require.NoError(t, repo.Create(second))

	var cardCount, relationshipCount int
	require.NoError(t, postgres.db.QueryRow(`SELECT COUNT(*) FROM cards WHERE oracle_id=$1`, card.OracleID).Scan(&cardCount))
	require.NoError(t, postgres.db.QueryRow(`SELECT COUNT(*) FROM deck_cards WHERE oracle_id=$1`, card.OracleID).Scan(&relationshipCount))
	assert.Equal(t, 1, cardCount)
	assert.Equal(t, 2, relationshipCount)

	require.NoError(t, repo.Delete(first.ID))
	require.NoError(t, postgres.db.QueryRow(`SELECT COUNT(*) FROM deck_cards WHERE oracle_id=$1`, card.OracleID).Scan(&relationshipCount))
	require.NoError(t, postgres.db.QueryRow(`SELECT COUNT(*) FROM cards WHERE oracle_id=$1`, card.OracleID).Scan(&cardCount))
	assert.Equal(t, 1, relationshipCount)
	assert.Equal(t, 1, cardCount)
}

func TestPostgresRepo_PatchCardsIsTransactionalAndIdempotent(t *testing.T) {
	repo := setupPostgresRepo(t)
	existing := deckEntity.Card{OracleID: "oracle-existing", Name: "Existing", Quantity: 1, ImageURI: "https://example.com/existing.jpg"}
	d := &deckEntity.Deck{Name: "Deck", Color: "U", Format: "commander", Commander: "Thassa", OwnerID: 1, Cards: []deckEntity.Card{existing}}
	require.NoError(t, repo.Create(d))

	invalid := deckEntity.Card{Name: "Missing Oracle ID", Quantity: 1}
	err := repo.PatchCards(d.ID, []deckEntity.Card{invalid}, []string{existing.OracleID})
	require.Error(t, err)
	unchanged, err := repo.GetByID(d.ID)
	require.NoError(t, err)
	require.Len(t, unchanged.Cards, 1)
	assert.Equal(t, existing.OracleID, unchanged.Cards[0].OracleID)

	replacement := deckEntity.Card{OracleID: "oracle-new", Name: "New", Quantity: 2, ImageURI: "https://example.com/new.jpg"}
	require.NoError(t, repo.PatchCards(d.ID, []deckEntity.Card{replacement}, []string{existing.OracleID}))
	require.NoError(t, repo.PatchCards(d.ID, []deckEntity.Card{replacement}, nil))
	updated, err := repo.GetByID(d.ID)
	require.NoError(t, err)
	require.Len(t, updated.Cards, 1)
	assert.Equal(t, replacement.OracleID, updated.Cards[0].OracleID)
	assert.Equal(t, 2, updated.Cards[0].Quantity)
}
