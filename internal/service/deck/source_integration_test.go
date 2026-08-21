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

const elvesVisionsURL = "https://archidekt.com/decks/22559444/elves_visions"

func TestArchidektImporter_ImportElvesVisions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/decks/22559444/", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name":"Elves visions", "deckFormat":3,
			"categories":[{"name":"Commander","includedInDeck":true},{"name":"Mainboard","includedInDeck":true}],
			"cards":[
				{"categories":["Commander"],"quantity":1,"card":{"oracleCard":{"name":"Elrond, Master of Healing","uid":"id-elrond","colorIdentity":["Blue","Green"]}}},
				{"categories":["Mainboard"],"quantity":2,"card":{"oracleCard":{"name":"Llanowar Elves","uid":"id-elves","colorIdentity":["Green"]}}}
			]
		}`))
	}))
	defer server.Close()

	deck, err := NewArchidektImporterWithBaseURL(server.Client(), server.URL).Import(elvesVisionsURL)
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
