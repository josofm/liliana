package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	userEntity "github.com/josofm/liliana/internal/entity/user"
	userRepo "github.com/josofm/liliana/internal/repository/user"
	userService "github.com/josofm/liliana/internal/service/user"
)

type UserHandler struct {
	service *userService.Service
}

func NewUserHandler(r *gin.Engine, repo userRepo.Repository) {
	service := userService.NewService(repo)
	h := &UserHandler{service: service}

	group := r.Group("/users")
	{
		group.POST("/", h.create)
		group.GET("/", h.getAll)
		group.GET("/:id", h.getByID)
		group.PUT("/:id", h.update)
		group.DELETE("/:id", h.delete)
		group.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "funcionando"})
		})

	}
}

func (h *UserHandler) create(c *gin.Context) {
	var input userEntity.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.service.Create(&input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func (h *UserHandler) getAll(c *gin.Context) {
	users, _ := h.service.GetAll()
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) getByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	user, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var input userEntity.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.service.Update(id, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update user"})
		return
	}
	c.JSON(http.StatusOK, input)
}

func (h *UserHandler) delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.service.Delete(id)
	c.Status(http.StatusNoContent)
}
