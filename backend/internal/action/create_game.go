package action

import (
	"context"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateGameAction handles the business logic for creating new games
// MIGRATION: Uses new architecture (GameRepository only, no board repository)
type CreateGameAction struct {
	gameRepo game.GameRepository
	eventBus *events.EventBusImpl
	logger   *zap.Logger
}

// NewCreateGameAction creates a new create game action
func NewCreateGameAction(
	gameRepo game.GameRepository,
	eventBus *events.EventBusImpl,
	logger *zap.Logger,
) *CreateGameAction {
	return &CreateGameAction{
		gameRepo: gameRepo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Execute performs the create game action
func (a *CreateGameAction) Execute(
	ctx context.Context,
	settings game.GameSettings,
) (*game.Game, error) {
	log := a.logger.With(
		zap.Int("max_players", settings.MaxPlayers),
		zap.Strings("card_packs", settings.CardPacks),
	)
	log.Info("🎮 Creating new game")

	// 1. Generate game ID
	gameID := uuid.New().String()

	// 2. Apply default settings
	if settings.MaxPlayers == 0 {
		settings.MaxPlayers = game.DefaultMaxPlayers
	}
	if len(settings.CardPacks) == 0 {
		settings.CardPacks = game.DefaultCardPacks()
	}

	// 3. Create game entity
	// Note: hostPlayerID is empty initially, will be set when first player joins
	// Board is automatically created by NewGame
	newGame := game.NewGame(gameID, "", settings, a.eventBus)

	// 4. Store game in repository
	err := a.gameRepo.Create(ctx, newGame)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		return nil, err
	}

	log.Info("✅ Game created successfully with board", zap.String("game_id", gameID))
	return newGame, nil
}
