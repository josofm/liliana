package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// LoginRequest representa a requisição de login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest representa a requisição de registro
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// AuthResponse representa a resposta de autenticação
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	User         User   `json:"user"`
}

// User representa o usuário na resposta de auth (sem senha)
type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Claims representa as claims do JWT e implementa jwt.Claims
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

// GetExpirationTime implementa jwt.Claims
func (c Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.Exp, 0)), nil
}

// GetNotBefore implementa jwt.Claims
func (c Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

// GetIssuedAt implementa jwt.Claims
func (c Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.Iat, 0)), nil
}

// GetIssuer implementa jwt.Claims
func (c Claims) GetIssuer() (string, error) {
	return "", nil
}

// GetSubject implementa jwt.Claims
func (c Claims) GetSubject() (string, error) {
	return "", nil
}

// GetAudience implementa jwt.Claims
func (c Claims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

// TokenPair representa um par de tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}
