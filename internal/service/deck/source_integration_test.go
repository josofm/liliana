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
const ladroLadrozinhoURL = "https://archidekt.com/decks/7375252/ladro_ladrozinho"

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

func TestArchidektImporter_ImportLadroLadrozinho(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/decks/7375252/", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name":"Ladro ladrozinho", "deckFormat":3,
			"categories":[
				{"name":"Commander","includedInDeck":true},
				{"name":"Mainboard","includedInDeck":true},
				{"name":"Maybeboard","includedInDeck":false},
				{"name":"Sideboard","includedInDeck":true}
			],
			"cards":[
				{"categories":["Commander"],"quantity":1,"card":{"oracleCard":{"name":"Don Andres, the Renegade","uid":"id-commander","colorIdentity":["Blue","Black","Red"]}}},
				{"categories":["Mainboard"],"quantity":99,"card":{"oracleCard":{"name":"Main deck fixture","uid":"id-main","colorIdentity":[]}}},
				{"categories":["Maybeboard"],"quantity":6,"card":{"oracleCard":{"name":"Ignored maybe fixture","uid":"id-maybe","colorIdentity":[]}}},
				{"categories":["Sideboard"],"quantity":3,"card":{"oracleCard":{"name":"Ignored side fixture","uid":"id-side","colorIdentity":[]}}}
			]
		}`))
	}))
	defer server.Close()

	imported, err := NewArchidektImporterWithBaseURL(server.Client(), server.URL).Import(ladroLadrozinhoURL)
	require.NoError(t, err)
	assert.Equal(t, 100, deckQuantity(imported.Cards))
	assert.False(t, hasCardNamed("Ignored maybe fixture", imported.Cards))
	assert.False(t, hasCardNamed("Ignored side fixture", imported.Cards))
}

func deckQuantity(cards []deckEntity.Card) int {
	total := 0
	for _, card := range cards {
		total += card.Quantity
	}
	return total
}
