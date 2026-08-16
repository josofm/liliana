//go:build integration

package service

import (
	"testing"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScryfallValidator_ValidateRealCards(t *testing.T) {
	cards, err := NewScryfallValidator().Validate([]deckEntity.Card{
		{Name: "Aqueous Form", Quantity: 1},
		{Name: "Vorrac Battlehorns", Quantity: 1},
	})
	require.NoError(t, err)
	require.Len(t, cards, 2)

	for _, card := range cards {
		assert.NotEmpty(t, card.OracleID)
		assert.NotEmpty(t, card.Name)
		assert.NotEmpty(t, card.TypeLine)
		assert.Equal(t, 1, card.Quantity)
	}
	assert.Equal(t, "Aqueous Form", cards[0].Name)
	assert.Equal(t, "{U}", cards[0].ManaCost)
}
