package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	userRepo "github.com/josofm/liliana/internal/repository/user"
	"github.com/josofm/liliana/pkg/logger"
)

func NewRouter(handler *gin.Engine, l logger.Interface, userRepo userRepo.Repository, deckRepo deckRepo.Repository) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	NewUserHandler(handler, userRepo)
	NewDeckHandler(handler, deckRepo)
}
