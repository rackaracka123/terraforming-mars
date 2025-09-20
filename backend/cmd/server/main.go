package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpHandler "terraforming-mars-backend/internal/delivery/http"
	wsHandler "terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/store"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	logLevel := os.Getenv("TM_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "debug"
	}

	// Initialize logger
	if err := logger.Init(&logLevel); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Shutdown()

	logger.Info("🚀 Starting Terraforming Mars backend server")
	logger.Info("Log level set to " + logLevel)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Initialize event system
	eventBus := events.NewInMemoryEventBus()
	logger.Info("Event bus initialized")

	// Initialize the store with reducer pattern and load cards
	appStore, err := store.InitializeStore(eventBus)
	if err != nil {
		logger.Fatal("Failed to initialize store", zap.Error(err))
	}
	logger.Info("✅ Application store initialized with reducer pattern")

	// Initialize WebSocket service with store
	webSocketService := wsHandler.NewWebSocketService(appStore, eventBus)

	// Start WebSocket service in background
	wsCtx, wsCancel := context.WithCancel(context.Background())
	defer wsCancel()
	go webSocketService.Run(wsCtx)
	logger.Info("WebSocket hub started with store-based architecture")

	// Setup main router without middleware for WebSocket
	mainRouter := mux.NewRouter()

	// Setup API router with store
	apiRouter := httpHandler.SetupRouter(appStore)

	// Mount API router
	mainRouter.PathPrefix("/api/v1").Handler(apiRouter)

	// Add WebSocket endpoint directly to main router (no middleware)
	mainRouter.HandleFunc("/ws", webSocketService.ServeWS)

	// Setup HTTP server
	server := &http.Server{
		Addr:         ":3001",
		Handler:      mainRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in background
	go func() {
		logger.Info("Starting HTTP server on :3001")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	logger.Info("✅ Server started")
	logger.Info("🌍 HTTP server listening on :3001")
	logger.Info("🔌 WebSocket endpoint available at /ws")

	// Wait for shutdown signal
	<-quit

	logger.Info("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to gracefully shutdown HTTP server", zap.Error(err))
	} else {
		logger.Info("✅ HTTP server stopped")
	}

	// Cancel WebSocket service context
	wsCancel()
	logger.Info("✅ WebSocket service stopped")

	logger.Info("✅ Server shutdown complete")
}
