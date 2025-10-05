package service

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// EffectProcessor handles triggering passive effects when game events occur
type EffectProcessor interface {
	// TriggerEffects finds and executes all passive effects matching the given event type
	TriggerEffects(ctx context.Context, gameID string, eventType model.TriggerType, eventContext model.EffectContext) error
}

// EffectProcessorImpl implements EffectProcessor
type EffectProcessorImpl struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
}

// NewEffectProcessor creates a new EffectProcessor instance
func NewEffectProcessor(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) EffectProcessor {
	return &EffectProcessorImpl{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
	}
}

// TriggerEffects finds all players with passive effects matching the event type and executes them
func (ep *EffectProcessorImpl) TriggerEffects(ctx context.Context, gameID string, eventType model.TriggerType, eventContext model.EffectContext) error {
	log := logger.Get()
	log.Debug("🎆 Triggering passive effects",
		zap.String("game_id", gameID),
		zap.String("event_type", string(eventType)),
		zap.String("triggering_player", eventContext.TriggeringPlayerID))

	// Get all players in the game
	players, err := ep.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	// Track total effects triggered for logging
	var totalEffectsTriggered int

	// Check each player for matching passive effects
	for _, player := range players {
		for _, effect := range player.Effects {
			// Check if this effect matches the event type
			if len(effect.Behavior.Triggers) == 0 {
				continue
			}

			trigger := effect.Behavior.Triggers[0]

			// Verify this is an auto trigger with a condition
			if trigger.Type != model.ResourceTriggerAuto || trigger.Condition == nil {
				continue
			}

			// Check if the trigger condition matches the event type
			if trigger.Condition.Type != eventType {
				continue
			}

			// Effect matches! Execute it
			log.Info("✨ Triggering passive effect",
				zap.String("player_id", player.ID),
				zap.String("card_name", effect.CardName),
				zap.String("event_type", string(eventType)))

			if err := ep.executeEffect(ctx, gameID, player.ID, effect, eventContext); err != nil {
				log.Error("Failed to execute passive effect",
					zap.String("player_id", player.ID),
					zap.String("card_name", effect.CardName),
					zap.Error(err))
				return fmt.Errorf("failed to execute effect for player %s: %w", player.ID, err)
			}

			totalEffectsTriggered++
		}
	}

	if totalEffectsTriggered > 0 {
		log.Info("🎊 Passive effects triggered successfully",
			zap.String("game_id", gameID),
			zap.String("event_type", string(eventType)),
			zap.Int("effects_count", totalEffectsTriggered))
	}

	return nil
}

// executeEffect executes a single passive effect's outputs
func (ep *EffectProcessorImpl) executeEffect(ctx context.Context, gameID, playerID string, effect model.PlayerEffect, eventContext model.EffectContext) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player state
	player, err := ep.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Process each output from the effect
	for _, output := range effect.Behavior.Outputs {
		switch output.Type {
		case model.ResourceCredits:
			// Award credits
			newResources := player.Resources
			newResources.Credits += output.Amount
			if err := ep.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
				return fmt.Errorf("failed to update credits: %w", err)
			}
			log.Info("💰 Passive effect: credits gained",
				zap.String("card_name", effect.CardName),
				zap.Int("amount", output.Amount))

		case model.ResourceSteel:
			newResources := player.Resources
			newResources.Steel += output.Amount
			if err := ep.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
				return fmt.Errorf("failed to update steel: %w", err)
			}
			log.Info("🔩 Passive effect: steel gained",
				zap.String("card_name", effect.CardName),
				zap.Int("amount", output.Amount))

		case model.ResourceTitanium:
			newResources := player.Resources
			newResources.Titanium += output.Amount
			if err := ep.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
				return fmt.Errorf("failed to update titanium: %w", err)
			}
			log.Info("⚙️ Passive effect: titanium gained",
				zap.String("card_name", effect.CardName),
				zap.Int("amount", output.Amount))

		case model.ResourcePlants:
			newResources := player.Resources
			newResources.Plants += output.Amount
			if err := ep.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
				return fmt.Errorf("failed to update plants: %w", err)
			}
			log.Info("🌱 Passive effect: plants gained",
				zap.String("card_name", effect.CardName),
				zap.Int("amount", output.Amount))

		case model.ResourceEnergy:
			newResources := player.Resources
			newResources.Energy += output.Amount
			if err := ep.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
				return fmt.Errorf("failed to update energy: %w", err)
			}
			log.Info("⚡ Passive effect: energy gained",
				zap.String("card_name", effect.CardName),
				zap.Int("amount", output.Amount))

		case model.ResourceHeat:
			newResources := player.Resources
			newResources.Heat += output.Amount
			if err := ep.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
				return fmt.Errorf("failed to update heat: %w", err)
			}
			log.Info("🔥 Passive effect: heat gained",
				zap.String("card_name", effect.CardName),
				zap.Int("amount", output.Amount))

		case model.ResourceCreditsProduction:
			newProduction := player.Production
			newProduction.Credits += output.Amount
			if err := ep.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
				return fmt.Errorf("failed to update credits production: %w", err)
			}
			log.Info("💰📈 Passive effect: credits production gained",
				zap.String("card_name", effect.CardName),
				zap.Int("amount", output.Amount))

		case model.ResourceSteelProduction:
			newProduction := player.Production
			newProduction.Steel += output.Amount
			if err := ep.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
				return fmt.Errorf("failed to update steel production: %w", err)
			}
			log.Info("🔩📈 Passive effect: steel production gained",
				zap.String("card_name", effect.CardName),
				zap.Int("amount", output.Amount))

		case model.ResourceTitaniumProduction:
			newProduction := player.Production
			newProduction.Titanium += output.Amount
			if err := ep.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
				return fmt.Errorf("failed to update titanium production: %w", err)
			}
			log.Info("⚙️📈 Passive effect: titanium production gained",
				zap.String("card_name", effect.CardName),
				zap.Int("amount", output.Amount))

		case model.ResourcePlantsProduction:
			newProduction := player.Production
			newProduction.Plants += output.Amount
			if err := ep.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
				return fmt.Errorf("failed to update plants production: %w", err)
			}
			log.Info("🌱📈 Passive effect: plants production gained",
				zap.String("card_name", effect.CardName),
				zap.Int("amount", output.Amount))

		case model.ResourceEnergyProduction:
			newProduction := player.Production
			newProduction.Energy += output.Amount
			if err := ep.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
				return fmt.Errorf("failed to update energy production: %w", err)
			}
			log.Info("⚡📈 Passive effect: energy production gained",
				zap.String("card_name", effect.CardName),
				zap.Int("amount", output.Amount))

		case model.ResourceHeatProduction:
			newProduction := player.Production
			newProduction.Heat += output.Amount
			if err := ep.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
				return fmt.Errorf("failed to update heat production: %w", err)
			}
			log.Info("🔥📈 Passive effect: heat production gained",
				zap.String("card_name", effect.CardName),
				zap.Int("amount", output.Amount))

		// Add more resource types as needed (microbes, animals, etc.)
		default:
			log.Warn("⚠️ Unsupported output type in passive effect",
				zap.String("card_name", effect.CardName),
				zap.String("output_type", string(output.Type)))
		}

		// Refresh player state after each update
		player, err = ep.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to refresh player state: %w", err)
		}
	}

	return nil
}
