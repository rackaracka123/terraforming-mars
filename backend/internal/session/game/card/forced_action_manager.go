package card

import (
	"context"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	sessionGame "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/game/deck"
	"terraforming-mars-backend/internal/session/types"
)

// ForcedActionManager manages forced first turn actions for corporations
type ForcedActionManager interface {
	// SubscribeToPhaseChanges subscribes to game phase change events
	SubscribeToPhaseChanges()

	// SubscribeToCardDrawEvents subscribes to card draw confirmation events
	SubscribeToCardDrawEvents()

	// MarkComplete marks a player's forced action as complete
	MarkComplete(ctx context.Context, gameID, playerID string) error

	// TriggerForcedFirstAction manually triggers a player's forced first action
	TriggerForcedFirstAction(ctx context.Context, gameID, playerID string, player types.Player) error
}

// ForcedActionManagerImpl implements ForcedActionManager
type ForcedActionManagerImpl struct {
	eventBus *events.EventBusImpl
	cardRepo Repository             // Session card repository
	gameRepo sessionGame.Repository // Session game repository
	deckRepo deck.Repository        // Session deck repository
	// TODO: Full implementation needs refactoring for new architecture
	// Event handlers need access to session to fetch player/game data when events trigger
}

// NewForcedActionManager creates a new forced action manager
func NewForcedActionManager(
	eventBus *events.EventBusImpl,
	cardRepo Repository,
	gameRepo sessionGame.Repository,
	deckRepo deck.Repository,
) ForcedActionManager {
	return &ForcedActionManagerImpl{
		eventBus: eventBus,
		cardRepo: cardRepo,
		gameRepo: gameRepo,
		deckRepo: deckRepo,
	}
}

// SubscribeToPhaseChanges subscribes to game phase change events
// TODO: Full implementation pending architecture refactoring
func (m *ForcedActionManagerImpl) SubscribeToPhaseChanges() {
	log := logger.Get()
	log.Warn("⚠️  ForcedActionManager not yet fully implemented - corporation forced first actions will not work")
}

// SubscribeToCardDrawEvents subscribes to card draw confirmation events
// TODO: Full implementation pending architecture refactoring
func (m *ForcedActionManagerImpl) SubscribeToCardDrawEvents() {
	log := logger.Get()
	log.Warn("⚠️  ForcedActionManager not yet fully implemented - corporation forced first actions will not work")
}

// MarkComplete marks a player's forced action as complete
// TODO: Full implementation pending architecture refactoring
func (m *ForcedActionManagerImpl) MarkComplete(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Warn("⚠️  ForcedActionManager.MarkComplete not yet implemented")
	return nil
}

// TriggerForcedFirstAction manually triggers a player's forced first action
// TODO: Full implementation pending architecture refactoring
func (m *ForcedActionManagerImpl) TriggerForcedFirstAction(ctx context.Context, gameID, playerID string, player types.Player) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Warn("⚠️  ForcedActionManager.TriggerForcedFirstAction not yet implemented")
	// Corporation forced first actions (like Inventrix or Helion's starting bonuses) won't work until this is implemented
	return nil
}
