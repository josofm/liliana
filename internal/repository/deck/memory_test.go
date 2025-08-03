package deck

import (
	"testing"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	"github.com/stretchr/testify/assert"
)

func TestNewInMemoryRepo(t *testing.T) {
	repo := NewInMemoryRepo()
	assert.NotNil(t, repo)
}

func TestInMemoryRepo_Create(t *testing.T) {
	repo := NewInMemoryRepo()
	
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

func TestInMemoryRepo_GetAll(t *testing.T) {
	repo := NewInMemoryRepo()
	
	// Create test decks
	deck1 := &deckEntity.Deck{Name: "Deck 1", Color: "WU", Commander: "Azorius", OwnerID: 1}
	deck2 := &deckEntity.Deck{Name: "Deck 2", Color: "BR", Commander: "Rakdos", OwnerID: 2}
	
	repo.Create(deck1)
	repo.Create(deck2)
	
	decks, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, decks, 2)
}

func TestInMemoryRepo_GetByID(t *testing.T) {
	repo := NewInMemoryRepo()
	
	deck := &deckEntity.Deck{Name: "Test Deck", Color: "WUBRG", Commander: "Atraxa", OwnerID: 1}
	repo.Create(deck)
	
	// Test successful retrieval
	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, deck.Name, found.Name)
	assert.Equal(t, deck.Color, found.Color)
	assert.Equal(t, deck.Commander, found.Commander)
	
	// Test not found
	notFound, err := repo.GetByID(999)
	assert.Error(t, err)
	assert.Nil(t, notFound)
	assert.Equal(t, "deck not found", err.Error())
}

func TestInMemoryRepo_Update(t *testing.T) {
	repo := NewInMemoryRepo()
	
	// Create deck
	deck := &deckEntity.Deck{Name: "Original Deck", Color: "WU", Commander: "Azorius", OwnerID: 1}
	repo.Create(deck)
	
	// Update deck
	updatedDeck := &deckEntity.Deck{Name: "Updated Deck", Color: "BR", Commander: "Rakdos", OwnerID: 2}
	err := repo.Update(1, updatedDeck)
	assert.NoError(t, err)
	
	// Verify update
	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Deck", found.Name)
	assert.Equal(t, "BR", found.Color)
	assert.Equal(t, "Rakdos", found.Commander)
	
	// Test update non-existent deck
	err = repo.Update(999, updatedDeck)
	assert.Error(t, err)
	assert.Equal(t, "deck not found", err.Error())
}

func TestInMemoryRepo_Delete(t *testing.T) {
	repo := NewInMemoryRepo()
	
	// Create deck
	deck := &deckEntity.Deck{Name: "Test Deck", Color: "WUBRG", Commander: "Atraxa", OwnerID: 1}
	repo.Create(deck)
	
	// Verify deck exists
	found, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	
	// Delete deck
	err = repo.Delete(1)
	assert.NoError(t, err)
	
	// Verify deck is deleted
	found, err = repo.GetByID(1)
	assert.Error(t, err)
	assert.Nil(t, found)
} 