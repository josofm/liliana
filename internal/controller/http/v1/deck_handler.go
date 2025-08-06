package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	deckEntity "github.com/josofm/liliana/internal/entity/deck"
	deckRepo "github.com/josofm/liliana/internal/repository/deck"
	deckService "github.com/josofm/liliana/internal/service/deck"
	"github.com/josofm/liliana/internal/validator"
)

// DeckRequest represents the incoming deck data for validation
type DeckRequest struct {
	Name       string `json:"name" validate:"required,min=1,max=100"`
	Color      string `json:"color" validate:"required,oneof=W U B R G WU WB WR WG UB UR UG BR BG RG WUB WUR WUG WBR WBG WRG UBR UBG URG BRG WUBR WUBG WURG WBRG UBRG WUBRG"`
	Commander  string `json:"commander" validate:"required,min=1,max=100"`
	OwnerID    int64  `json:"owner_id" validate:"required,gt=0"`
	SourceLink string `json:"source_link" validate:"omitempty,url"`
}

type DeckHandler struct {
	service   *deckService.Service
	validator *validator.Validator
}

func NewDeckHandler(r *gin.Engine, repo deckRepo.Repository) {
	service := deckService.NewService(repo)
	validator := validator.New()
	h := &DeckHandler{service: service, validator: validator}

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
	var request DeckRequest
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
	deck := deckEntity.Deck{
		Name:       request.Name,
		Color:      request.Color,
		Commander:  request.Commander,
		OwnerID:    request.OwnerID,
		SourceLink: request.SourceLink,
	}

	err := h.service.Create(&deck)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create deck"})
		return
	}
	c.JSON(http.StatusCreated, deck)
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

	var request DeckRequest
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
	deck := deckEntity.Deck{
		Name:       request.Name,
		Color:      request.Color,
		Commander:  request.Commander,
		OwnerID:    request.OwnerID,
		SourceLink: request.SourceLink,
	}

	err := h.service.Update(id, &deck)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update deck"})
		return
	}
	c.JSON(http.StatusOK, deck)
}

func (h *DeckHandler) delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.service.Delete(id)
	c.Status(http.StatusNoContent)
}
