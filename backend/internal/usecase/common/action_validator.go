package common

import (
	"context"
	"fmt"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// ActionValidatorImpl implements ActionValidator interface
type ActionValidatorImpl struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
}

// NewActionValidator creates a new ActionValidator instance
func NewActionValidator(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
) ActionValidator {
	return &ActionValidatorImpl{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
	}
}

// ValidatePlayerAction checks if a player can perform an action
func (v *ActionValidatorImpl) ValidatePlayerAction(ctx context.Context, gameID, playerID string, cost ActionCost) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get game
	game, err := v.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for action validation", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Check if game is active
	if game.Status != model.GameStatusActive {
		log.Warn("Action attempted on non-active game", zap.String("game_status", string(game.Status)))
		return fmt.Errorf("game is not active, current status: %s", game.Status)
	}

	// Check if it's the player's turn
	if game.CurrentTurn == nil || *game.CurrentTurn != playerID {
		currentTurn := "none"
		if game.CurrentTurn != nil {
			currentTurn = *game.CurrentTurn
		}
		log.Warn("Action attempted by player when not their turn",
			zap.String("current_turn", currentTurn))
		return fmt.Errorf("not player's turn, current turn: %s", currentTurn)
	}

	// Get player to check resources and available actions
	player, err := v.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for resource validation", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Check if player has available actions
	if player.AvailableActions <= 0 {
		log.Warn("Action attempted with no remaining actions",
			zap.Int("available_actions", player.AvailableActions))
		return fmt.Errorf("no remaining actions, current count: %d", player.AvailableActions)
	}

	// Validate resource costs
	if err := v.validateResourceCost(player, cost); err != nil {
		log.Warn("Player has insufficient resources",
			zap.Any("required_cost", cost),
			zap.Any("player_resources", player.Resources))
		return fmt.Errorf("insufficient resources: %w", err)
	}

	log.Debug("Action validation passed",
		zap.Any("cost", cost),
		zap.Int("available_actions", player.AvailableActions))

	return nil
}

// validateResourceCost checks if player has sufficient resources
func (v *ActionValidatorImpl) validateResourceCost(player model.Player, cost ActionCost) error {
	if player.Resources.Credits < cost.Credits {
		return fmt.Errorf("need %d credits, have %d", cost.Credits, player.Resources.Credits)
	}
	if player.Resources.Steel < cost.Steel {
		return fmt.Errorf("need %d steel, have %d", cost.Steel, player.Resources.Steel)
	}
	if player.Resources.Titanium < cost.Titanium {
		return fmt.Errorf("need %d titanium, have %d", cost.Titanium, player.Resources.Titanium)
	}
	if player.Resources.Plants < cost.Plants {
		return fmt.Errorf("need %d plants, have %d", cost.Plants, player.Resources.Plants)
	}
	if player.Resources.Energy < cost.Energy {
		return fmt.Errorf("need %d energy, have %d", cost.Energy, player.Resources.Energy)
	}
	if player.Resources.Heat < cost.Heat {
		return fmt.Errorf("need %d heat, have %d", cost.Heat, player.Resources.Heat)
	}
	return nil
}
