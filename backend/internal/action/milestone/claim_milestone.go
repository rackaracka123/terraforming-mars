package milestone

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	baseaction "terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
)

// ClaimMilestoneAction handles the business logic for claiming a milestone
type ClaimMilestoneAction struct {
	baseaction.BaseAction
}

// NewClaimMilestoneAction creates a new claim milestone action
func NewClaimMilestoneAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
) *ClaimMilestoneAction {
	return &ClaimMilestoneAction{
		BaseAction: baseaction.NewBaseAction(gameRepo, cardRegistry),
	}
}

// Execute claims a milestone for the player
func (a *ClaimMilestoneAction) Execute(ctx context.Context, gameID string, playerID string, milestoneType string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "claim_milestone"), zap.String("milestone", milestoneType))
	log.Info("🏆 Claiming milestone")

	// 1. Validate milestone type
	if !game.ValidMilestoneType(milestoneType) {
		log.Warn("Invalid milestone type", zap.String("milestone_type", milestoneType))
		return fmt.Errorf("invalid milestone type: %s", milestoneType)
	}

	// 2. Fetch game from repository and validate it's active
	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	// 3. Validate it's the player's turn
	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 4. Get player from game
	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	// 5. Validate milestone is not already claimed
	milestones := g.Milestones()
	mt := game.MilestoneType(milestoneType)
	if milestones.IsClaimed(mt) {
		log.Warn("Milestone already claimed", zap.String("milestone", milestoneType))
		return fmt.Errorf("milestone %s is already claimed", milestoneType)
	}

	// 6. Validate max milestones not reached
	if !milestones.CanClaimMore() {
		log.Warn("Maximum milestones already claimed", zap.Int("max", game.MaxClaimedMilestones))
		return fmt.Errorf("maximum milestones (%d) already claimed", game.MaxClaimedMilestones)
	}

	// 7. Validate player has enough credits (8 MC)
	resources := player.Resources().Get()
	if resources.Credits < game.MilestoneClaimCost {
		log.Warn("Insufficient credits for milestone",
			zap.Int("cost", game.MilestoneClaimCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", game.MilestoneClaimCost, resources.Credits)
	}

	// 8. Validate player meets milestone requirements
	milestoneTypeForValidator := gamecards.MilestoneType(milestoneType)
	if !gamecards.CanClaimMilestone(milestoneTypeForValidator, player, g.Board(), a.CardRegistry()) {
		requirement := gamecards.GetMilestoneRequirement(milestoneTypeForValidator)
		progress := gamecards.GetPlayerMilestoneProgress(milestoneTypeForValidator, player, g.Board(), a.CardRegistry())
		log.Warn("Player does not meet milestone requirements",
			zap.String("requirement", requirement.Description),
			zap.Int("required", requirement.Required),
			zap.Int("current", progress))
		return fmt.Errorf("requirements not met: %s (have %d, need %d)", requirement.Description, progress, requirement.Required)
	}

	// 9. Deduct cost
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -game.MilestoneClaimCost,
	})
	log.Info("💰 Deducted milestone cost",
		zap.Int("cost", game.MilestoneClaimCost),
		zap.Int("remaining_credits", player.Resources().Get().Credits))

	// 10. Claim the milestone
	if err := milestones.ClaimMilestone(ctx, mt, playerID, g.Generation()); err != nil {
		log.Error("Failed to claim milestone", zap.Error(err))
		return fmt.Errorf("failed to claim milestone: %w", err)
	}

	// 11. Consume action
	a.ConsumePlayerAction(g, log)

	log.Info("✅ Milestone claimed successfully",
		zap.String("milestone", milestoneType),
		zap.Int("total_claimed", milestones.ClaimedCount()))

	return nil
}
