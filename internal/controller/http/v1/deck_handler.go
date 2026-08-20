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
	Name      string `json:"name"`
	Color     string `json:"color"`
	Format    string `json:"format"`
	Commander string `json:"commander"`
	// OwnerID is kept only for source compatibility with internal callers.
	// JSON clients cannot set it; ownership always comes from the authenticated JWT.
	OwnerID    int64  `json:"-"`
	SourceLink string `json:"source_link" validate:"omitempty,url"`
	Cards      string `json:"cards"`
}

type DeckCardsRequest struct {
	Cards string `json:"cards" validate:"required"`
}

type DeckHandler struct {
	service   *deckService.Service
	validator *validator.Validator
}

func NewDeckHandler(r *gin.Engine, repo deckRepo.Repository) {
	service := deckService.NewService(repo)
	NewDeckHandlerWithService(r, service)
}

func NewDeckHandlerWithService(r *gin.Engine, service *deckService.Service) {
	validator := validator.New()
	h := &DeckHandler{service: service, validator: validator}

	group := r.Group("/decks")
	{
		group.GET("/commanders", h.searchCommanders)
		group.POST("/", h.create)
		group.GET("/", h.getAll)
		group.GET("/:id", h.getByID)
		group.PUT("/:id", h.update)
		group.POST("/:id/cards", h.addCards)
		group.DELETE("/:id", h.delete)
	}
}

func (h *DeckHandler) searchCommanders(c *gin.Context) {
	query := c.Query("q")
	if len([]rune(query)) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query must contain at least 2 characters"})
		return
	}
	commanders, err := h.service.SearchCommanders(query)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not search commanders"})
		return
	}
	c.JSON(http.StatusOK, commanders)
}

func (h *DeckHandler) create(c *gin.Context) {
	var request DeckRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if validationErrors := h.validator.ValidateAndGetErrors(&request); validationErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}
	ownerID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Convert to entity
	deck := deckEntity.Deck{
		Name:       request.Name,
		Color:      request.Color,
		Format:     request.Format,
		Commander:  request.Commander,
		OwnerID:    ownerID,
		SourceLink: request.SourceLink,
	}
	if request.Cards != "" {
		cards, err := deckService.ParseCardList(request.Cards)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		deck.Cards = cards
	}
	if err := h.service.Prepare(&deck); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	if validationErrors := h.validator.ValidateAndGetErrors(&deck); validationErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
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
	ownerID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var request DeckRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if validationErrors := h.validator.ValidateAndGetErrors(&request); validationErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}

	// Convert to entity
	deck := deckEntity.Deck{
		Name:       request.Name,
		Color:      request.Color,
		Format:     request.Format,
		Commander:  request.Commander,
		OwnerID:    ownerID,
		SourceLink: request.SourceLink,
	}
	if err := h.service.Prepare(&deck); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	if validationErrors := h.validator.ValidateAndGetErrors(&deck); validationErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
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

func (h *DeckHandler) addCards(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deck id"})
		return
	}
	var request DeckCardsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if validationErrors := h.validator.ValidateAndGetErrors(&request); validationErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
		return
	}
	cards, err := deckService.ParseCardList(request.Cards)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	d, err := h.service.AddCards(id, cards)
	if err != nil {
		if err.Error() == "deck not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "deck not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, d)
}
