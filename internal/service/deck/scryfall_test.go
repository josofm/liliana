package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScryfallValidator_ResolveCommanderByExactName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/cards/named", r.URL.Path)
		assert.Equal(t, "Atraxa, Praetors' Voice", r.URL.Query().Get("exact"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"oracle_id":"id-1","name":"Atraxa, Praetors' Voice","type_line":"Legendary Creature — Phyrexian Angel Horror","color_identity":["W","U","B","G"],"image_uris":{"normal":"https://example.com/atraxa.jpg"}}`))
	}))
	defer server.Close()

	validator := NewScryfallValidatorWithBaseURL(server.Client(), server.URL)
	commander, err := validator.ResolveCommander("Atraxa, Praetors' Voice")
	require.NoError(t, err)
	assert.Equal(t, "id-1", commander.OracleID)
	assert.Equal(t, []string{"W", "U", "B", "G"}, commander.ColorIdentity)
}

func TestScryfallValidator_RejectsCardThatCannotBeCommander(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"oracle_id":"id-1","name":"Sol Ring","type_line":"Artifact","oracle_text":"{T}: Add {C}{C}.","color_identity":[]}`))
	}))
	defer server.Close()

	validator := NewScryfallValidatorWithBaseURL(server.Client(), server.URL)
	_, err := validator.ResolveCommander("Sol Ring")
	assert.EqualError(t, err, "card cannot be a commander: Sol Ring")
}

func TestScryfallValidator_Validate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/cards/collection", r.URL.Path)
		var request scryfallCollectionRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		assert.Equal(t, "Aqueous Form", request.Identifiers[0].Name)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"oracle_id":"oracle-1","name":"Aqueous Form","mana_cost":"{U}","type_line":"Enchantment — Aura","color_identity":["U"],"image_uris":{"normal":"https://example.com/card.jpg"}}],"not_found":[]}`))
	}))
	defer server.Close()

	validator := NewScryfallValidatorWithBaseURL(server.Client(), server.URL)
	cards, err := validator.Validate([]deckEntity.Card{{Name: "Aqueous Form", Quantity: 2}})
	require.NoError(t, err)
	require.Len(t, cards, 1)
	assert.Equal(t, "oracle-1", cards[0].OracleID)
	assert.Equal(t, 2, cards[0].Quantity)
	assert.Equal(t, "{U}", cards[0].ManaCost)
}

func TestScryfallValidator_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[],"not_found":[{"name":"Not a card"}]}`))
	}))
	defer server.Close()

	validator := NewScryfallValidatorWithBaseURL(server.Client(), server.URL)
	_, err := validator.Validate([]deckEntity.Card{{Name: "Not a card", Quantity: 1}})
	assert.EqualError(t, err, "cards not found: Not a card")
}

func TestScryfallValidator_SkipsCardsWithOracleID(t *testing.T) {
	validator := NewScryfallValidatorWithBaseURL(http.DefaultClient, "http://invalid")
	cards, err := validator.Validate([]deckEntity.Card{{OracleID: "oracle-1", Name: "Imported", Quantity: 1}})
	require.NoError(t, err)
	assert.Equal(t, "oracle-1", cards[0].OracleID)
}
