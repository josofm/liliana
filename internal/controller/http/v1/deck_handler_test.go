//go:build integration

package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	"github.com/stretchr/testify/assert"
)

func setupDeckHandler() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	repo := deckRepo.NewInMemoryRepo()

	// Use the real handler
	NewDeckHandler(router, repo)

	return router
}

func TestDeckHandler_Create(t *testing.T) {
	router := setupDeckHandler()

	deck := deckEntity.Deck{
		Name:       "Test Deck",
		Color:      "WUBRG",
		Commander:  "Atraxa, Praetors' Voice",
		OwnerID:    1,
		SourceLink: "https://archidekt.com/decks/123456",
	}

	body, err := json.Marshal(deck)
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
	assert.Equal(t, deck.Name, response.Name)
	assert.Equal(t, deck.Color, response.Color)
	assert.Equal(t, deck.Commander, response.Commander)
	assert.Equal(t, int64(1), response.ID)
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
	router := setupDeckHandler()

	// Create test decks via HTTP
	deck1 := deckEntity.Deck{Name: "Deck 1", Color: "WU", Commander: "Azorius", OwnerID: 1}
	deck2 := deckEntity.Deck{Name: "Deck 2", Color: "BR", Commander: "Rakdos", OwnerID: 2}

	// Create first deck
	body1, err := json.Marshal(deck1)
	checkErr(t, err)
	req1, err := http.NewRequest("POST", "/decks/", bytes.NewBuffer(body1))
	checkErr(t, err)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Create second deck
	body2, err := json.Marshal(deck2)
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
	router := setupDeckHandler()

	// Create test deck via HTTP
	deck := deckEntity.Deck{Name: "Test Deck", Color: "WUBRG", Commander: "Atraxa", OwnerID: 1}
	body, err := json.Marshal(deck)
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
	assert.Equal(t, deck.Name, response.Name)
	assert.Equal(t, deck.Color, response.Color)
	assert.Equal(t, deck.Commander, response.Commander)
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
	router := setupDeckHandler()

	// Create deck via HTTP
	deck := deckEntity.Deck{Name: "Original Deck", Color: "WU", Commander: "Azorius", OwnerID: 1}
	body1, err := json.Marshal(deck)
	checkErr(t, err)
	req1, err := http.NewRequest("POST", "/decks/", bytes.NewBuffer(body1))
	checkErr(t, err)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Update deck
	updatedDeck := deckEntity.Deck{Name: "Updated Deck", Color: "BR", Commander: "Rakdos", OwnerID: 2}
	body, err := json.Marshal(updatedDeck)
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
	assert.Equal(t, "BR", response.Color)
	assert.Equal(t, "Rakdos", response.Commander)
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
	router := setupDeckHandler()

	// Create deck via HTTP
	deck := deckEntity.Deck{Name: "Test Deck", Color: "WUBRG", Commander: "Atraxa", OwnerID: 1}
	body, err := json.Marshal(deck)
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
