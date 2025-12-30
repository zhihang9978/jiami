package api

import (
	"net/http"
	"strconv"

	"github.com/feiji/feiji-backend/internal/channels"
	"github.com/gin-gonic/gin"
)

type ChannelsHandler struct {
	service *channels.Service
}

func NewChannelsHandler(service *channels.Service) *ChannelsHandler {
	return &ChannelsHandler{service: service}
}

// CreateChannel creates a new channel or supergroup
func (h *ChannelsHandler) CreateChannel(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Title       string `json:"title" binding:"required"`
		About       string `json:"about"`
		Broadcast   bool   `json:"broadcast"`
		Megagroup   bool   `json:"megagroup"`
		ForImport   bool   `json:"for_import"`
		GeoPoint    interface{} `json:"geo_point"`
		Address     string `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.service.CreateChannel(c.Request.Context(), userID, req.Title, req.About, req.Broadcast, req.Megagroup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channel.ToTL())
}

// GetFullChannel retrieves full channel info
func (h *ChannelsHandler) GetFullChannel(c *gin.Context) {
	channelIDStr := c.Query("channel_id")
	if channelIDStr == "" {
		var req struct {
			ChannelID int64 `json:"channel_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		channelIDStr = strconv.FormatInt(req.ChannelID, 10)
	}

	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel_id"})
		return
	}

	fullChannel, err := h.service.GetFullChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fullChannel.ToTL())
}

// EditTitle updates the channel title
func (h *ChannelsHandler) EditTitle(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChannelID int64  `json:"channel_id" binding:"required"`
		Title     string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.service.EditTitle(c.Request.Context(), req.ChannelID, userID, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channel.ToTL())
}

// EditAbout updates the channel about
func (h *ChannelsHandler) EditAbout(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChannelID int64  `json:"channel_id" binding:"required"`
		About     string `json:"about" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.service.EditAbout(c.Request.Context(), req.ChannelID, userID, req.About)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channel.ToTL())
}

// UpdateUsername updates the channel username
func (h *ChannelsHandler) UpdateUsername(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChannelID int64  `json:"channel_id" binding:"required"`
		Username  string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.service.UpdateUsername(c.Request.Context(), req.ChannelID, userID, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channel.ToTL())
}

// JoinChannel joins a public channel
func (h *ChannelsHandler) JoinChannel(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChannelID int64 `json:"channel_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.service.JoinChannel(c.Request.Context(), req.ChannelID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channel.ToTL())
}

// LeaveChannel leaves a channel
func (h *ChannelsHandler) LeaveChannel(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChannelID int64 `json:"channel_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.service.LeaveChannel(c.Request.Context(), req.ChannelID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channel.ToTL())
}

// InviteToChannel invites a user to a channel
func (h *ChannelsHandler) InviteToChannel(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChannelID int64   `json:"channel_id" binding:"required"`
		UserIDs   []int64 `json:"users" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var channel *channels.Channel
	var err error
	for _, inviteUserID := range req.UserIDs {
		channel, err = h.service.InviteToChannel(c.Request.Context(), req.ChannelID, userID, inviteUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, channel.ToTL())
}

// KickFromChannel kicks a user from a channel
func (h *ChannelsHandler) KickFromChannel(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChannelID int64 `json:"channel_id" binding:"required"`
		UserID    int64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.service.KickFromChannel(c.Request.Context(), req.ChannelID, userID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channel.ToTL())
}

// GetParticipants retrieves channel participants
func (h *ChannelsHandler) GetParticipants(c *gin.Context) {
	var req struct {
		ChannelID int64  `json:"channel_id" binding:"required"`
		Filter    string `json:"filter"`
		Offset    int    `json:"offset"`
		Limit     int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	participants, err := h.service.GetParticipants(c.Request.Context(), req.ChannelID, req.Filter, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	participantsTL := make([]map[string]interface{}, len(participants))
	for i, p := range participants {
		participantsTL[i] = p.ToTL()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":            "channels.channelParticipants",
		"participants": participantsTL,
		"count":        len(participantsTL),
		"users":        []interface{}{},
	})
}

// GetChannels retrieves user's channels
func (h *ChannelsHandler) GetChannels(c *gin.Context) {
	userID := c.GetInt64("user_id")

	channelsList, err := h.service.GetUserChannels(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	channelsTL := make([]map[string]interface{}, len(channelsList))
	for i, ch := range channelsList {
		channelsTL[i] = ch.ToTL()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":        "messages.chats",
		"chats":    channelsTL,
		"count":    len(channelsTL),
	})
}

// CheckUsername checks if a username is available
func (h *ChannelsHandler) CheckUsername(c *gin.Context) {
	var req struct {
		ChannelID int64  `json:"channel_id" binding:"required"`
		Username  string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.service.GetChannelByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	available := existing == nil || existing.ID == req.ChannelID

	c.JSON(http.StatusOK, gin.H{
		"_":      "boolTrue",
		"result": available,
	})
}
