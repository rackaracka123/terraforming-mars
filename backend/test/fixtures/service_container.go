package fixtures

import (
	"context"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/internal/session/board"
	"terraforming-mars-backend/internal/session/card"
	sessionCard "terraforming-mars-backend/internal/session/card"
	sessionDeck "terraforming-mars-backend/internal/session/deck"
	sessionGame "terraforming-mars-backend/internal/session/game"
	sessionPlayer "terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/tile"
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
	StandardProjectService service.StandardProjectService

	// Card system
	EffectSubscriber    card.CardEffectSubscriber
	ForcedActionManager card.ForcedActionManager

	// Mocks
	SessionManager *test.MockSessionManager
}

// NewServiceContainer creates a complete service dependency graph for testing
func NewServiceContainer() *ServiceContainer {
	// Core infrastructure
	eventBus := events.NewEventBus()

	// OLD Repositories (still used by some services during migration)
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()

	// NEW session repositories
	newDeckRepo, err := sessionDeck.NewRepository(context.Background())
	if err != nil {
		panic("Failed to initialize deck repository: " + err.Error())
	}
	newGameRepo := sessionGame.NewRepository(eventBus)
	newPlayerRepo := sessionPlayer.NewRepository(eventBus)
	newCardRepo := sessionCard.NewRepository(newDeckRepo, cardRepo)
	newBoardRepo := board.NewRepository(eventBus)

	// Mocks
	sessionManager := test.NewMockSessionManager()

	// Services (build dependency chain)
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)

	// NEW session-based tile processor
	boardProcessor := board.NewBoardProcessor()
	tileProcessor := tile.NewProcessor(newGameRepo, newPlayerRepo, newBoardRepo, boardProcessor)

	effectSubscriber := card.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo, cardRepo)
	forcedActionManager := card.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	forcedActionManager.SubscribeToPhaseChanges()
	cardService := service.NewCardService(newGameRepo, newPlayerRepo, newCardRepo, cardDeckRepo, sessionManager, tileProcessor, effectSubscriber, forcedActionManager)
	playerService := service.NewPlayerService(gameRepo, playerRepo, sessionManager, boardService, tileService, forcedActionManager, eventBus)
	gameService := service.NewGameService(gameRepo, playerRepo, cardRepo, cardService, cardDeckRepo, boardService, sessionManager)
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
		StandardProjectService: standardProjectService,
		EffectSubscriber:       effectSubscriber,
		ForcedActionManager:    forcedActionManager,
		SessionManager:         sessionManager,
	}
}
