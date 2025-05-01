package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	userRepo "github.com/josofm/liliana/internal/repository/user"
	"github.com/josofm/liliana/pkg/logger"
)

func NewRouter(handler *gin.Engine, l logger.Interface, userRepo userRepo.Repository) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	NewUserHandler(handler, userRepo)

}
