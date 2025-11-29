package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// PlayCardAction handles the business logic for playing a project card from hand
// Card playing involves: validating requirements, calculating costs (with discounts),
// moving card to played cards, applying immediate effects, and deducting payment
type PlayCardAction struct {
	BaseAction
}

// NewPlayCardAction creates a new play card action
func NewPlayCardAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *PlayCardAction {
	return &PlayCardAction{
		BaseAction: BaseAction{
			gameRepo:     gameRepo,
			cardRegistry: cardRegistry,
			logger:       logger,
		},
	}
}

// PaymentRequest represents the payment resources provided by the player
type PaymentRequest struct {
	Credits  int `json:"credits"`
	Steel    int `json:"steel"`
	Titanium int `json:"titanium"`
}

// Execute performs the play card action
func (a *PlayCardAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
	cardID string,
	payment PaymentRequest,
) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("card_id", cardID),
		zap.String("action", "play_card"),
	)
	log.Info("🃏 Player attempting to play card")

	// 1. Validate game exists and is active
	g, err := ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate game is in action phase
	if err := ValidateGamePhase(g, game.GamePhaseAction, log); err != nil {
		return err
	}

	// 3. Validate it's the player's turn
	if err := ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 4. Get player from game
	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	// 5. BUSINESS LOGIC: Validate card is in player's hand
	if !player.Hand().HasCard(cardID) {
		log.Error("Card not in player's hand")
		return fmt.Errorf("card %s not in hand", cardID)
	}

	// 6. BUSINESS LOGIC: Get card data from registry
	card, err := a.cardRegistry.GetByID(cardID)
	if err != nil {
		log.Error("Card not found in registry", zap.Error(err))
		return fmt.Errorf("card not found: %w", err)
	}

	log.Debug("Card data retrieved",
		zap.String("card_name", card.Name),
		zap.Int("base_cost", card.Cost))

	// 7. BUSINESS LOGIC: Validate card requirements (temperature, oxygen, tags, etc.)
	// TODO: Implement full requirement validation when card requirements system is ready
	// For now, skip requirement validation

	// 8. BUSINESS LOGIC: Calculate effective cost with discounts
	effectiveCost := card.Cost
	steelDiscount := 0
	titaniumDiscount := 0

	// Steel reduces building costs by 2 MC per steel
	if hasTag(card, shared.TagBuilding) {
		steelDiscount = payment.Steel * 2
	}

	// Titanium reduces space costs by 3 MC per titanium
	if hasTag(card, shared.TagSpace) {
		titaniumDiscount = payment.Titanium * 3
	}

	effectiveCost -= steelDiscount + titaniumDiscount

	if effectiveCost < 0 {
		effectiveCost = 0
	}

	log.Debug("Cost calculated",
		zap.Int("base_cost", card.Cost),
		zap.Int("steel_discount", steelDiscount),
		zap.Int("titanium_discount", titaniumDiscount),
		zap.Int("effective_cost", effectiveCost))

	// 9. BUSINESS LOGIC: Validate payment covers effective cost
	if payment.Credits < effectiveCost {
		log.Error("Insufficient credits",
			zap.Int("required", effectiveCost),
			zap.Int("provided", payment.Credits))
		return fmt.Errorf("insufficient credits: need %d, provided %d", effectiveCost, payment.Credits)
	}

	// 10. BUSINESS LOGIC: Validate player has the resources they're trying to spend
	resources := player.Resources().Get()
	if resources.Credits < payment.Credits {
		log.Error("Player doesn't have enough credits",
			zap.Int("has", resources.Credits),
			zap.Int("trying_to_spend", payment.Credits))
		return fmt.Errorf("insufficient credits: have %d, trying to spend %d", resources.Credits, payment.Credits)
	}

	if resources.Steel < payment.Steel {
		log.Error("Player doesn't have enough steel",
			zap.Int("has", resources.Steel),
			zap.Int("trying_to_spend", payment.Steel))
		return fmt.Errorf("insufficient steel: have %d, trying to spend %d", resources.Steel, payment.Steel)
	}

	if resources.Titanium < payment.Titanium {
		log.Error("Player doesn't have enough titanium",
			zap.Int("has", resources.Titanium),
			zap.Int("trying_to_spend", payment.Titanium))
		return fmt.Errorf("insufficient titanium: have %d, trying to spend %d", resources.Titanium, payment.Titanium)
	}

	// 11. STATE UPDATE: Remove card from hand
	if !player.Hand().RemoveCard(cardID) {
		log.Error("Failed to remove card from hand - card not found")
		return fmt.Errorf("failed to remove card from hand: card not found")
	}

	log.Info("✅ Card removed from hand")

	// 12. STATE UPDATE: Add card to played cards
	player.PlayedCards().AddCard(cardID)

	log.Info("✅ Card added to played cards")

	// 13. STATE UPDATE: Deduct payment from player resources (using negative values)
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits:  -payment.Credits,
		shared.ResourceSteel:    -payment.Steel,
		shared.ResourceTitanium: -payment.Titanium,
	})

	log.Info("✅ Payment deducted",
		zap.Int("credits", payment.Credits),
		zap.Int("steel", payment.Steel),
		zap.Int("titanium", payment.Titanium))

	// 14. BUSINESS LOGIC: Apply card immediate effects
	// TODO: Implement card effect application when effect system is ready
	// For now, skip immediate effects
	log.Debug("⚠️  Card effect application not yet implemented")

	// 15. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//     - player.Hand().RemoveCard() publishes CardHandUpdatedEvent
	//     - player.PlayedCards().AddCard() publishes CardPlayedEvent
	//     - player.Resources().Subtract() publishes ResourcesChangedEvent
	//     Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("🎉 Card played successfully",
		zap.String("card_name", card.Name),
		zap.Int("cost_paid", effectiveCost))

	return nil
}

// hasTag checks if a card has a specific tag
func hasTag(card *game.Card, tag shared.CardTag) bool {
	for _, cardTag := range card.Tags {
		if cardTag == tag {
			return true
		}
	}
	return false
}
