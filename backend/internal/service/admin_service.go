package service

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/features/card"
	cardPkg "terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"
	playerPkg "terraforming-mars-backend/internal/player"
	"terraforming-mars-backend/internal/shared/types"

	"go.uber.org/zap"
)

// AdminService handles administrative operations for development mode
type AdminService interface {
	// OnAdminGiveCard gives a specific card to a player
	OnAdminGiveCard(ctx context.Context, gameID, playerID, cardID string) error

	// OnAdminSetPhase sets the game phase
	OnAdminSetPhase(ctx context.Context, gameID string, phase game.GamePhase) error

	// OnAdminSetResources sets a player's resources
	OnAdminSetResources(ctx context.Context, gameID, playerID string, resources resources.Resources) error

	// OnAdminSetProduction sets a player's production
	OnAdminSetProduction(ctx context.Context, gameID, playerID string, production resources.Production) error

	// OnAdminSetGlobalParameters sets the global terraforming parameters
	OnAdminSetGlobalParameters(ctx context.Context, gameID string, params parameters.GlobalParameters) error

	// OnAdminStartTileSelection starts tile selection for testing
	OnAdminStartTileSelection(ctx context.Context, gameID, playerID, tileType string) error

	// OnAdminSetCurrentTurn sets the current player turn
	OnAdminSetCurrentTurn(ctx context.Context, gameID, playerID string) error

	// OnAdminSetCorporation sets a player's corporation directly (bypassing selection validation)
	OnAdminSetCorporation(ctx context.Context, gameID, playerID, corporationID string) error
}

// AdminServiceImpl implements AdminService interface
type AdminServiceImpl struct {
	gameRepo            game.Repository
	playerRepo          player.Repository
	cardRepo            card.CardRepository
	cardDeckRepo        card.CardDeckRepository
	sessionManager      session.SessionManager
	effectSubscriber    CardEffectSubscriber
	forcedActionManager ForcedActionManager
}

// NewAdminService creates a new AdminService instance
func NewAdminService(
	gameRepo game.Repository,
	playerRepo player.Repository,
	cardRepo card.CardRepository,
	cardDeckRepo card.CardDeckRepository,
	sessionManager session.SessionManager,
	effectSubscriber CardEffectSubscriber,
	forcedActionManager ForcedActionManager,
) AdminService {
	return &AdminServiceImpl{
		gameRepo:            gameRepo,
		playerRepo:          playerRepo,
		cardRepo:            cardRepo,
		cardDeckRepo:        cardDeckRepo,
		sessionManager:      sessionManager,
		effectSubscriber:    effectSubscriber,
		forcedActionManager: forcedActionManager,
	}
}

// OnAdminGiveCard gives a specific card to a player
func (s *AdminServiceImpl) OnAdminGiveCard(ctx context.Context, gameID, playerID, cardID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🎴 Admin giving card to player", zap.String("card_id", cardID))

	// Verify card exists
	card, err := s.cardRepo.GetCardByID(ctx, cardID)
	if err != nil || card == nil {
		log.Error("Card not found", zap.String("card_id", cardID), zap.Error(err))
		return fmt.Errorf("card not found: %s", cardID)
	}

	// Verify player exists
	_, err = s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %w", err)
	}

	// Add card to player's hand using repository method
	if err := s.playerRepo.AddCard(ctx, gameID, playerID, cardID); err != nil {
		log.Error("Failed to add card to player hand", zap.Error(err))
		return fmt.Errorf("failed to add card to player hand: %w", err)
	}

	log.Info("✅ Card given to player successfully",
		zap.String("card_id", cardID),
		zap.String("card_name", card.Name))

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after giving card", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// OnAdminSetPhase sets the game phase
func (s *AdminServiceImpl) OnAdminSetPhase(ctx context.Context, gameID string, phase game.GamePhase) error {
	log := logger.WithGameContext(gameID, "")
	log.Info("🔄 Admin setting game phase", zap.String("phase", string(phase)))

	// Verify game exists
	_, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return fmt.Errorf("game not found: %w", err)
	}

	// Update game phase
	if err := s.gameRepo.UpdatePhase(ctx, gameID, phase); err != nil {
		log.Error("Failed to update game phase", zap.Error(err))
		return fmt.Errorf("failed to update game phase: %w", err)
	}

	log.Info("✅ Game phase set successfully", zap.String("phase", string(phase)))

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after phase change", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// OnAdminSetResources sets a player's resources
func (s *AdminServiceImpl) OnAdminSetResources(ctx context.Context, gameID, playerID string, resources resources.Resources) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("💰 Admin setting player resources", zap.Any("resources", resources))

	// Verify player exists
	_, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %w", err)
	}

	// Update player resources
	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, resources); err != nil {
		log.Error("Failed to update player resources", zap.Error(err))
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	log.Info("✅ Player resources set successfully",
		zap.Int("credits", resources.Credits),
		zap.Int("steel", resources.Steel),
		zap.Int("titanium", resources.Titanium),
		zap.Int("plants", resources.Plants),
		zap.Int("energy", resources.Energy),
		zap.Int("heat", resources.Heat))

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after resources change", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// OnAdminSetProduction sets a player's production
func (s *AdminServiceImpl) OnAdminSetProduction(ctx context.Context, gameID, playerID string, production resources.Production) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🏭 Admin setting player production", zap.Any("production", production))

	// Verify player exists
	_, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %w", err)
	}

	// Update player production
	if err := s.playerRepo.UpdateProduction(ctx, gameID, playerID, production); err != nil {
		log.Error("Failed to update player production", zap.Error(err))
		return fmt.Errorf("failed to update player production: %w", err)
	}

	log.Info("✅ Player production set successfully",
		zap.Int("credits", production.Credits),
		zap.Int("steel", production.Steel),
		zap.Int("titanium", production.Titanium),
		zap.Int("plants", production.Plants),
		zap.Int("energy", production.Energy),
		zap.Int("heat", production.Heat))

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after production change", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// OnAdminSetGlobalParameters sets the global terraforming parameters
func (s *AdminServiceImpl) OnAdminSetGlobalParameters(ctx context.Context, gameID string, params parameters.GlobalParameters) error {
	log := logger.WithGameContext(gameID, "")
	log.Info("🌍 Admin setting global parameters",
		zap.Int("temperature", params.Temperature),
		zap.Int("oxygen", params.Oxygen),
		zap.Int("oceans", params.Oceans))

	// Verify game exists
	_, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return fmt.Errorf("game not found: %w", err)
	}

	// Validate parameters are within game bounds
	if params.Temperature < -30 || params.Temperature > 8 {
		return fmt.Errorf("temperature must be between -30 and 8, got %d", params.Temperature)
	}
	if params.Oxygen < 0 || params.Oxygen > 14 {
		return fmt.Errorf("oxygen must be between 0 and 14, got %d", params.Oxygen)
	}
	if params.Oceans < 0 || params.Oceans > 9 {
		return fmt.Errorf("oceans must be between 0 and 9, got %d", params.Oceans)
	}

	// Update global parameters
	if err := s.gameRepo.UpdateGlobalParameters(ctx, gameID, params); err != nil {
		log.Error("Failed to update global parameters", zap.Error(err))
		return fmt.Errorf("failed to update global parameters: %w", err)
	}

	log.Info("✅ Global parameters set successfully",
		zap.Int("temperature", params.Temperature),
		zap.Int("oxygen", params.Oxygen),
		zap.Int("oceans", params.Oceans))

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after global parameters change", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// OnAdminStartTileSelection starts tile selection for testing
func (s *AdminServiceImpl) OnAdminStartTileSelection(ctx context.Context, gameID, playerID, tileType string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🎯 Admin starting tile selection", zap.String("tile_type", tileType))

	// Verify player exists
	_, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %w", err)
	}

	// Get game to access the board
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return fmt.Errorf("game not found: %w", err)
	}

	// Calculate available hexes based on tile type (demo logic)
	availableHexes, err := s.calculateDemoAvailableHexes(game, tileType)
	if err != nil {
		return fmt.Errorf("failed to calculate available hexes: %w", err)
	}

	if len(availableHexes) == 0 {
		log.Warn("No valid positions available for tile type", zap.String("tile_type", tileType))
		return fmt.Errorf("no valid positions available for %s placement", tileType)
	}

	// Set pending tile selection
	pendingSelection := &player.PendingTileSelection{
		TileType:       tileType,
		AvailableHexes: availableHexes,
		Source:         "admin_demo",
	}

	if err := s.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, pendingSelection); err != nil {
		log.Error("Failed to set pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to set pending tile selection: %w", err)
	}

	log.Info("✅ Tile selection started successfully",
		zap.String("tile_type", tileType),
		zap.Int("available_positions", len(availableHexes)))

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after starting tile selection", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// OnAdminSetCurrentTurn sets the current player turn
func (s *AdminServiceImpl) OnAdminSetCurrentTurn(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🔄 Admin setting current turn")

	// Verify player exists
	_, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %w", err)
	}

	// Set current turn
	if err := s.gameRepo.SetCurrentTurn(ctx, gameID, &playerID); err != nil {
		log.Error("Failed to set current turn", zap.Error(err))
		return fmt.Errorf("failed to set current turn: %w", err)
	}

	log.Info("✅ Current turn set successfully")

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after setting current turn", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// calculateDemoAvailableHexes calculates available positions for demo tile placement
func (s *AdminServiceImpl) calculateDemoAvailableHexes(game game.Game, tileType string) ([]string, error) {
	var availableHexes []string

	// Get board from game
	board, err := game.GetBoard()
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	// Demo logic: just find empty tiles based on type
	for _, tile := range board.Tiles {
		// Tile must be empty
		if tile.OccupiedBy != nil {
			continue
		}

		switch tileType {
		case "ocean":
			// Ocean tiles can only be placed on ocean-designated spaces
			if tile.Type == tiles.ResourceOceanTile {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}

		case "city", "greenery":
			// Cities and greenery can be placed on any empty land space
			if tile.Type != tiles.ResourceOceanTile {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}
		}
	}

	return availableHexes, nil
}

// OnAdminSetCorporation sets a player's corporation directly (bypassing selection validation)
func (s *AdminServiceImpl) OnAdminSetCorporation(ctx context.Context, gameID, playerID, corporationID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🏢 Admin setting player corporation", zap.String("corporation_id", corporationID))

	// Get corporation card from card repository
	corporationCard, err := s.cardRepo.GetCardByID(ctx, corporationID)
	if err != nil {
		log.Error("Corporation card not found", zap.String("corporation_id", corporationID), zap.Error(err))
		return fmt.Errorf("corporation card not found: %s", corporationID)
	}

	// Verify player exists
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %w", err)
	}

	// Clear any selection phases to prevent stuck modals
	if player.SelectStartingCardsPhase != nil {
		if err := s.playerRepo.UpdateSelectStartingCardsPhase(ctx, gameID, playerID, nil); err != nil {
			log.Error("Failed to clear starting cards phase", zap.Error(err))
			return fmt.Errorf("failed to clear starting cards phase: %w", err)
		}
		log.Debug("Cleared SelectStartingCardsPhase to prevent stuck modal")
	}

	// Clear any pending selections
	if err := s.playerRepo.ClearPendingCardSelection(ctx, gameID, playerID); err != nil {
		log.Error("Failed to clear pending card selection", zap.Error(err))
		// Don't fail, just log
	}

	if err := s.playerRepo.ClearPendingTileSelection(ctx, gameID, playerID); err != nil {
		log.Error("Failed to clear pending tile selection", zap.Error(err))
		// Don't fail, just log
	}

	// Update player's corporation
	if err := s.playerRepo.UpdateCorporation(ctx, gameID, playerID, *corporationCard); err != nil {
		log.Error("Failed to update player corporation", zap.Error(err))
		return fmt.Errorf("failed to update player corporation: %w", err)
	}

	// Apply starting resources and production from corporation
	if corporationCard.StartingResources != nil {
		resources := resources.Resources{
			Credits:  corporationCard.StartingResources.Credits,
			Steel:    corporationCard.StartingResources.Steel,
			Titanium: corporationCard.StartingResources.Titanium,
			Plants:   corporationCard.StartingResources.Plants,
			Energy:   corporationCard.StartingResources.Energy,
			Heat:     corporationCard.StartingResources.Heat,
		}

		if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, resources); err != nil {
			log.Error("Failed to update starting resources", zap.Error(err))
			return fmt.Errorf("failed to update starting resources: %w", err)
		}
		log.Debug("Applied starting resources from corporation", zap.Any("resources", resources))
	}

	if corporationCard.StartingProduction != nil {
		production := resources.Production{
			Credits:  corporationCard.StartingProduction.Credits,
			Steel:    corporationCard.StartingProduction.Steel,
			Titanium: corporationCard.StartingProduction.Titanium,
			Plants:   corporationCard.StartingProduction.Plants,
			Energy:   corporationCard.StartingProduction.Energy,
			Heat:     corporationCard.StartingProduction.Heat,
		}

		if err := s.playerRepo.UpdateProduction(ctx, gameID, playerID, production); err != nil {
			log.Error("Failed to update starting production", zap.Error(err))
			return fmt.Errorf("failed to update starting production: %w", err)
		}
		log.Debug("Applied starting production from corporation", zap.Any("production", production))
	}

	// Extract and apply payment substitutes from corporation behaviors
	paymentSubstitutes := []types.PaymentSubstitute{}
	for _, behavior := range corporationCard.Behaviors {
		// Look for auto-trigger behaviors without conditions (starting bonuses)
		hasAutoTrigger := false
		hasCondition := false
		for _, trigger := range behavior.Triggers {
			if trigger.Type == card.ResourceTriggerAuto {
				hasAutoTrigger = true
				if trigger.Condition != nil {
					hasCondition = true
				}
			}
		}

		// Only process auto behaviors without conditions (starting bonuses)
		if !hasAutoTrigger || hasCondition {
			continue
		}

		// Extract payment-substitute outputs
		for _, output := range behavior.Outputs {
			if output.Type == "payment-substitute" {
				// Extract the resource type from affectedResources
				if len(output.AffectedResources) > 0 {
					resourceTypeStr := output.AffectedResources[0]
					substitute := types.PaymentSubstitute{
						ResourceType:   types.ResourceType(resourceTypeStr),
						ConversionRate: output.Amount,
					}
					paymentSubstitutes = append(paymentSubstitutes, substitute)
					log.Debug("💰 Extracted payment substitute from corporation",
						zap.String("resource_type", resourceTypeStr),
						zap.Int("conversion_rate", output.Amount))
				}
			}
		}
	}

	// Apply payment substitutes to player if any were found
	if len(paymentSubstitutes) > 0 {
		if err := s.playerRepo.UpdatePaymentSubstitutes(ctx, gameID, playerID, paymentSubstitutes); err != nil {
			log.Error("Failed to update player payment substitutes", zap.Error(err))
			return fmt.Errorf("failed to update player payment substitutes: %w", err)
		}
		log.Debug("💰 Payment substitutes applied",
			zap.Int("substitutes_count", len(paymentSubstitutes)))
	}

	// Subscribe corporation passive effects using CardEffectSubscriber (event-driven system)
	if s.effectSubscriber != nil {
		if err := s.effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, corporationCard.ID, corporationCard); err != nil {
			log.Error("Failed to subscribe corporation effects", zap.Error(err))
			return fmt.Errorf("failed to subscribe corporation effects: %w", err)
		}
		log.Debug("🎆 Corporation passive effects subscribed")
	}

	// Extract and register corporation manual actions (manual triggers)
	var manualActions []card.PlayerAction
	for behaviorIndex, behavior := range corporationCard.Behaviors {
		// Check if this behavior has manual triggers
		hasManualTrigger := false
		for _, trigger := range behavior.Triggers {
			if trigger.Type == card.ResourceTriggerManual {
				hasManualTrigger = true
				break
			}
		}

		// If behavior has manual triggers, create a PlayerAction
		if hasManualTrigger {
			action := card.PlayerAction{
				CardID:        corporationCard.ID,
				CardName:      corporationCard.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			}
			manualActions = append(manualActions, action)
			log.Debug("🎯 Extracted manual action from corporation",
				zap.Int("behavior_index", behaviorIndex))
		}
	}

	// Add manual actions to player if any were found
	if len(manualActions) > 0 {
		// Get current player state to append to existing actions
		currentPlayer, err := s.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to get player for actions update: %w", err)
		}

		// Create new actions list with existing + new corporation actions
		newActions := make([]card.PlayerAction, len(currentPlayer.Actions)+len(manualActions))
		copy(newActions, currentPlayer.Actions)
		copy(newActions[len(currentPlayer.Actions):], manualActions)

		// Update player with new actions
		if err := s.playerRepo.UpdatePlayerActions(ctx, gameID, playerID, newActions); err != nil {
			return fmt.Errorf("failed to update player manual actions: %w", err)
		}

		log.Debug("🎯 Corporation manual actions registered",
			zap.Int("actions_count", len(manualActions)))
	}

	// Extract and set forced first action if corporation has one
	forcedAction := s.extractForcedFirstAction(corporationCard)
	if forcedAction != nil {
		if err := s.playerRepo.UpdateForcedFirstAction(ctx, gameID, playerID, forcedAction); err != nil {
			log.Error("Failed to set forced first action", zap.Error(err))
			return fmt.Errorf("failed to set forced first action: %w", err)
		}

		log.Info("🎯 Corporation requires forced first turn action",
			zap.String("corporation_id", corporationID),
			zap.String("action_type", forcedAction.ActionType),
			zap.String("description", forcedAction.Description))

		// Trigger the forced action immediately (admin tools don't wait for phase changes)
		// Get updated player with forced action set
		updatedPlayer, err := s.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			log.Error("Failed to get player for triggering forced action", zap.Error(err))
			return fmt.Errorf("failed to get player for triggering forced action: %w", err)
		}

		// Trigger the forced action via ForcedActionManager
		if err := s.forcedActionManager.TriggerForcedFirstAction(ctx, gameID, playerID, updatedPlayer); err != nil {
			log.Error("Failed to trigger forced first action", zap.Error(err))
			return fmt.Errorf("failed to trigger forced first action: %w", err)
		}

		log.Info("✅ Forced first action triggered",
			zap.String("action_type", forcedAction.ActionType))
	}

	log.Info("✅ Corporation set successfully",
		zap.String("corporation_id", corporationID),
		zap.String("corporation_name", corporationCard.Name))

	// Broadcast updated game state
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after setting corporation", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
}

// extractForcedFirstAction parses a corporation's forced first action from its behaviors
// Forced first actions use auto-first-action trigger type (e.g., Inventrix: draw 3 cards, Tharsis: place city)
func (s *AdminServiceImpl) extractForcedFirstAction(corporation *cardPkg.Card) *playerPkg.ForcedFirstAction {
	for _, behavior := range corporation.Behaviors {
		// Check if this behavior has a forced first action trigger
		for _, trigger := range behavior.Triggers {
			if trigger.Type == cardPkg.ResourceTriggerAutoFirstAction {
				// Determine action type from outputs
				actionType := s.determineActionType(behavior.Outputs)
				description := s.generateForcedActionDescription(behavior.Outputs)

				return &playerPkg.ForcedFirstAction{
					ActionType:    actionType,
					CorporationID: corporation.ID,
					Completed:     false,
					Description:   description,
				}
			}
		}
	}
	return nil
}

// determineActionType extracts the primary action type from behavior outputs
func (s *AdminServiceImpl) determineActionType(outputs []cardPkg.ResourceCondition) string {
	if len(outputs) == 0 {
		return "unknown"
	}

	// Return the first output type as the action type
	switch outputs[0].Type {
	case "card-draw":
		return "card_draw"
	case "city-placement":
		return "city_placement"
	case "greenery-placement":
		return "greenery_placement"
	case "card-take":
		return "card_selection"
	default:
		return string(outputs[0].Type)
	}
}

// generateForcedActionDescription creates a human-readable description from outputs
func (s *AdminServiceImpl) generateForcedActionDescription(outputs []cardPkg.ResourceCondition) string {
	if len(outputs) == 0 {
		return "Complete forced action"
	}

	// Generate description based on output type
	switch outputs[0].Type {
	case "card-draw":
		return fmt.Sprintf("Draw %d cards", outputs[0].Amount)
	case "city-placement":
		return "Place a city tile"
	case "greenery-placement":
		return "Place a greenery tile"
	case "card-take":
		return fmt.Sprintf("Select %d cards", outputs[0].Amount)
	default:
		return fmt.Sprintf("Complete %s action", outputs[0].Type)
	}
}
