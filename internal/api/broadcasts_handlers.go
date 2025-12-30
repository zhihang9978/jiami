package api

import (
	"fmt"
	"net/http"

	"github.com/feiji/feiji-backend/internal/broadcasts"
	"github.com/gin-gonic/gin"
)

type BroadcastsHandler struct {
	service *broadcasts.Service
}

func NewBroadcastsHandler(service *broadcasts.Service) *BroadcastsHandler {
	return &BroadcastsHandler{service: service}
}

// CreateBroadcast handles POST /api/v1/broadcasts/create
func (h *BroadcastsHandler) CreateBroadcast(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": err.Error()})
		return
	}

	broadcast, err := h.service.CreateBroadcast(c.Request.Context(), userID, req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":           true,
		"broadcast_id": broadcast.ID,
	})
}

// GetBroadcastList handles GET /api/v1/broadcasts/list
func (h *BroadcastsHandler) GetBroadcastList(c *gin.Context) {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil {
			page = 1
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if _, err := fmt.Sscanf(ps, "%d", &pageSize); err != nil {
			pageSize = 20
		}
	}

	broadcastList, err := h.service.GetBroadcasts(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": err.Error()})
		return
	}

	result := make([]map[string]interface{}, 0, len(broadcastList))
	for _, b := range broadcastList {
		result = append(result, b.ToTL())
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"broadcasts": result,
	})
}

// GetBroadcast handles GET /api/v1/broadcasts/:id
func (h *BroadcastsHandler) GetBroadcast(c *gin.Context) {
	var broadcastID int64
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &broadcastID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": "invalid broadcast_id"})
		return
	}

	broadcast, err := h.service.GetBroadcast(c.Request.Context(), broadcastID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error_code": "BROADCAST_NOT_FOUND", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"broadcast": broadcast.ToTL(),
	})
}

// UpdateBroadcast handles PUT /api/v1/broadcasts/:id
func (h *BroadcastsHandler) UpdateBroadcast(c *gin.Context) {
	var broadcastID int64
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &broadcastID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": "invalid broadcast_id"})
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": err.Error()})
		return
	}

	broadcast, err := h.service.UpdateBroadcast(c.Request.Context(), broadcastID, req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"broadcast": broadcast.ToTL(),
	})
}

// DeleteBroadcast handles DELETE /api/v1/broadcasts/:id
func (h *BroadcastsHandler) DeleteBroadcast(c *gin.Context) {
	var broadcastID int64
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &broadcastID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": "invalid broadcast_id"})
		return
	}

	if err := h.service.DeleteBroadcast(c.Request.Context(), broadcastID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
	})
}

// SendBroadcast handles POST /api/v1/broadcasts/:id/send
func (h *BroadcastsHandler) SendBroadcast(c *gin.Context) {
	var broadcastID int64
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &broadcastID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": "invalid broadcast_id"})
		return
	}

	broadcast, err := h.service.SendBroadcast(c.Request.Context(), broadcastID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"broadcast": broadcast.ToTL(),
	})
}
