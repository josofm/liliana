package service

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchidektImporter_Import(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/decks/123/", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name":"Partners", "deckFormat":3,
			"categories":[
				{"name":"Commander","includedInDeck":true},
				{"name":"Mainboard","includedInDeck":true},
				{"name":"Maybeboard","includedInDeck":false}
			],
			"cards":[
				{"categories":["Commander"],"quantity":1,"card":{"oracleCard":{"name":"Tymna the Weaver","uid":"id-1","colorIdentity":["White","Black"]}}},
				{"categories":["Commander"],"quantity":1,"card":{"oracleCard":{"name":"Kraum, Ludevic's Opus","uid":"id-2","colorIdentity":["Blue","Red"]}}},
				{"categories":["Mainboard"],"quantity":2,"card":{"oracleCard":{"name":"Forest","uid":"id-3","colorIdentity":["Green"]}}},
				{"categories":["Maybeboard"],"quantity":1,"card":{"oracleCard":{"name":"Ignored Card","uid":"id-4","colorIdentity":[]}}}
			]
		}`))
	}))
	defer server.Close()

	importer := NewArchidektImporterWithBaseURL(server.Client(), server.URL)
	deck, err := importer.Import("https://archidekt.com/decks/123/partners")
	require.NoError(t, err)
	assert.Equal(t, "Partners", deck.Name)
	assert.Equal(t, "commander", deck.Format)
	assert.Equal(t, "WUBRG", deck.Color)
	assert.Equal(t, "Tymna the Weaver / Kraum, Ludevic's Opus", deck.Commander)
	assert.Len(t, deck.Cards, 3)
	assert.Equal(t, "Forest", deck.Cards[0].Name)
	assert.Equal(t, 2, deck.Cards[0].Quantity)
}

func TestArchidektImporter_RejectsUnsupportedSource(t *testing.T) {
	importer := NewArchidektImporter()
	_, err := importer.Import("https://example.com/decks/123")
	assert.ErrorIs(t, err, ErrUnsupportedSource)
}

func TestColorCode_Colorless(t *testing.T) {
	assert.Equal(t, "C", colorCode(nil))
}

type failingImporter struct{ err error }

func (f failingImporter) Import(string) (*deckEntity.Deck, error) { return nil, f.err }

func TestServicePrepare_FallsBackForCompleteManualDeck(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewServiceWithImporter(repo, failingImporter{err: errors.New("offline")})
	d := &deckEntity.Deck{Name: "Manual", Color: "W", Format: "commander", Commander: "Sram", OwnerID: 1, SourceLink: "https://archidekt.com/decks/123"}
	assert.NoError(t, service.Prepare(d))
}

func TestServicePrepare_SourceOnlyRequiresSuccessfulImport(t *testing.T) {
	repo := deckRepo.NewInMemoryRepo()
	service := NewServiceWithImporter(repo, failingImporter{err: errors.New("offline")})
	d := &deckEntity.Deck{OwnerID: 1, SourceLink: "https://archidekt.com/decks/123"}
	assert.EqualError(t, service.Prepare(d), "offline")
}
