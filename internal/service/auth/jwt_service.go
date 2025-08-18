package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/josofm/liliana/internal/entity/auth"
	"golang.org/x/crypto/bcrypt"
)

// TimeProvider interface para permitir mock de time nos testes
type TimeProvider interface {
	Now() time.Time
	Unix() int64
}

// DefaultTimeProvider implementação padrão
type DefaultTimeProvider struct{}

func (d *DefaultTimeProvider) Now() time.Time { return time.Now() }
func (d *DefaultTimeProvider) Unix() int64    { return time.Now().Unix() }

// JWTConfig configuração do serviço JWT
type JWTConfig struct {
	SecretKey     string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// JWTService serviço para gerenciar tokens JWT
type JWTService struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	timeProvider  TimeProvider
}

// NewJWTService cria uma nova instância do serviço JWT
func NewJWTService(config JWTConfig) *JWTService {
	return &JWTService{
		secretKey:     []byte(config.SecretKey),
		accessExpiry:  config.AccessExpiry,
		refreshExpiry: config.RefreshExpiry,
		timeProvider:  &DefaultTimeProvider{},
	}
}

// NewJWTServiceWithTimeProvider cria uma nova instância com time provider customizado (para testes)
func NewJWTServiceWithTimeProvider(config JWTConfig, timeProvider TimeProvider) *JWTService {
	return &JWTService{
		secretKey:     []byte(config.SecretKey),
		accessExpiry:  config.AccessExpiry,
		refreshExpiry: config.RefreshExpiry,
		timeProvider:  timeProvider,
	}
}

// GenerateTokenPair gera um par de tokens (access + refresh)
func (s *JWTService) GenerateTokenPair(userID int64, email string) (*auth.TokenPair, error) {
	now := s.timeProvider.Now()
	nowUnix := s.timeProvider.Unix()

	// Access token
	accessClaims := auth.Claims{
		UserID: userID,
		Email:  email,
		Exp:    now.Add(s.accessExpiry).Unix(),
		Iat:    nowUnix,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token
	refreshClaims := auth.Claims{
		UserID: userID,
		Email:  email,
		Exp:    now.Add(s.refreshExpiry).Unix(),
		Iat:    nowUnix,
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &auth.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    now.Add(s.accessExpiry),
	}, nil
}

// ValidateToken valida um token JWT e retorna as claims
func (s *JWTService) ValidateToken(tokenString string) (*auth.Claims, error) {
	// Parse sem validação automática usando jwt.Parse
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extrair claims manualmente
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Converter para nossa estrutura Claims
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid user_id claim")
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid email claim")
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid exp claim")
		}

		iat, ok := claims["iat"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid iat claim")
		}

		// Validar expiração usando o timeProvider
		now := s.timeProvider.Now()
		if int64(exp) <= now.Unix() {
			return nil, fmt.Errorf("token is expired")
		}

		return &auth.Claims{
			UserID: int64(userID),
			Email:  email,
			Exp:    int64(exp),
			Iat:    int64(iat),
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken renova um access token usando o refresh token
func (s *JWTService) RefreshToken(refreshTokenString string) (*auth.TokenPair, error) {
	claims, err := s.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	return s.GenerateTokenPair(claims.UserID, claims.Email)
}

// HashPassword criptografa uma senha
func (s *JWTService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword verifica se uma senha corresponde ao hash
func (s *JWTService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomString gera uma string aleatória para refresh tokens
func (s *JWTService) GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
