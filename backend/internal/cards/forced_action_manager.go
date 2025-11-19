package cards

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	sessionGame "terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/deck"
	"terraforming-mars-backend/internal/session/game/player"

	"go.uber.org/zap"
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
	TriggerForcedFirstAction(ctx context.Context, gameID, playerID string, player model.Player) error
}

// ForcedActionManagerImpl implements ForcedActionManager
type ForcedActionManagerImpl struct {
	eventBus   *events.EventBusImpl
	cardRepo   repository.CardRepository
	playerRepo player.Repository      // NEW: Session player repository
	gameRepo   sessionGame.Repository // NEW: Session game repository
	deckRepo   deck.Repository        // NEW: Session deck repository
}

// NewForcedActionManager creates a new forced action manager
func NewForcedActionManager(
	eventBus *events.EventBusImpl,
	cardRepo repository.CardRepository,
	playerRepo player.Repository, // NEW: Session player repository
	gameRepo sessionGame.Repository, // NEW: Session game repository
	deckRepo deck.Repository, // NEW: Session deck repository
) ForcedActionManager {
	return &ForcedActionManagerImpl{
		eventBus:   eventBus,
		cardRepo:   cardRepo,
		playerRepo: playerRepo,
		gameRepo:   gameRepo,
		deckRepo:   deckRepo,
	}
}

// SubscribeToPhaseChanges subscribes to game phase change events
func (m *ForcedActionManagerImpl) SubscribeToPhaseChanges() {
	events.Subscribe(m.eventBus, func(event repository.GamePhaseChangedEvent) {
		ctx := context.Background()
		if err := m.onPhaseChanged(ctx, event); err != nil {
			logger.Get().Error("Failed to handle phase change event",
				zap.Error(err),
				zap.String("game_id", event.GameID),
				zap.String("old_phase", event.OldPhase),
				zap.String("new_phase", event.NewPhase))
		}
	})
}

// SubscribeToCardDrawEvents subscribes to card draw confirmation events
func (m *ForcedActionManagerImpl) SubscribeToCardDrawEvents() {
	events.Subscribe(m.eventBus, func(event player.CardDrawConfirmedEvent) {
		ctx := context.Background()
		if err := m.onCardDrawConfirmed(ctx, event); err != nil {
			logger.Get().Error("Failed to handle card draw confirmation event",
				zap.Error(err),
				zap.String("game_id", event.GameID),
				zap.String("player_id", event.PlayerID),
				zap.String("source", event.Source))
		}
	})
}

// onPhaseChanged handles game phase change events
func (m *ForcedActionManagerImpl) onPhaseChanged(ctx context.Context, event repository.GamePhaseChangedEvent) error {
	log := logger.WithGameContext(event.GameID, "")

	// Only trigger forced actions when transitioning to Action phase
	if event.NewPhase != string(model.GamePhaseAction) {
		return nil
	}

	log.Debug("🎯 Phase changed to Action, checking for forced actions")

	// Get current game to find whose turn it is
	game, err := m.gameRepo.GetByID(ctx, event.GameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// If no current turn set, nothing to check
	if game.CurrentTurn == nil {
		log.Debug("No current turn set, skipping forced action check")
		return nil
	}

	playerID := *game.CurrentTurn

	// Get player to check for forced action
	sessionPlayer, err := m.playerRepo.GetByID(ctx, event.GameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Check if player has an uncompleted forced action
	if sessionPlayer.ForcedFirstAction == nil || sessionPlayer.ForcedFirstAction.Completed {
		log.Debug("No uncompleted forced action for current player", zap.String("player_id", playerID))
		return nil
	}

	log.Info("🎯 Triggering forced first action",
		zap.String("player_id", playerID),
		zap.String("corporation_id", sessionPlayer.ForcedFirstAction.CorporationID))

	// Convert session Player to model.Player for TriggerForcedFirstAction
	modelPlayer := convertSessionPlayerToModel(sessionPlayer)

	// Trigger the forced action
	if err := m.TriggerForcedFirstAction(ctx, event.GameID, playerID, modelPlayer); err != nil {
		return fmt.Errorf("failed to trigger forced action: %w", err)
	}

	return nil
}

// convertSessionPlayerToModel converts a session player to model player
// This is a temporary helper during migration - only copies fields needed by forced actions
func convertSessionPlayerToModel(sp *player.Player) model.Player {
	return model.Player{
		ID:                       sp.ID,
		Name:                     sp.Name,
		Resources:                sp.Resources,
		Production:               sp.Production,
		TerraformRating:          sp.TerraformRating,
		IsConnected:              sp.IsConnected,
		ForcedFirstAction:        sp.ForcedFirstAction,
		PendingCardDrawSelection: sp.PendingCardDrawSelection,
		PendingTileSelection:     sp.PendingTileSelection,
		Cards:                    sp.Cards,
		ResourceStorage:          sp.ResourceStorage,
	}
}

// onCardDrawConfirmed handles card draw confirmation events
func (m *ForcedActionManagerImpl) onCardDrawConfirmed(ctx context.Context, event player.CardDrawConfirmedEvent) error {
	log := logger.WithGameContext(event.GameID, event.PlayerID)

	// Get player to check if this was a forced action
	sessionPlayer, err := m.playerRepo.GetByID(ctx, event.GameID, event.PlayerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Check if player has a forced action and if the source matches
	if sessionPlayer.ForcedFirstAction == nil {
		log.Debug("🎯 Card draw confirmed, but no forced action present")
		return nil
	}

	if sessionPlayer.ForcedFirstAction.Source != event.Source {
		log.Debug("🎯 Card draw confirmed, but source doesn't match forced action",
			zap.String("event_source", event.Source),
			zap.String("forced_action_source", sessionPlayer.ForcedFirstAction.Source))
		return nil
	}

	// This was a forced action - mark it complete
	log.Info("🎯 Card draw confirmed from forced action, marking complete",
		zap.String("corporation_id", sessionPlayer.ForcedFirstAction.CorporationID),
		zap.String("source", sessionPlayer.ForcedFirstAction.Source),
		zap.Int("cards_confirmed", len(event.Cards)))

	if err := m.MarkComplete(ctx, event.GameID, event.PlayerID); err != nil {
		return fmt.Errorf("failed to mark forced action complete: %w", err)
	}

	return nil
}

// TriggerForcedFirstAction triggers the forced action for a player
func (m *ForcedActionManagerImpl) TriggerForcedFirstAction(ctx context.Context, gameID, playerID string, player model.Player) error {
	// Look up the corporation card to get the behavior details
	corporation, err := m.cardRepo.GetCardByID(ctx, player.ForcedFirstAction.CorporationID)
	if err != nil {
		return fmt.Errorf("failed to get corporation card: %w", err)
	}
	if corporation == nil {
		return fmt.Errorf("corporation card not found: %s", player.ForcedFirstAction.CorporationID)
	}

	// Find the forced action behavior (auto-first-action trigger)
	var forcedBehavior *model.CardBehavior
	for _, behavior := range corporation.Behaviors {
		for _, trigger := range behavior.Triggers {
			if trigger.Type == model.ResourceTriggerAutoCorporationFirstAction {
				forcedBehavior = &behavior
				break
			}
		}
		if forcedBehavior != nil {
			break
		}
	}

	if forcedBehavior == nil {
		return fmt.Errorf("forced action behavior not found for corporation: %s", corporation.ID)
	}

	if len(forcedBehavior.Outputs) == 0 {
		return fmt.Errorf("forced action behavior has no outputs")
	}

	output := forcedBehavior.Outputs[0] // Forced actions should have a single primary output

	// Trigger the appropriate action based on output type
	switch output.Type {
	case model.ResourceCardDraw:
		return m.triggerCardDrawAction(ctx, gameID, playerID, player, output.Amount)
	case model.ResourceCityPlacement, model.ResourceGreeneryPlacement, model.ResourceOceanPlacement:
		return m.triggerTilePlacementAction(ctx, gameID, playerID, player, output.Type)
	default:
		return fmt.Errorf("unsupported forced action type: %s", output.Type)
	}
}

// triggerCardDrawAction creates a PendingCardDrawSelection for the player
func (m *ForcedActionManagerImpl) triggerCardDrawAction(ctx context.Context, gameID, playerID string, player model.Player, amount int) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("📚 Triggering card draw forced action", zap.Int("amount", amount))

	// Draw cards from the deck using deck repository
	drawnCards, err := m.deckRepo.DrawProjectCards(ctx, gameID, amount)
	if err != nil {
		return fmt.Errorf("failed to draw cards: %w", err)
	}

	// Create pending card draw selection
	pendingCardDrawSelection := &model.PendingCardDrawSelection{
		AvailableCards: drawnCards,
		FreeTakeCount:  amount, // All cards are free for forced actions
		MaxBuyCount:    0,      // Cannot buy additional cards
		CardBuyCost:    0,
		Source:         player.ForcedFirstAction.CorporationID,
	}

	// Update player with pending selection
	if err := m.playerRepo.UpdatePendingCardDrawSelection(ctx, gameID, playerID, pendingCardDrawSelection); err != nil {
		return fmt.Errorf("failed to update pending card draw selection: %w", err)
	}

	// Set Source on ForcedFirstAction so we can track completion
	updatedForcedAction := *player.ForcedFirstAction
	updatedForcedAction.Source = player.ForcedFirstAction.CorporationID
	if err := m.playerRepo.UpdateForcedFirstAction(ctx, gameID, playerID, &updatedForcedAction); err != nil {
		return fmt.Errorf("failed to update forced action source: %w", err)
	}

	log.Info("✅ Card draw forced action triggered", zap.Strings("cards", drawnCards))
	return nil
}

// triggerTilePlacementAction creates a PendingTileSelection for the player
func (m *ForcedActionManagerImpl) triggerTilePlacementAction(ctx context.Context, gameID, playerID string, player model.Player, tileType model.ResourceType) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🏙️ Triggering tile placement forced action", zap.String("tileType", string(tileType)))

	// Get available hexes for tile placement
	game, err := m.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	availableHexes := m.getAvailableHexesForTileType(game, tileType, playerID)
	if len(availableHexes) == 0 {
		return fmt.Errorf("no available hexes for tile type: %s", tileType)
	}

	// Map resource type to tile type string
	tileTypeStr := m.mapResourceTypeToTileType(tileType)

	// Create pending tile selection
	pendingTileSelection := &model.PendingTileSelection{
		TileType:       tileTypeStr,
		AvailableHexes: availableHexes,
		Source:         player.ForcedFirstAction.CorporationID,
	}

	// Update player with pending selection
	if err := m.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, pendingTileSelection); err != nil {
		return fmt.Errorf("failed to update pending tile selection: %w", err)
	}

	// Set Source on ForcedFirstAction so we can track completion
	updatedForcedAction := *player.ForcedFirstAction
	updatedForcedAction.Source = player.ForcedFirstAction.CorporationID
	if err := m.playerRepo.UpdateForcedFirstAction(ctx, gameID, playerID, &updatedForcedAction); err != nil {
		return fmt.Errorf("failed to update forced action source: %w", err)
	}

	log.Info("✅ Tile placement forced action triggered", zap.String("tileType", tileTypeStr))
	return nil
}

// getAvailableHexesForTileType returns available hexes for the given tile type
func (m *ForcedActionManagerImpl) getAvailableHexesForTileType(game *sessionGame.Game, tileType model.ResourceType, playerID string) []string {
	var availableHexes []string

	for _, tile := range game.Board.Tiles {
		// Skip occupied tiles
		if tile.OccupiedBy != nil {
			continue
		}

		// For city placement, any empty hex is valid
		if tileType == model.ResourceCityPlacement {
			availableHexes = append(availableHexes, tile.Coordinates.String())
			continue
		}

		// For greenery, check if adjacent to player's tile
		if tileType == model.ResourceGreeneryPlacement {
			if m.isAdjacentToPlayerTile(game, tile.Coordinates, playerID) {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}
			continue
		}

		// For ocean, check if it's marked as ocean-compatible (tile.Type == "ocean-tile")
		if tileType == model.ResourceOceanPlacement {
			if tile.Type == model.ResourceOceanTile {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}
			continue
		}
	}

	return availableHexes
}

// isAdjacentToPlayerTile checks if a hex is adjacent to any of the player's tiles
func (m *ForcedActionManagerImpl) isAdjacentToPlayerTile(game *sessionGame.Game, coord model.HexPosition, playerID string) bool {
	// For first action, if player has no tiles yet, any empty hex is valid
	hasPlayerTile := false
	for _, tile := range game.Board.Tiles {
		if tile.OwnerID != nil && *tile.OwnerID == playerID {
			hasPlayerTile = true
			break
		}
	}

	if !hasPlayerTile {
		return true // Any hex is valid for first tile
	}

	// Check adjacency to player tiles
	// TODO: Implement proper hex adjacency checking
	// For now, return true to allow placement
	return true
}

// mapResourceTypeToTileType converts ResourceType to tile type string
func (m *ForcedActionManagerImpl) mapResourceTypeToTileType(resourceType model.ResourceType) string {
	switch resourceType {
	case model.ResourceCityPlacement:
		return "city"
	case model.ResourceGreeneryPlacement:
		return "greenery"
	case model.ResourceOceanPlacement:
		return "ocean"
	default:
		return string(resourceType)
	}
}

// MarkComplete marks the player's forced action as completed
func (m *ForcedActionManagerImpl) MarkComplete(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	player, err := m.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	if player.ForcedFirstAction == nil {
		log.Debug("⚠️ No forced action to mark complete")
		return nil
	}

	log.Info("✅ Marking forced action complete", zap.String("corporationId", player.ForcedFirstAction.CorporationID))

	// Update forced action to completed
	completedAction := *player.ForcedFirstAction
	completedAction.Completed = true

	if err := m.playerRepo.UpdateForcedFirstAction(ctx, gameID, playerID, &completedAction); err != nil {
		return fmt.Errorf("failed to update forced action: %w", err)
	}

	return nil
}
