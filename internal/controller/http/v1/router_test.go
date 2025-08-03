package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	userRepo "github.com/josofm/liliana/internal/repository/user"
	"github.com/josofm/liliana/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	userRepo := userRepo.NewInMemoryRepo()
	deckRepo := deckRepo.NewInMemoryRepo()
	l := logger.New("debug")

	NewRouter(router, l, userRepo, deckRepo)

	return router
}

func TestRouter_HealthCheck(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_UserEndpoints(t *testing.T) {
	router := setupTestRouter()

	// Test user creation
	userData := map[string]interface{}{
		"name":     "Test User",
		"email":    "test@example.com",
		"password": "password123",
	}

	body, _ := json.Marshal(userData)
	req, _ := http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Test get all users
	req, _ = http.NewRequest("GET", "/users/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test get user by ID
	req, _ = http.NewRequest("GET", "/users/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test update user
	updateData := map[string]interface{}{
		"name":     "Updated User",
		"email":    "updated@example.com",
		"password": "newpassword",
	}

	body, _ = json.Marshal(updateData)
	req, _ = http.NewRequest("PUT", "/users/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test delete user
	req, _ = http.NewRequest("DELETE", "/users/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRouter_DeckEndpoints(t *testing.T) {
	router := setupTestRouter()

	// Test deck creation
	deckData := map[string]interface{}{
		"name":        "Test Deck",
		"color":       "WUBRG",
		"commander":   "Atraxa, Praetors' Voice",
		"owner_id":    1,
		"source_link": "https://archidekt.com/decks/123456",
	}

	body, _ := json.Marshal(deckData)
	req, _ := http.NewRequest("POST", "/decks/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Test get all decks
	req, _ = http.NewRequest("GET", "/decks/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test get deck by ID
	req, _ = http.NewRequest("GET", "/decks/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test update deck
	updateData := map[string]interface{}{
		"name":        "Updated Deck",
		"color":       "BR",
		"commander":   "Rakdos, Lord of Riots",
		"owner_id":    2,
		"source_link": "https://archidekt.com/decks/654321",
	}

	body, _ = json.Marshal(updateData)
	req, _ = http.NewRequest("PUT", "/decks/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test delete deck
	req, _ = http.NewRequest("DELETE", "/decks/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRouter_NotFound(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
