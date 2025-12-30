package api

import (
	"net/http"

	"github.com/feiji/feiji-backend/internal/push"
	"github.com/gin-gonic/gin"
)

type PushHandler struct {
	service *push.Service
}

func NewPushHandler(service *push.Service) *PushHandler {
	return &PushHandler{service: service}
}

// RegisterDevice registers a device for push notifications
func (h *PushHandler) RegisterDevice(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Token         string `json:"token" binding:"required"`
		TokenType     int    `json:"token_type" binding:"required"`
		DeviceModel   string `json:"device_model"`
		SystemVersion string `json:"system_version"`
		AppVersion    string `json:"app_version"`
		AppSandbox    bool   `json:"app_sandbox"`
		Secret        []byte `json:"secret"`
		OtherUIDs     []int64 `json:"other_uids"`
		NoMuted       bool   `json:"no_muted"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := h.service.RegisterDevice(c.Request.Context(), userID, req.Token, req.TokenType,
		req.DeviceModel, req.SystemVersion, req.AppVersion, req.AppSandbox, req.Secret, req.NoMuted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":      "boolTrue",
		"result": true,
		"device": device.ToTL(),
	})
}

// UnregisterDevice removes a device
func (h *PushHandler) UnregisterDevice(c *gin.Context) {
	var req struct {
		Token     string `json:"token" binding:"required"`
		TokenType int    `json:"token_type" binding:"required"`
		OtherUIDs []int64 `json:"other_uids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UnregisterDevice(c.Request.Context(), req.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":      "boolTrue",
		"result": true,
	})
}

// GetNotifySettings retrieves notification settings for a peer
func (h *PushHandler) GetNotifySettings(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		PeerID   int64  `json:"peer_id" binding:"required"`
		PeerType string `json:"peer_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service.GetNotificationSettings(c.Request.Context(), userID, req.PeerID, req.PeerType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings.ToTL())
}

// UpdateNotifySettings updates notification settings
func (h *PushHandler) UpdateNotifySettings(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		PeerID            int64  `json:"peer_id" binding:"required"`
		PeerType          string `json:"peer_type" binding:"required"`
		ShowPreviews      *bool  `json:"show_previews"`
		Silent            *bool  `json:"silent"`
		MuteUntil         *int   `json:"mute_until"`
		Sound             string `json:"sound"`
		StoriesMuted      *bool  `json:"stories_muted"`
		StoriesHideSender *bool  `json:"stories_hide_sender"`
		StoriesSound      string `json:"stories_sound"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current settings
	settings, err := h.service.GetNotificationSettings(c.Request.Context(), userID, req.PeerID, req.PeerType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.ShowPreviews != nil {
		settings.ShowPreviews = *req.ShowPreviews
	}
	if req.Silent != nil {
		settings.Silent = *req.Silent
	}
	if req.MuteUntil != nil {
		settings.MuteUntil = *req.MuteUntil
	}
	if req.Sound != "" {
		settings.Sound = req.Sound
	}
	if req.StoriesMuted != nil {
		settings.StoriesMuted = *req.StoriesMuted
	}
	if req.StoriesHideSender != nil {
		settings.StoriesHideSender = *req.StoriesHideSender
	}
	if req.StoriesSound != "" {
		settings.StoriesSound = req.StoriesSound
	}

	if err := h.service.UpdateNotificationSettings(c.Request.Context(), settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":      "boolTrue",
		"result": true,
	})
}

// ResetNotifySettings resets notification settings to default
func (h *PushHandler) ResetNotifySettings(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		PeerID   int64  `json:"peer_id" binding:"required"`
		PeerType string `json:"peer_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ResetNotificationSettings(c.Request.Context(), userID, req.PeerID, req.PeerType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":      "boolTrue",
		"result": true,
	})
}

// GetAllNotifySettings retrieves all notification settings
func (h *PushHandler) GetAllNotifySettings(c *gin.Context) {
	userID := c.GetInt64("user_id")

	settings, err := h.service.GetAllNotificationSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	settingsTL := make([]map[string]interface{}, len(settings))
	for i, s := range settings {
		settingsTL[i] = s.ToTL()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":        "account.notifySettings",
		"settings": settingsTL,
	})
}
