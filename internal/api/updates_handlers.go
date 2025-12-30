package api

import (
	"net/http"

	"github.com/feiji/feiji-backend/internal/models"
	"github.com/feiji/feiji-backend/internal/updates"
	"github.com/gin-gonic/gin"
)

// UpdatesHandler handles updates-related requests
type UpdatesHandler struct {
	updatesService *updates.Service
}

func NewUpdatesHandler(updatesService *updates.Service) *UpdatesHandler {
	return &UpdatesHandler{
		updatesService: updatesService,
	}
}

// GetState handles updates.getState
func (h *UpdatesHandler) GetState(c *gin.Context) {
	user := getUpdatesUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	state, err := h.updatesService.GetState(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, state.ToTL())
}

// GetDifference handles updates.getDifference
func (h *UpdatesHandler) GetDifference(c *gin.Context) {
	user := getUpdatesUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Pts  int `json:"pts" binding:"required"`
		Qts  int `json:"qts"`
		Date int `json:"date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.updatesService.GetDifference(c.Request.Context(), user.ID, req.Pts, req.Qts, req.Date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result.ToTL())
}

// GetChannelDifference handles updates.getChannelDifference (stub for now)
func (h *UpdatesHandler) GetChannelDifference(c *gin.Context) {
	// For Phase 1, channels are not implemented
	c.JSON(http.StatusOK, gin.H{
		"_":       "updates.channelDifferenceEmpty",
		"final":   true,
		"pts":     1,
		"timeout": 0,
	})
}

// Helper function
func getUpdatesUser(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}
