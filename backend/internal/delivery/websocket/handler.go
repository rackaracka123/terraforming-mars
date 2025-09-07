package websocket

import (
	"net/http"
	"time"

	"terraforming-mars-backend/internal/logger"

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
	h.logger.Info("🔗 WebSocket connection request received", zap.String("remote_addr", r.RemoteAddr))

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("❌ Failed to upgrade connection to WebSocket", zap.Error(err))
		return
	}

	// Create connection ID
	connectionID := uuid.New().String()

	h.logger.Info("✅ New WebSocket connection established",
		zap.String("connection_id", connectionID),
		zap.String("remote_addr", r.RemoteAddr))

	// Create new connection
	connection := NewConnection(connectionID, conn, h.hub)

	// Register connection with hub
	h.logger.Info("📤 Registering connection with hub", zap.String("connection_id", connectionID))
	h.hub.Register <- connection
	h.logger.Info("✅ Connection sent to Register channel successfully", zap.String("connection_id", connectionID))
	// Configure connection timeouts
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// Handle pong messages to keep connection alive
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	h.logger.Debug("🚀 Starting connection pumps", zap.String("connection_id", connectionID))
	// Start read and write pumps without context (they'll manage their own lifecycle)
	go connection.WritePump()
	go connection.ReadPump()

	// Send periodic pings to keep connection alive
	go h.pingLoop(connection)

	h.logger.Info("🎉 WebSocket connection fully initialized", zap.String("connection_id", connectionID))
}

// pingLoop sends periodic ping messages to keep the connection alive
func (h *Handler) pingLoop(connection *Connection) {
	ticker := time.NewTicker(54 * time.Second) // Ping every 54 seconds
	defer ticker.Stop()

	for {
		select {
		case <-connection.Done:
			h.logger.Debug("Ping loop stopping - connection closed", zap.String("connection_id", connection.ID))
			return
		case <-ticker.C:
			if err := connection.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				h.logger.Error("Failed to set write deadline for ping",
					zap.Error(err),
					zap.String("connection_id", connection.ID))
				return
			}
			if err := connection.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.logger.Error("Failed to send ping message",
					zap.Error(err),
					zap.String("connection_id", connection.ID))
				return
			}
		}
	}
}
