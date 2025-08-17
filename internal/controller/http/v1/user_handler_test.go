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
	NewUserHandler(router, repo)
	return router
}

func TestUserHandler_Create(t *testing.T) {
	router := setupUserHandler()

	userRequest := UserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	body, err := json.Marshal(userRequest)
	checkErr(t, err)

	req, err := http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response userEntity.User
	err = json.Unmarshal(w.Body.Bytes(), &response)
	checkErr(t, err)
	assert.Equal(t, userRequest.Name, response.Name)
	assert.Equal(t, userRequest.Email, response.Email)
	assert.Equal(t, int64(1), response.ID)
}

func TestUserHandler_Create_InvalidJSON(t *testing.T) {
	router := setupUserHandler()

	req, err := http.NewRequest("POST", "/users/", bytes.NewBuffer([]byte("invalid json")))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_GetAll(t *testing.T) {
	router := setupUserHandler()

	// Create test users via HTTP
	userRequest1 := UserRequest{Name: "User 1", Email: "user1@example.com", Password: "pass123"}
	userRequest2 := UserRequest{Name: "User 2", Email: "user2@example.com", Password: "pass456"}

	// Create first user
	body1, err := json.Marshal(userRequest1)
	checkErr(t, err)
	req1, err := http.NewRequest("POST", "/users/", bytes.NewBuffer(body1))
	checkErr(t, err)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Create second user
	body2, err := json.Marshal(userRequest2)
	checkErr(t, err)
	req2, err := http.NewRequest("POST", "/users/", bytes.NewBuffer(body2))
	checkErr(t, err)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusCreated, w2.Code)

	// Get all users
	req, err := http.NewRequest("GET", "/users/", nil)
	checkErr(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []userEntity.User
	err = json.Unmarshal(w.Body.Bytes(), &response)
	checkErr(t, err)
	assert.Len(t, response, 2)
}

func TestUserHandler_GetByID(t *testing.T) {
	router := setupUserHandler()

	// Create test user via HTTP
	userRequest := UserRequest{Name: "Test User", Email: "test@example.com", Password: "password123"}
	body, err := json.Marshal(userRequest)
	checkErr(t, err)
	req1, err := http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
	checkErr(t, err)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Get user by ID
	req, err := http.NewRequest("GET", "/users/1", nil)
	checkErr(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response userEntity.User
	err = json.Unmarshal(w.Body.Bytes(), &response)
	checkErr(t, err)
	assert.Equal(t, userRequest.Name, response.Name)
	assert.Equal(t, userRequest.Email, response.Email)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	router := setupUserHandler()

	req, err := http.NewRequest("GET", "/users/999", nil)
	checkErr(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_Update(t *testing.T) {
	router := setupUserHandler()

	// Create user via HTTP
	userRequest := UserRequest{Name: "Original Name", Email: "original@example.com", Password: "password123"}
	body1, err := json.Marshal(userRequest)
	checkErr(t, err)
	req1, err := http.NewRequest("POST", "/users/", bytes.NewBuffer(body1))
	checkErr(t, err)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Update user
	updatedUserRequest := UserRequest{Name: "Updated Name", Email: "updated@example.com", Password: "newpass"}
	body, err := json.Marshal(updatedUserRequest)
	checkErr(t, err)

	req, err := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(body))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response userEntity.User
	err = json.Unmarshal(w.Body.Bytes(), &response)
	checkErr(t, err)
	assert.Equal(t, "Updated Name", response.Name)
	assert.Equal(t, "updated@example.com", response.Email)
}

func TestUserHandler_Update_InvalidJSON(t *testing.T) {
	router := setupUserHandler()

	req, err := http.NewRequest("PUT", "/users/1", bytes.NewBuffer([]byte("invalid json")))
	checkErr(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Delete(t *testing.T) {
	router := setupUserHandler()

	// Create user via HTTP
	userRequest := UserRequest{Name: "Test User", Email: "test@example.com", Password: "password123"}
	body, err := json.Marshal(userRequest)
	checkErr(t, err)
	req1, err := http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
	checkErr(t, err)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Verify user exists via HTTP
	req2, err := http.NewRequest("GET", "/users/1", nil)
	checkErr(t, err)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Delete user
	req, err := http.NewRequest("DELETE", "/users/1", nil)
	checkErr(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify user is deleted via HTTP
	req3, err := http.NewRequest("GET", "/users/1", nil)
	checkErr(t, err)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusNotFound, w3.Code)
}
