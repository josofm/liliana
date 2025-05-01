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
	"github.com/josofm/liliana/internal/repository/user"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Repositório em memória
	repo := user.NewInMemoryRepo()
	v1.NewUserHandler(r, repo)

	return r
}

func TestCreateUser(t *testing.T) {
	router := setupRouter()

	input := map[string]string{
		"name":     "Liliana",
		"email":    "lili@mtg.com",
		"password": "necromancer",
	}
	body, _ := json.Marshal(input)

	req, _ := http.NewRequest(http.MethodPost, "/users/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
	assert.Contains(t, resp.Body.String(), `"name":"Liliana"`)
}

func TestGetUserByID(t *testing.T) {
	router := setupRouter()

	body := []byte(`{"name":"Liliana","email":"lili@mtg.com","password":"necromancer"}`)
	req, _ := http.NewRequest(http.MethodPost, "/users/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	req2, _ := http.NewRequest(http.MethodGet, "/users/1", nil)
	resp2 := httptest.NewRecorder()
	router.ServeHTTP(resp2, req2)

	assert.Equal(t, http.StatusOK, resp2.Code)
	assert.Contains(t, resp2.Body.String(), `"email":"lili@mtg.com"`)
}

func TestCreateUserGetBadRequest(t *testing.T) {
	router := setupRouter()

	body := []byte(`{name: Liliana}`)

	req, _ := http.NewRequest(http.MethodPost, "/users/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), "error")
}

func TestGetUserByIDNotFound(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest(http.MethodGet, "/users/999", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.Contains(t, resp.Body.String(), "not found")
}

func TestDeleteUser(t *testing.T) {
	router := setupRouter()

	body := []byte(`{"name":"Liliana","email":"lili@mtg.com","password":"necromancer"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/users/", bytes.NewBuffer(body))
	reqCreate.Header.Set("Content-Type", "application/json")
	respCreate := httptest.NewRecorder()
	router.ServeHTTP(respCreate, reqCreate)
	assert.Equal(t, http.StatusCreated, respCreate.Code)

	reqDelete, _ := http.NewRequest(http.MethodDelete, "/users/1", nil)
	respDelete := httptest.NewRecorder()
	router.ServeHTTP(respDelete, reqDelete)

	assert.Equal(t, http.StatusNoContent, respDelete.Code)

	reqGet, _ := http.NewRequest(http.MethodGet, "/users/1", nil)
	respGet := httptest.NewRecorder()
	router.ServeHTTP(respGet, reqGet)

	assert.Equal(t, http.StatusNotFound, respGet.Code)
}
