package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"terraforming-mars-backend/internal/cards"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	wsHandler "terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/actions"
	"terraforming-mars-backend/internal/actions/card_selection"
	"terraforming-mars-backend/internal/actions/standard_projects"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/production"
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/features/turn"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/player"
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

	// Initialize event bus for domain events
	eventBus := events.NewEventBus()
	log.Info("🎆 Event bus initialized")

	// Initialize individual repositories with event bus
	playerRepo := player.NewRepository(eventBus)
	log.Info("Player repository initialized")

	gameRepo := game.NewRepository(eventBus)
	log.Info("Game repository initialized")

	// Initialize card repository and load cards
	cardRepo := game.NewCardRepository()
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
	cardDeckRepo := game.NewCardDeckRepository()

	// Create Hub first (no dependencies)
	hub := core.NewHub()

	// Initialize SessionManager for WebSocket broadcasting with Hub
	sessionManager := session.NewSessionManager(gameRepo, playerRepo, cardRepo, hub)

	// Initialize services in dependency order
	boardService := service.NewBoardService()

	// Initialize card effect subscriber for passive effects
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	log.Info("🎆 Card effect subscriber initialized")

	// Initialize forced action manager for corporation forced first actions
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	forcedActionManager.SubscribeToPhaseChanges()
	log.Info("🎯 Forced action manager initialized and subscribed to phase changes")

	// Initialize card manager for card validation and playing
	cardManager := cards.NewCardManager(gameRepo, playerRepo, cardRepo, cardDeckRepo, effectSubscriber)
	log.Info("🃏 Card manager initialized")

	// Initialize selection manager for card selection operations
	selectionManager := cards.NewSelectionManager(gameRepo, playerRepo, cardRepo, cardDeckRepo, effectSubscriber)
	log.Info("📋 Selection manager initialized")

	// Initialize game features (isolated, self-contained modules)
	// Create feature-specific repositories
	resourcesRepo := resources.NewRepository(playerRepo)
	parametersRepo := parameters.NewRepository(gameRepo, playerRepo)
	tilesRepo := tiles.NewRepository(gameRepo, playerRepo)
	turnRepo := turn.NewRepository(gameRepo, playerRepo)
	productionRepo := production.NewRepository(gameRepo, playerRepo, cardDeckRepo)

	// Create feature services
	resourcesFeature := resources.NewService(resourcesRepo)
	parametersFeature := parameters.NewService(parametersRepo)
	tilesBoardAdapter := tiles.NewBoardServiceAdapter(boardService)
	tilesFeature := tiles.NewService(tilesRepo, tilesBoardAdapter, eventBus)
	turnFeature := turn.NewService(turnRepo)
	productionFeature := production.NewService(productionRepo)
	log.Info("🔧 Game features initialized (resources, parameters, tiles, turn, production)")

	// CardService needs tilesFeature for tile queue processing and effect subscriber for passive effects
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tilesFeature, effectSubscriber, forcedActionManager)
	log.Info("SessionManager initialized for service-level broadcasting")

	// Initialize actions (orchestration layer)
	buildAquiferAction := standard_projects.NewBuildAquiferAction(
		playerRepo,
		gameRepo,
		resourcesFeature,
		parametersFeature,
		tilesFeature,
		sessionManager,
	)
	launchAsteroidAction := standard_projects.NewLaunchAsteroidAction(
		playerRepo,
		gameRepo,
		resourcesFeature,
		parametersFeature,
		sessionManager,
	)
	buildPowerPlantAction := standard_projects.NewBuildPowerPlantAction(
		playerRepo,
		gameRepo,
		resourcesFeature,
		sessionManager,
	)
	plantGreeneryAction := standard_projects.NewPlantGreeneryAction(
		playerRepo,
		gameRepo,
		resourcesFeature,
		parametersFeature,
		tilesFeature,
		sessionManager,
	)
	buildCityAction := standard_projects.NewBuildCityAction(
		playerRepo,
		gameRepo,
		resourcesFeature,
		tilesFeature,
		sessionManager,
	)
	skipAction := actions.NewSkipAction(
		turnFeature,
		productionFeature,
		sessionManager,
	)
	convertHeatAction := actions.NewConvertHeatToTemperatureAction(
		playerRepo,
		gameRepo,
		resourcesFeature,
		parametersFeature,
		sessionManager,
	)
	convertPlantsAction := actions.NewConvertPlantsToGreeneryAction(
		playerRepo,
		gameRepo,
		resourcesFeature,
		parametersFeature,
		tilesFeature,
		sessionManager,
	)

	// Card-related actions
	sellPatentsAction := standard_projects.NewSellPatentsAction(
		playerRepo,
		sessionManager,
	)
	playCardAction := actions.NewPlayCardAction(
		cardManager,
		tilesFeature,
		gameRepo,
		playerRepo,
		cardRepo,
		sessionManager,
	)
	selectTileAction := actions.NewSelectTileAction(
		playerRepo,
		gameRepo,
		tilesFeature,
		parametersFeature,
		forcedActionManager,
		sessionManager,
	)
	playCardActionAction := actions.NewPlayCardActionAction(
		cardService,
		gameRepo,
		playerRepo,
		sessionManager,
	)

	// Card selection actions (not needing gameService)
	submitSellPatentsAction := card_selection.NewSubmitSellPatentsAction(
		playerRepo,
		resourcesFeature,
		sessionManager,
	)
	confirmCardDrawAction := card_selection.NewConfirmCardDrawAction(
		playerRepo,
		resourcesFeature,
		forcedActionManager,
		sessionManager,
	)

	log.Info("🎬 Actions initialized (aquifer, asteroid, power plant, greenery, city, skip, convert heat, convert plants, sell patents, play card, select tile, play card action, submit sell patents, confirm card draw)")

	// PlayerService needs tilesFeature and parametersFeature for processing queues after tile placement
	playerService := service.NewPlayerService(gameRepo, playerRepo, sessionManager, tilesFeature, parametersFeature, forcedActionManager)

	gameService := service.NewGameService(gameRepo, playerRepo, cardRepo, cardService, cardDeckRepo, boardService, sessionManager, turnFeature, productionFeature, tilesFeature)

	// Card selection actions that need gameService
	selectStartingCardsAction := card_selection.NewSelectStartingCardsAction(
		selectionManager,
		gameRepo,
		sessionManager,
	)
	selectProductionCardsAction := card_selection.NewSelectProductionCardsAction(
		selectionManager,
		gameService,
		sessionManager,
	)
	log.Info("🎬 Card selection actions initialized (select starting cards, select production cards)")

	// Initialize lobby service for pre-game operations
	lobbyService := lobby.NewService(gameRepo, playerRepo, cardRepo, cardService, cardDeckRepo, boardService, sessionManager)
	log.Info("Lobby service initialized for game creation and joining")

	standardProjectService := service.NewStandardProjectService(gameRepo, playerRepo, sessionManager, tilesFeature)
	adminService := service.NewAdminService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, effectSubscriber, forcedActionManager)

	log.Info("Services initialized with new architecture and reconnection system")

	// Log service initialization
	log.Info("Player service ready", zap.Any("service", playerService != nil))
	log.Info("Game service ready", zap.Any("service", gameService != nil))

	log.Info("Game management service initialized and ready")
	log.Info("Consolidated repositories working correctly")

	// Show that the service is working by testing it
	ctx := context.Background()
	testGame, err := lobbyService.CreateGame(ctx, model.GameSettings{MaxPlayers: 4})
	if err != nil {
		log.Error("Failed to create test game", zap.Error(err))
	} else {
		log.Info("Test game created", zap.String("game_id", testGame.ID))
	}

	// Initialize WebSocket service with shared Hub
	webSocketService := wsHandler.NewWebSocketService(
		gameService,
		lobbyService,
		playerService,
		standardProjectService,
		cardService,
		adminService,
		gameRepo,
		playerRepo,
		cardRepo,
		hub,
		buildAquiferAction,
		launchAsteroidAction,
		buildPowerPlantAction,
		plantGreeneryAction,
		buildCityAction,
		skipAction,
		convertHeatAction,
		convertPlantsAction,
		sellPatentsAction,
		submitSellPatentsAction,
		selectStartingCardsAction,
		selectProductionCardsAction,
		confirmCardDrawAction,
		playCardAction,
		selectTileAction,
		playCardActionAction,
	)

	// Start WebSocket service in background
	wsCtx, wsCancel := context.WithCancel(ctx)
	defer wsCancel()
	go webSocketService.Run(wsCtx)
	log.Info("WebSocket hub started")

	// Setup main router without middleware for WebSocket
	mainRouter := mux.NewRouter()

	// Setup API router with middleware
	apiRouter := httpHandler.SetupRouter(gameService, lobbyService, playerService, cardService, playerRepo, cardRepo)

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
