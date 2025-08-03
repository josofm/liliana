//go:build integration

package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	userEntity "github.com/josofm/liliana/internal/entity/user"
	userRepo "github.com/josofm/liliana/internal/repository/user"
	"github.com/stretchr/testify/assert"
)

func setupUserHandler() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	repo := userRepo.NewInMemoryRepo()

	// Use the real handler
	NewUserHandler(router, repo)

	return router
}

func TestUserHandler_Create(t *testing.T) {
	router := setupUserHandler()

	user := userEntity.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response userEntity.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, user.Name, response.Name)
	assert.Equal(t, user.Email, response.Email)
	assert.Equal(t, int64(1), response.ID)
}

func TestUserHandler_Create_InvalidJSON(t *testing.T) {
	router := setupUserHandler()

	req, _ := http.NewRequest("POST", "/users/", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_GetAll(t *testing.T) {
	router := setupUserHandler()

	// Create test users via HTTP
	user1 := userEntity.User{Name: "User 1", Email: "user1@example.com", Password: "pass1"}
	user2 := userEntity.User{Name: "User 2", Email: "user2@example.com", Password: "pass2"}

	// Create first user
	body1, _ := json.Marshal(user1)
	req1, _ := http.NewRequest("POST", "/users/", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Create second user
	body2, _ := json.Marshal(user2)
	req2, _ := http.NewRequest("POST", "/users/", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusCreated, w2.Code)

	// Get all users
	req, _ := http.NewRequest("GET", "/users/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []userEntity.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Len(t, response, 2)
}

func TestUserHandler_GetByID(t *testing.T) {
	router := setupUserHandler()

	// Create test user via HTTP
	user := userEntity.User{Name: "Test User", Email: "test@example.com", Password: "password"}
	body, _ := json.Marshal(user)
	req1, _ := http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Get user by ID
	req, _ := http.NewRequest("GET", "/users/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response userEntity.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, user.Name, response.Name)
	assert.Equal(t, user.Email, response.Email)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	router := setupUserHandler()

	req, _ := http.NewRequest("GET", "/users/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_Update(t *testing.T) {
	router := setupUserHandler()

	// Create user via HTTP
	user := userEntity.User{Name: "Original Name", Email: "original@example.com", Password: "pass"}
	body1, _ := json.Marshal(user)
	req1, _ := http.NewRequest("POST", "/users/", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Update user
	updatedUser := userEntity.User{Name: "Updated Name", Email: "updated@example.com", Password: "newpass"}
	body, _ := json.Marshal(updatedUser)

	req, _ := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response userEntity.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Updated Name", response.Name)
	assert.Equal(t, "updated@example.com", response.Email)
}

func TestUserHandler_Update_InvalidJSON(t *testing.T) {
	router := setupUserHandler()

	req, _ := http.NewRequest("PUT", "/users/1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Delete(t *testing.T) {
	router := setupUserHandler()

	// Create user via HTTP
	user := userEntity.User{Name: "Test User", Email: "test@example.com", Password: "password"}
	body, _ := json.Marshal(user)
	req1, _ := http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Verify user exists via HTTP
	req2, _ := http.NewRequest("GET", "/users/1", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Delete user
	req, _ := http.NewRequest("DELETE", "/users/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify user is deleted via HTTP
	req3, _ := http.NewRequest("GET", "/users/1", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusNotFound, w3.Code)
}
