package websocket

import (
	"context"
	"net/http"
	"terraforming-mars-backend/internal/logger"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development - should be restricted in production
		return true
	},
}

// Handler handles WebSocket connections
type Handler struct {
	hub    *Hub
	logger *zap.Logger
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub:    hub,
		logger: logger.Get(),
	}
}

// ServeWS handles WebSocket requests from clients
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection to WebSocket", zap.Error(err))
		return
	}

	// Create connection ID
	connectionID := uuid.New().String()
	
	h.logger.Info("New WebSocket connection established",
		zap.String("connection_id", connectionID),
		zap.String("remote_addr", r.RemoteAddr))

	// Create new connection
	connection := NewConnection(connectionID, conn, h.hub)
	
	// Register connection with hub
	h.hub.Register <- connection
	
	// Start goroutines with context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Configure connection timeouts
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	
	// Handle pong messages to keep connection alive
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	// Start read and write pumps
	go connection.WritePump(ctx)
	go connection.ReadPump(ctx)
	
	// Send periodic pings to keep connection alive
	go h.pingLoop(ctx, connection)
}

// pingLoop sends periodic ping messages to keep the connection alive
func (h *Handler) pingLoop(ctx context.Context, connection *Connection) {
	ticker := time.NewTicker(54 * time.Second) // Ping every 54 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			connection.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := connection.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.logger.Error("Failed to send ping message",
					zap.Error(err),
					zap.String("connection_id", connection.ID))
				return
			}
		}
	}
}