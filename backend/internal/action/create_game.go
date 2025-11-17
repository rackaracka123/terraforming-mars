package action

import (
	"context"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/session/game"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateGameAction handles the business logic for creating new games
type CreateGameAction struct {
	gameRepo game.Repository
	logger   *zap.Logger
}

// NewCreateGameAction creates a new create game action
func NewCreateGameAction(gameRepo game.Repository) *CreateGameAction {
	return &CreateGameAction{
		gameRepo: gameRepo,
		logger:   logger.Get(),
	}
}

// Execute performs the create game action
func (a *CreateGameAction) Execute(ctx context.Context, settings game.GameSettings) (*game.Game, error) {
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
	newGame := game.NewGame(gameID, settings)

	// 4. Initialize board (simplified - just create empty board)
	// In full implementation, we'd populate tiles here
	newGame.Board = model.Board{Tiles: []model.Tile{}}

	// 5. Store game in repository
	err := a.gameRepo.Create(ctx, newGame)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		return nil, err
	}

	log.Info("✅ Game created successfully", zap.String("game_id", gameID))
	return newGame, nil
}
