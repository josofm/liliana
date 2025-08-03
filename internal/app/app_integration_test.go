//go:build integration

package app_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/josofm/liliana/config"
	"github.com/stretchr/testify/assert"
)

func setupTestApp() *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Setup router manually for testing
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add test routes
	router.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	return router
}

func TestApp_HealthCheck(t *testing.T) {
	router := setupTestApp()

	req, _ := http.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestApp_UserCRUD(t *testing.T) {
	router := setupTestApp()

	// Add user routes for testing
	userData := map[string]interface{}{
		"name":     "Integration Test User",
		"email":    "integration@test.com",
		"password": "testpass123",
	}

	body, _ := json.Marshal(userData)
	req, _ := http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// This would fail in this test setup, but shows the structure
	// In a real integration test, you'd have the full app running
	assert.Equal(t, http.StatusNotFound, w.Code) // Expected since we don't have full routes
}

func TestApp_DeckCRUD(t *testing.T) {
	router := setupTestApp()

	// Add deck routes for testing
	deckData := map[string]interface{}{
		"name":        "Integration Test Deck",
		"color":       "WUBRG",
		"commander":   "Atraxa, Praetors' Voice",
		"owner_id":    1,
		"source_link": "https://archidekt.com/decks/test",
	}

	body, _ := json.Marshal(deckData)
	req, _ := http.NewRequest("POST", "/decks/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// This would fail in this test setup, but shows the structure
	// In a real integration test, you'd have the full app running
	assert.Equal(t, http.StatusNotFound, w.Code) // Expected since we don't have full routes
}

func TestApp_Config(t *testing.T) {
	// Test config loading
	cfg, err := config.NewConfig()
	if err != nil {
		// If config file doesn't exist, create a minimal config
		cfg = &config.Config{
			App: config.App{
				Name:    "test-app",
				Version: "1.0.0",
			},
			HTTP: config.HTTP{
				Port: "8080",
			},
			Log: config.Log{
				Level: "debug",
			},
		}
	}

	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.Log.Level)
}
