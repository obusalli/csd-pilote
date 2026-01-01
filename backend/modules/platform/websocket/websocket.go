package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"csd-pilote/backend/modules/platform/config"
	"csd-pilote/backend/modules/platform/events"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Validate origin against CORS allowed origins to prevent CSWSH attacks
		origin := r.Header.Get("Origin")
		if origin == "" {
			// No origin header (same-origin request or non-browser client)
			return true
		}

		cfg := config.GetConfig()
		if cfg == nil {
			return false
		}

		for _, allowed := range cfg.CORS.AllowedOrigins {
			if allowed == "*" || allowed == origin {
				return true
			}
		}

		log.Printf("[WebSocket] Rejected connection from unauthorized origin: %s", origin)
		return false
	},
}

// Client represents a WebSocket client connection
type Client struct {
	ID       string
	TenantID uuid.UUID
	UserID   uuid.UUID
	Conn     *websocket.Conn
	Send     chan []byte
}

// Hub manages all WebSocket clients
type Hub struct {
	mu         sync.RWMutex
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	TenantID uuid.UUID
	Data     []byte
}

var globalHub *Hub
var hubOnce sync.Once

// GetHub returns the global WebSocket hub singleton
func GetHub() *Hub {
	hubOnce.Do(func() {
		globalHub = &Hub{
			clients:    make(map[string]*Client),
			register:   make(chan *Client),
			unregister: make(chan *Client),
			broadcast:  make(chan *BroadcastMessage, 256),
		}
		go globalHub.run()
		globalHub.subscribeToEvents()
	})
	return globalHub
}

// run starts the hub's main loop
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("[WebSocket] Client connected: %s (tenant: %s)", client.ID, client.TenantID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("[WebSocket] Client disconnected: %s", client.ID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				// Only send to clients in the same tenant
				if client.TenantID == message.TenantID {
					select {
					case client.Send <- message.Data:
					default:
						// Client send buffer full, skip
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// subscribeToEvents subscribes to all events and broadcasts them
func (h *Hub) subscribeToEvents() {
	bus := events.GetEventBus()

	// Subscribe to all event types
	bus.SubscribeAll(func(ctx context.Context, event events.Event) {
		data, err := event.ToJSON()
		if err != nil {
			log.Printf("[WebSocket] Failed to marshal event: %v", err)
			return
		}

		h.broadcast <- &BroadcastMessage{
			TenantID: event.TenantID,
			Data:     data,
		}
	})
}

// Broadcast sends a message to all clients in a tenant
func (h *Hub) Broadcast(tenantID uuid.UUID, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	h.broadcast <- &BroadcastMessage{
		TenantID: tenantID,
		Data:     jsonData,
	}
	return nil
}

// ClientCount returns the number of connected clients
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// ClientCountByTenant returns the number of connected clients for a tenant
func (h *Hub) ClientCountByTenant(tenantID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, client := range h.clients {
		if client.TenantID == tenantID {
			count++
		}
	}
	return count
}

// HandleWebSocket handles WebSocket upgrade and connection
func HandleWebSocket(w http.ResponseWriter, r *http.Request, tenantID, userID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	hub := GetHub()
	hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump(hub)
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("[WebSocket] Write error: %v", err)
			return
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Read error: %v", err)
			}
			break
		}
		// Handle incoming messages if needed (e.g., subscription filters)
	}
}
