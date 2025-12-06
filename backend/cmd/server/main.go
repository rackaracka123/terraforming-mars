package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"terraforming-mars-backend/internal/action"
	admin "terraforming-mars-backend/internal/action/admin"
	query "terraforming-mars-backend/internal/action/query"
	"terraforming-mars-backend/internal/cards"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	wsHandler "terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	httpmiddleware "terraforming-mars-backend/internal/middleware/http"

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

	// ========== Initialize Card Registry ==========
	// Get working directory to build absolute path
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working directory", zap.Error(err))
	}

	cardPath := filepath.Join(wd, "assets", "terraforming_mars_cards.json")
	log.Info("📂 Loading cards from", zap.String("path", cardPath))

	cardData, err := cards.LoadCardsFromJSON(cardPath)
	if err != nil {
		log.Fatal("Failed to load cards", zap.Error(err))
	}
	cardRegistry := cards.NewInMemoryCardRegistry(cardData)
	log.Info("🃏 Card registry initialized", zap.Int("card_count", len(cardData)))

	// ========== Initialize Game Repository (Single Source of Truth) ==========
	gameRepo := game.NewInMemoryGameRepository()
	log.Info("🎮 Game repository initialized")

	// ========== Initialize WebSocket Hub ==========
	hub := core.NewHub()
	log.Info("🔌 WebSocket hub initialized")

	// ========== Initialize Game State Broadcaster (Automatic Broadcasting) ==========
	broadcaster := wsHandler.NewBroadcaster(gameRepo, hub, cardRegistry)
	log.Info("📡 Game state broadcaster initialized (provides automatic broadcasting for all games)")

	// ========== Initialize Game Actions ==========

	// Game lifecycle (2)
	createGameAction := action.NewCreateGameAction(gameRepo, cardRegistry, log)
	joinGameAction := action.NewJoinGameAction(gameRepo, cardRegistry, log)

	// Card actions (2)
	playCardAction := action.NewPlayCardAction(gameRepo, cardRegistry, log)
	useCardActionAction := action.NewUseCardActionAction(gameRepo, cardRegistry, log)

	// Standard projects (6)
	launchAsteroidAction := action.NewLaunchAsteroidAction(gameRepo, log)
	buildPowerPlantAction := action.NewBuildPowerPlantAction(gameRepo, log)
	buildAquiferAction := action.NewBuildAquiferAction(gameRepo, log)
	buildCityAction := action.NewBuildCityAction(gameRepo, log)
	plantGreeneryAction := action.NewPlantGreeneryAction(gameRepo, log)
	sellPatentsAction := action.NewSellPatentsAction(gameRepo, log)

	// Resource conversions (2)
	convertHeatAction := action.NewConvertHeatToTemperatureAction(gameRepo, log)
	convertPlantsAction := action.NewConvertPlantsToGreeneryAction(gameRepo, log)

	// Tile selection (1)
	selectTileAction := action.NewSelectTileAction(gameRepo, log)

	// Turn management (3)
	startGameAction := action.NewStartGameAction(gameRepo, log)
	skipActionAction := action.NewSkipActionAction(gameRepo, log)
	selectStartingCardsAction := action.NewSelectStartingCardsAction(gameRepo, cardRegistry, log)

	// Confirmations (3)
	confirmSellPatentsAction := action.NewConfirmSellPatentsAction(gameRepo, log)
	confirmProductionCardsAction := action.NewConfirmProductionCardsAction(gameRepo, log)
	confirmCardDrawAction := action.NewConfirmCardDrawAction(gameRepo, log)

	// Connection management (2)
	playerReconnectedAction := action.NewPlayerReconnectedAction(gameRepo, log)
	playerDisconnectedAction := action.NewPlayerDisconnectedAction(gameRepo, log)

	// Admin actions (8)
	adminSetPhaseAction := admin.NewSetPhaseAction(gameRepo, log)
	adminSetCurrentTurnAction := admin.NewSetCurrentTurnAction(gameRepo, log)
	adminSetResourcesAction := admin.NewSetResourcesAction(gameRepo, log)
	adminSetProductionAction := admin.NewSetProductionAction(gameRepo, log)
	adminSetGlobalParametersAction := admin.NewSetGlobalParametersAction(gameRepo, log)
	adminGiveCardAction := admin.NewGiveCardAction(gameRepo, log)
	adminSetCorporationAction := admin.NewSetCorporationAction(gameRepo, log)
	adminStartTileSelectionAction := admin.NewStartTileSelectionAction(gameRepo, log)

	// Query actions for HTTP (4)
	getGameAction := query.NewGetGameAction(gameRepo, log)
	listGamesAction := query.NewListGamesAction(gameRepo, log)
	listCardsAction := query.NewListCardsAction(cardRegistry, log)
	getPlayerAction := query.NewGetPlayerAction(gameRepo, log)

	log.Info("✅ All migration actions initialized")
	log.Info("   📌 Game Lifecycle (2): CreateGame, JoinGame")
	log.Info("   📌 Card Actions (2): PlayCard, UseCardAction")
	log.Info("   📌 Standard Projects (6): LaunchAsteroid, BuildPowerPlant, BuildAquifer, BuildCity, PlantGreenery, SellPatents")
	log.Info("   📌 Resource Conversions (2): ConvertHeat, ConvertPlants")
	log.Info("   📌 Tile Selection (1): SelectTile")
	log.Info("   📌 Turn Management (3): StartGame, SkipAction, SelectStartingCards")
	log.Info("   📌 Confirmations (3): ConfirmSellPatents, ConfirmProductionCards, ConfirmCardDraw")
	log.Info("   📌 Connection Management (2): PlayerReconnected, PlayerDisconnected")
	log.Info("   📌 Admin Actions (8): SetPhase, SetCurrentTurn, SetResources, SetProduction, SetGlobalParameters, GiveCard, SetCorporation, StartTileSelection")
	log.Info("   📌 Query Actions (4): GetGame, ListGames, ListCards, GetPlayer")

	// ========== Register Migration Handlers with WebSocket Hub ==========
	wsHandler.RegisterHandlers(
		hub,
		broadcaster,
		// Game lifecycle
		createGameAction,
		joinGameAction,
		// Card actions
		playCardAction,
		useCardActionAction,
		// Standard projects
		launchAsteroidAction,
		buildPowerPlantAction,
		buildAquiferAction,
		buildCityAction,
		plantGreeneryAction,
		sellPatentsAction,
		// Resource conversions
		convertHeatAction,
		convertPlantsAction,
		// Tile selection
		selectTileAction,
		// Turn management
		startGameAction,
		skipActionAction,
		selectStartingCardsAction,
		// Confirmations
		confirmSellPatentsAction,
		confirmProductionCardsAction,
		confirmCardDrawAction,
		// Connection
		playerReconnectedAction,
		playerDisconnectedAction,
		// Admin actions
		adminSetPhaseAction,
		adminSetCurrentTurnAction,
		adminSetResourcesAction,
		adminSetProductionAction,
		adminSetGlobalParametersAction,
		adminGiveCardAction,
		adminSetCorporationAction,
		adminStartTileSelectionAction,
	)

	log.Info("🎯 Migration handlers registered with WebSocket hub (21 handlers)")

	// ========== Start WebSocket Hub ==========
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)
	log.Info("🔌 WebSocket hub running")

	// ========== Setup HTTP Router ==========
	mainRouter := mux.NewRouter()
	mainRouter.Use(httpmiddleware.CORS) // Apply CORS to all routes

	// Setup API router with migration actions
	apiRouter := httpHandler.SetupRouter(
		createGameAction,
		getGameAction,
		listGamesAction,
		listCardsAction,
		getPlayerAction,
		cardRegistry,
	)

	// Mount API router
	mainRouter.PathPrefix("/api/v1").Handler(apiRouter)

	// Create WebSocket handler
	wsHttpHandler := core.NewHandler(hub)

	// Add WebSocket endpoint
	mainRouter.HandleFunc("/ws", wsHttpHandler.ServeWS)

	log.Info("🌐 HTTP routes configured")
	log.Info("   📌 POST /api/v1/games - Create game")
	log.Info("   📌 GET  /api/v1/games - List games")
	log.Info("   📌 GET  /api/v1/games/{gameId} - Get game")
	log.Info("   📌 GET  /api/v1/cards - List cards")
	log.Info("   📌 GET  /api/v1/games/{gameId}/players/{playerId} - Get player")
	log.Info("   📌 WS   /ws - WebSocket endpoint")
	log.Info("   ℹ️  Game creation available via both HTTP POST and WebSocket 'create-game'")

	// ========== Setup HTTP Server ==========
	server := &http.Server{
		Addr:         ":3001",
		Handler:      mainRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in background
	go func() {
		log.Info("🌍 HTTP server listening on :3001")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	log.Info("✅ Server started successfully")
	log.Info("🎮 Using migration architecture - all old code removed")

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

	// Cancel WebSocket hub context
	cancel()
	log.Info("✅ WebSocket hub stopped")

	log.Info("✅ Server shutdown complete")
}
