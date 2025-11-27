package action

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// BaseAction provides common dependencies for all migrated actions
// Following the new architecture: actions use ONLY GameRepository (+ logger + card registry)
// Broadcasting happens automatically via events published by Game methods
type BaseAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewBaseAction creates a new BaseAction with minimal dependencies
func NewBaseAction(gameRepo game.GameRepository, cardRegistry cards.CardRegistry) BaseAction {
	return BaseAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		logger:       logger.Get(),
	}
}

// InitLogger creates a logger with game and player context
// This should be called at the start of every Execute method
func (b *BaseAction) InitLogger(gameID, playerID string) *zap.Logger {
	return b.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
	)
}

// GetLogger returns the base logger for actions that don't have game/player context
func (b *BaseAction) GetLogger() *zap.Logger {
	return b.logger
}

// GameRepository returns the game repository
func (b *BaseAction) GameRepository() game.GameRepository {
	return b.gameRepo
}

// CardRegistry returns the card registry
func (b *BaseAction) CardRegistry() cards.CardRegistry {
	return b.cardRegistry
}
