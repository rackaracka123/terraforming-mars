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
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"

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

	log := logger.Get()
	log.Info("🚀 Starting Terraforming Mars backend server")
	log.Info("Log level set to " + logLevel)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Initialize individual repositories
	playerRepo := repository.NewPlayerRepository()
	log.Info("Player repository initialized")

	gameRepo := repository.NewGameRepository()
	log.Info("Game repository initialized")

	// Initialize card repository and load cards
	cardRepo := repository.NewCardRepository()
	if err := cardRepo.LoadCards(context.Background()); err != nil {
		log.Warn("Failed to load card data, using fallback cards", zap.Error(err))
	} else {
		allCards, _ := cardRepo.GetAllCards(context.Background())
		projectCards, _ := cardRepo.GetProjectCards(context.Background())
		corporationCards, _ := cardRepo.GetCorporationCards(context.Background())
		preludeCards, _ := cardRepo.GetPreludeCards(context.Background())
		log.Info("📚 Card data loaded successfully",
			zap.Int("project_cards", len(projectCards)),
			zap.Int("corporation_cards", len(corporationCards)),
			zap.Int("prelude_cards", len(preludeCards)),
			zap.Int("total_cards", len(allCards)))
	}

	// Initialize new service architecture
	cardDeckRepo := repository.NewCardDeckRepository()

	// Create Hub first (no dependencies)
	hub := core.NewHub()

	// Initialize SessionManager for WebSocket broadcasting with Hub
	sessionManager := session.NewSessionManager(gameRepo, playerRepo, cardRepo, hub)

	// Initialize CardService with SessionManager
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager)
	log.Info("SessionManager initialized for service-level broadcasting")

	boardService := service.NewBoardService()
	gameService := service.NewGameService(gameRepo, playerRepo, cardService.(*service.CardServiceImpl), boardService, sessionManager)
	playerService := service.NewPlayerService(gameRepo, playerRepo, sessionManager)
	standardProjectService := service.NewStandardProjectService(gameRepo, playerRepo, gameService)

	log.Info("Services initialized with new architecture and reconnection system")

	// Log service initialization
	log.Info("Player service ready", zap.Any("service", playerService != nil))
	log.Info("Game service ready", zap.Any("service", gameService != nil))

	log.Info("Game management service initialized and ready")
	log.Info("Consolidated repositories working correctly")

	// Show that the service is working by testing it
	ctx := context.Background()
	testGame, err := gameService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		log.Error("Failed to create test game", zap.Error(err))
	} else {
		log.Info("Test game created", zap.String("game_id", testGame.ID))
	}

	// Initialize WebSocket service with shared Hub
	webSocketService := wsHandler.NewWebSocketService(gameService, playerService, standardProjectService, cardService, gameRepo, playerRepo, hub)

	// Start WebSocket service in background
	wsCtx, wsCancel := context.WithCancel(ctx)
	defer wsCancel()
	go webSocketService.Run(wsCtx)
	log.Info("WebSocket hub started")

	// Setup main router without middleware for WebSocket
	mainRouter := mux.NewRouter()

	// Setup API router with middleware
	apiRouter := httpHandler.SetupRouter(gameService, playerService, cardService)

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
		log.Info("Starting HTTP server on :3001")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	log.Info("✅ Server started")
	log.Info("🌍 HTTP server listening on :3001")
	log.Info("🔌 WebSocket endpoint available at /ws")

	// Wait for shutdown signal
	<-quit

	log.Info("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Failed to gracefully shutdown HTTP server", zap.Error(err))
	} else {
		log.Info("✅ HTTP server stopped")
	}

	// Cancel WebSocket service context
	wsCancel()
	log.Info("✅ WebSocket service stopped")

	log.Info("✅ Server shutdown complete")
}
