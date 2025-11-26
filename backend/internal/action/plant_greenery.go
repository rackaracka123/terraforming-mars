package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"
	playerTypes "terraforming-mars-backend/internal/session/game/player"

	"go.uber.org/zap"
)

const (
	// PlantGreeneryStandardProjectCost is the megacredit cost to plant greenery via standard project
	PlantGreeneryStandardProjectCost = 23
)

// PlantGreeneryAction handles the business logic for the plant greenery standard project
type PlantGreeneryAction struct {
	BaseAction
	gameRepo game.Repository
}

// NewPlantGreeneryAction creates a new plant greenery action
func NewPlantGreeneryAction(
	gameRepo game.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *PlantGreeneryAction {
	return &PlantGreeneryAction{
		BaseAction: NewBaseAction(sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the plant greenery action
func (a *PlantGreeneryAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID)
	log.Info("🌱 Planting greenery (standard project)")

	// 1. Validate game is active
	g, err := ValidateActiveGame(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate it's the player's turn
	if err := ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 3. Get session and player
	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. Validate cost (23 M€)
	resources := player.Resources().Get()
	if resources.Credits < PlantGreeneryStandardProjectCost {
		log.Warn("Insufficient credits for greenery",
			zap.Int("cost", PlantGreeneryStandardProjectCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", PlantGreeneryStandardProjectCost, resources.Credits)
	}

	// 5. Deduct cost
	resources.Credits -= PlantGreeneryStandardProjectCost
	player.Resources().Set(resources)

	log.Info("💰 Deducted greenery cost",
		zap.Int("cost", PlantGreeneryStandardProjectCost),
		zap.Int("remaining_credits", resources.Credits))

	// 6. Create tile queue with "greenery" type on Game (phase state managed by Game)
	queue := &playerTypes.PendingTileSelectionQueue{
		Items:  []string{"greenery"},
		Source: "standard-project-greenery",
	}
	if err := sess.Game().SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
		return fmt.Errorf("failed to queue tile placement: %w", err)
	}

	log.Info("📋 Created tile queue for greenery placement")

	// Note: Terraform rating increase happens when the greenery is placed (via SelectTileAction)

	// 7. Consume action (only if not unlimited actions)
	availableActions := player.Turn().AvailableActions()
	if availableActions > 0 {
		player.Turn().SetAvailableActions(availableActions - 1)
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", availableActions-1))
	}

	// 8. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Greenery queued successfully, tile awaiting placement",
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
