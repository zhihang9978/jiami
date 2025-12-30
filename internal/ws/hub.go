package ws

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to clients
type Hub struct {
	// Registered clients by user ID
	clients map[int64]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast message to specific user
	broadcast chan *UserMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Database connection for call lookups
	db *sql.DB
}

// UserMessage represents a message to be sent to a specific user
type UserMessage struct {
	UserID  int64
	Message []byte
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *UserMessage, 256),
	}
}

// NewHubWithDB creates a new Hub with database connection for call lookups
func NewHubWithDB(db *sql.DB) *Hub {
	return &Hub{
		clients:    make(map[int64]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *UserMessage, 256),
		db:         db,
	}
}

// SetDB sets the database connection for call lookups
func (h *Hub) SetDB(db *sql.DB) {
	h.db = db
}

// GetCallInfo retrieves call information for signaling validation
func (h *Hub) GetCallInfo(callID int64) *CallInfo {
	if h.db == nil {
		log.Printf("Hub: database connection not set")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var callInfo CallInfo
	err := h.db.QueryRowContext(ctx,
		"SELECT id, caller_id, callee_id, status FROM calls WHERE id = ?",
		callID,
	).Scan(&callInfo.ID, &callInfo.CallerID, &callInfo.CalleeID, &callInfo.State)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Hub: failed to get call info: %v", err)
		}
		return nil
	}

	return &callInfo
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.userID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.clients[message.UserID]; ok {
				for client := range clients {
					select {
					case client.send <- message.Message:
					default:
						// Client buffer full, skip
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToUser sends a message to all connections of a specific user
func (h *Hub) SendToUser(userID int64, data interface{}) error {
	message, err := json.Marshal(data)
	if err != nil {
		return err
	}

	h.broadcast <- &UserMessage{
		UserID:  userID,
		Message: message,
	}
	return nil
}

// IsUserOnline checks if a user has any active connections
func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

// GetOnlineUsers returns a list of online user IDs
func (h *Hub) GetOnlineUsers() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	users := make([]int64, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// GetUserConnectionCount returns the number of connections for a user
func (h *Hub) GetUserConnectionCount(userID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if clients, ok := h.clients[userID]; ok {
		return len(clients)
	}
	return 0
}

// Register registers a client
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
