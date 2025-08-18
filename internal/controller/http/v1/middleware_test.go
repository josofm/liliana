package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/josofm/liliana/internal/service/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouterMiddleware() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	router := setupTestRouterMiddleware()

	// Criar JWT service para testes
	jwtService := auth.NewJWTService(auth.JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	})

	// Criar auth service real para testes
	authService := auth.NewService(nil, jwtService) // nil repo para este teste

	// Adicionar middleware
	router.Use(AuthMiddleware(authService))

	// Rota de teste
	router.GET("/test", func(c *gin.Context) {
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	// Gerar token válido
	tokenPair, err := jwtService.GenerateTokenPair(123, "test@example.com")
	require.NoError(t, err)

	// Fazer request com token válido
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "123")
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	router := setupTestRouterMiddleware()

	jwtService := auth.NewJWTService(auth.JWTConfig{
		SecretKey: "test-secret",
	})
	authService := auth.NewService(nil, jwtService)

	router.Use(AuthMiddleware(authService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authorization header required")
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	router := setupTestRouterMiddleware()

	jwtService := auth.NewJWTService(auth.JWTConfig{
		SecretKey: "test-secret",
	})
	authService := auth.NewService(nil, jwtService)

	router.Use(AuthMiddleware(authService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Header sem "Bearer"
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidToken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid authorization header format")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := setupTestRouterMiddleware()

	jwtService := auth.NewJWTService(auth.JWTConfig{
		SecretKey: "test-secret",
	})
	authService := auth.NewService(nil, jwtService)

	router.Use(AuthMiddleware(authService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Token inválido
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestGetUserIDFromContext(t *testing.T) {
	router := setupTestRouterMiddleware()

	jwtService := auth.NewJWTService(auth.JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	})
	authService := auth.NewService(nil, jwtService)

	router.Use(AuthMiddleware(authService))
	router.GET("/test", func(c *gin.Context) {
		userID, exists := GetUserIDFromContext(c)
		email, emailExists := GetUserEmailFromContext(c)

		c.JSON(http.StatusOK, gin.H{
			"user_id":      userID,
			"exists":       exists,
			"email":        email,
			"email_exists": emailExists,
		})
	})

	tokenPair, err := jwtService.GenerateTokenPair(123, "test@example.com")
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "123")
	assert.Contains(t, w.Body.String(), "test@example.com")
}
