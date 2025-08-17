package validator

import (
	"testing"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	userEntity "github.com/josofm/liliana/internal/entity/user"
	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateUser(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		user    userEntity.User
		wantErr bool
	}{
		{
			name: "valid user",
			user: userEntity.User{
				Name:     "John Doe",
				Password: "password123",
				Email:    "john@example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			user: userEntity.User{
				Name:     "John Doe",
				Password: "password123",
				Email:    "invalid-email",
			},
			wantErr: true,
		},
		{
			name: "name too short",
			user: userEntity.User{
				Name:     "J",
				Password: "password123",
				Email:    "john@example.com",
			},
			wantErr: true,
		},
		{
			name: "password too short",
			user: userEntity.User{
				Name:     "John Doe",
				Password: "123",
				Email:    "john@example.com",
			},
			wantErr: false, // Password validation is now handled separately
		},
		{
			name: "missing required fields",
			user: userEntity.User{
				Name:     "",
				Password: "",
				Email:    "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(&tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateDeck(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		deck    deckEntity.Deck
		wantErr bool
	}{
		{
			name: "valid deck",
			deck: deckEntity.Deck{
				Name:      "My Commander Deck",
				Color:     "WUBRG",
				Commander: "Atraxa, Praetors' Voice",
				OwnerID:   1,
			},
			wantErr: false,
		},
		{
			name: "invalid color",
			deck: deckEntity.Deck{
				Name:      "My Commander Deck",
				Color:     "INVALID",
				Commander: "Atraxa, Praetors' Voice",
				OwnerID:   1,
			},
			wantErr: true,
		},
		{
			name: "valid color W",
			deck: deckEntity.Deck{
				Name:      "White Deck",
				Color:     "W",
				Commander: "Sram, Senior Edificer",
				OwnerID:   1,
			},
			wantErr: false,
		},
		{
			name: "invalid owner_id",
			deck: deckEntity.Deck{
				Name:      "My Commander Deck",
				Color:     "WUBRG",
				Commander: "Atraxa, Praetors' Voice",
				OwnerID:   0,
			},
			wantErr: true,
		},
		{
			name: "valid URL",
			deck: deckEntity.Deck{
				Name:       "My Commander Deck",
				Color:      "WUBRG",
				Commander:  "Atraxa, Praetors' Voice",
				OwnerID:    1,
				SourceLink: "https://archidekt.com/decks/123456",
			},
			wantErr: false,
		},
		{
			name: "invalid URL",
			deck: deckEntity.Deck{
				Name:       "My Commander Deck",
				Color:      "WUBRG",
				Commander:  "Atraxa, Praetors' Voice",
				OwnerID:    1,
				SourceLink: "not-a-url",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(&tt.deck)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateAndGetErrors(t *testing.T) {
	v := New()

	user := userEntity.User{
		Name:     "",
		Password: "123",
		Email:    "invalid-email",
	}

	errors := v.ValidateAndGetErrors(&user)

	assert.NotNil(t, errors)
	assert.Contains(t, errors, "name")
	assert.Contains(t, errors, "email")
	// Password validation is now handled separately in handlers
}

func TestValidator_FormatError(t *testing.T) {
	v := New()

	// Test formatError indirectly through ValidateAndGetErrors
	user := userEntity.User{
		Name:     "",
		Password: "123",
		Email:    "invalid-email",
	}

	errors := v.ValidateAndGetErrors(&user)

	// Check that error messages are in English and contain expected keywords
	assert.NotNil(t, errors)
	assert.Contains(t, errors, "name")
	assert.Contains(t, errors, "email")

	// Check specific error messages
	assert.Contains(t, errors["name"], "required")
	assert.Contains(t, errors["email"], "Invalid")
}
