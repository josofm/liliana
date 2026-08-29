//go:build integration

package v1_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "github.com/josofm/liliana/internal/controller/http/v1"
	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	deckService "github.com/josofm/liliana/internal/service/deck"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCardValidator struct{}

func (testCardValidator) ResolveCommander(name string) (deckEntity.Card, error) {
	colors := []string{"U"}
	if name == "Atraxa, Praetors' Voice" || name == "Atraxa" {
		colors = []string{"W", "U", "B", "G"}
	}
	return deckEntity.Card{Name: name, ColorIdentity: colors, ImageURI: "https://example.com/commander.jpg"}, nil
}

func (testCardValidator) SearchCommanders(string) ([]deckService.CommanderSuggestion, error) {
	return []deckService.CommanderSuggestion{{Name: "Thassa, God of the Sea", ColorIdentity: []string{"U"}}}, nil
}

func (testCardValidator) SearchCards(string) ([]deckEntity.Card, error) {
	return []deckEntity.Card{{OracleID: "oracle-sol-ring", Name: "Sol Ring", ImageURI: "https://example.com/sol-ring.jpg"}}, nil
}

func (testCardValidator) Validate(cards []deckEntity.Card) ([]deckEntity.Card, error) {
	for index := range cards {
		if cards[index].OracleID == "" {
			cards[index].OracleID = "oracle-" + cards[index].Name
		}
		switch cards[index].Name {
		case "Aqueous Form":
			cards[index].ManaCost = "{U}"
			cards[index].TypeLine = "Enchantment — Aura"
			cards[index].ColorIdentity = []string{"U"}
			cards[index].ImageURI = "https://example.com/aqueous-form.jpg"
		case "Vorrac Battlehorns":
			cards[index].ManaCost = "{2}"
			cards[index].TypeLine = "Artifact — Equipment"
			cards[index].ColorIdentity = []string{}
			cards[index].ImageURI = "https://example.com/vorrac-battlehorns.jpg"
		}
	}
	return cards, nil
}

func setupDeckHandler() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", int64(1)); c.Next() })
	repo := deckRepo.NewInMemoryRepo()
	v1.NewDeckHandler(router, repo)
	return router
}

func setupDeckHandlerWithCardValidation() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", int64(1)); c.Next() })
	repo := deckRepo.NewInMemoryRepo()
	service := deckService.NewServiceWithDependencies(repo, deckService.NewArchidektImporter(), testCardValidator{})
	v1.NewDeckHandlerWithService(router, service)
	return router
}

func TestDeckHandler_Create(t *testing.T) {
	router := setupDeckHandlerWithCardValidation()

	deckRequest := v1.DeckRequest{
		Name:      "Test Deck",
		Color:     "WUBRG",
		Format:    "commander",
		Commander: "Atraxa, Praetors' Voice",
		OwnerID:   1,
	}

	body, err := json.Marshal(deckRequest)
	checkErr(t, err)

	req, err := http.NewRequest("POST", "/decks/", bytes.NewBuffer(body))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response deckEntity.Deck
	err = json.Unmarshal(w.Body.Bytes(), &response)
	checkErr(t, err)
	assert.Equal(t, deckRequest.Name, response.Name)
	assert.Equal(t, "WUBG", response.Color)
	assert.Equal(t, deckRequest.Format, response.Format)
	assert.Equal(t, deckRequest.Commander, response.Commander)
	assert.Equal(t, "https://example.com/commander.jpg", response.CommanderImageURI)
	assert.Equal(t, int64(1), response.OwnerID)
	assert.Equal(t, int64(1), response.ID)
}

func TestDeckHandler_Create_IgnoresOwnerIDFromJSON(t *testing.T) {
	router := setupDeckHandlerWithCardValidation()
	body := []byte(`{"name":"Test Deck","format":"commander","commander":"Atraxa, Praetors' Voice","owner_id":999}`)
	req, err := http.NewRequest(http.MethodPost, "/decks/", bytes.NewReader(body))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var response deckEntity.Deck
	checkErr(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, int64(1), response.OwnerID)
}

func TestDeckHandler_Create_RequiresAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	repo := deckRepo.NewInMemoryRepo()
	service := deckService.NewServiceWithDependencies(repo, deckService.NewArchidektImporter(), testCardValidator{})
	v1.NewDeckHandlerWithService(router, service)
	body := []byte(`{"name":"Test Deck","format":"commander","commander":"Atraxa, Praetors' Voice"}`)
	req, err := http.NewRequest(http.MethodPost, "/decks/", bytes.NewReader(body))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"error":"user not authenticated"}`, w.Body.String())
}

func TestDeckHandler_SearchCommanders(t *testing.T) {
	router := setupDeckHandlerWithCardValidation()
	req, err := http.NewRequest(http.MethodGet, "/decks/commanders?q=thassa", nil)
	checkErr(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.JSONEq(t, `[{"name":"Thassa, God of the Sea","color_identity":["U"]}]`, w.Body.String())
}

func TestDeckHandler_SearchCards(t *testing.T) {
	router := setupDeckHandlerWithCardValidation()
	req, err := http.NewRequest(http.MethodGet, "/decks/cards/search?q=sol", nil)
	checkErr(t, err)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var cards []deckEntity.Card
	checkErr(t, json.Unmarshal(response.Body.Bytes(), &cards))
	require.Len(t, cards, 1)
	assert.Equal(t, "oracle-sol-ring", cards[0].OracleID)
	assert.Equal(t, "Sol Ring", cards[0].Name)
}

func TestDeckHandler_SearchCardsRequiresTwoCharacters(t *testing.T) {
	router := setupDeckHandlerWithCardValidation()
	req, err := http.NewRequest(http.MethodGet, "/decks/cards/search?q=s", nil)
	checkErr(t, err)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestDeckHandler_Create_NonCommanderWithoutCommander(t *testing.T) {
	router := setupDeckHandler()

	deckRequest := v1.DeckRequest{
		Name:    "Standard Deck",
		Color:   "WU",
		Format:  "standard",
		OwnerID: 1,
	}

	body, err := json.Marshal(deckRequest)
	checkErr(t, err)

	req, err := http.NewRequest("POST", "/decks/", bytes.NewBuffer(body))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response deckEntity.Deck
	err = json.Unmarshal(w.Body.Bytes(), &response)
	checkErr(t, err)
	assert.Equal(t, deckRequest.Format, response.Format)
	assert.Empty(t, response.Commander)
}

func TestDeckHandler_CreateManualWithCards(t *testing.T) {
	router := setupDeckHandlerWithCardValidation()
	deckRequest := v1.DeckRequest{
		Name: "Auras", Color: "U", Format: "commander", Commander: "Thassa", OwnerID: 1,
		Cards: "1 Aqueous Form\n1 Vorrac Battlehorns",
	}
	body, err := json.Marshal(deckRequest)
	checkErr(t, err)
	req, err := http.NewRequest(http.MethodPost, "/decks/", bytes.NewBuffer(body))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var response deckEntity.Deck
	checkErr(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Len(t, response.Cards, 2)
	assert.Equal(t, "Aqueous Form", response.Cards[0].Name)
}

func TestDeckHandler_AddCards(t *testing.T) {
	router := setupDeckHandlerWithCardValidation()
	deckRequest := v1.DeckRequest{Name: "Auras", Color: "U", Format: "commander", Commander: "Thassa", OwnerID: 1}
	body, err := json.Marshal(deckRequest)
	checkErr(t, err)
	createRequest, err := http.NewRequest(http.MethodPost, "/decks/", bytes.NewBuffer(body))
	checkErr(t, err)
	createRequest.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, createRequest)
	assert.Equal(t, http.StatusCreated, createResponse.Code)

	body, err = json.Marshal(v1.DeckCardsRequest{Cards: "1 Aqueous Form\n1 Vorrac Battlehorns"})
	checkErr(t, err)
	request, err := http.NewRequest(http.MethodPost, "/decks/1/cards", bytes.NewBuffer(body))
	checkErr(t, err)
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	assert.Equal(t, http.StatusOK, responseRecorder.Code)

	var response deckEntity.Deck
	checkErr(t, json.Unmarshal(responseRecorder.Body.Bytes(), &response))
	assert.Len(t, response.Cards, 2)
}

func TestDeckHandler_PatchCards(t *testing.T) {
	router := setupDeckHandlerWithCardValidation()
	createBody, err := json.Marshal(v1.DeckRequest{Name: "Auras", Format: "commander", Commander: "Thassa", Cards: "1 Aqueous Form"})
	checkErr(t, err)
	createRequest, err := http.NewRequest(http.MethodPost, "/decks/", bytes.NewReader(createBody))
	checkErr(t, err)
	createRequest.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, createRequest)
	require.Equal(t, http.StatusCreated, createResponse.Code, createResponse.Body.String())

	patchBody, err := json.Marshal(v1.DeckCardsPatchRequest{
		Upsert: []v1.DeckCardUpsertRequest{{Name: "Vorrac Battlehorns", Quantity: 2}},
		Remove: []string{"oracle-Aqueous Form"},
	})
	checkErr(t, err)
	patchRequest, err := http.NewRequest(http.MethodPatch, "/decks/1/cards", bytes.NewReader(patchBody))
	checkErr(t, err)
	patchRequest.Header.Set("Content-Type", "application/json")
	patchResponse := httptest.NewRecorder()
	router.ServeHTTP(patchResponse, patchRequest)

	require.Equal(t, http.StatusOK, patchResponse.Code, patchResponse.Body.String())
	var response deckEntity.Deck
	checkErr(t, json.Unmarshal(patchResponse.Body.Bytes(), &response))
	require.Len(t, response.Cards, 1)
	assert.Equal(t, "Vorrac Battlehorns", response.Cards[0].Name)
	assert.Equal(t, 2, response.Cards[0].Quantity)
}

func TestDeckHandler_CreateManualWithValidatedCards(t *testing.T) {
	router := setupDeckHandlerWithCardValidation()
	requestBody := v1.DeckRequest{
		Name: "Validated Auras", Color: "U", Format: "commander", Commander: "Thassa, God of the Sea", OwnerID: 1,
		Cards: "1 Aqueous Form\n1 Vorrac Battlehorns",
	}
	body, err := json.Marshal(requestBody)
	checkErr(t, err)
	request, err := http.NewRequest(http.MethodPost, "/decks/", bytes.NewBuffer(body))
	checkErr(t, err)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())

	var response deckEntity.Deck
	checkErr(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Cards, 2)
	assert.NotEmpty(t, response.Cards[0].OracleID)
	assert.Equal(t, "{U}", response.Cards[0].ManaCost)
}

func TestDeckHandler_CreateFromArchidektWithCards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name":"Elves visions", "deckFormat":3,
			"categories":[{"name":"Commander","includedInDeck":true},{"name":"Mainboard","includedInDeck":true}],
			"cards":[
				{"categories":["Commander"],"quantity":1,"card":{"oracleCard":{"name":"Elrond, Master of Healing","uid":"id-elrond","colorIdentity":["Blue","Green"]}}},
				{"categories":["Mainboard"],"quantity":1,"card":{"oracleCard":{"name":"Llanowar Elves","uid":"id-elves","colorIdentity":["Green"]}}}
			]
		}`))
	}))
	defer server.Close()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", int64(1)); c.Next() })
	repo := deckRepo.NewInMemoryRepo()
	importer := deckService.NewArchidektImporterWithBaseURL(server.Client(), server.URL)
	service := deckService.NewServiceWithDependencies(repo, importer, testCardValidator{})
	v1.NewDeckHandlerWithService(router, service)
	requestBody := v1.DeckRequest{OwnerID: 1, SourceLink: "https://archidekt.com/decks/22559444/elves_visions"}
	body, err := json.Marshal(requestBody)
	checkErr(t, err)
	request, err := http.NewRequest(http.MethodPost, "/decks/", bytes.NewBuffer(body))
	checkErr(t, err)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())

	var response deckEntity.Deck
	checkErr(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, "Elves visions", response.Name)
	assert.Equal(t, "Elrond, Master of Healing", response.Commander)
	assert.NotEmpty(t, response.Cards)
	for _, card := range response.Cards {
		assert.NotEmpty(t, card.OracleID)
	}
}

func TestDeckHandler_Create_InvalidJSON(t *testing.T) {
	router := setupDeckHandler()

	req, err := http.NewRequest("POST", "/decks/", bytes.NewBuffer([]byte("invalid json")))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeckHandler_GetAll(t *testing.T) {
	router := setupDeckHandlerWithCardValidation()

	// Create test decks via HTTP
	deckRequest1 := v1.DeckRequest{Name: "Deck 1", Color: "W", Format: "commander", Commander: "Sram", OwnerID: 1}
	deckRequest2 := v1.DeckRequest{Name: "Deck 2", Color: "U", Format: "commander", Commander: "Baral", OwnerID: 1}

	// Create first deck
	body1, err := json.Marshal(deckRequest1)
	checkErr(t, err)
	req1, err := http.NewRequest("POST", "/decks/", bytes.NewBuffer(body1))
	checkErr(t, err)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Create second deck
	body2, err := json.Marshal(deckRequest2)
	checkErr(t, err)
	req2, err := http.NewRequest("POST", "/decks/", bytes.NewBuffer(body2))
	checkErr(t, err)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusCreated, w2.Code)

	// Get all decks
	req, err := http.NewRequest("GET", "/decks/", nil)
	checkErr(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []deckEntity.Deck
	err = json.Unmarshal(w.Body.Bytes(), &response)
	checkErr(t, err)
	assert.Len(t, response, 2)
}

func TestDeckHandler_GetByID(t *testing.T) {
	router := setupDeckHandlerWithCardValidation()

	// Create test deck via HTTP
	deckRequest := v1.DeckRequest{Name: "Test Deck", Color: "WUBRG", Format: "commander", Commander: "Atraxa", OwnerID: 1}
	body, err := json.Marshal(deckRequest)
	checkErr(t, err)
	req1, err := http.NewRequest("POST", "/decks/", bytes.NewBuffer(body))
	checkErr(t, err)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Get deck by ID
	req, err := http.NewRequest("GET", "/decks/1", nil)
	checkErr(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response deckEntity.Deck
	err = json.Unmarshal(w.Body.Bytes(), &response)
	checkErr(t, err)
	assert.Equal(t, deckRequest.Name, response.Name)
	assert.Equal(t, "WUBG", response.Color)
	assert.Equal(t, deckRequest.Format, response.Format)
	assert.Equal(t, deckRequest.Commander, response.Commander)
}

func TestDeckHandler_GetByID_NotFound(t *testing.T) {
	router := setupDeckHandler()

	req, err := http.NewRequest("GET", "/decks/999", nil)
	checkErr(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeckHandler_Update(t *testing.T) {
	router := setupDeckHandlerWithCardValidation()

	// Create deck via HTTP
	deckRequest := v1.DeckRequest{Name: "Original Deck", Color: "W", Format: "commander", Commander: "Sram", OwnerID: 1}
	body1, err := json.Marshal(deckRequest)
	checkErr(t, err)
	req1, err := http.NewRequest("POST", "/decks/", bytes.NewBuffer(body1))
	checkErr(t, err)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Update deck
	updatedDeckRequest := v1.DeckRequest{Name: "Updated Deck", Color: "U", Format: "commander", Commander: "Baral", OwnerID: 1}
	body, err := json.Marshal(updatedDeckRequest)
	checkErr(t, err)

	req, err := http.NewRequest("PUT", "/decks/1", bytes.NewBuffer(body))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response deckEntity.Deck
	err = json.Unmarshal(w.Body.Bytes(), &response)
	checkErr(t, err)
	assert.Equal(t, "Updated Deck", response.Name)
	assert.Equal(t, "U", response.Color)
	assert.Equal(t, "commander", response.Format)
	assert.Equal(t, "Baral", response.Commander)
}

func TestDeckHandler_Update_InvalidJSON(t *testing.T) {
	router := setupDeckHandler()

	req, err := http.NewRequest("PUT", "/decks/1", bytes.NewBuffer([]byte("invalid json")))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeckHandler_Delete(t *testing.T) {
	router := setupDeckHandlerWithCardValidation()

	// Create deck via HTTP
	deckRequest := v1.DeckRequest{Name: "Test Deck", Color: "WUBRG", Format: "commander", Commander: "Atraxa", OwnerID: 1}
	body, err := json.Marshal(deckRequest)
	checkErr(t, err)
	req1, err := http.NewRequest("POST", "/decks/", bytes.NewBuffer(body))
	checkErr(t, err)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Verify deck exists via HTTP
	req2, err := http.NewRequest("GET", "/decks/1", nil)
	checkErr(t, err)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Delete deck
	req, err := http.NewRequest("DELETE", "/decks/1", nil)
	checkErr(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify deck is deleted via HTTP
	req3, err := http.NewRequest("GET", "/decks/1", nil)
	checkErr(t, err)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusNotFound, w3.Code)
}
