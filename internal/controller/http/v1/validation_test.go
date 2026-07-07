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
	userEntity "github.com/josofm/liliana/internal/entity/user"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	userRepo "github.com/josofm/liliana/internal/repository/user"

	"github.com/stretchr/testify/assert"
)

func setupValidationTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	userRepository := userRepo.NewInMemoryRepo()
	deckRepository := deckRepo.NewInMemoryRepo()
	v1.NewUserHandler(router, userRepository)
	v1.NewDeckHandler(router, deckRepository)
	return router
}

func TestUserHandler_Validation(t *testing.T) {
	router := setupValidationTest()

	tests := []struct {
		name           string
		userRequest    v1.UserRequest
		expectedStatus int
		shouldHaveID   bool
	}{
		{
			name: "valid_user",
			userRequest: v1.UserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusCreated,
			shouldHaveID:   true,
		},
		{
			name: "invalid_email",
			userRequest: v1.UserRequest{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveID:   false,
		},
		{
			name: "name_too_short",
			userRequest: v1.UserRequest{
				Name:     "J",
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveID:   false,
		},
		{
			name: "password_too_short",
			userRequest: v1.UserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "123",
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveID:   false,
		},
		{
			name: "missing_required_fields",
			userRequest: v1.UserRequest{
				Name:     "",
				Email:    "",
				Password: "",
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveID:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.userRequest)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/users/", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.shouldHaveID {
				var response userEntity.User
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotZero(t, response.ID)
				assert.Equal(t, tt.userRequest.Name, response.Name)
				assert.Equal(t, tt.userRequest.Email, response.Email)
			}
		})
	}
}

func TestDeckHandler_Validation(t *testing.T) {
	router := setupValidationTest()

	tests := []struct {
		name           string
		deckRequest    v1.DeckRequest
		expectedStatus int
		shouldHaveID   bool
	}{
		{
			name: "valid_deck",
			deckRequest: v1.DeckRequest{
				Name:      "My Commander Deck",
				Color:     "WUBRG",
				Format:    "commander",
				Commander: "Atraxa, Praetors' Voice",
				OwnerID:   1,
			},
			expectedStatus: http.StatusCreated,
			shouldHaveID:   true,
		},
		{
			name: "invalid_color",
			deckRequest: v1.DeckRequest{
				Name:      "My Commander Deck",
				Color:     "INVALID",
				Format:    "commander",
				Commander: "Atraxa, Praetors' Voice",
				OwnerID:   1,
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveID:   false,
		},
		{
			name: "valid_color_W",
			deckRequest: v1.DeckRequest{
				Name:      "White Deck",
				Color:     "W",
				Format:    "commander",
				Commander: "Sram, Senior Edificer",
				OwnerID:   1,
			},
			expectedStatus: http.StatusCreated,
			shouldHaveID:   true,
		},
		{
			name: "invalid_owner_id",
			deckRequest: v1.DeckRequest{
				Name:      "My Commander Deck",
				Color:     "WUBRG",
				Format:    "commander",
				Commander: "Atraxa, Praetors' Voice",
				OwnerID:   0,
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveID:   false,
		},
		{
			name: "valid_URL",
			deckRequest: v1.DeckRequest{
				Name:       "My Commander Deck",
				Color:      "WUBRG",
				Format:     "commander",
				Commander:  "Atraxa, Praetors' Voice",
				OwnerID:    1,
				SourceLink: "https://archidekt.com/decks/123456",
			},
			expectedStatus: http.StatusCreated,
			shouldHaveID:   true,
		},
		{
			name: "invalid_URL",
			deckRequest: v1.DeckRequest{
				Name:       "My Commander Deck",
				Color:      "WUBRG",
				Format:     "commander",
				Commander:  "Atraxa, Praetors' Voice",
				OwnerID:    1,
				SourceLink: "not-a-url",
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveID:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.deckRequest)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/decks/", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.shouldHaveID {
				var response deckEntity.Deck
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotZero(t, response.ID)
				assert.Equal(t, tt.deckRequest.Name, response.Name)
				assert.Equal(t, tt.deckRequest.Color, response.Color)
				assert.Equal(t, tt.deckRequest.Format, response.Format)
				assert.Equal(t, tt.deckRequest.Commander, response.Commander)
			}
		})
	}
}
