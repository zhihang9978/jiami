package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/feiji/feiji-backend/internal/auth"
	"github.com/feiji/feiji-backend/internal/messages"
	"github.com/feiji/feiji-backend/internal/models"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService    *auth.Service
	messageService *messages.Service
}

func NewHandler(authService *auth.Service, messageService *messages.Service) *Handler {
	return &Handler{
		authService:    authService,
		messageService: messageService,
	}
}

// Health check
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

// ==================== Auth Handlers ====================

// SendCode handles auth.sendCode
func (h *Handler) SendCode(c *gin.Context) {
	var req struct {
		PhoneNumber string `json:"phone_number" binding:"required"`
		APIId       int    `json:"api_id" binding:"required"`
		APIHash     string `json:"api_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.authService.SendCode(c.Request.Context(), req.PhoneNumber, req.APIId, req.APIHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":               "auth.sentCode",
		"phone_code_hash": result.PhoneCodeHash,
		"type": gin.H{
			"_":      result.Type.Type,
			"length": result.Type.Length,
		},
		"next_type": result.NextType,
		"timeout":   result.Timeout,
	})
}

// SignIn handles auth.signIn
func (h *Handler) SignIn(c *gin.Context) {
	var req struct {
		PhoneNumber   string `json:"phone_number" binding:"required"`
		PhoneCodeHash string `json:"phone_code_hash" binding:"required"`
		PhoneCode     string `json:"phone_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.SignIn(c.Request.Context(), req.PhoneNumber, req.PhoneCodeHash, req.PhoneCode)
	if err != nil {
		if err.Error() == "user not registered" {
			c.JSON(http.StatusOK, gin.H{
				"_":               "auth.authorizationSignUpRequired",
				"phone_code_hash": req.PhoneCodeHash,
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get auth key ID from header or generate session
	authKeyID := c.GetHeader("X-Auth-Key-ID")
	if authKeyID != "" {
		h.authService.BindAuthKey(c.Request.Context(), authKeyID, user.ID)
		h.authService.CreateSession(c.Request.Context(), authKeyID, user.ID)
	}

	c.JSON(http.StatusOK, gin.H{
		"_":    "auth.authorization",
		"user": user.ToTLUser(),
	})
}

// SignUp handles auth.signUp
func (h *Handler) SignUp(c *gin.Context) {
	var req struct {
		PhoneNumber   string `json:"phone_number" binding:"required"`
		PhoneCodeHash string `json:"phone_code_hash" binding:"required"`
		FirstName     string `json:"first_name" binding:"required"`
		LastName      string `json:"last_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.SignUp(c.Request.Context(), req.PhoneNumber, req.PhoneCodeHash, req.FirstName, req.LastName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get auth key ID from header or generate session
	authKeyID := c.GetHeader("X-Auth-Key-ID")
	if authKeyID != "" {
		h.authService.BindAuthKey(c.Request.Context(), authKeyID, user.ID)
		h.authService.CreateSession(c.Request.Context(), authKeyID, user.ID)
	}

	c.JSON(http.StatusOK, gin.H{
		"_":    "auth.authorization",
		"user": user.ToTLUser(),
	})
}

// LogOut handles auth.logOut
func (h *Handler) LogOut(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"_": "auth.loggedOut",
	})
}

// ==================== Messages Handlers ====================

// SendMessage handles messages.sendMessage
func (h *Handler) SendMessage(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Peer struct {
			UserID    int64  `json:"user_id"`
			ChatID    int64  `json:"chat_id"`
			ChannelID int64  `json:"channel_id"`
			Type      string `json:"_"`
		} `json:"peer" binding:"required"`
		Message      string `json:"message" binding:"required"`
		RandomID     int64  `json:"random_id"`
		ReplyToMsgID *int   `json:"reply_to_msg_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine peer type and ID
	var peerID int64
	var peerType string
	switch {
	case req.Peer.UserID > 0:
		peerID = req.Peer.UserID
		peerType = "user"
	case req.Peer.ChatID > 0:
		peerID = req.Peer.ChatID
		peerType = "chat"
	case req.Peer.ChannelID > 0:
		peerID = req.Peer.ChannelID
		peerType = "channel"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid peer"})
		return
	}

	msg, err := h.messageService.SendMessage(c.Request.Context(), user.ID, peerID, peerType, req.Message, req.RandomID, req.ReplyToMsgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_": "updates",
		"updates": []interface{}{
			gin.H{
				"_":       "updateNewMessage",
				"message": msg.ToTLMessage(),
				"pts":     1,
				"pts_count": 1,
			},
		},
		"users": []interface{}{},
		"chats": []interface{}{},
		"date":  time.Now().Unix(),
		"seq":   0,
	})
}

// GetHistory handles messages.getHistory
func (h *Handler) GetHistory(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Peer struct {
			UserID    int64  `json:"user_id"`
			ChatID    int64  `json:"chat_id"`
			ChannelID int64  `json:"channel_id"`
			Type      string `json:"_"`
		} `json:"peer" binding:"required"`
		OffsetID   int `json:"offset_id"`
		OffsetDate int `json:"offset_date"`
		AddOffset  int `json:"add_offset"`
		Limit      int `json:"limit"`
		MaxID      int `json:"max_id"`
		MinID      int `json:"min_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine peer type and ID
	var peerID int64
	var peerType string
	switch {
	case req.Peer.UserID > 0:
		peerID = req.Peer.UserID
		peerType = "user"
	case req.Peer.ChatID > 0:
		peerID = req.Peer.ChatID
		peerType = "chat"
	case req.Peer.ChannelID > 0:
		peerID = req.Peer.ChannelID
		peerType = "channel"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid peer"})
		return
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	messages, err := h.messageService.GetHistory(c.Request.Context(), user.ID, peerID, peerType, req.OffsetID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert messages to TL format
	tlMessages := make([]interface{}, len(messages))
	for i, msg := range messages {
		tlMessages[i] = msg.ToTLMessage()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":        "messages.messages",
		"messages": tlMessages,
		"chats":    []interface{}{},
		"users":    []interface{}{},
	})
}

// GetDialogs handles messages.getDialogs
func (h *Handler) GetDialogs(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		OffsetDate int `json:"offset_date"`
		OffsetID   int `json:"offset_id"`
		Limit      int `json:"limit"`
		FolderID   int `json:"folder_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Use query params as fallback
		req.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "50"))
		req.OffsetDate, _ = strconv.Atoi(c.DefaultQuery("offset_date", "0"))
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	dialogs, err := h.messageService.GetDialogs(c.Request.Context(), user.ID, req.OffsetDate, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert dialogs to TL format
	tlDialogs := make([]interface{}, len(dialogs))
	for i, d := range dialogs {
		tlDialogs[i] = d.ToTLDialog()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":        "messages.dialogs",
		"dialogs":  tlDialogs,
		"messages": []interface{}{},
		"chats":    []interface{}{},
		"users":    []interface{}{},
	})
}

// ReadHistory handles messages.readHistory
func (h *Handler) ReadHistory(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Peer struct {
			UserID    int64  `json:"user_id"`
			ChatID    int64  `json:"chat_id"`
			ChannelID int64  `json:"channel_id"`
		} `json:"peer" binding:"required"`
		MaxID int `json:"max_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var peerID int64
	var peerType string
	switch {
	case req.Peer.UserID > 0:
		peerID = req.Peer.UserID
		peerType = "user"
	case req.Peer.ChatID > 0:
		peerID = req.Peer.ChatID
		peerType = "chat"
	case req.Peer.ChannelID > 0:
		peerID = req.Peer.ChannelID
		peerType = "channel"
	}

	if err := h.messageService.MarkAsRead(c.Request.Context(), user.ID, peerID, peerType, req.MaxID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":         "messages.affectedMessages",
		"pts":       1,
		"pts_count": 1,
	})
}

// DeleteMessages handles messages.deleteMessages
func (h *Handler) DeleteMessages(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ID     []int `json:"id" binding:"required"`
		Revoke bool  `json:"revoke"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.messageService.DeleteMessages(c.Request.Context(), user.ID, req.ID, req.Revoke); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":         "messages.affectedMessages",
		"pts":       1,
		"pts_count": len(req.ID),
	})
}

// EditMessage handles messages.editMessage
func (h *Handler) EditMessage(c *gin.Context) {
	user := h.getCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Peer struct {
			UserID    int64 `json:"user_id"`
			ChatID    int64 `json:"chat_id"`
			ChannelID int64 `json:"channel_id"`
		} `json:"peer" binding:"required"`
		ID       int             `json:"id" binding:"required"`
		Message  string          `json:"message"`
		Entities json.RawMessage `json:"entities"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var peerID int64
	var peerType string
	switch {
	case req.Peer.UserID > 0:
		peerID = req.Peer.UserID
		peerType = "user"
	case req.Peer.ChatID > 0:
		peerID = req.Peer.ChatID
		peerType = "chat"
	case req.Peer.ChannelID > 0:
		peerID = req.Peer.ChannelID
		peerType = "channel"
	}

	msg, err := h.messageService.EditMessage(c.Request.Context(), user.ID, peerID, peerType, req.ID, req.Message, req.Entities)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_": "updates",
		"updates": []interface{}{
			gin.H{
				"_":         "updateEditMessage",
				"message":   msg.ToTLMessage(),
				"pts":       1,
				"pts_count": 1,
			},
		},
		"users": []interface{}{},
		"chats": []interface{}{},
		"date":  time.Now().Unix(),
		"seq":   0,
	})
}

// Helper to get current user from context
func (h *Handler) getCurrentUser(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}
