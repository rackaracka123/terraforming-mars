package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"terraforming-mars-backend/internal/action"
	adminaction "terraforming-mars-backend/internal/action/admin"
	executecardaction "terraforming-mars-backend/internal/action/execute_card_action"
	queryaction "terraforming-mars-backend/internal/action/query"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	wsHandler "terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	httpmiddleware "terraforming-mars-backend/internal/middleware/http"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/board"
	"terraforming-mars-backend/internal/session/card"
	sessionCard "terraforming-mars-backend/internal/session/card"
	"terraforming-mars-backend/internal/session/deck"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/tile"

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

	// Initialize event bus for domain events
	eventBus := events.NewEventBus()
	log.Info("🎆 Event bus initialized")

	// Initialize NEW deck repository with card loading
	newDeckRepo, err := deck.NewRepository(context.Background())
	if err != nil {
		log.Fatal("Failed to initialize deck repository", zap.Error(err))
	}
	log.Info("🎴 NEW deck repository initialized with card definitions")

	// Create Hub first (no dependencies)
	hub := core.NewHub()

	// ========== NEW ARCHITECTURE: Initialize Action Pattern ==========
	// Initialize SessionFactory for game-scoped sessions
	sessionFactory := session.NewSessionFactory(eventBus)
	log.Info("🎮 SessionFactory initialized")

	// Initialize legacy repositories for components not yet migrated
	newGameRepo := game.NewRepository(eventBus)
	newCardRepo := sessionCard.NewRepository(newDeckRepo) // Use NEW deck repository
	newBoardRepo := board.NewRepository(eventBus)
	log.Info("🗺️  Legacy repositories initialized")

	// Initialize BoardProcessor for hex calculations
	boardProcessor := board.NewBoardProcessor()
	log.Info("🎲 Board processor initialized")

	// Initialize TileProcessor for tile queue processing
	tileProcessor := tile.NewProcessor(newGameRepo, newPlayerRepo, newBoardRepo, boardProcessor, eventBus)
	log.Info("🎯 Tile processor initialized")

	// Subscribe TileProcessor to events for automatic queue processing
	tileProcessor.SubscribeToEvents()
	log.Info("🎆 TileProcessor subscribed to TileQueueCreatedEvent")

	// Initialize BroadcasterFactory (creates session-aware SessionManagers)
	// Each game gets its own SessionManager instance bound to that specific gameID
	// Factory subscribes to domain events and automatically broadcasts on state changes
	sessionManagerFactory := wsHandler.NewBroadcasterFactory(newGameRepo, newPlayerRepo, newCardRepo, newBoardRepo, hub, eventBus)
	log.Info("📡 BroadcasterFactory initialized and subscribed to domain events")

	// Initialize actions with SessionManagerFactory
	// Actions will call sessionManagerFactory.GetOrCreate(gameID) to get game-specific broadcasters
	startGameAction := action.NewStartGameAction(newGameRepo, newPlayerRepo, newCardRepo, newDeckRepo, sessionManagerFactory)
	createGameAction := action.NewCreateGameAction(newGameRepo, newBoardRepo)
	joinGameAction := action.NewJoinGameAction(newGameRepo, newPlayerRepo) // Event-driven: no SessionManager needed
	playerReconnectedAction := action.NewPlayerReconnectedAction(sessionFactory, sessionManagerFactory)
	playerDisconnectedAction := action.NewPlayerDisconnectedAction(sessionFactory, sessionManagerFactory)
	selectStartingCardsAction := action.NewSelectStartingCardsAction(newGameRepo, newPlayerRepo, newCardRepo, sessionManagerFactory)
	skipActionAction := action.NewSkipActionAction(newGameRepo, newPlayerRepo, newDeckRepo, sessionManagerFactory)
	confirmProductionCardsAction := action.NewConfirmProductionCardsAction(newGameRepo, newPlayerRepo, newCardRepo, sessionManagerFactory)
	buildCityAction := action.NewBuildCityAction(newGameRepo, newPlayerRepo, tileProcessor, sessionManagerFactory)

	// Initialize BonusCalculator for tile placement bonuses
	bonusCalculator := tile.NewBonusCalculator(newGameRepo, newPlayerRepo, newBoardRepo, newDeckRepo)
	log.Info("🎁 Bonus calculator initialized")

	// Initialize SelectTileAction for tile placement
	selectTileAction := action.NewSelectTileAction(newGameRepo, newPlayerRepo, newBoardRepo, tileProcessor, bonusCalculator, sessionManagerFactory)

	log.Info("🎯 New architecture initialized: start_game, create_game, join_game, player_reconnected, player_disconnected, select_starting_cards, skip_action, confirm_production_cards, build_city, select_tile actions ready")
	// ================================================================

	// Initialize services in dependency order
	// Initialize card effect subscriber for passive effects (session-scoped)
	effectSubscriber := sessionCard.NewCardEffectSubscriber(eventBus, newPlayerRepo, newGameRepo, newCardRepo)
	log.Info("🎆 Card effect subscriber initialized (session-scoped)")

	// Initialize CardManager for card playing logic (session-based)
	// UPDATED: Now uses NEW deck repository instead of OLD CardDeckRepository
	cardManager := sessionCard.NewCardManager(newGameRepo, newPlayerRepo, newCardRepo, newDeckRepo, effectSubscriber)
	log.Info("🎴 Card manager initialized")

	// Initialize PlayCardAction for playing cards from hand
	playCardAction := action.NewPlayCardAction(newGameRepo, newPlayerRepo, cardManager, tileProcessor, sessionManagerFactory)
	log.Info("✅ PlayCardAction initialized")

	// Initialize standard project actions
	launchAsteroidAction := action.NewLaunchAsteroidAction(newGameRepo, sessionFactory, sessionManagerFactory)
	buildPowerPlantAction := action.NewBuildPowerPlantAction(newGameRepo, sessionFactory, sessionManagerFactory)
	buildAquiferAction := action.NewBuildAquiferAction(newGameRepo, sessionFactory, sessionManagerFactory)
	plantGreeneryAction := action.NewPlantGreeneryAction(newGameRepo, sessionFactory, sessionManagerFactory)
	sellPatentsAction := action.NewSellPatentsAction(newGameRepo, sessionFactory, sessionManagerFactory)
	confirmSellPatentsAction := action.NewConfirmSellPatentsAction(newGameRepo, sessionFactory, sessionManagerFactory)
	log.Info("✅ Standard project actions initialized")

	// Initialize resource conversion actions
	convertHeatAction := action.NewConvertHeatToTemperatureAction(newGameRepo, sessionFactory, sessionManagerFactory)
	convertPlantsAction := action.NewConvertPlantsToGreeneryAction(newGameRepo, sessionFactory, sessionManagerFactory)
	log.Info("✅ Resource conversion actions initialized")

	// Initialize forced action manager for corporation forced first actions
	// UPDATED: Now uses NEW session repositories (newPlayerRepo, newGameRepo, newCardRepo) to match event sources
	forcedActionManager := card.NewForcedActionManager(eventBus, newCardRepo, newPlayerRepo, newGameRepo, newDeckRepo)
	forcedActionManager.SubscribeToPhaseChanges()
	forcedActionManager.SubscribeToCardDrawEvents()
	log.Info("🎯 Forced action manager initialized and subscribed to events (phase changes + card draw confirmations)")

	// Initialize card selection confirmation actions
	confirmCardDrawAction := action.NewConfirmCardDrawAction(newGameRepo, newPlayerRepo, sessionManagerFactory, eventBus)
	log.Info("✅ Card selection confirmation actions initialized")

	// Initialize admin actions
	giveCardAdminAction := adminaction.NewGiveCardAction(newGameRepo, newPlayerRepo, newCardRepo, sessionManagerFactory)
	setPhaseAdminAction := adminaction.NewSetPhaseAction(newGameRepo, newPlayerRepo, sessionManagerFactory)
	setResourcesAdminAction := adminaction.NewSetResourcesAction(newGameRepo, newPlayerRepo, sessionManagerFactory)
	setProductionAdminAction := adminaction.NewSetProductionAction(newGameRepo, newPlayerRepo, sessionManagerFactory)
	setGlobalParametersAdminAction := adminaction.NewSetGlobalParametersAction(newGameRepo, newPlayerRepo, sessionManagerFactory)
	startTileSelectionAdminAction := adminaction.NewStartTileSelectionAction(newGameRepo, newPlayerRepo, newBoardRepo, boardProcessor, sessionManagerFactory)
	setCurrentTurnAdminAction := adminaction.NewSetCurrentTurnAction(newGameRepo, newPlayerRepo, sessionManagerFactory)
	setCorporationAdminAction := adminaction.NewSetCorporationAction(newGameRepo, newPlayerRepo, sessionManagerFactory)
	log.Info("✅ Admin actions initialized")

	// Initialize query actions for HTTP handlers
	getGameAction := queryaction.NewGetGameAction(newGameRepo, newPlayerRepo, newCardRepo, sessionManagerFactory)
	listGamesAction := queryaction.NewListGamesAction(newGameRepo, newPlayerRepo, sessionManagerFactory)
	getPlayerAction := queryaction.NewGetPlayerAction(newGameRepo, newPlayerRepo, sessionManagerFactory)
	listCardsAction := queryaction.NewListCardsAction(newGameRepo, newPlayerRepo, newCardRepo, sessionManagerFactory)
	getCorporationsAction := queryaction.NewGetCorporationsAction(newGameRepo, newPlayerRepo, newCardRepo, sessionManagerFactory)
	log.Info("✅ Query actions initialized for HTTP handlers")

	// Initialize CardProcessor for card action execution
	cardProcessor := sessionCard.NewCardProcessor(newGameRepo, newPlayerRepo, newDeckRepo)
	log.Info("🎴 Card processor initialized")

	// Initialize card action execution action (fully migrated to session-based architecture)
	executeCardActionAction := executecardaction.NewExecuteCardActionAction(newGameRepo, newPlayerRepo, sessionManagerFactory, cardProcessor, newDeckRepo)
	log.Info("✅ Execute card action action fully migrated to session-based architecture")

	// Initialize WebSocket service with shared Hub and new actions
	ctx := context.Background()
	webSocketService := wsHandler.NewWebSocketService(
		newGameRepo, newPlayerRepo, hub, sessionManagerFactory,
		startGameAction, joinGameAction, playerReconnectedAction, playerDisconnectedAction, selectStartingCardsAction, skipActionAction, confirmProductionCardsAction,
		buildCityAction, selectTileAction, playCardAction, executeCardActionAction,
		launchAsteroidAction, buildPowerPlantAction, buildAquiferAction, plantGreeneryAction,
		sellPatentsAction, confirmSellPatentsAction,
		convertHeatAction, convertPlantsAction,
		confirmCardDrawAction,
		giveCardAdminAction, setPhaseAdminAction, setResourcesAdminAction, setProductionAdminAction,
		setGlobalParametersAdminAction, startTileSelectionAdminAction, setCurrentTurnAdminAction, setCorporationAdminAction,
	)

	// Start WebSocket service in background
	wsCtx, wsCancel := context.WithCancel(ctx)
	defer wsCancel()
	go webSocketService.Run(wsCtx)
	log.Info("WebSocket hub started")

	// Setup main router with CORS middleware
	mainRouter := mux.NewRouter()
	mainRouter.Use(httpmiddleware.CORS) // Apply CORS to all routes (API + WebSocket)

	// Setup API router with middleware
	apiRouter := httpHandler.SetupRouter(
		createGameAction,
		joinGameAction,
		getGameAction,
		listGamesAction,
		getPlayerAction,
		listCardsAction,
		getCorporationsAction,
	)

	// Mount API router
	mainRouter.PathPrefix("/api/v1").Handler(apiRouter)

	// Add WebSocket endpoint to main router (inherits CORS from mainRouter)
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
