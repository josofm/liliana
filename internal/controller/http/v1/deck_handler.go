package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	deckService "github.com/josofm/liliana/internal/service/deck"
)

type DeckHandler struct {
	service *deckService.Service
}

func NewDeckHandler(r *gin.Engine, repo deckRepo.Repository) {
	service := deckService.NewService(repo)
	h := &DeckHandler{service: service}

	group := r.Group("/decks")
	{
		group.POST("/", h.create)
		group.GET("/", h.getAll)
		group.GET("/:id", h.getByID)
		group.PUT("/:id", h.update)
		group.DELETE("/:id", h.delete)
	}
}

func (h *DeckHandler) create(c *gin.Context) {
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

func (h *DeckHandler) getAll(c *gin.Context) {
	decks, _ := h.service.GetAll()
	c.JSON(http.StatusOK, decks)
}

func (h *DeckHandler) getByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	deck, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, deck)
}

func (h *DeckHandler) update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var input deckEntity.Deck
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.service.Update(id, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update deck"})
		return
	}
	c.JSON(http.StatusOK, input)
}

func (h *DeckHandler) delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.service.Delete(id)
	c.Status(http.StatusNoContent)
}
