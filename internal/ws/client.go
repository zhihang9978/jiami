package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// CallService interface for call operations
type CallService interface {
	GetCall(ctx context.Context, callID int64) (*CallInfo, error)
}

// CallInfo represents basic call information
type CallInfo struct {
	ID       int64
	CallerID int64
	CalleeID int64
	State    string
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// Client represents a WebSocket client connection
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int64
}

// NewClient creates a new client
func NewClient(hub *Hub, conn *websocket.Conn, userID int64) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming message
		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse WebSocket message: %v", err)
		return
	}

	// Handle different message types
	msgType, _ := msg["type"].(string)
	switch msgType {
	case "ping":
		// Respond with pong
		c.Send(map[string]interface{}{
			"type": "pong",
			"time": time.Now().Unix(),
		})
	case "ack":
		// Client acknowledged receiving an update
		// Could be used for delivery confirmation
	case "signaling":
		// Handle VoIP signaling data forwarding
		c.handleSignaling(msg)
	default:
		log.Printf("Unknown WebSocket message type: %s", msgType)
	}
}

// handleSignaling processes VoIP signaling messages and forwards them to the peer
func (c *Client) handleSignaling(msg map[string]interface{}) {
	// Extract call_id from message
	callIDFloat, ok := msg["call_id"].(float64)
	if !ok {
		log.Printf("Signaling message missing call_id")
		c.Send(map[string]interface{}{
			"type":  "error",
			"error": "missing call_id",
		})
		return
	}
	callID := int64(callIDFloat)

	// Extract signaling data
	data, ok := msg["data"].(string)
	if !ok {
		log.Printf("Signaling message missing data")
		c.Send(map[string]interface{}{
			"type":  "error",
			"error": "missing data",
		})
		return
	}

	// Get call info from hub's call service
	callInfo := c.hub.GetCallInfo(callID)
	if callInfo == nil {
		log.Printf("Call not found: %d", callID)
		c.Send(map[string]interface{}{
			"type":  "error",
			"error": "call not found",
		})
		return
	}

	// Verify sender is a participant in the call
	if callInfo.CallerID != c.userID && callInfo.CalleeID != c.userID {
		log.Printf("User %d is not a participant in call %d", c.userID, callID)
		c.Send(map[string]interface{}{
			"type":  "error",
			"error": "not authorized",
		})
		return
	}

	// Determine the peer (the other participant)
	var peerID int64
	if callInfo.CallerID == c.userID {
		peerID = callInfo.CalleeID
	} else {
		peerID = callInfo.CallerID
	}

	// Forward signaling data to peer
	forwardMsg := map[string]interface{}{
		"type":         "signaling",
		"call_id":      callID,
		"from_user_id": c.userID,
		"data":         data,
	}

	if err := c.hub.SendToUser(peerID, forwardMsg); err != nil {
		log.Printf("Failed to forward signaling to user %d: %v", peerID, err)
	} else {
		log.Printf("Forwarded signaling from user %d to user %d for call %d", c.userID, peerID, callID)
	}
}

// Send sends a message to the client
func (c *Client) Send(data interface{}) error {
	message, err := json.Marshal(data)
	if err != nil {
		return err
	}

	select {
	case c.send <- message:
		return nil
	default:
		return nil // Buffer full, skip
	}
}

// UserID returns the user ID of the client
func (c *Client) UserID() int64 {
	return c.userID
}
