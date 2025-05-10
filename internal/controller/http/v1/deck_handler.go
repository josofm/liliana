package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	deckService "github.com/josofm/liliana/internal/service/deck"
)

type deckHandler struct {
	service *deckService.Service
}

func NewDeckHandler(r *gin.Engine, repo deckRepo.Repository) {
	service := deckService.NewService(repo)
	h := &deckHandler{service: service}

	group := r.Group("/decks")
	{
		group.POST("/", h.create)
		group.GET("/", h.getAll)
		group.GET("/:id", h.getByID)
		group.DELETE("/:id", h.delete)
	}
}

func (h *deckHandler) create(c *gin.Context) {
	var input deckEntity.Deck
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.service.Create(&input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create deck"})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func (h *deckHandler) getAll(c *gin.Context) {
	decks, _ := h.service.GetAll()
	c.JSON(http.StatusOK, decks)
}

func (h *deckHandler) getByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	deck, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, deck)
}

func (h *deckHandler) delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.service.Delete(id)
	c.Status(http.StatusNoContent)
}
