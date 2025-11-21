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

	// OLD Repositories (being phased out)
	GameRepo     repository.GameRepository
	PlayerRepo   repository.PlayerRepository
	CardRepo     repository.CardRepository
	CardDeckRepo repository.CardDeckRepository

	// NEW Session Repositories
	NewGameRepo   sessionGame.Repository
	NewPlayerRepo sessionPlayer.Repository
	NewCardRepo   sessionCard.Repository
	NewBoardRepo  board.Repository
	NewDeckRepo   sessionDeck.Repository

	// Services (still using old repositories)
	CardService   service.CardService
	PlayerService service.PlayerService
	GameService   service.GameService

	// Session-based components
	BoardProcessor      *board.BoardProcessor
	TileProcessor       *tile.Processor
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

	// NEW session-based components
	boardProcessor := board.NewBoardProcessor()
	tileProcessor := tile.NewProcessor(newGameRepo, newPlayerRepo, newBoardRepo, boardProcessor, eventBus)

	// Use NEW session repositories for card system components
	effectSubscriber := card.NewCardEffectSubscriber(eventBus, newPlayerRepo, newGameRepo, newCardRepo)
	forcedActionManager := card.NewForcedActionManager(eventBus, cardRepo, newPlayerRepo, newGameRepo, newDeckRepo)
	forcedActionManager.SubscribeToPhaseChanges()

	// Services (still using mixed repositories during migration)
	cardService := service.NewCardService(newGameRepo, newPlayerRepo, newCardRepo, newDeckRepo, sessionManager, tileProcessor, effectSubscriber, forcedActionManager)
	playerService := service.NewPlayerService(gameRepo, playerRepo, sessionManager, forcedActionManager, eventBus)
	gameService := service.NewGameService(gameRepo, playerRepo, cardRepo, cardService, newDeckRepo, sessionManager)

	return &ServiceContainer{
		EventBus:            eventBus,
		GameRepo:            gameRepo,
		PlayerRepo:          playerRepo,
		CardRepo:            cardRepo,
		CardDeckRepo:        cardDeckRepo,
		NewGameRepo:         newGameRepo,
		NewPlayerRepo:       newPlayerRepo,
		NewCardRepo:         newCardRepo,
		NewBoardRepo:        newBoardRepo,
		NewDeckRepo:         newDeckRepo,
		CardService:         cardService,
		PlayerService:       playerService,
		GameService:         gameService,
		BoardProcessor:      boardProcessor,
		TileProcessor:       tileProcessor,
		EffectSubscriber:    effectSubscriber,
		ForcedActionManager: forcedActionManager,
		SessionManager:      sessionManager,
	}
}
