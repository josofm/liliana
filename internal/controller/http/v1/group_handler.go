package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	repository "github.com/josofm/liliana/internal/repository/group"
	service "github.com/josofm/liliana/internal/service/group"
	"net/http"
	"strconv"
)

type GroupHandler struct{ service *service.Service }

func setupGroupRoutes(rg RouterGroup, s *service.Service) {
	h := &GroupHandler{service: s}
	g := rg.Group("/groups")
	// Also guard direct handler registration; identity always comes from middleware.
	g.Use(func(c *gin.Context) {
		if id, ok := GetUserIDFromContext(c); !ok || id <= 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		}
	})
	g.POST("/", h.create)
	g.GET("/", h.list)
	g.GET("/:id", h.get)
	g.PATCH("/:id", h.update)
	g.GET("/:id/members", h.members)
	g.DELETE("/:id/members/:playerID", h.removeMember)
}
func groupError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "could not process group operation"
	switch {
	case errors.Is(err, repository.ErrNotFound), errors.Is(err, repository.ErrMemberNotFound):
		status = http.StatusNotFound
		message = err.Error()
	case errors.Is(err, service.ErrForbidden):
		status = http.StatusForbidden
		message = err.Error()
	case errors.Is(err, service.ErrInvalidName):
		status = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, repository.ErrOwnerCannotLeave), errors.Is(err, repository.ErrAlreadyMember):
		status = http.StatusConflict
		message = err.Error()
	}
	c.JSON(status, gin.H{"error": message})
}
func groupID(c *gin.Context, key string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + key})
		return 0, false
	}
	return id, true
}
func (h *GroupHandler) create(c *gin.Context) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	actor, _ := GetUserIDFromContext(c)
	g, err := h.service.Create(actor, body.Name, body.Description)
	if err != nil {
		groupError(c, err)
		return
	}
	c.JSON(http.StatusCreated, g)
}
func (h *GroupHandler) list(c *gin.Context) {
	actor, _ := GetUserIDFromContext(c)
	groups, err := h.service.GetByPlayerID(actor)
	if err != nil {
		groupError(c, err)
		return
	}
	c.JSON(http.StatusOK, groups)
}
func (h *GroupHandler) get(c *gin.Context) {
	id, ok := groupID(c, "id")
	if !ok {
		return
	}
	actor, _ := GetUserIDFromContext(c)
	g, err := h.service.GetByID(actor, id)
	if err != nil {
		groupError(c, err)
		return
	}
	c.JSON(http.StatusOK, g)
}
func (h *GroupHandler) update(c *gin.Context) {
	id, ok := groupID(c, "id")
	if !ok {
		return
	}
	actor, _ := GetUserIDFromContext(c)
	var body struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if body.Name == nil && body.Description == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name or description is required"})
		return
	}
	g, err := h.service.Update(actor, id, body.Name, body.Description)
	if err != nil {
		groupError(c, err)
		return
	}
	c.JSON(http.StatusOK, g)
}
func (h *GroupHandler) members(c *gin.Context) {
	id, ok := groupID(c, "id")
	if !ok {
		return
	}
	actor, _ := GetUserIDFromContext(c)
	members, err := h.service.GetMembers(actor, id)
	if err != nil {
		groupError(c, err)
		return
	}
	c.JSON(http.StatusOK, members)
}
func (h *GroupHandler) removeMember(c *gin.Context) {
	id, ok := groupID(c, "id")
	if !ok {
		return
	}
	playerID, ok := groupID(c, "playerID")
	if !ok {
		return
	}
	actor, _ := GetUserIDFromContext(c)
	if err := h.service.RemoveMember(actor, id, playerID); err != nil {
		groupError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
