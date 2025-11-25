package execute_card_action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// Validator handles validation logic for card action execution
type Validator struct {
	resourceMgr    *game.ResourceManager
	sessionFactory session.SessionFactory
}

// NewValidator creates a new Validator instance
func NewValidator(sessionFactory session.SessionFactory) *Validator {
	return &Validator{
		resourceMgr:    game.NewResourceManager(),
		sessionFactory: sessionFactory,
	}
}

// ValidateActionInputs validates that the player has sufficient resources for the action inputs
// choiceIndex is optional and used when the action has choices between different effects
func (v *Validator) ValidateActionInputs(ctx context.Context, gameID, playerID string, action *types.PlayerAction, choiceIndex *int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get session and player to check resources
	sess := v.sessionFactory.Get(gameID)
	if sess == nil {
		return fmt.Errorf("game session not found: %s", gameID)
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		return fmt.Errorf("player not found: %s", playerID)
	}

	// Get current resources for validation
	currentResources, err := player.Resources.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get player resources: %w", err)
	}

	// Aggregate all inputs: behavior.Inputs + choice[choiceIndex].Inputs
	allInputs := action.Behavior.Inputs

	// If choiceIndex is provided and this action has choices, add choice inputs
	if choiceIndex != nil && len(action.Behavior.Choices) > 0 && *choiceIndex < len(action.Behavior.Choices) {
		selectedChoice := action.Behavior.Choices[*choiceIndex]
		allInputs = append(allInputs, selectedChoice.Inputs...)
		log.Debug("🎯 Validating choice inputs",
			zap.Int("choice_index", *choiceIndex),
			zap.Int("choice_inputs_count", len(selectedChoice.Inputs)))
	}

	// Check each input requirement
	for _, input := range allInputs {
		switch input.Type {
		case types.ResourceCredits, types.ResourceSteel, types.ResourceTitanium,
			types.ResourcePlants, types.ResourceEnergy, types.ResourceHeat:
			// Use ResourceManager for standard resource validation
			if err := v.resourceMgr.ValidateHasResource(currentResources, input.Type, input.Amount); err != nil {
				return err
			}

		// Card storage resources (animals, microbes, floaters, science, asteroid)
		case types.ResourceAnimals, types.ResourceMicrobes, types.ResourceFloaters, types.ResourceScience, types.ResourceAsteroid:
			// Validate card storage resource inputs
			if input.Target == types.TargetSelfCard {
				// Initialize resource storage map if nil (for checking)
				if player.ResourceStorage == nil {
					player.ResourceStorage = make(map[string]int)
				}

				currentStorage := player.ResourceStorage[action.CardID]
				if currentStorage < input.Amount {
					return fmt.Errorf("insufficient %s storage on card %s: need %d, have %d",
						input.Type, action.CardID, input.Amount, currentStorage)
				}
			}

		default:
			log.Warn("Unknown input resource type", zap.String("type", string(input.Type)))
			// For unknown types, we'll allow the action to proceed
		}
	}

	log.Debug("✅ Action input validation passed")
	return nil
}
