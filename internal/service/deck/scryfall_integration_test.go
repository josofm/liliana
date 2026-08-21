//go:build integration

package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScryfallValidator_ValidateCardsThroughHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[
			{"oracle_id":"id-aqueous","name":"Aqueous Form","mana_cost":"{U}","type_line":"Enchantment — Aura","color_identity":["U"]},
			{"oracle_id":"id-vorrac","name":"Vorrac Battlehorns","mana_cost":"{2}","type_line":"Artifact — Equipment","color_identity":[]}
		],"not_found":[]}`))
	}))
	defer server.Close()

	cards, err := NewScryfallValidatorWithBaseURL(server.Client(), server.URL).Validate([]deckEntity.Card{
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
