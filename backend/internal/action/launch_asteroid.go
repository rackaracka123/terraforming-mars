package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

const (
	// LaunchAsteroidCost is the megacredit cost to launch an asteroid via standard project
	LaunchAsteroidCost = 14
)

// LaunchAsteroidAction handles the business logic for the launch asteroid standard project
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type LaunchAsteroidAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewLaunchAsteroidAction creates a new launch asteroid action
func NewLaunchAsteroidAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *LaunchAsteroidAction {
	return &LaunchAsteroidAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the launch asteroid action
func (a *LaunchAsteroidAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "launch_asteroid"),
	)
	log.Info("☄️ Launching asteroid")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. BUSINESS LOGIC: Validate game is active
	if g.Status() != game.GameStatusActive {
		log.Warn("Game is not active", zap.String("status", string(g.Status())))
		return fmt.Errorf("game is not active: %s", g.Status())
	}

	// 3. BUSINESS LOGIC: Validate it's the player's turn
	currentTurn := g.CurrentTurn()
	if currentTurn == nil || currentTurn.PlayerID() != playerID {
		var turnPlayerID string
		if currentTurn != nil {
			turnPlayerID = currentTurn.PlayerID()
		}
		log.Warn("Not player's turn",
			zap.String("current_turn_player", turnPlayerID),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("not your turn")
	}

	// 4. Get player from game
	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 5. BUSINESS LOGIC: Validate cost (14 M€)
	resources := player.Resources().Get()
	if resources.Credits < LaunchAsteroidCost {
		log.Warn("Insufficient credits for asteroid",
			zap.Int("cost", LaunchAsteroidCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", LaunchAsteroidCost, resources.Credits)
	}

	// 6. BUSINESS LOGIC: Deduct cost using domain method
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: -LaunchAsteroidCost,
	})

	resources = player.Resources().Get() // Refresh after update
	log.Info("💰 Deducted asteroid cost",
		zap.Int("cost", LaunchAsteroidCost),
		zap.Int("remaining_credits", resources.Credits))

	// 7. BUSINESS LOGIC: Increase temperature by 1 step
	oldTemp := g.GlobalParameters().Temperature()
	stepsRaised, err := g.GlobalParameters().IncreaseTemperature(ctx, 1)
	if err != nil {
		log.Error("Failed to increase temperature", zap.Error(err))
		return fmt.Errorf("failed to increase temperature: %w", err)
	}
	newTemp := g.GlobalParameters().Temperature()

	if stepsRaised > 0 {
		log.Info("🌡️ Increased temperature",
			zap.Int("old_temperature", oldTemp),
			zap.Int("new_temperature", newTemp),
			zap.Int("steps_raised", stepsRaised))
	}

	// 8. BUSINESS LOGIC: Increase terraform rating (only if temperature actually increased)
	if stepsRaised > 0 {
		oldTR := player.Resources().TerraformRating()
		newTR := oldTR + 1
		player.Resources().SetTerraformRating(newTR)

		log.Info("🏆 Increased terraform rating",
			zap.Int("old_tr", oldTR),
			zap.Int("new_tr", newTR))
	}

	// 9. BUSINESS LOGIC: Consume action (only if not unlimited actions)
	availableActions := player.Turn().AvailableActions()
	if availableActions > 0 {
		player.Turn().SetAvailableActions(availableActions - 1)
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", availableActions-1))
	}

	// 10. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - player.Resources().Add() publishes ResourcesChangedEvent
	//    - g.GlobalParameters().IncreaseTemperature() publishes TemperatureChangedEvent + BroadcastEvent
	//    - player.Resources().SetTerraformRating() publishes TerraformRatingChangedEvent
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("✅ Asteroid launched successfully",
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
