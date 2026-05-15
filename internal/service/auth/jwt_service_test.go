package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTimeProvider para testes
type MockTimeProvider struct {
	currentTime time.Time
}

func NewMockTimeProvider() *MockTimeProvider {
	// Usar uma data futura fixa para garantir que os tokens sejam válidos
	return &MockTimeProvider{
		currentTime: time.Date(2030, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

func (m *MockTimeProvider) Now() time.Time {
	return m.currentTime
}

func (m *MockTimeProvider) Unix() int64 {
	return m.currentTime.Unix()
}

func (m *MockTimeProvider) Advance(duration time.Duration) {
	m.currentTime = m.currentTime.Add(duration)
}

func TestNewJWTService(t *testing.T) {
	config := JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	}

	service := NewJWTService(config)
	assert.NotNil(t, service)
	assert.Equal(t, []byte("test-secret"), service.secretKey)
	assert.Equal(t, 15*time.Minute, service.accessExpiry)
	assert.Equal(t, 24*time.Hour, service.refreshExpiry)
}

func TestGenerateTokenPair(t *testing.T) {
	mockTime := NewMockTimeProvider()
	service := NewJWTServiceWithTimeProvider(JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	}, mockTime)

	userID := int64(123)
	email := "test@example.com"

	tokenPair, err := service.GenerateTokenPair(userID, email)
	require.NoError(t, err)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.True(t, tokenPair.ExpiresAt.After(mockTime.Now()))
}

func TestValidateToken(t *testing.T) {
	mockTime := NewMockTimeProvider()
	service := NewJWTServiceWithTimeProvider(JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	}, mockTime)

	userID := int64(123)
	email := "test@example.com"

	tokenPair, err := service.GenerateTokenPair(userID, email)
	require.NoError(t, err)

	// Validar access token
	claims, err := service.ValidateToken(tokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.True(t, claims.Exp > mockTime.Unix())
	assert.True(t, claims.Iat <= mockTime.Unix())

	// Validar refresh token
	claims, err = service.ValidateToken(tokenPair.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	service := NewJWTService(JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	})

	// Token inválido
	_, err := service.ValidateToken("invalid-token")
	assert.Error(t, err)

	// Token vazio
	_, err = service.ValidateToken("")
	assert.Error(t, err)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	mockTime := NewMockTimeProvider()
	service := NewJWTServiceWithTimeProvider(JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  -1 * time.Hour, // Token já expirado
		RefreshExpiry: 24 * time.Hour,
	}, mockTime)

	userID := int64(123)
	email := "test@example.com"

	tokenPair, err := service.GenerateTokenPair(userID, email)
	require.NoError(t, err)

	// Token expirado deve falhar
	_, err = service.ValidateToken(tokenPair.AccessToken)
	assert.Error(t, err)
}

func TestRefreshToken(t *testing.T) {
	mockTime := NewMockTimeProvider()
	service := NewJWTServiceWithTimeProvider(JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	}, mockTime)

	userID := int64(123)
	email := "test@example.com"

	tokenPair, err := service.GenerateTokenPair(userID, email)
	require.NoError(t, err)

	// Avançar o tempo para garantir timestamps diferentes
	mockTime.Advance(1 * time.Second)

	// Renovar token
	newTokenPair, err := service.RefreshToken(tokenPair.RefreshToken)
	require.NoError(t, err)
	assert.NotNil(t, newTokenPair)
	assert.NotEmpty(t, newTokenPair.AccessToken)
	assert.NotEmpty(t, newTokenPair.RefreshToken)

	// Verificar que os tokens são diferentes (devido ao timestamp diferente)
	assert.NotEqual(t, tokenPair.AccessToken, newTokenPair.AccessToken)
	assert.NotEqual(t, tokenPair.RefreshToken, newTokenPair.RefreshToken)

	// Verificar que os novos tokens são válidos
	claims, err := service.ValidateToken(newTokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	service := NewJWTService(JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	})

	// Refresh token inválido
	_, err := service.RefreshToken("invalid-refresh-token")
	assert.Error(t, err)
}

func TestHashPassword(t *testing.T) {
	service := NewJWTService(JWTConfig{
		SecretKey: "test-secret",
	})

	password := "my-secret-password"
	hash, err := service.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestCheckPassword(t *testing.T) {
	service := NewJWTService(JWTConfig{
		SecretKey: "test-secret",
	})

	password := "my-secret-password"
	hash, err := service.HashPassword(password)
	require.NoError(t, err)

	// Senha correta
	assert.True(t, service.CheckPassword(password, hash))

	// Senha incorreta
	assert.False(t, service.CheckPassword("wrong-password", hash))
	assert.False(t, service.CheckPassword("", hash))
}

func TestGenerateRandomString(t *testing.T) {
	service := NewJWTService(JWTConfig{
		SecretKey: "test-secret",
	})

	// Gerar strings de diferentes tamanhos
	lengths := []int{16, 32, 64}
	for _, length := range lengths {
		randomString, err := service.GenerateRandomString(length)
		require.NoError(t, err)

		// Base64 pode adicionar padding (=) e o tamanho pode variar
		// Para 16 bytes: 16 * 4/3 = ~21.33, arredondado para 22 + possível padding
		// Para 32 bytes: 32 * 4/3 = ~42.67, arredondado para 43 + possível padding
		// Para 64 bytes: 64 * 4/3 = ~85.33, arredondado para 86 + possível padding
		expectedMin := length * 4 / 3
		expectedMax := expectedMin + 3 // +3 para possível padding e arredondamento

		assert.GreaterOrEqual(t, len(randomString), expectedMin)
		assert.LessOrEqual(t, len(randomString), expectedMax)
		assert.NotEmpty(t, randomString)
	}
}
