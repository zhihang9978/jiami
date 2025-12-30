package api

import (
	"encoding/json"
	"net/http"

	"github.com/feiji/feiji-backend/internal/calls"
	"github.com/gin-gonic/gin"
)

type CallsHandler struct {
	service *calls.Service
}

func NewCallsHandler(service *calls.Service) *CallsHandler {
	return &CallsHandler{service: service}
}

// RequestCall initiates a new call
func (h *CallsHandler) RequestCall(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		UserID   int64           `json:"user_id" binding:"required"`
		Video    bool            `json:"video"`
		Protocol json.RawMessage `json:"protocol"`
		GAHash   []byte          `json:"g_a_hash"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	call, err := h.service.RequestCall(c.Request.Context(), userID, req.UserID, req.Video, req.Protocol, req.GAHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":           "phone.phoneCall",
		"phone_call":  call.ToTL(),
		"users":       []interface{}{},
	})
}

// AcceptCall accepts an incoming call
func (h *CallsHandler) AcceptCall(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Peer struct {
			ID         int64 `json:"id"`
			AccessHash int64 `json:"access_hash"`
		} `json:"peer"`
		GB             []byte          `json:"g_b"`
		Protocol       json.RawMessage `json:"protocol"`
		KeyFingerprint int64           `json:"key_fingerprint"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	call, err := h.service.AcceptCall(c.Request.Context(), req.Peer.ID, userID, req.GB, req.KeyFingerprint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":           "phone.phoneCall",
		"phone_call":  call.ToTL(),
		"users":       []interface{}{},
	})
}

// DiscardCall ends or declines a call
func (h *CallsHandler) DiscardCall(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Peer struct {
			ID         int64 `json:"id"`
			AccessHash int64 `json:"access_hash"`
		} `json:"peer"`
		Duration       int    `json:"duration"`
		Reason         string `json:"reason"`
		ConnectionID   int64  `json:"connection_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	call, err := h.service.DiscardCall(c.Request.Context(), req.Peer.ID, userID, req.Reason, req.Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":           "updates",
		"updates":     []interface{}{call.ToTL()},
		"users":       []interface{}{},
		"chats":       []interface{}{},
		"date":        call.Date,
		"seq":         0,
	})
}

// ConfirmCall confirms call parameters
func (h *CallsHandler) ConfirmCall(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Peer struct {
			ID         int64 `json:"id"`
			AccessHash int64 `json:"access_hash"`
		} `json:"peer"`
		GA             []byte          `json:"g_a"`
		KeyFingerprint int64           `json:"key_fingerprint"`
		Protocol       json.RawMessage `json:"protocol"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	call, err := h.service.ConfirmCall(c.Request.Context(), req.Peer.ID, userID, req.KeyFingerprint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":           "phone.phoneCall",
		"phone_call":  call.ToTL(),
		"users":       []interface{}{},
	})
}

// ReceivedCall marks a call as received
func (h *CallsHandler) ReceivedCall(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Peer struct {
			ID         int64 `json:"id"`
			AccessHash int64 `json:"access_hash"`
		} `json:"peer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.service.ReceivedCall(c.Request.Context(), req.Peer.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":      "boolTrue",
		"result": true,
	})
}

// SetCallRating sets the rating for a call
func (h *CallsHandler) SetCallRating(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Peer struct {
			ID         int64 `json:"id"`
			AccessHash int64 `json:"access_hash"`
		} `json:"peer"`
		Rating  int    `json:"rating" binding:"required"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SetCallRating(c.Request.Context(), req.Peer.ID, userID, req.Rating, req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":      "boolTrue",
		"result": true,
	})
}

// SaveCallDebug saves debug information for a call
func (h *CallsHandler) SaveCallDebug(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Peer struct {
			ID         int64 `json:"id"`
			AccessHash int64 `json:"access_hash"`
		} `json:"peer"`
		Debug json.RawMessage `json:"debug"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SaveCallDebug(c.Request.Context(), req.Peer.ID, userID, req.Debug); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":      "boolTrue",
		"result": true,
	})
}

// GetCallConfig retrieves call configuration
func (h *CallsHandler) GetCallConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"_":                          "phone.phoneCallConfig",
		"default_p2p_contacts":       true,
		"preload_prefix_kb":          1024,
		"me_url_prefix":              "https://t.me/",
		"suggested_lang_code":        "en",
		"lang_pack_version":          0,
		"base_lang_pack_version":     0,
		"call_receive_timeout_ms":    20000,
		"call_ring_timeout_ms":       90000,
		"call_connect_timeout_ms":    30000,
		"call_packet_timeout_ms":     10000,
	})
}
