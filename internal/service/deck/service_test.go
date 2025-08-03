package service

import (
	"testing"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewService(repo)
	assert.NotNil(t, service)
}

func TestService_Create(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewService(repo)

	deck := &deckEntity.Deck{
		Name:      "Test Deck",
		Color:     "WUBRG",
		Commander: "Atraxa",
		OwnerID:   1,
	}

	err := service.Create(deck)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deck.ID)
}

func TestService_GetAll(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewService(repo)

	// Create test decks
	deck1 := &deckEntity.Deck{Name: "Deck 1", Color: "WU", Commander: "Azorius", OwnerID: 1}
	deck2 := &deckEntity.Deck{Name: "Deck 2", Color: "BR", Commander: "Rakdos", OwnerID: 2}

	service.Create(deck1)
	service.Create(deck2)

	decks, err := service.GetAll()
	assert.NoError(t, err)
	assert.Len(t, decks, 2)
}

func TestService_GetByID(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewService(repo)

	deck := &deckEntity.Deck{Name: "Test Deck", Color: "WUBRG", Commander: "Atraxa", OwnerID: 1}
	service.Create(deck)

	// Test successful retrieval
	found, err := service.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, deck.Name, found.Name)
	assert.Equal(t, deck.Color, found.Color)
	assert.Equal(t, deck.Commander, found.Commander)

	// Test not found
	notFound, err := service.GetByID(999)
	assert.Error(t, err)
	assert.Nil(t, notFound)
}

func TestService_Update(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewService(repo)

	// Create deck
	deck := &deckEntity.Deck{Name: "Original Deck", Color: "WU", Commander: "Azorius", OwnerID: 1}
	service.Create(deck)

	// Update deck
	updatedDeck := &deckEntity.Deck{Name: "Updated Deck", Color: "BR", Commander: "Rakdos", OwnerID: 2}
	err := service.Update(1, updatedDeck)
	assert.NoError(t, err)

	// Verify update
	found, err := service.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Deck", found.Name)
	assert.Equal(t, "BR", found.Color)
	assert.Equal(t, "Rakdos", found.Commander)
}

func TestService_Delete(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewService(repo)

	// Create deck
	deck := &deckEntity.Deck{Name: "Test Deck", Color: "WUBRG", Commander: "Atraxa", OwnerID: 1}
	service.Create(deck)

	// Verify deck exists
	found, err := service.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, found)

	// Delete deck
	err = service.Delete(1)
	assert.NoError(t, err)

	// Verify deck is deleted
	found, err = service.GetByID(1)
	assert.Error(t, err)
	assert.Nil(t, found)
}
