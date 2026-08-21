package v1

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/josofm/liliana/config"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	userRepo "github.com/josofm/liliana/internal/repository/user"
	"github.com/josofm/liliana/internal/service/auth"
	deckService "github.com/josofm/liliana/internal/service/deck"
	userService "github.com/josofm/liliana/internal/service/user"
	"github.com/josofm/liliana/internal/validator"
	"github.com/josofm/liliana/pkg/logger"
)

// RouterGroup é uma interface que tanto gin.Engine quanto gin.RouterGroup implementam
type RouterGroup interface {
	Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup
	Use(middleware ...gin.HandlerFunc) gin.IRoutes
}

func NewRouter(handler *gin.Engine, l logger.Interface, userRepo userRepo.Repository, deckRepo deckRepo.Repository, cfg *config.Config) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())
	handler.Use(corsMiddleware(cfg.HTTP.CORSAllowedOrigins))

	// Health check
	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Validator
	validator := validator.New()

	// JWT Service
	jwtService := auth.NewJWTService(auth.JWTConfig{
		SecretKey:     cfg.JWT.SecretKey,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
	})

	// Auth Service
	authService := auth.NewService(userRepo, jwtService)

	// Auth Handler
	authHandler := NewAuthHandler(authService, validator)

	// Auth routes (públicas)
	auth := handler.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Rotas protegidas
	protected := handler.Group("/")
	protected.Use(AuthMiddleware(authService))
	{
		// User profile
		protected.GET("/me", authHandler.Me)

		// User management (protegido)
		setupUserRoutes(protected, userRepo)

		// Deck management (protegido)
		setupDeckRoutes(protected, deckRepo)
	}
}

func corsMiddleware(allowedOriginsConfig string) gin.HandlerFunc {
	allowedOrigins := splitCSV(allowedOriginsConfig)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && isAllowedOrigin(origin, allowedOrigins) {
			if contains(allowedOrigins, "*") {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			items = append(items, item)
		}
	}

	return items
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	return contains(allowedOrigins, "*") || contains(allowedOrigins, origin)
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}

	return false
}

// setupUserRoutes configura as rotas de usuário
func setupUserRoutes(rg RouterGroup, userRepo userRepo.Repository) {
	service := userService.NewService(userRepo)
	validator := validator.New()
	h := &UserHandler{service: service, validator: validator}

	group := rg.Group("/users")
	{
		group.POST("/", h.create)
		group.GET("/", h.getAll)
		group.GET("/:id", h.getByID)
		group.PUT("/:id", h.update)
		group.DELETE("/:id", h.delete)
	}
}

// setupDeckRoutes configura as rotas de deck
func setupDeckRoutes(rg RouterGroup, deckRepo deckRepo.Repository) {
	service := deckService.NewService(deckRepo)
	validator := validator.New()
	h := &DeckHandler{service: service, validator: validator}

	group := rg.Group("/decks")
	{
		group.GET("/commanders", h.searchCommanders)
		group.POST("/", h.create)
		group.GET("/", h.getAll)
		group.GET("/:id", h.getByID)
		group.PUT("/:id", h.update)
		group.POST("/:id/cards", h.addCards)
		group.PATCH("/:id/cards", h.patchCards)
		group.DELETE("/:id", h.delete)
	}
}
