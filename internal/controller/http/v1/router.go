package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/josofm/liliana/pkg/logger"
)

func NewRouter(handler *gin.Engine, l logger.Interface) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	//TODO: Routers

}
