package core

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

// Handler handles WebSocket HTTP upgrade requests
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

// ServeWS handles WebSocket upgrade requests from clients
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	logger.Info("🔗 WebSocket connection request received", zap.String("remote_addr", r.RemoteAddr))

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("❌ Failed to upgrade connection to WebSocket", zap.Error(err))
		return
	}

	// Create connection ID and connection object
	connectionID := uuid.New().String()
	connection := NewConnection(connectionID, conn, h.hub)

	logger.Info("✅ New WebSocket connection established",
		zap.String("connection_id", connectionID),
		zap.String("remote_addr", r.RemoteAddr))

	// Register connection with hub
	h.hub.Register <- connection

	// Configure connection timeouts
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// Handle pong messages to keep connection alive
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start connection pumps
	go connection.WritePump()
	go connection.ReadPump()

	logger.Info("🎉 WebSocket connection fully initialized", zap.String("connection_id", connectionID))
}
