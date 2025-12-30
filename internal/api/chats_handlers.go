package api

import (
	"net/http"
	"strconv"

	"github.com/feiji/feiji-backend/internal/chats"
	"github.com/gin-gonic/gin"
)

type ChatsHandler struct {
	service *chats.Service
}

func NewChatsHandler(service *chats.Service) *ChatsHandler {
	return &ChatsHandler{service: service}
}

// CreateChat creates a new chat group
func (h *ChatsHandler) CreateChat(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Title   string  `json:"title" binding:"required"`
		UserIDs []int64 `json:"users" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, err := h.service.CreateChat(c.Request.Context(), userID, req.Title, req.UserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chat.ToTL())
}

// GetFullChat retrieves full chat info
func (h *ChatsHandler) GetFullChat(c *gin.Context) {
	chatIDStr := c.Query("chat_id")
	if chatIDStr == "" {
		var req struct {
			ChatID int64 `json:"chat_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		chatIDStr = strconv.FormatInt(req.ChatID, 10)
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}

	fullChat, err := h.service.GetFullChat(c.Request.Context(), chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fullChat.ToTL())
}

// EditChatTitle updates the chat title
func (h *ChatsHandler) EditChatTitle(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChatID int64  `json:"chat_id" binding:"required"`
		Title  string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, err := h.service.EditChatTitle(c.Request.Context(), req.ChatID, userID, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chat.ToTL())
}

// EditChatPhoto updates the chat photo
func (h *ChatsHandler) EditChatPhoto(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChatID  int64 `json:"chat_id" binding:"required"`
		PhotoID int64 `json:"photo_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, err := h.service.EditChatPhoto(c.Request.Context(), req.ChatID, userID, req.PhotoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chat.ToTL())
}

// AddChatUser adds a user to a chat
func (h *ChatsHandler) AddChatUser(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChatID   int64 `json:"chat_id" binding:"required"`
		UserID   int64 `json:"user_id" binding:"required"`
		FwdLimit int   `json:"fwd_limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, err := h.service.AddChatUser(c.Request.Context(), req.ChatID, userID, req.UserID, req.FwdLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chat.ToTL())
}

// DeleteChatUser removes a user from a chat
func (h *ChatsHandler) DeleteChatUser(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChatID int64 `json:"chat_id" binding:"required"`
		UserID int64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, err := h.service.DeleteChatUser(c.Request.Context(), req.ChatID, userID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chat.ToTL())
}

// LeaveChat removes the user from a chat
func (h *ChatsHandler) LeaveChat(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChatID int64 `json:"chat_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, err := h.service.LeaveChat(c.Request.Context(), req.ChatID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chat.ToTL())
}

// EditChatAdmin updates admin status for a user
func (h *ChatsHandler) EditChatAdmin(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ChatID  int64 `json:"chat_id" binding:"required"`
		UserID  int64 `json:"user_id" binding:"required"`
		IsAdmin bool  `json:"is_admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, err := h.service.EditChatAdmin(c.Request.Context(), req.ChatID, userID, req.UserID, req.IsAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chat.ToTL())
}

// GetCommonChats retrieves common chats with a user
func (h *ChatsHandler) GetCommonChats(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		UserID int64 `json:"user_id" binding:"required"`
		MaxID  int64 `json:"max_id"`
		Limit  int   `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user's chats
	chats, err := h.service.GetUserChats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter to common chats (simplified - would need to check other user's membership)
	chatsTL := make([]map[string]interface{}, len(chats))
	for i, chat := range chats {
		chatsTL[i] = chat.ToTL()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":      "messages.chats",
		"chats":  chatsTL,
		"count":  len(chatsTL),
	})
}
