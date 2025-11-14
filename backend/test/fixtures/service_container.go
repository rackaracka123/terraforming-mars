package fixtures

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/production"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/features/turn"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"
)

// ServiceContainer holds all service dependencies for testing
// This centralizes dependency injection to eliminate boilerplate in tests
type ServiceContainer struct {
	// Core infrastructure
	EventBus *events.EventBusImpl

	// Repositories
	GameRepo     game.Repository
	PlayerRepo   player.Repository
	CardRepo     game.CardRepository
	CardDeckRepo game.CardDeckRepository

	// Services
	BoardService           service.BoardService
	CardService            service.CardService
	PlayerService          service.PlayerService
	GameService            service.GameService
	LobbyService           lobby.Service
	StandardProjectService service.StandardProjectService

	// Card system
	EffectSubscriber    cards.CardEffectSubscriber
	ForcedActionManager cards.ForcedActionManager

	// Mocks
	SessionManager *test.MockSessionManager
}

// NewServiceContainer creates a complete service dependency graph for testing
func NewServiceContainer() *ServiceContainer {
	// Core infrastructure
	eventBus := events.NewEventBus()

	// Repositories
	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)
	cardRepo := game.NewCardRepository()
	cardDeckRepo := game.NewCardDeckRepository()

	// Mocks
	sessionManager := test.NewMockSessionManager()

	// Services (build dependency chain)
	boardService := service.NewBoardService()
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	forcedActionManager.SubscribeToPhaseChanges()

	// Initialize game features
	tilesRepo := tiles.NewRepository(gameRepo, playerRepo)
	tilesBoardAdapter := tiles.NewBoardServiceAdapter(boardService)
	tilesFeature := tiles.NewService(tilesRepo, tilesBoardAdapter, eventBus)
	turnRepo := turn.NewRepository(gameRepo, playerRepo)
	turnFeature := turn.NewService(turnRepo)
	productionRepo := production.NewRepository(gameRepo, playerRepo, cardDeckRepo)
	productionFeature := production.NewService(productionRepo)
	parametersRepo := parameters.NewRepository(gameRepo, playerRepo)
	parametersFeature := parameters.NewService(parametersRepo)

	// Create services that depend on mechanics
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tilesFeature, effectSubscriber, forcedActionManager)
	playerService := service.NewPlayerService(gameRepo, playerRepo, sessionManager, tilesFeature, parametersFeature, forcedActionManager)
	gameService := service.NewGameService(gameRepo, playerRepo, cardRepo, cardService, cardDeckRepo, boardService, sessionManager, turnFeature, productionFeature, tilesFeature)
	lobbyService := lobby.NewService(gameRepo, playerRepo, cardRepo, cardService, cardDeckRepo, boardService, sessionManager)
	standardProjectService := service.NewStandardProjectService(gameRepo, playerRepo, sessionManager, tilesFeature)

	return &ServiceContainer{
		EventBus:               eventBus,
		GameRepo:               gameRepo,
		PlayerRepo:             playerRepo,
		CardRepo:               cardRepo,
		CardDeckRepo:           cardDeckRepo,
		BoardService:           boardService,
		CardService:            cardService,
		PlayerService:          playerService,
		GameService:            gameService,
		LobbyService:           lobbyService,
		StandardProjectService: standardProjectService,
		EffectSubscriber:       effectSubscriber,
		ForcedActionManager:    forcedActionManager,
		SessionManager:         sessionManager,
	}
}
