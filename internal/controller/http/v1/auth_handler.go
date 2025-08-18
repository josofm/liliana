package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/josofm/liliana/internal/entity/auth"
	authService "github.com/josofm/liliana/internal/service/auth"
	"github.com/josofm/liliana/internal/validator"
)

type AuthHandler struct {
	service   *authService.Service
	validator *validator.Validator
}

func NewAuthHandler(service *authService.Service, validator *validator.Validator) *AuthHandler {
	return &AuthHandler{
		service:   service,
		validator: validator,
	}
}

// Register registra um novo usuário
func (h *AuthHandler) Register(c *gin.Context) {
	var request auth.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validar request
	if validationErrors := h.validator.ValidateAndGetErrors(&request); validationErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}

	// Registrar usuário
	response, err := h.service.Register(&request)
	if err != nil {
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login autentica um usuário existente
func (h *AuthHandler) Login(c *gin.Context) {
	var request auth.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validar request
	if validationErrors := h.validator.ValidateAndGetErrors(&request); validationErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}

	// Fazer login
	response, err := h.service.Login(&request)
	if err != nil {
		if err.Error() == "invalid credentials" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken renova um access token usando o refresh token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validar request
	if validationErrors := h.validator.ValidateAndGetErrors(&request); validationErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}

	// Renovar token
	response, err := h.service.RefreshToken(request.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Me retorna informações do usuário autenticado
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Buscar usuário
	user, err := h.service.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	})
}
