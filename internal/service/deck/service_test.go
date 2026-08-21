package service

import (
	"fmt"
	"strings"
	"testing"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCardValidator struct{}

func (testCardValidator) ResolveCommander(name string) (deckEntity.Card, error) {
	return deckEntity.Card{Name: name, ColorIdentity: []string{"U"}, ImageURI: "https://example.com/commander.jpg"}, nil
}

func (testCardValidator) SearchCommanders(string) ([]CommanderSuggestion, error) {
	return []CommanderSuggestion{{Name: "Thassa, God of the Sea", ColorIdentity: []string{"U"}}}, nil
}

func (testCardValidator) Validate(cards []deckEntity.Card) ([]deckEntity.Card, error) {
	result := make([]deckEntity.Card, len(cards))
	copy(result, cards)
	for index := range result {
		if result[index].Name == "Invalid" {
			return nil, fmt.Errorf("cards not found: Invalid")
		}
		if result[index].OracleID == "" {
			result[index].OracleID = "oracle-" + result[index].Name
		}
		if result[index].ImageURI == "" {
			result[index].ImageURI = "https://example.com/" + strings.ReplaceAll(strings.ToLower(result[index].Name), " ", "-") + ".jpg"
		}
	}
	return result, nil
}

type testSourceImporter struct{ deck *deckEntity.Deck }

func (i testSourceImporter) Import(string) (*deckEntity.Deck, error) {
	copy := *i.deck
	copy.Cards = append([]deckEntity.Card(nil), i.deck.Cards...)
	return &copy, nil
}

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
		Format:    "commander",
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
	deck1 := &deckEntity.Deck{Name: "Deck 1", Color: "WU", Format: "commander", Commander: "Azorius", OwnerID: 1}
	deck2 := &deckEntity.Deck{Name: "Deck 2", Color: "BR", Format: "commander", Commander: "Rakdos", OwnerID: 2}

	service.Create(deck1)
	service.Create(deck2)

	decks, err := service.GetAll()
	assert.NoError(t, err)
	assert.Len(t, decks, 2)
}

func TestService_GetByID(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewService(repo)

	deck := &deckEntity.Deck{Name: "Test Deck", Color: "WUBRG", Format: "commander", Commander: "Atraxa", OwnerID: 1}
	service.Create(deck)

	// Test successful retrieval
	found, err := service.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, deck.Name, found.Name)
	assert.Equal(t, deck.Color, found.Color)
	assert.Equal(t, deck.Format, found.Format)
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
	deck := &deckEntity.Deck{Name: "Original Deck", Color: "WU", Format: "commander", Commander: "Azorius", OwnerID: 1, Cards: []deckEntity.Card{{OracleID: "oracle-existing", Name: "Existing", Quantity: 1}}}
	service.Create(deck)

	// Update deck
	updatedDeck := &deckEntity.Deck{Name: "Updated Deck", Color: "BR", Format: "commander", Commander: "Rakdos", OwnerID: 2}
	err := service.Update(1, updatedDeck)
	assert.NoError(t, err)

	// Verify update
	found, err := service.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Deck", found.Name)
	assert.Equal(t, "BR", found.Color)
	assert.Equal(t, "commander", found.Format)
	assert.Equal(t, "Rakdos", found.Commander)
	require.Len(t, found.Cards, 1)
	assert.Equal(t, "oracle-existing", found.Cards[0].OracleID)
}

func TestService_Delete(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewService(repo)

	// Create deck
	deck := &deckEntity.Deck{Name: "Test Deck", Color: "WUBRG", Format: "commander", Commander: "Atraxa", OwnerID: 1}
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

func TestParseCardList(t *testing.T) {
	cards, err := ParseCardList("1 Aqueous Form\n1 Vorrac Battlehorns\n2 aqueous form\n")
	assert.NoError(t, err)
	assert.Equal(t, []deckEntity.Card{
		{Name: "Aqueous Form", Quantity: 3},
		{Name: "Vorrac Battlehorns", Quantity: 1},
	}, cards)
}

func TestParseCardList_InvalidLine(t *testing.T) {
	_, err := ParseCardList("Aqueous Form")
	assert.EqualError(t, err, "invalid card at line 1: expected '<quantity> <card name>'")
}

func TestService_PrepareManualCommanderDerivesColorAndKeepsEmptyCards(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewServiceWithDependencies(repo, NewArchidektImporter(), testCardValidator{})
	d := &deckEntity.Deck{Name: "Manual", Format: "commander", Commander: "Atraxa", OwnerID: 1}

	require.NoError(t, service.Prepare(d))
	assert.Equal(t, "U", d.Color)
	assert.Equal(t, "https://example.com/commander.jpg", d.CommanderImageURI)
	assert.NotNil(t, d.Cards)
	assert.Empty(t, d.Cards)
	assert.NoError(t, service.Create(d))
}

func TestService_PrepareImportedDeckEnrichesCardImages(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	importer := testSourceImporter{deck: &deckEntity.Deck{
		Name: "Imported", Color: "U", Format: "commander", Commander: "Thassa",
		Cards: []deckEntity.Card{{OracleID: "oracle-aura", Name: "Aqueous Form", Quantity: 1}},
	}}
	service := NewServiceWithDependencies(repo, importer, testCardValidator{})
	d := &deckEntity.Deck{OwnerID: 7, SourceLink: "https://archidekt.com/decks/123"}

	require.NoError(t, service.Prepare(d))
	require.Len(t, d.Cards, 1)
	assert.Equal(t, "https://example.com/aqueous-form.jpg", d.Cards[0].ImageURI)
	assert.Equal(t, "https://example.com/commander.jpg", d.CommanderImageURI)
	assert.Equal(t, int64(7), d.OwnerID)
}

func TestColorIdentityCode(t *testing.T) {
	assert.Equal(t, "WUBG", colorIdentityCode([]string{"G", "B", "W", "U"}))
	assert.Equal(t, "C", colorIdentityCode(nil))
}

func TestService_AddCards(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewServiceWithDependencies(repo, NewArchidektImporter(), testCardValidator{})
	d := &deckEntity.Deck{Name: "Manual", Color: "U", Format: "commander", Commander: "Thassa", OwnerID: 1, Cards: []deckEntity.Card{{OracleID: "oracle-aqueous form", Name: "Aqueous Form", Quantity: 1}}}
	assert.NoError(t, service.Create(d))

	updated, err := service.AddCards(d.ID, []deckEntity.Card{{Name: "aqueous form", Quantity: 2}, {Name: "Vorrac Battlehorns", Quantity: 1}})
	assert.NoError(t, err)
	assert.Equal(t, 3, updated.Cards[0].Quantity)
	assert.Equal(t, "Vorrac Battlehorns", updated.Cards[1].Name)
}

func TestService_PatchCardsUpsertsRemovesAndIsIdempotent(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewServiceWithDependencies(repo, NewArchidektImporter(), testCardValidator{})
	d := &deckEntity.Deck{Name: "Manual", Color: "U", Format: "commander", Commander: "Thassa", OwnerID: 1, Cards: []deckEntity.Card{{OracleID: "oracle-Old Card", Name: "Old Card", Quantity: 1}}}
	require.NoError(t, service.Create(d))

	updated, err := service.PatchCards(d.ID, []deckEntity.Card{{Name: "Aqueous Form", Quantity: 2}}, []string{"oracle-Old Card"})
	require.NoError(t, err)
	require.Len(t, updated.Cards, 1)
	assert.Equal(t, "Aqueous Form", updated.Cards[0].Name)
	assert.Equal(t, 2, updated.Cards[0].Quantity)

	updated, err = service.PatchCards(d.ID, []deckEntity.Card{{Name: "Aqueous Form", Quantity: 3}}, nil)
	require.NoError(t, err)
	updated, err = service.PatchCards(d.ID, []deckEntity.Card{{Name: "Aqueous Form", Quantity: 3}}, nil)
	require.NoError(t, err)
	require.Len(t, updated.Cards, 1)
	assert.Equal(t, 3, updated.Cards[0].Quantity)
}

func TestService_PatchCardsDoesNotChangeDeckWhenValidationFails(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewServiceWithDependencies(repo, NewArchidektImporter(), testCardValidator{})
	d := &deckEntity.Deck{Name: "Manual", Color: "U", Format: "commander", Commander: "Thassa", OwnerID: 1, Cards: []deckEntity.Card{{OracleID: "oracle-existing", Name: "Existing", Quantity: 1}}}
	require.NoError(t, service.Create(d))

	_, err := service.PatchCards(d.ID, []deckEntity.Card{{Name: "Invalid", Quantity: 1}}, []string{"oracle-existing"})
	require.Error(t, err)
	found, getErr := service.GetByID(d.ID)
	require.NoError(t, getErr)
	require.Len(t, found.Cards, 1)
	assert.Equal(t, "oracle-existing", found.Cards[0].OracleID)
}
