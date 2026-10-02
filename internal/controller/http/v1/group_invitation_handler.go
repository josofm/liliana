package v1

import (
	"github.com/gin-gonic/gin"
	entity "github.com/josofm/liliana/internal/entity/group"
	"net/http"
)

func (h *GroupHandler) createInvitation(c *gin.Context) {
	id, ok := groupID(c, "id")
	if !ok {
		return
	}
	actor, _ := GetUserIDFromContext(c)
	var body struct {
		InvitedPlayerID int64 `json:"invited_player_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	i, err := h.service.CreateInvitation(actor, id, body.InvitedPlayerID)
	if err != nil {
		groupError(c, err)
		return
	}
	c.JSON(http.StatusCreated, i)
}
func (h *GroupHandler) groupInvitations(c *gin.Context) {
	id, ok := groupID(c, "id")
	if !ok {
		return
	}
	actor, _ := GetUserIDFromContext(c)
	items, err := h.service.GetGroupInvitations(actor, id)
	if err != nil {
		groupError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}
func (h *GroupHandler) playerInvitations(c *gin.Context) {
	actor, _ := GetUserIDFromContext(c)
	items, err := h.service.GetPlayerInvitations(actor)
	if err != nil {
		groupError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}
func (h *GroupHandler) respondInvitation(c *gin.Context, status entity.InvitationStatus) {
	id, ok := groupID(c, "id")
	if !ok {
		return
	}
	actor, _ := GetUserIDFromContext(c)
	i, err := h.service.RespondInvitation(actor, id, status)
	if err != nil {
		groupError(c, err)
		return
	}
	if status == entity.InvitationCancelled {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, i)
}
func (h *GroupHandler) acceptInvitation(c *gin.Context) {
	h.respondInvitation(c, entity.InvitationAccepted)
}
func (h *GroupHandler) declineInvitation(c *gin.Context) {
	h.respondInvitation(c, entity.InvitationDeclined)
}
func (h *GroupHandler) cancelInvitation(c *gin.Context) {
	h.respondInvitation(c, entity.InvitationCancelled)
}
