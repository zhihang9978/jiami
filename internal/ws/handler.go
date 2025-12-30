package ws

import (
	"context"
	"log"
	"net/http"

	"github.com/feiji/feiji-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// Handler handles WebSocket connections
type Handler struct {
	hub         *Hub
	authService AuthService
}

// AuthService interface for authentication
type AuthService interface {
	GetUserByAuthKey(ctx context.Context, authKeyID string) (*models.User, error)
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, authService AuthService) *Handler {
	return &Handler{
		hub:         hub,
		authService: authService,
	}
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Get auth key from query or header
	authKeyID := c.Query("auth_key_id")
	if authKeyID == "" {
		authKeyID = c.GetHeader("X-Auth-Key-ID")
	}
	if authKeyID == "" {
		authKeyID = c.GetHeader("Authorization")
	}

	if authKeyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Validate auth key and get user
	user, err := h.authService.GetUserByAuthKey(c.Request.Context(), authKeyID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get user ID from user object
	userID := getUserID(user)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create client and register
	client := NewClient(h.hub, conn, userID)
	h.hub.Register(client)

	// Start read and write pumps
	go client.WritePump()
	go client.ReadPump()

	log.Printf("WebSocket client connected: user_id=%d", userID)
}

// getUserID extracts user ID from user object
func getUserID(user *models.User) int64 {
	if user == nil {
		return 0
	}
	return user.ID
}
