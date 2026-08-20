package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	assert.Equal(t, "https://example.com/atraxa.jpg", commander.ImageURI)
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

func TestScryfallValidator_SearchCommandersFiltersLegendaryCreatures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/cards/search", r.URL.Path)
		assert.Equal(t, "atraxa t:legendary t:creature", r.URL.Query().Get("q"))
		assert.Equal(t, "name", r.URL.Query().Get("order"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"name":"Atraxa, Praetors' Voice","type_line":"Legendary Creature — Phyrexian Angel Horror","color_identity":["W","U","B","G"]},{"name":"Jace","type_line":"Legendary Planeswalker — Jace","color_identity":["U"]}]}`))
	}))
	defer server.Close()

	validator := NewScryfallValidatorWithBaseURL(server.Client(), server.URL)
	result, err := validator.SearchCommanders("atraxa")
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Atraxa, Praetors' Voice", result[0].Name)
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

func TestScryfallValidator_ValidateModalDoubleFacedCardByFaceName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		var request scryfallCollectionRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.Len(t, request.Identifiers, 1)
		assert.Equal(t, "Boggart Trawler", request.Identifiers[0].Name)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data":[{
				"oracle_id":"1c5f9d7d-52b2-4d46-85ad-example",
				"name":"Boggart Trawler // Boggart Bog",
				"type_line":"Creature — Goblin Rogue // Land",
				"color_identity":["B"],
				"card_faces":[
					{"name":"Boggart Trawler","mana_cost":"{2}{B}","type_line":"Creature — Goblin Rogue","oracle_text":"When Boggart Trawler enters...","colors":["B"],"image_uris":{"normal":"https://example.com/trawler.jpg"}},
					{"name":"Boggart Bog","mana_cost":"","type_line":"Land","oracle_text":"As Boggart Bog enters...","colors":[],"image_uris":{"normal":"https://example.com/bog.jpg"}}
				]
			}],
			"not_found":[]
		}`))
	}))
	defer server.Close()

	validator := NewScryfallValidatorWithBaseURL(server.Client(), server.URL)
	cards, err := validator.Validate([]deckEntity.Card{{Name: "Boggart Trawler", Quantity: 2}})
	require.NoError(t, err)
	require.Len(t, cards, 1)
	assert.Equal(t, "Boggart Trawler // Boggart Bog", cards[0].Name)
	assert.Equal(t, 2, cards[0].Quantity)
	assert.Equal(t, "{2}{B}", cards[0].ManaCost)
	assert.Equal(t, "Creature — Goblin Rogue // Land", cards[0].TypeLine)
	assert.Equal(t, []string{"B"}, cards[0].ColorIdentity)
	assert.Equal(t, "https://example.com/trawler.jpg", cards[0].ImageURI)
}

func TestScryfallValidator_LimitsRequestsToTwoPerSecond(t *testing.T) {
	requestTimes := make([]time.Time, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestTimes = append(requestTimes, time.Now())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	validator := NewScryfallValidatorWithBaseURL(server.Client(), server.URL)
	_, err := validator.SearchCommanders("atraxa")
	require.NoError(t, err)
	_, err = validator.SearchCommanders("thassa")
	require.NoError(t, err)
	require.Len(t, requestTimes, 2)
	assert.GreaterOrEqual(t, requestTimes[1].Sub(requestTimes[0]), 450*time.Millisecond)
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

func TestScryfallValidator_SkipsCardsAlreadyEnriched(t *testing.T) {
	validator := NewScryfallValidatorWithBaseURL(http.DefaultClient, "http://invalid")
	cards, err := validator.Validate([]deckEntity.Card{{OracleID: "oracle-1", Name: "Imported", Quantity: 1, ImageURI: "https://example.com/imported.jpg"}})
	require.NoError(t, err)
	assert.Equal(t, "oracle-1", cards[0].OracleID)
}

func TestScryfallValidator_EnrichesImportedCardMissingImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request scryfallCollectionRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		assert.Equal(t, "Aqueous Form", request.Identifiers[0].Name)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"oracle_id":"oracle-1","name":"Aqueous Form","mana_cost":"{U}","type_line":"Enchantment — Aura","color_identity":["U"],"image_uris":{"normal":"https://example.com/aura.jpg"}}],"not_found":[]}`))
	}))
	defer server.Close()

	validator := NewScryfallValidatorWithBaseURL(server.Client(), server.URL)
	cards, err := validator.Validate([]deckEntity.Card{{OracleID: "oracle-1", Name: "Aqueous Form", Quantity: 2}})
	require.NoError(t, err)
	require.Len(t, cards, 1)
	assert.Equal(t, "https://example.com/aura.jpg", cards[0].ImageURI)
	assert.Equal(t, 2, cards[0].Quantity)
}
