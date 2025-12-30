package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/feiji/feiji-backend/internal/admin/models"
	"github.com/gorilla/websocket"
)

// AdminHub manages admin WebSocket connections
type AdminHub struct {
	// Registered clients
	clients map[*AdminClient]bool

	// Inbound messages from clients
	broadcast chan *models.WebSocketMessage

	// Register requests from clients
	register chan *AdminClient

	// Unregister requests from clients
	unregister chan *AdminClient

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// AdminClient represents a WebSocket client
type AdminClient struct {
	hub      *AdminHub
	conn     *websocket.Conn
	send     chan []byte
	adminID  int64
	username string
}

// NewAdminHub creates a new admin hub
func NewAdminHub() *AdminHub {
	return &AdminHub{
		clients:    make(map[*AdminClient]bool),
		broadcast:  make(chan *models.WebSocketMessage, 256),
		register:   make(chan *AdminClient),
		unregister: make(chan *AdminClient),
	}
}

// Run starts the hub
func (h *AdminHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Admin client connected: %s (ID: %d)", client.username, client.adminID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Admin client disconnected: %s (ID: %d)", client.username, client.adminID)

		case message := <-h.broadcast:
			h.mu.RLock()
			data, _ := json.Marshal(message)
			for client := range h.clients {
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastNotification broadcasts a notification to all admin clients
func (h *AdminHub) BroadcastNotification(notification *models.AdminNotification) {
	message := &models.WebSocketMessage{
		Type:      "notification",
		Category:  notification.Category,
		Title:     notification.Title,
		Message:   notification.Message,
		Data:      notification.Data,
		Priority:  notification.Priority,
		Sound:     notification.Priority == "high" || notification.Priority == "urgent",
		Timestamp: time.Now().Unix(),
	}
	h.broadcast <- message
}

// BroadcastToAdmin broadcasts a message to a specific admin
func (h *AdminHub) BroadcastToAdmin(adminID int64, message *models.WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, _ := json.Marshal(message)
	for client := range h.clients {
		if client.adminID == adminID {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

// GetOnlineAdminCount returns the number of online admins
func (h *AdminHub) GetOnlineAdminCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetOnlineAdmins returns list of online admin IDs
func (h *AdminHub) GetOnlineAdmins() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	adminIDs := make([]int64, 0, len(h.clients))
	seen := make(map[int64]bool)
	for client := range h.clients {
		if !seen[client.adminID] {
			adminIDs = append(adminIDs, client.adminID)
			seen[client.adminID] = true
		}
	}
	return adminIDs
}

// SendUserRegisterNotification sends user registration notification
func (h *AdminHub) SendUserRegisterNotification(userID int64, phone string) {
	notification := &models.AdminNotification{
		Category: "user_register",
		Title:    "New User Registration",
		Message:  "User " + phone + " just registered",
		Priority: "normal",
	}
	notification.Data, _ = json.Marshal(map[string]interface{}{
		"user_id": userID,
		"phone":   phone,
	})
	h.BroadcastNotification(notification)
}

// SendUserLoginNotification sends user login notification
func (h *AdminHub) SendUserLoginNotification(userID int64, phone string) {
	notification := &models.AdminNotification{
		Category: "user_login",
		Title:    "User Login",
		Message:  "User " + phone + " logged in",
		Priority: "low",
	}
	notification.Data, _ = json.Marshal(map[string]interface{}{
		"user_id": userID,
		"phone":   phone,
	})
	h.BroadcastNotification(notification)
}

// SendCallStartNotification sends call start notification
func (h *AdminHub) SendCallStartNotification(callID int64, callerID, calleeID int64, isVideo bool) {
	callType := "voice"
	if isVideo {
		callType = "video"
	}
	notification := &models.AdminNotification{
		Category: "call_start",
		Title:    "Call Started",
		Message:  "A " + callType + " call has started",
		Priority: "low",
	}
	notification.Data, _ = json.Marshal(map[string]interface{}{
		"call_id":   callID,
		"caller_id": callerID,
		"callee_id": calleeID,
		"is_video":  isVideo,
	})
	h.BroadcastNotification(notification)
}

// SendCallEndNotification sends call end notification
func (h *AdminHub) SendCallEndNotification(callID int64, duration int) {
	notification := &models.AdminNotification{
		Category: "call_end",
		Title:    "Call Ended",
		Message:  "Call ended after " + formatDuration(duration),
		Priority: "low",
	}
	notification.Data, _ = json.Marshal(map[string]interface{}{
		"call_id":  callID,
		"duration": duration,
	})
	h.BroadcastNotification(notification)
}

// SendServiceDownNotification sends service down notification
func (h *AdminHub) SendServiceDownNotification(serviceName string) {
	notification := &models.AdminNotification{
		Category: "service_down",
		Title:    "Service Error",
		Message:  serviceName + " service has stopped",
		Priority: "urgent",
	}
	notification.Data, _ = json.Marshal(map[string]interface{}{
		"service": serviceName,
	})
	h.BroadcastNotification(notification)
}

// SendServiceUpNotification sends service up notification
func (h *AdminHub) SendServiceUpNotification(serviceName string) {
	notification := &models.AdminNotification{
		Category: "service_up",
		Title:    "Service Recovered",
		Message:  serviceName + " service has recovered",
		Priority: "normal",
	}
	notification.Data, _ = json.Marshal(map[string]interface{}{
		"service": serviceName,
	})
	h.BroadcastNotification(notification)
}

// SendSystemErrorNotification sends system error notification
func (h *AdminHub) SendSystemErrorNotification(errorMsg string) {
	notification := &models.AdminNotification{
		Category: "system_error",
		Title:    "System Error",
		Message:  errorMsg,
		Priority: "high",
	}
	h.BroadcastNotification(notification)
}

// SendBroadcastCompleteNotification sends broadcast complete notification
func (h *AdminHub) SendBroadcastCompleteNotification(broadcastID int64, totalUsers, successCount, failedCount int) {
	notification := &models.AdminNotification{
		Category: "broadcast_complete",
		Title:    "Broadcast Complete",
		Message:  "Broadcast sent successfully",
		Priority: "normal",
	}
	notification.Data, _ = json.Marshal(map[string]interface{}{
		"broadcast_id":  broadcastID,
		"total_users":   totalUsers,
		"success_count": successCount,
		"failed_count":  failedCount,
	})
	h.BroadcastNotification(notification)
}

// SendBroadcastFailedNotification sends broadcast failed notification
func (h *AdminHub) SendBroadcastFailedNotification(broadcastID int64, errorMsg string) {
	notification := &models.AdminNotification{
		Category: "broadcast_failed",
		Title:    "Broadcast Failed",
		Message:  errorMsg,
		Priority: "high",
	}
	notification.Data, _ = json.Marshal(map[string]interface{}{
		"broadcast_id": broadcastID,
		"error":        errorMsg,
	})
	h.BroadcastNotification(notification)
}

// formatDuration formats duration in seconds to human readable string
func formatDuration(seconds int) string {
	if seconds < 60 {
		return string(rune(seconds)) + " seconds"
	}
	minutes := seconds / 60
	secs := seconds % 60
	if secs == 0 {
		return string(rune(minutes)) + " minutes"
	}
	return string(rune(minutes)) + " minutes " + string(rune(secs)) + " seconds"
}

// Global admin hub instance
var GlobalAdminHub *AdminHub

// InitAdminHub initializes the global admin hub
func InitAdminHub() {
	GlobalAdminHub = NewAdminHub()
	go GlobalAdminHub.Run()
}
