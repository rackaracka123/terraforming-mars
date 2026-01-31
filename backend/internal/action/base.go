package action

import (
	"context"
	"fmt"
	"time"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// BaseAction provides common dependencies for all migrated actions
// Following the new architecture: actions use ONLY GameRepository (+ logger + card registry)
// Broadcasting happens automatically via events published by Game methods
type BaseAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	stateRepo    game.GameStateRepository
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

// NewBaseActionWithStateRepo creates a new BaseAction with state repository for logging
func NewBaseActionWithStateRepo(gameRepo game.GameRepository, cardRegistry cards.CardRegistry, stateRepo game.GameStateRepository) BaseAction {
	return BaseAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		stateRepo:    stateRepo,
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

// StateRepository returns the game state repository (may be nil)
func (b *BaseAction) StateRepository() game.GameStateRepository {
	return b.stateRepo
}

// WriteStateLog writes a state diff to the state repository if configured
func (b *BaseAction) WriteStateLog(ctx context.Context, g *game.Game, source string, sourceType game.SourceType, playerID, description string) {
	b.WriteStateLogWithChoice(ctx, g, source, sourceType, playerID, description, nil)
}

// WriteStateLogWithChoice writes a state diff with an optional choice index
func (b *BaseAction) WriteStateLogWithChoice(ctx context.Context, g *game.Game, source string, sourceType game.SourceType, playerID, description string, choiceIndex *int) {
	b.WriteStateLogWithChoiceAndOutputs(ctx, g, source, sourceType, playerID, description, choiceIndex, nil)
}

// WriteStateLogWithChoiceAndOutputs writes a state diff with optional choice index and calculated outputs
func (b *BaseAction) WriteStateLogWithChoiceAndOutputs(ctx context.Context, g *game.Game, source string, sourceType game.SourceType, playerID, description string, choiceIndex *int, calculatedOutputs []game.CalculatedOutput) {
	if b.stateRepo == nil {
		return
	}
	_, err := b.stateRepo.WriteWithChoiceAndOutputs(ctx, g.ID(), g, source, sourceType, playerID, description, choiceIndex, calculatedOutputs)
	if err != nil {
		b.logger.Warn("Failed to write state log",
			zap.String("game_id", g.ID()),
			zap.String("source", source),
			zap.Error(err))
	}
}

// GetPlayerFromGame fetches a player from the game with consistent error handling
func (b *BaseAction) GetPlayerFromGame(g *game.Game, playerID string, log *zap.Logger) (*player.Player, error) {
	p, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return nil, fmt.Errorf("player not found: %s", playerID)
	}
	return p, nil
}

// ConsumePlayerAction consumes an action from the game's current turn
// Returns true if an action was consumed, false if unlimited (-1) or no actions remaining (0)
// This properly handles unlimited actions by not consuming them
func (b *BaseAction) ConsumePlayerAction(g *game.Game, log *zap.Logger) bool {
	currentTurn := g.CurrentTurn()
	if currentTurn == nil {
		log.Warn("⚠️ No current turn set, cannot consume action")
		return false
	}

	consumed := currentTurn.ConsumeAction()
	if consumed {
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", currentTurn.ActionsRemaining()))

		// Publish GameStateChangedEvent to trigger broadcast
		// This ensures all clients see the updated action count immediately
		if eventBus := g.EventBus(); eventBus != nil {
			events.Publish(eventBus, events.GameStateChangedEvent{
				GameID:    g.ID(),
				Timestamp: time.Now(),
			})
		}
	}

	return consumed
}
