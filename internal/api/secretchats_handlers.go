package api

import (
	"net/http"

	"github.com/feiji/feiji-backend/internal/secretchats"
	"github.com/gin-gonic/gin"
)

type SecretChatsHandler struct {
	service *secretchats.Service
}

func NewSecretChatsHandler(service *secretchats.Service) *SecretChatsHandler {
	return &SecretChatsHandler{service: service}
}

// CreateSecretChat handles POST /api/v1/secret_chats/create
func (h *SecretChatsHandler) CreateSecretChat(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		PeerID int64  `json:"peer_id" binding:"required"`
		GA     string `json:"g_a" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": err.Error()})
		return
	}

	chat, err := h.service.CreateSecretChat(c.Request.Context(), userID, req.PeerID, req.GA)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":             true,
		"secret_chat_id": chat.ID,
		"status":         chat.Status,
	})
}

// UpdateSecretChatStatus handles POST /api/v1/secret_chats/status
func (h *SecretChatsHandler) UpdateSecretChatStatus(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		SecretChatID int64  `json:"secret_chat_id" binding:"required"`
		Status       string `json:"status" binding:"required"`
		GB           string `json:"g_b"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": err.Error()})
		return
	}

	chat, err := h.service.UpdateSecretChatStatus(c.Request.Context(), req.SecretChatID, userID, req.Status, req.GB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":     true,
		"status": chat.Status,
	})
}

// GetSecretChatList handles GET /api/v1/secret_chats/list
func (h *SecretChatsHandler) GetSecretChatList(c *gin.Context) {
	userID := c.GetInt64("user_id")

	chats, err := h.service.GetUserSecretChats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": err.Error()})
		return
	}

	chatList := make([]map[string]interface{}, 0, len(chats))
	for _, chat := range chats {
		chatList = append(chatList, chat.ToTL())
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":           true,
		"secret_chats": chatList,
	})
}

// SendSecretMessage handles POST /api/v1/secret_messages/send
func (h *SecretChatsHandler) SendSecretMessage(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		SecretChatID     int64  `json:"secret_chat_id" binding:"required"`
		EncryptedMessage string `json:"encrypted_message" binding:"required"`
		RandomID         int64  `json:"random_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": err.Error()})
		return
	}

	msg, err := h.service.SendSecretMessage(c.Request.Context(), req.SecretChatID, userID, req.EncryptedMessage, req.RandomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"message_id": msg.ID,
		"date":       msg.Date,
	})
}

// GetSecretMessageHistory handles GET /api/v1/secret_messages/history
func (h *SecretChatsHandler) GetSecretMessageHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		SecretChatID int64 `form:"secret_chat_id" binding:"required"`
		Limit        int   `form:"limit"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": err.Error()})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 50
	}

	messages, err := h.service.GetSecretMessages(c.Request.Context(), req.SecretChatID, userID, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": err.Error()})
		return
	}

	msgList := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		msgList = append(msgList, msg.ToTL())
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"messages": msgList,
	})
}

// CloseSecretChat handles POST /api/v1/secret_chats/close
func (h *SecretChatsHandler) CloseSecretChat(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		SecretChatID int64 `json:"secret_chat_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": err.Error()})
		return
	}

	if err := h.service.CloseSecretChat(c.Request.Context(), req.SecretChatID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
	})
}
