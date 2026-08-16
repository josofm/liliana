//go:build integration

package service

import (
	"testing"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const elvesVisionsURL = "https://archidekt.com/decks/22559444/elves_visions"

func TestArchidektImporter_ImportElvesVisions(t *testing.T) {
	deck, err := NewArchidektImporter().Import(elvesVisionsURL)
	require.NoError(t, err)

	assert.Equal(t, "Elves visions", deck.Name)
	assert.Equal(t, "commander", deck.Format)
	assert.Equal(t, "UG", deck.Color)
	assert.Equal(t, "Elrond, Master of Healing", deck.Commander)
	assert.NotEmpty(t, deck.Cards)

	assert.True(t, hasCardNamed("Elrond, Master of Healing", deck.Cards))
}

func hasCardNamed(name string, cards []deckEntity.Card) bool {
	for _, card := range cards {
		if card.Name == name {
			return true
		}
	}
	return false
}
