package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	userEntity "github.com/josofm/liliana/internal/entity/user"
	userRepo "github.com/josofm/liliana/internal/repository/user"
	userService "github.com/josofm/liliana/internal/service/user"
	"github.com/josofm/liliana/internal/validator"
)

// UserRequest represents the incoming user data for validation
type UserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Password string `json:"password" validate:"required,min=6"`
	Email    string `json:"email" validate:"required,email"`
}

type UserHandler struct {
	service   *userService.Service
	validator *validator.Validator
}

func NewUserHandler(r *gin.Engine, repo userRepo.Repository) {
	service := userService.NewService(repo)
	validator := validator.New()
	h := &UserHandler{service: service, validator: validator}

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
	var request UserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if validationErrors := h.validator.ValidateAndGetErrors(&request); validationErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}

	// Convert to entity
	user := userEntity.User{
		Name:     request.Name,
		Password: request.Password,
		Email:    request.Email,
	}

	err := h.service.Create(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}
	c.JSON(http.StatusCreated, user)
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

	var request UserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if validationErrors := h.validator.ValidateAndGetErrors(&request); validationErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}

	// Convert to entity
	user := userEntity.User{
		Name:     request.Name,
		Password: request.Password,
		Email:    request.Email,
	}

	err := h.service.Update(id, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update user"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.service.Delete(id)
	c.Status(http.StatusNoContent)
}
