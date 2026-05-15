package v1

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	authService "github.com/josofm/liliana/internal/service/auth"
)

// AuthMiddleware verifica se o usuário está autenticado
func AuthMiddleware(authService *authService.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extrair token do header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Verificar formato "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Validar token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Adicionar claims ao contexto para uso posterior
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("claims", claims)

		c.Next()
	}
}

// OptionalAuthMiddleware verifica autenticação mas não falha se não houver token
func OptionalAuthMiddleware(authService *authService.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		token := tokenParts[1]

		// Tentar validar token, mas não falhar se inválido
		if claims, err := authService.ValidateToken(token); err == nil {
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("claims", claims)
		}

		c.Next()
	}
}

// GetUserIDFromContext extrai o ID do usuário do contexto
func GetUserIDFromContext(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	if id, ok := userID.(int64); ok {
		return id, true
	}

	return 0, false
}

// GetUserEmailFromContext extrai o email do usuário do contexto
func GetUserEmailFromContext(c *gin.Context) (string, bool) {
	email, exists := c.Get("user_email")
	if !exists {
		return "", false
	}

	if e, ok := email.(string); ok {
		return e, true
	}

	return "", false
}
