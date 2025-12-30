package api

import (
	"net/http"

	"github.com/feiji/feiji-backend/internal/search"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	service *search.Service
}

func NewSearchHandler(service *search.Service) *SearchHandler {
	return &SearchHandler{service: service}
}

// SearchGlobal performs a global search
func (h *SearchHandler) SearchGlobal(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Q           string `json:"q" binding:"required"`
		OffsetRate  int    `json:"offset_rate"`
		OffsetPeer  int64  `json:"offset_peer"`
		OffsetID    int64  `json:"offset_id"`
		Limit       int    `json:"limit"`
		FolderID    int    `json:"folder_id"`
		MinDate     int    `json:"min_date"`
		MaxDate     int    `json:"max_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.SearchGlobal(c.Request.Context(), userID, req.Q, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result.ToTL())
}

// SearchMessages searches for messages
func (h *SearchHandler) SearchMessages(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		PeerID    int64  `json:"peer_id"`
		PeerType  string `json:"peer_type"`
		Q         string `json:"q"`
		FromID    int64  `json:"from_id"`
		Filter    string `json:"filter"`
		MinDate   int    `json:"min_date"`
		MaxDate   int    `json:"max_date"`
		OffsetID  int64  `json:"offset_id"`
		AddOffset int    `json:"add_offset"`
		Limit     int    `json:"limit"`
		MaxID     int64  `json:"max_id"`
		MinID     int64  `json:"min_id"`
		Hash      int64  `json:"hash"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.service.SearchMessages(c.Request.Context(), userID, req.Q, req.PeerID, req.PeerType, req.OffsetID, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	messages := make([]map[string]interface{}, len(results))
	for i, r := range results {
		messages[i] = r.ToMessageTL()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":                "messages.messages",
		"messages":         messages,
		"chats":            []interface{}{},
		"users":            []interface{}{},
		"count":            len(messages),
		"inexact":          false,
		"next_rate":        0,
		"offset_id_offset": 0,
	})
}

// SearchHashtag searches for messages with a hashtag
func (h *SearchHandler) SearchHashtag(c *gin.Context) {
	var req struct {
		Hashtag  string `json:"hashtag" binding:"required"`
		OffsetID int64  `json:"offset_id"`
		Limit    int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.service.SearchHashtag(c.Request.Context(), req.Hashtag, req.OffsetID, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	messages := make([]map[string]interface{}, len(results))
	for i, r := range results {
		messages[i] = r.ToMessageTL()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":        "messages.messages",
		"messages": messages,
		"chats":    []interface{}{},
		"users":    []interface{}{},
		"count":    len(messages),
	})
}

// GetRecentSearch retrieves recent search queries
func (h *SearchHandler) GetRecentSearch(c *gin.Context) {
	userID := c.GetInt64("user_id")

	queries, err := h.service.GetRecentSearch(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":       "contacts.found",
		"queries": queries,
	})
}

// ClearRecentSearch clears search history
func (h *SearchHandler) ClearRecentSearch(c *gin.Context) {
	userID := c.GetInt64("user_id")

	if err := h.service.ClearRecentSearch(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":      "boolTrue",
		"result": true,
	})
}

// SearchUsers searches for users
func (h *SearchHandler) SearchUsers(c *gin.Context) {
	var req struct {
		Q     string `json:"q" binding:"required"`
		Limit int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.service.SearchUsers(c.Request.Context(), req.Q, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	users := make([]map[string]interface{}, len(results))
	for i, r := range results {
		users[i] = r.ToUserTL()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":     "contacts.found",
		"users": users,
		"chats": []interface{}{},
	})
}

// SearchChannels searches for channels
func (h *SearchHandler) SearchChannels(c *gin.Context) {
	var req struct {
		Q     string `json:"q" binding:"required"`
		Limit int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.service.SearchChannels(c.Request.Context(), req.Q, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	channels := make([]map[string]interface{}, len(results))
	for i, r := range results {
		channels[i] = r.ToChannelTL()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":        "contacts.found",
		"channels": channels,
		"users":    []interface{}{},
		"chats":    []interface{}{},
	})
}
