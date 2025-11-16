package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/turn"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// PlayCardActionAction handles playing a card action (blue card ability)
// This action orchestrates:
// - Turn and action validation
// - Player action availability validation
// - Manual action execution (inline, following ARCHITECTURE_FLOW.md)
// - Action consumption
type PlayCardActionAction struct {
	cardRepo          card.CardRepository
	playerRepo        player.Repository
	parametersFeature parameters.Service
	turnOrderService  turn.TurnOrderService
	sessionManager    session.SessionManager
}

// NewPlayCardActionAction creates a new play card action action
func NewPlayCardActionAction(
	cardRepo card.CardRepository,
	playerRepo player.Repository,
	parametersFeature parameters.Service,
	turnOrderService turn.TurnOrderService,
	sessionManager session.SessionManager,
) *PlayCardActionAction {
	return &PlayCardActionAction{
		cardRepo:          cardRepo,
		playerRepo:        playerRepo,
		parametersFeature: parametersFeature,
		turnOrderService:  turnOrderService,
		sessionManager:    sessionManager,
	}
}

// Execute performs the play card action action
// Steps:
// 1. Validate turn is player's turn
// 2. Validate player has available actions
// 3. Execute manual action via PlayService (handles validation, inputs, outputs, play count)
// 4. Consume player action (if not infinite)
// 5. Broadcast state
func (a *PlayCardActionAction) Execute(ctx context.Context, gameID string, playerID string, cardID string, behaviorIndex int, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🎯 Executing play card action action",
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex))

	// Validate turn
	currentTurn, err := a.turnOrderService.GetCurrentTurn(ctx)
	if err != nil {
		log.Error("Failed to get current turn", zap.Error(err))
		return fmt.Errorf("failed to get current turn: %w", err)
	}

	if currentTurn == nil {
		log.Error("No current player turn set", zap.String("requesting_player", playerID))
		return fmt.Errorf("no current player turn set, requesting player is %s", playerID)
	}

	if *currentTurn != playerID {
		log.Error("Not current player's turn",
			zap.String("current_turn", *currentTurn),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("not current player's turn: current turn is %s, requesting player is %s", *currentTurn, playerID)
	}

	// Get the player to validate they exist
	_, err = a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for card action play", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Get available actions via repository
	availableActions, err := a.playerRepo.GetAvailableActions(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get available actions", zap.Error(err))
		return fmt.Errorf("failed to get available actions: %w", err)
	}

	// Check if player has available actions
	if availableActions <= 0 && availableActions != -1 {
		log.Warn("Player has no available actions", zap.Int("available_actions", availableActions))
		return fmt.Errorf("no available actions remaining")
	}

	// Execute the card action (inline implementation)
	if err := a.executeCardAction(ctx, gameID, playerID, cardID, behaviorIndex, choiceIndex, cardStorageTarget, log); err != nil {
		log.Error("Failed to execute card action", zap.Error(err))
		return fmt.Errorf("failed to execute card action: %w", err)
	}

	log.Debug("✅ Card action executed successfully")

	// Consume one action now that all steps have succeeded
	if availableActions > 0 {
		newActions := availableActions - 1
		if err := a.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, newActions); err != nil {
			log.Error("Failed to consume player action", zap.Error(err))
			// Note: Action has already been applied, but we couldn't consume the action
			// This is a critical error but we don't rollback the entire action
			return fmt.Errorf("action applied but failed to consume available action: %w", err)
		}
		log.Debug("🎯 Action consumed", zap.Int("remaining_actions", newActions))
	} else {
		log.Debug("🎯 Action consumed (unlimited actions)", zap.Int("available_actions", -1))
	}

	// Broadcast game state update
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after card action play", zap.Error(err))
		// Don't fail the action, just log the error
	}

	log.Info("✅ Play card action action completed successfully",
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex))

	return nil
}

// executeCardAction processes a manual card action behavior inline
// This implements the logic that was previously in PlayService.PlayCardAction stub
func (a *PlayCardActionAction) executeCardAction(ctx context.Context, gameID, playerID, cardID string, behaviorIndex int, choiceIndex *int, cardStorageTarget *string, log *zap.Logger) error {
	// Get the card to access its behaviors
	cardObj, err := a.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("failed to get card: %w", err)
	}

	// Validate behavior index
	if behaviorIndex < 0 || behaviorIndex >= len(cardObj.Behaviors) {
		return fmt.Errorf("invalid behavior index %d for card %s (has %d behaviors)", behaviorIndex, cardID, len(cardObj.Behaviors))
	}

	behavior := cardObj.Behaviors[behaviorIndex]

	// Validate this is a manual trigger behavior
	hasManualTrigger := false
	for _, trigger := range behavior.Triggers {
		if trigger.Type == card.ResourceTriggerManual {
			hasManualTrigger = true
			break
		}
	}

	if !hasManualTrigger {
		return fmt.Errorf("behavior %d on card %s is not a manual action", behaviorIndex, cardID)
	}

	log.Debug("🎯 Processing manual action behavior",
		zap.String("card_name", cardObj.Name),
		zap.Int("behavior_index", behaviorIndex))

	// Get current player for resource/production updates
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	playerResources := player.Resources
	playerProduction := player.Production

	var resourcesChanged, productionChanged bool
	var tilePlacementQueue []string
	var trChange int

	// Aggregate all outputs: behavior.Outputs + choice[choiceIndex].Outputs
	allOutputs := behavior.Outputs

	// If choiceIndex is provided and this behavior has choices, add choice outputs
	if choiceIndex != nil && len(behavior.Choices) > 0 && *choiceIndex < len(behavior.Choices) {
		selectedChoice := behavior.Choices[*choiceIndex]
		allOutputs = append(allOutputs, selectedChoice.Outputs...)
		log.Debug("🎯 Applying choice outputs",
			zap.Int("choice_index", *choiceIndex),
			zap.Int("choice_outputs_count", len(selectedChoice.Outputs)))
	}

	// Process all aggregated outputs (same pattern as processCardEffects in play_card.go)
	for _, output := range allOutputs {
		switch output.Type {
		// Resource outputs
		case domain.ResourceCredits:
			playerResources.Credits += output.Amount
			resourcesChanged = true
		case domain.ResourceSteel:
			playerResources.Steel += output.Amount
			resourcesChanged = true
		case domain.ResourceTitanium:
			playerResources.Titanium += output.Amount
			resourcesChanged = true
		case domain.ResourcePlants:
			playerResources.Plants += output.Amount
			resourcesChanged = true
		case domain.ResourceEnergy:
			playerResources.Energy += output.Amount
			resourcesChanged = true
		case domain.ResourceHeat:
			playerResources.Heat += output.Amount
			resourcesChanged = true

		// Production outputs
		case domain.ResourceCreditsProduction:
			playerProduction.Credits += output.Amount
			productionChanged = true
		case domain.ResourceSteelProduction:
			playerProduction.Steel += output.Amount
			productionChanged = true
		case domain.ResourceTitaniumProduction:
			playerProduction.Titanium += output.Amount
			productionChanged = true
		case domain.ResourcePlantsProduction:
			playerProduction.Plants += output.Amount
			productionChanged = true
		case domain.ResourceEnergyProduction:
			playerProduction.Energy += output.Amount
			productionChanged = true
		case domain.ResourceHeatProduction:
			playerProduction.Heat += output.Amount
			productionChanged = true

		// TR output
		case domain.ResourceTR:
			trChange += output.Amount

		// Global parameters - orchestrate via parameters feature
		case domain.ResourceTemperature:
			steps := output.Amount / 2 // Temperature increases in 2°C steps
			if steps > 0 {
				actualSteps, err := a.parametersFeature.RaiseTemperature(ctx, steps)
				if err != nil {
					return fmt.Errorf("failed to raise temperature: %w", err)
				}
				// Award TR for actual steps raised
				if actualSteps > 0 {
					trChange += actualSteps
					log.Info("🌡️ Temperature raised (TR awarded)",
						zap.Int("steps", actualSteps),
						zap.Int("tr_change", actualSteps))
				}
			}

		case domain.ResourceOxygen:
			if output.Amount > 0 {
				actualSteps, err := a.parametersFeature.RaiseOxygen(ctx, output.Amount)
				if err != nil {
					return fmt.Errorf("failed to raise oxygen: %w", err)
				}
				// Award TR for actual steps raised
				if actualSteps > 0 {
					trChange += actualSteps
					log.Info("💨 Oxygen raised (TR awarded)",
						zap.Int("steps", actualSteps),
						zap.Int("tr_change", actualSteps))
				}
			}

		case domain.ResourceOceans:
			for i := 0; i < output.Amount; i++ {
				if err := a.parametersFeature.PlaceOcean(ctx); err != nil {
					log.Warn("Failed to place ocean (may be at max)", zap.Error(err))
					break
				}
				trChange++ // Award TR for ocean placement
			}
			if output.Amount > 0 {
				log.Info("🌊 Oceans placed (TR awarded)",
					zap.Int("count", output.Amount),
					zap.Int("tr_change", output.Amount))
			}

		// Tile placements
		case domain.ResourceCityPlacement:
			for i := 0; i < output.Amount; i++ {
				tilePlacementQueue = append(tilePlacementQueue, "city")
			}
		case domain.ResourceOceanPlacement:
			for i := 0; i < output.Amount; i++ {
				tilePlacementQueue = append(tilePlacementQueue, "ocean")
			}
		case domain.ResourceGreeneryPlacement:
			for i := 0; i < output.Amount; i++ {
				tilePlacementQueue = append(tilePlacementQueue, "greenery")
			}

		// Card storage resources
		case domain.ResourceAnimals, domain.ResourceMicrobes, domain.ResourceFloaters, domain.ResourceScience, domain.ResourceAsteroid:
			if err := a.applyCardStorageResource(ctx, gameID, playerID, cardID, output, cardStorageTarget, log); err != nil {
				return fmt.Errorf("failed to apply card storage resource: %w", err)
			}
		}
	}

	// Update resources if changed
	if resourcesChanged {
		if err := a.playerRepo.UpdateResources(ctx, gameID, playerID, playerResources); err != nil {
			return fmt.Errorf("failed to update player resources: %w", err)
		}
		log.Debug("💰 Resources updated")
	}

	// Update production if changed
	if productionChanged {
		if err := a.playerRepo.UpdateProduction(ctx, gameID, playerID, playerProduction); err != nil {
			return fmt.Errorf("failed to update player production: %w", err)
		}
		log.Debug("📈 Production updated")
	}

	// Update TR if changed
	if trChange != 0 {
		newTR := player.TerraformRating + trChange
		if err := a.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR); err != nil {
			return fmt.Errorf("failed to update terraform rating: %w", err)
		}
		log.Info("⭐ Terraform Rating changed",
			zap.Int("change", trChange),
			zap.Int("new_tr", newTR))
	}

	// Create tile placement queue if needed
	if len(tilePlacementQueue) > 0 {
		if err := a.playerRepo.CreateTileQueue(ctx, gameID, playerID, cardID, tilePlacementQueue); err != nil {
			return fmt.Errorf("failed to create tile placement queue: %w", err)
		}
		log.Debug("🏗️ Tile placement queue created", zap.Int("tiles", len(tilePlacementQueue)))
	}

	log.Debug("✅ Manual action effects processed")
	return nil
}

// applyCardStorageResource applies a card storage resource to the target card
func (a *PlayCardActionAction) applyCardStorageResource(ctx context.Context, gameID, playerID, cardID string, output card.ResourceCondition, cardStorageTarget *string, log *zap.Logger) error {
	targetCardID := cardID // Default to self
	if cardStorageTarget != nil && *cardStorageTarget != "" {
		targetCardID = *cardStorageTarget
	}

	// Get current player to access ResourceStorage
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Get or initialize resource storage map
	storage := player.ResourceStorage
	if storage == nil {
		storage = make(map[string]int)
	}

	// Update the storage for this card
	storage[targetCardID] += output.Amount

	// Save updated storage
	if err := a.playerRepo.UpdateResourceStorage(ctx, gameID, playerID, storage); err != nil {
		return fmt.Errorf("failed to update resource storage: %w", err)
	}

	log.Debug("📦 Card storage updated",
		zap.String("resource_type", string(output.Type)),
		zap.Int("amount", output.Amount),
		zap.String("target_card", targetCardID))

	return nil
}
