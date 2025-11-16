package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"terraforming-mars-backend/internal/actions"
	"terraforming-mars-backend/internal/actions/card_selection"
	"terraforming-mars-backend/internal/actions/standard_projects"
	"terraforming-mars-backend/internal/admin"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	wsHandler "terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"
	sessionPkg "terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/subscriptions"

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

	// Initialize individual repositories with event bus
	playerRepo := player.NewRepository(eventBus)
	log.Info("Player repository initialized")

	gameRepo := game.NewRepository(eventBus)
	log.Info("Game repository initialized")

	// Initialize card repository and load cards
	cardRepo := card.NewCardRepository()
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
	cardDeckRepo := card.NewCardDeckRepository()

	// Create Hub first (no dependencies)
	hub := core.NewHub()

	// Initialize SessionRepository for runtime game sessions
	sessionRepo := sessionPkg.NewRepository()
	log.Info("🎮 Session repository initialized")

	// Initialize SessionManager for WebSocket broadcasting with Hub
	sessionManager := session.NewSessionManager(gameRepo, playerRepo, cardRepo, sessionRepo, hub)

	// Initialize BroadcastSubscriber for automatic event-driven broadcasting
	broadcastSubscriber := subscriptions.NewBroadcastSubscriber(sessionManager, eventBus)
	log.Info("📢 Automatic broadcast subscriber initialized")
	_ = broadcastSubscriber // Keep reference to prevent garbage collection

	// Initialize services in dependency order
	// Note: Feature repositories (board, parameters) are created per-game by game repository,
	// not as singletons here. They are scoped by gameID.

	// TODO: Initialize effect trigger service for reactive card effects
	// The subscriptions package needs to be recreated after restructuring
	// effectTriggerService := subscriptions.NewEffectTriggerService(eventBus, playerRepo, gameRepo)
	// effectTriggerService.Initialize()
	// log.Info("🎆 Effect trigger service initialized")

	// TODO: Re-initialize forced action manager after restructuring
	// forcedActionManager := actions.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	// forcedActionManager.SubscribeToPhaseChanges()
	log.Info("🎯 Forced action manager temporarily disabled during restructuring")

	// Initialize card feature services (effect processing, play, selection)
	// Note: ValidationService removed - validation happens inline in Actions per ARCHITECTURE_FLOW.md
	// Note: PlayService, DrawService, ActionService stubs removed - logic inlined in Actions per ARCHITECTURE_FLOW.md
	_ = card.NewEffectProcessor(cardRepo, playerRepo, gameRepo, cardDeckRepo)
	log.Info("✨ Card effect processor initialized")

	// Note: ResourceConversionService has been eliminated - logic moved into actions
	// Feature services (resources, parameters, tiles, etc.) are created per-game by
	// game repository, not as global singletons. Actions access them via repositories.

	// Initialize actions (orchestration layer)
	// Note: Feature services are now injected from session, not accessed via game objects
	// BuildAquiferAction needs parametersService and tilesService - these will be nil for now
	// and should be passed from session when actions are triggered
	buildAquiferAction := standard_projects.NewBuildAquiferAction(
		playerRepo,
		nil, // parametersService - will be injected from session
		nil, // placementService - will be injected from session
		sessionManager,
	)
	launchAsteroidAction := standard_projects.NewLaunchAsteroidAction(
		playerRepo,
		nil, // resourcesService - not yet implemented
		nil, // parametersService - will be injected from session
		sessionManager,
	)
	buildPowerPlantAction := standard_projects.NewBuildPowerPlantAction(
		playerRepo,
		nil, // resourcesService - not yet implemented
		sessionManager,
	)
	plantGreeneryAction := standard_projects.NewPlantGreeneryAction(
		playerRepo,
		gameRepo,
		nil, // placementService - will be injected from session
		sessionManager,
	)
	buildCityAction := standard_projects.NewBuildCityAction(
		playerRepo,
		nil, // resourcesService - not yet implemented
		nil, // tilesService - will be injected from session
		sessionManager,
	)
	skipAction := actions.NewSkipAction(
		gameRepo,
		sessionManager,
	)
	convertHeatAction := actions.NewConvertHeatToTemperatureAction(
		playerRepo,
		nil, // parametersService - will be injected from session
		sessionManager,
	)
	convertPlantsAction := actions.NewConvertPlantsToGreeneryAction(
		playerRepo,
		nil, // placementService - not yet initialized, actions will calculate hexes internally
		sessionManager,
	)

	// Card-related actions
	sellPatentsAction := standard_projects.NewSellPatentsAction(
		playerRepo,
		sessionManager,
	)
	playCardAction := actions.NewPlayCardAction(
		cardRepo,
		gameRepo,
		playerRepo,
		cardDeckRepo,
		nil, // parametersService - will be injected from session
		nil, // turnOrderService - will be injected from session
		sessionManager,
	)
	selectTileAction := actions.NewSelectTileAction(
		playerRepo,
		gameRepo,
		nil, // selectionService - accessed via session
		nil, // parametersService - accessed via session
		sessionManager,
	)
	playCardActionAction := actions.NewPlayCardActionAction(
		cardRepo,
		playerRepo,
		nil, // parametersService - will be injected from session
		nil, // turnOrderService - will be injected from session
		sessionManager,
	)

	// Card selection actions (not needing gameService)
	submitSellPatentsAction := card_selection.NewSubmitSellPatentsAction(
		playerRepo,
		sessionManager,
	)
	confirmCardDrawAction := card_selection.NewConfirmCardDrawAction(
		cardRepo,
		cardDeckRepo,
		playerRepo,
		sessionManager,
	)

	log.Info("🎬 Actions initialized (aquifer, asteroid, power plant, greenery, city, skip, convert heat, convert plants, sell patents, play card, select tile, play card action, submit sell patents, confirm card draw)")

	// NOTE: PlayerService and GameService have been eliminated.
	// All production code now uses repositories and actions directly.
	// Feature services are accessed via game/player objects when needed.

	// Card selection actions
	selectStartingCardsAction := card_selection.NewSelectStartingCardsAction(
		playerRepo,
		cardRepo,
		gameRepo,
		sessionManager,
	)
	selectProductionCardsAction := card_selection.NewSelectProductionCardsAction(
		playerRepo,
		cardRepo,
		sessionManager,
	)
	// Unified select cards action (routes between sell patents and production card selection)
	selectCardsAction := card_selection.NewSelectCardsAction(
		playerRepo,
		submitSellPatentsAction,
		selectProductionCardsAction,
	)
	log.Info("🎬 Card selection actions initialized (select starting cards, select production cards, select cards)")

	// Initialize lobby service for pre-game operations
	// Note: BoardService is created per-game in StartGame(), not as a global singleton
	lobbyService := lobby.NewService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, sessionRepo, eventBus)
	log.Info("Lobby service initialized for game creation and joining")

	// Connection management actions
	connectPlayerAction := actions.NewConnectPlayerAction(
		lobbyService,
		playerRepo,
		sessionManager,
	)
	disconnectPlayerAction := actions.NewDisconnectPlayerAction(
		playerRepo,
		sessionManager,
	)
	log.Info("🔗 Connection actions initialized (connect player, disconnect player)")
	log.Info("📋 All actions initialized successfully")

	adminService := admin.NewAdminService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, sessionRepo)

	log.Info("Services initialized with new architecture")
	log.Info("All repositories and actions initialized successfully")

	// Show that the service is working by testing it
	ctx := context.Background()
	testGame, err := lobbyService.CreateGame(ctx, game.GameSettings{MaxPlayers: 4})
	if err != nil {
		log.Error("Failed to create test game", zap.Error(err))
	} else {
		log.Info("Test game created", zap.String("game_id", testGame.ID))
	}

	// Initialize WebSocket service with shared Hub and all actions
	webSocketService := wsHandler.NewWebSocketService(
		lobbyService,
		adminService,
		gameRepo,
		cardRepo,
		hub,
		connectPlayerAction,
		disconnectPlayerAction,
		buildAquiferAction,
		launchAsteroidAction,
		buildPowerPlantAction,
		plantGreeneryAction,
		buildCityAction,
		sellPatentsAction,
		skipAction,
		convertHeatAction,
		convertPlantsAction,
		playCardAction,
		selectTileAction,
		playCardActionAction,
		submitSellPatentsAction,
		selectStartingCardsAction,
		selectProductionCardsAction,
		confirmCardDrawAction,
		selectCardsAction,
	)

	// Start WebSocket service in background
	wsCtx, wsCancel := context.WithCancel(ctx)
	defer wsCancel()
	go webSocketService.Run(wsCtx)
	log.Info("WebSocket hub started")

	// Setup main router without middleware for WebSocket
	mainRouter := mux.NewRouter()

	// Setup API router with middleware
	apiRouter := httpHandler.SetupRouter(lobbyService, cardRepo, playerRepo, gameRepo, cardRepo, sessionRepo)

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
