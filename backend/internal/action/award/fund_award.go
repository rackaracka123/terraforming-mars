package award

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	baseaction "terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// FundAwardAction handles the business logic for funding an award
type FundAwardAction struct {
	baseaction.BaseAction
}

// NewFundAwardAction creates a new fund award action
func NewFundAwardAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
) *FundAwardAction {
	return &FundAwardAction{
		BaseAction: baseaction.NewBaseAction(gameRepo, cardRegistry),
	}
}

// Execute funds an award for the player
func (a *FundAwardAction) Execute(ctx context.Context, gameID string, playerID string, awardType string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "fund_award"), zap.String("award", awardType))
	log.Info("🏅 Funding award")

	// 1. Validate award type
	if !shared.ValidAwardType(awardType) {
		log.Warn("Invalid award type", zap.String("award_type", awardType))
		return fmt.Errorf("invalid award type: %s", awardType)
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

	// 5. Validate award is not already funded
	awards := g.Awards()
	at := shared.AwardType(awardType)
	if awards.IsFunded(at) {
		log.Warn("Award already funded", zap.String("award", awardType))
		return fmt.Errorf("award %s is already funded", awardType)
	}

	// 6. Validate max awards not reached
	if !awards.CanFundMore() {
		log.Warn("Maximum awards already funded", zap.Int("max", game.MaxFundedAwards))
		return fmt.Errorf("maximum awards (%d) already funded", game.MaxFundedAwards)
	}

	// 7. Get funding cost and validate player has enough credits
	fundingCost := awards.GetCurrentFundingCost()
	resources := player.Resources().Get()
	if resources.Credits < fundingCost {
		log.Warn("Insufficient credits for award",
			zap.Int("cost", fundingCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", fundingCost, resources.Credits)
	}

	// 8. Deduct cost
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -fundingCost,
	})
	log.Info("💰 Deducted award funding cost",
		zap.Int("cost", fundingCost),
		zap.Int("remaining_credits", player.Resources().Get().Credits))

	// 9. Fund the award
	if err := awards.FundAward(ctx, at, playerID); err != nil {
		log.Error("Failed to fund award", zap.Error(err))
		return fmt.Errorf("failed to fund award: %w", err)
	}

	// 10. Consume action
	a.ConsumePlayerAction(g, log)

	log.Info("✅ Award funded successfully",
		zap.String("award", awardType),
		zap.Int("total_funded", awards.FundedCount()))

	return nil
}
