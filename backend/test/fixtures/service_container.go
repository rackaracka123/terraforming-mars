package fixtures

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"
)

// ServiceContainer holds all service dependencies for testing
// This centralizes dependency injection to eliminate boilerplate in tests
type ServiceContainer struct {
	// Core infrastructure
	EventBus *events.EventBusImpl

	// Repositories
	GameRepo     repository.GameRepository
	PlayerRepo   repository.PlayerRepository
	CardRepo     repository.CardRepository
	CardDeckRepo repository.CardDeckRepository

	// Services
	BoardService           service.BoardService
	TileService            service.TileService
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
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()

	// Mocks
	sessionManager := test.NewMockSessionManager()

	// Services (build dependency chain)
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	forcedActionManager.SubscribeToPhaseChanges()
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService, effectSubscriber, forcedActionManager)
	playerService := service.NewPlayerService(gameRepo, playerRepo, sessionManager, boardService, tileService, forcedActionManager, eventBus)
	gameService := service.NewGameService(gameRepo, playerRepo, cardRepo, cardService, cardDeckRepo, boardService, sessionManager)
	lobbyService := lobby.NewService(gameRepo, playerRepo, cardRepo, cardService, cardDeckRepo, boardService, sessionManager)
	standardProjectService := service.NewStandardProjectService(gameRepo, playerRepo, sessionManager, tileService)

	return &ServiceContainer{
		EventBus:               eventBus,
		GameRepo:               gameRepo,
		PlayerRepo:             playerRepo,
		CardRepo:               cardRepo,
		CardDeckRepo:           cardDeckRepo,
		BoardService:           boardService,
		TileService:            tileService,
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
