//go:build integration

package v1_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/josofm/liliana/config"
	v1 "github.com/josofm/liliana/internal/controller/http/v1"
	authEntity "github.com/josofm/liliana/internal/entity/auth"
	userEntity "github.com/josofm/liliana/internal/entity/user"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	userRepo "github.com/josofm/liliana/internal/repository/user"
	"github.com/josofm/liliana/internal/service/auth"
	"github.com/josofm/liliana/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func setupTestRouterV1() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	userRepo := userRepo.NewInMemoryRepo()
	deckRepo := deckRepo.NewInMemoryRepo()
	l := logger.New("debug")

	// Criar config de teste
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:     "test-secret",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 24 * time.Hour,
		},
	}

	v1.NewRouter(router, l, userRepo, deckRepo, cfg)

	// Criar usuário de teste para autenticação
	testUser := &userEntity.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Criar usuário usando o auth service para garantir que a senha seja hasheada
	jwtService := auth.NewJWTService(auth.JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	})

	authService := auth.NewService(userRepo, jwtService)

	// Registrar usuário usando o auth service
	registerReq := &authEntity.RegisterRequest{
		Name:     testUser.Name,
		Email:    testUser.Email,
		Password: testUser.Password,
	}

	_, err := authService.Register(registerReq)
	if err != nil {
		panic(err) // Falhar o teste se não conseguir criar o usuário
	}

	return router
}

func TestRouter_HealthCheck(t *testing.T) {
	router := setupTestRouterV1()

	req, err := http.NewRequest("GET", "/healthz", nil)
	checkErr(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_UserEndpoints(t *testing.T) {
	router := setupTestRouterV1()

	// Primeiro, fazer login para obter token
	loginData := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}

	loginBody, err := json.Marshal(loginData)
	checkErr(t, err)
	loginReq, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginBody))
	checkErr(t, err)
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	// Debug: ver o que está retornando
	t.Logf("Login response status: %d", loginW.Code)
	t.Logf("Login response body: %s", loginW.Body.String())

	assert.Equal(t, http.StatusOK, loginW.Code)

	// Extrair token da resposta
	var loginResponse map[string]interface{}
	err = json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
	checkErr(t, err)
	accessToken := loginResponse["access_token"].(string)

	// Test user creation
	userData := map[string]interface{}{
		"name":     "Integration Test User",
		"email":    "integration@test.com",
		"password": "testpass123",
	}

	body, err := json.Marshal(userData)
	checkErr(t, err)
	req, err := http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Test get all users
	req, err = http.NewRequest("GET", "/users/", nil)
	checkErr(t, err)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test get user by ID
	req, err = http.NewRequest("GET", "/users/2", nil) // ID 2 para o novo usuário
	checkErr(t, err)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test update user
	updateData := map[string]interface{}{
		"name":     "Updated User",
		"email":    "updated@example.com",
		"password": "newpassword",
	}

	body, err = json.Marshal(updateData)
	checkErr(t, err)
	req, err = http.NewRequest("PUT", "/users/2", bytes.NewBuffer(body))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test delete user
	req, err = http.NewRequest("DELETE", "/users/2", nil)
	checkErr(t, err)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRouter_DeckEndpoints(t *testing.T) {
	router := setupTestRouterV1()

	// Primeiro, fazer login para obter token
	loginData := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}

	loginBody, err := json.Marshal(loginData)
	checkErr(t, err)
	loginReq, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginBody))
	checkErr(t, err)
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	// Debug: ver o que está retornando
	t.Logf("Login response status: %d", loginW.Code)
	t.Logf("Login response body: %s", loginW.Body.String())

	assert.Equal(t, http.StatusOK, loginW.Code)

	// Extrair token da resposta
	var loginResponse map[string]interface{}
	err = json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
	checkErr(t, err)
	accessToken := loginResponse["access_token"].(string)

	// Test deck creation
	deckData := map[string]interface{}{
		"name":        "Test Deck",
		"color":       "WUBRG",
		"format":      "commander",
		"commander":   "Atraxa, Praetors' Voice",
		"owner_id":    1,
		"source_link": "https://archidekt.com/decks/123456",
	}

	body, err := json.Marshal(deckData)
	checkErr(t, err)
	req, err := http.NewRequest("POST", "/decks/", bytes.NewBuffer(body))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Test get all decks
	req, err = http.NewRequest("GET", "/decks/", nil)
	checkErr(t, err)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test get deck by ID
	req, err = http.NewRequest("GET", "/decks/1", nil)
	checkErr(t, err)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test update deck
	updateData := map[string]interface{}{
		"name":        "Updated Deck",
		"color":       "BR",
		"format":      "commander",
		"commander":   "Rakdos, Lord of Riots",
		"owner_id":    2,
		"source_link": "https://archidekt.com/decks/654321",
	}

	body, err = json.Marshal(updateData)
	checkErr(t, err)
	req, err = http.NewRequest("PUT", "/decks/1", bytes.NewBuffer(body))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test delete deck
	req, err = http.NewRequest("DELETE", "/decks/1", nil)
	checkErr(t, err)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRouter_NotFound(t *testing.T) {
	router := setupTestRouterV1()

	req, err := http.NewRequest("GET", "/nonexistent", nil)
	checkErr(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
