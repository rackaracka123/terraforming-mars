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
	"terraforming-mars-backend/internal/session/game/board"
	"terraforming-mars-backend/internal/session/game/card"
	sessionCard "terraforming-mars-backend/internal/session/game/card"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/game/deck"

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
	// Initialize repositories first (needed by actions during migration)
	// NOTE: These are deprecated facades - Phase 4-5 will migrate actions to use Session directly
	newGameRepo := game.NewRepository(eventBus)
	newCardRepo := sessionCard.NewRepository(newDeckRepo) // Use NEW deck repository

	// Create shared board storage and global board repository for actions (temporary)
	// SessionFactory will create game-scoped repositories internally
	sharedBoards := make(map[string]*board.Board)
	newBoardRepo := board.NewRepository("", sharedBoards, eventBus) // Empty gameID = global instance
	log.Info("🗺️  Repositories initialized (facade for backwards compatibility)")

	// Initialize SessionFactory with card and deck repositories
	// SessionFactory manages game-scoped repositories internally (game, board, deck per game)
	sessionFactory := session.NewSessionFactory(eventBus, newCardRepo, newDeckRepo)
	log.Info("🎮 SessionFactory initialized - manages game-scoped repositories internally")

	// Initialize BoardProcessor for hex calculations
	boardProcessor := board.NewBoardProcessor()
	log.Info("🎲 Board processor initialized")

	// Initialize TileProcessor for tile queue processing
	tileProcessor := board.NewProcessor(newGameRepo, newBoardRepo, boardProcessor)
	log.Info("🎯 Tile processor initialized")

	// Initialize BroadcasterFactory (creates session-aware SessionManagers)
	// Each game gets its own SessionManager instance bound to that specific gameID
	// Factory subscribes to domain events and automatically broadcasts on state changes
	sessionManagerFactory := wsHandler.NewBroadcasterFactory(newGameRepo, sessionFactory, newCardRepo, newBoardRepo, hub, eventBus)
	log.Info("📡 BroadcasterFactory initialized and subscribed to domain events")

	// Initialize actions with SessionManagerFactory
	// Actions will call sessionManagerFactory.GetOrCreate(gameID) to get game-specific broadcasters
	startGameAction := action.NewStartGameAction(newGameRepo, newCardRepo, newDeckRepo, sessionManagerFactory)
	createGameAction := action.NewCreateGameAction(newGameRepo, newBoardRepo)
	joinGameAction := action.NewJoinGameAction(newGameRepo, sessionFactory)
	playerReconnectedAction := action.NewPlayerReconnectedAction(sessionManagerFactory)
	playerDisconnectedAction := action.NewPlayerDisconnectedAction(sessionManagerFactory)
	selectStartingCardsAction := action.NewSelectStartingCardsAction(newGameRepo, newCardRepo, sessionManagerFactory)
	skipActionAction := action.NewSkipActionAction(newGameRepo, newDeckRepo, sessionManagerFactory)
	confirmProductionCardsAction := action.NewConfirmProductionCardsAction(newGameRepo, sessionManagerFactory)
	buildCityAction := action.NewBuildCityAction(newGameRepo, tileProcessor, sessionManagerFactory)

	// Initialize BonusCalculator for tile placement bonuses
	bonusCalculator := board.NewBonusCalculator(newGameRepo, newBoardRepo, newDeckRepo)
	log.Info("🎁 Bonus calculator initialized")

	// Initialize SelectTileAction for tile placement
	selectTileAction := action.NewSelectTileAction(newGameRepo, newBoardRepo, tileProcessor, bonusCalculator, sessionManagerFactory)

	log.Info("🎯 New architecture initialized: start_game, create_game, join_game, player_reconnected, player_disconnected, select_starting_cards, skip_action, confirm_production_cards, build_city, select_tile actions ready")
	// ================================================================

	// Initialize services in dependency order
	// Initialize card effect subscriber for passive effects (session-scoped)
	effectSubscriber := sessionCard.NewCardEffectSubscriber(eventBus, newCardRepo)
	log.Info("🎆 Card effect subscriber initialized (session-scoped)")

	// Initialize CardManager for card playing logic (session-based)
	// UPDATED: Now uses NEW deck repository instead of OLD CardDeckRepository
	cardManager := sessionCard.NewCardManager(newCardRepo, newDeckRepo, effectSubscriber)
	log.Info("🎴 Card manager initialized")

	// Initialize PlayCardAction for playing cards from hand
	playCardAction := action.NewPlayCardAction(newGameRepo, cardManager, tileProcessor, sessionManagerFactory)
	log.Info("✅ PlayCardAction initialized")

	// Initialize standard project actions
	launchAsteroidAction := action.NewLaunchAsteroidAction(newGameRepo, sessionManagerFactory)
	buildPowerPlantAction := action.NewBuildPowerPlantAction(newGameRepo, sessionManagerFactory)
	buildAquiferAction := action.NewBuildAquiferAction(newGameRepo, sessionManagerFactory)
	plantGreeneryAction := action.NewPlantGreeneryAction(newGameRepo, sessionManagerFactory)
	sellPatentsAction := action.NewSellPatentsAction(newGameRepo, sessionManagerFactory)
	confirmSellPatentsAction := action.NewConfirmSellPatentsAction(newGameRepo, sessionManagerFactory)
	log.Info("✅ Standard project actions initialized")

	// Initialize resource conversion actions
	convertHeatAction := action.NewConvertHeatToTemperatureAction(newGameRepo, sessionManagerFactory)
	convertPlantsAction := action.NewConvertPlantsToGreeneryAction(newGameRepo, sessionManagerFactory)
	log.Info("✅ Resource conversion actions initialized")

	// Initialize forced action manager for corporation forced first actions
	// UPDATED: Now uses NEW session repositories to match event sources
	forcedActionManager := card.NewForcedActionManager(eventBus, newCardRepo, newGameRepo, newDeckRepo)
	forcedActionManager.SubscribeToPhaseChanges()
	forcedActionManager.SubscribeToCardDrawEvents()
	log.Info("🎯 Forced action manager initialized and subscribed to events (phase changes + card draw confirmations)")

	// Initialize card selection confirmation actions
	confirmCardDrawAction := action.NewConfirmCardDrawAction(newGameRepo, sessionManagerFactory, eventBus)
	log.Info("✅ Card selection confirmation actions initialized")

	// Initialize admin actions
	giveCardAdminAction := adminaction.NewGiveCardAction(newGameRepo, newCardRepo, sessionManagerFactory, sessionFactory)
	setPhaseAdminAction := adminaction.NewSetPhaseAction(newGameRepo, sessionManagerFactory)
	setResourcesAdminAction := adminaction.NewSetResourcesAction(newGameRepo, sessionManagerFactory)
	setProductionAdminAction := adminaction.NewSetProductionAction(newGameRepo, sessionManagerFactory)
	setGlobalParametersAdminAction := adminaction.NewSetGlobalParametersAction(newGameRepo, sessionManagerFactory)
	startTileSelectionAdminAction := adminaction.NewStartTileSelectionAction(newGameRepo, newBoardRepo, boardProcessor, sessionManagerFactory)
	setCurrentTurnAdminAction := adminaction.NewSetCurrentTurnAction(newGameRepo, sessionManagerFactory)
	setCorporationAdminAction := adminaction.NewSetCorporationAction(newGameRepo, sessionManagerFactory, sessionFactory)
	log.Info("✅ Admin actions initialized")

	// Initialize query actions for HTTP handlers
	getGameAction := queryaction.NewGetGameAction(newGameRepo, newCardRepo)
	listGamesAction := queryaction.NewListGamesAction(newGameRepo)
	getPlayerAction := queryaction.NewGetPlayerAction(newGameRepo)
	listCardsAction := queryaction.NewListCardsAction(newCardRepo)
	getCorporationsAction := queryaction.NewGetCorporationsAction(newCardRepo)
	log.Info("✅ Query actions initialized for HTTP handlers")

	// Initialize CardProcessor for card action execution
	cardProcessor := sessionCard.NewCardProcessor(newDeckRepo)
	log.Info("🎴 Card processor initialized")

	// Initialize card action execution action (fully migrated to session-based architecture)
	executeCardActionAction := executecardaction.NewExecuteCardActionAction(newGameRepo, sessionManagerFactory, sessionFactory, cardProcessor, newDeckRepo)
	log.Info("✅ Execute card action action fully migrated to session-based architecture")

	// Initialize WebSocket service with shared Hub and new actions
	ctx := context.Background()
	webSocketService := wsHandler.NewWebSocketService(
		newGameRepo, sessionFactory, hub, sessionManagerFactory,
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
		sessionFactory,
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
