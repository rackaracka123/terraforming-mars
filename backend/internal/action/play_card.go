package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
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
	if err := validateCardRequirements(card, g, player); err != nil {
		log.Error("Card requirements not met", zap.Error(err))
		return fmt.Errorf("cannot play card: %w", err)
	}

	log.Debug("✅ Card requirements validated")

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

	// 12. STATE UPDATE: Add card to played cards (publishes CardPlayedEvent)
	player.PlayedCards().AddCard(cardID, card.Name, string(card.Type))

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

	// 14. BUSINESS LOGIC: Apply card immediate effects and register behaviors
	if err := a.applyCardBehaviors(ctx, g, card, player, log); err != nil {
		log.Error("Failed to apply card behaviors", zap.Error(err))
		return fmt.Errorf("failed to apply card behaviors: %w", err)
	}

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
func hasTag(card *gamecards.Card, tag shared.CardTag) bool {
	for _, cardTag := range card.Tags {
		if cardTag == tag {
			return true
		}
	}
	return false
}

// validateCardRequirements validates that the player and game state meet all card requirements
func validateCardRequirements(card *gamecards.Card, g *game.Game, player *player.Player) error {
	if len(card.Requirements) == 0 {
		return nil // No requirements to validate
	}

	for _, req := range card.Requirements {
		switch req.Type {
		case gamecards.RequirementTemperature:
			temp := g.GlobalParameters().Temperature()
			if req.Min != nil && temp < *req.Min {
				return fmt.Errorf("temperature requirement not met: need %d°C, current %d°C", *req.Min, temp)
			}
			if req.Max != nil && temp > *req.Max {
				return fmt.Errorf("temperature requirement not met: max %d°C, current %d°C", *req.Max, temp)
			}

		case gamecards.RequirementOxygen:
			oxygen := g.GlobalParameters().Oxygen()
			if req.Min != nil && oxygen < *req.Min {
				return fmt.Errorf("oxygen requirement not met: need %d%%, current %d%%", *req.Min, oxygen)
			}
			if req.Max != nil && oxygen > *req.Max {
				return fmt.Errorf("oxygen requirement not met: max %d%%, current %d%%", *req.Max, oxygen)
			}

		case gamecards.RequirementOceans:
			oceans := g.GlobalParameters().Oceans()
			if req.Min != nil && oceans < *req.Min {
				return fmt.Errorf("ocean requirement not met: need %d, current %d", *req.Min, oceans)
			}
			if req.Max != nil && oceans > *req.Max {
				return fmt.Errorf("ocean requirement not met: max %d, current %d", *req.Max, oceans)
			}

		case gamecards.RequirementTR:
			tr := player.Resources().TerraformRating()
			if req.Min != nil && tr < *req.Min {
				return fmt.Errorf("terraform rating requirement not met: need %d, current %d", *req.Min, tr)
			}
			if req.Max != nil && tr > *req.Max {
				return fmt.Errorf("terraform rating requirement not met: max %d, current %d", *req.Max, tr)
			}

		case gamecards.RequirementTags:
			if req.Tag == nil {
				return fmt.Errorf("tag requirement missing tag specification")
			}

			// Count tags across all played cards (including corporation)
			tagCount := 0
			for _, playedCardID := range player.PlayedCards().Cards() {
				// TODO: Get card from registry and check if it has the tag
				// This requires injecting CardRegistry into this function
				// For now, skip tag validation
				_ = playedCardID
			}

			if req.Min != nil && tagCount < *req.Min {
				return fmt.Errorf("tag requirement not met: need %d %s tags, have %d", *req.Min, *req.Tag, tagCount)
			}
			if req.Max != nil && tagCount > *req.Max {
				return fmt.Errorf("tag requirement not met: max %d %s tags, have %d", *req.Max, *req.Tag, tagCount)
			}

		case gamecards.RequirementProduction:
			if req.Resource == nil {
				return fmt.Errorf("production requirement missing resource specification")
			}
			// TODO: Implement production requirement validation
			// This requires checking player's production values
			// For now, skip production validation

		case gamecards.RequirementResource:
			if req.Resource == nil {
				return fmt.Errorf("resource requirement missing resource specification")
			}
			resources := player.Resources().Get()
			var currentAmount int

			switch *req.Resource {
			case shared.ResourceCredits:
				currentAmount = resources.Credits
			case shared.ResourceSteel:
				currentAmount = resources.Steel
			case shared.ResourceTitanium:
				currentAmount = resources.Titanium
			case shared.ResourcePlants:
				currentAmount = resources.Plants
			case shared.ResourceEnergy:
				currentAmount = resources.Energy
			case shared.ResourceHeat:
				currentAmount = resources.Heat
			}

			if req.Min != nil && currentAmount < *req.Min {
				return fmt.Errorf("resource requirement not met: need %d %s, have %d", *req.Min, *req.Resource, currentAmount)
			}
			if req.Max != nil && currentAmount > *req.Max {
				return fmt.Errorf("resource requirement not met: max %d %s, have %d", *req.Max, *req.Resource, currentAmount)
			}

		case gamecards.RequirementCities, gamecards.RequirementGreeneries:
			// TODO: Implement tile-based requirements when Board tile counting is ready
			// For now, skip these validations

		case gamecards.RequirementVenus:
			// TODO: Implement Venus track when expansion is supported
			// For now, skip Venus validation
		}
	}

	return nil
}

// applyCardBehaviors processes all card behaviors and applies immediate effects or registers actions/effects
func (a *PlayCardAction) applyCardBehaviors(
	ctx context.Context,
	g *game.Game,
	card *gamecards.Card,
	p *player.Player,
	log *zap.Logger,
) error {
	if len(card.Behaviors) == 0 {
		log.Debug("No card behaviors to apply")
		return nil
	}

	log.Info("🎴 Processing card behaviors",
		zap.String("card_id", card.ID),
		zap.Int("behavior_count", len(card.Behaviors)))

	for behaviorIndex, behavior := range card.Behaviors {
		log.Debug("Processing behavior",
			zap.Int("index", behaviorIndex),
			zap.Int("trigger_count", len(behavior.Triggers)))

		// Apply auto-trigger behaviors immediately
		if gamecards.HasAutoTrigger(behavior) {
			log.Info("✨ Found auto-trigger behavior, applying outputs immediately",
				zap.Int("output_count", len(behavior.Outputs)))

			if err := a.applyBehaviorOutputs(ctx, g, behavior.Outputs, p, card.Name, log); err != nil {
				return fmt.Errorf("failed to apply auto behavior %d outputs: %w", behaviorIndex, err)
			}
		}

		// Register manual-trigger behaviors as player actions
		if gamecards.HasManualTrigger(behavior) {
			log.Info("🎯 Found manual-trigger behavior, registering as player action")

			p.Actions().AddAction(player.CardAction{
				CardID:        card.ID,
				CardName:      card.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
				PlayCount:     0,
			})
		}

		// Register conditional-trigger behaviors as passive effects
		if gamecards.HasConditionalTrigger(behavior) {
			log.Info("⚡ Found conditional-trigger behavior, registering as passive effect",
				zap.Int("trigger_count", len(behavior.Triggers)))

			effect := player.CardEffect{
				CardID:        card.ID,
				CardName:      card.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			}
			p.Effects().AddEffect(effect)

			// Subscribe passive effects to relevant events
			subscribePassiveEffectToEvents(ctx, g, p, effect, log)
		}
	}

	log.Info("✅ All card behaviors processed successfully")
	return nil
}

// applyBehaviorOutputs applies resource and production outputs to a player
func (a *PlayCardAction) applyBehaviorOutputs(
	ctx context.Context,
	g *game.Game,
	outputs []shared.ResourceCondition,
	p *player.Player,
	cardName string,
	log *zap.Logger,
) error {
	for _, output := range outputs {
		switch output.ResourceType {
		// Basic resources
		case shared.ResourceCredits:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceCredits: output.Amount,
			})
			log.Info("💰 Added credits", zap.Int("amount", output.Amount))

		case shared.ResourceSteel:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceSteel: output.Amount,
			})
			log.Info("🔩 Added steel", zap.Int("amount", output.Amount))

		case shared.ResourceTitanium:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceTitanium: output.Amount,
			})
			log.Info("⚙️ Added titanium", zap.Int("amount", output.Amount))

		case shared.ResourcePlants:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourcePlants: output.Amount,
			})
			log.Info("🌱 Added plants", zap.Int("amount", output.Amount))

		case shared.ResourceEnergy:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceEnergy: output.Amount,
			})
			log.Info("⚡ Added energy", zap.Int("amount", output.Amount))

		case shared.ResourceHeat:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceHeat: output.Amount,
			})
			log.Info("🔥 Added heat", zap.Int("amount", output.Amount))

		// Production resources
		case shared.ResourceCreditsProduction:
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceCreditsProduction: output.Amount,
			})
			log.Info("💰 Added credits production", zap.Int("amount", output.Amount))

		case shared.ResourceSteelProduction:
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceSteelProduction: output.Amount,
			})
			log.Info("🔩 Added steel production", zap.Int("amount", output.Amount))

		case shared.ResourceTitaniumProduction:
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceTitaniumProduction: output.Amount,
			})
			log.Info("⚙️ Added titanium production", zap.Int("amount", output.Amount))

		case shared.ResourcePlantsProduction:
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourcePlantsProduction: output.Amount,
			})
			log.Info("🌱 Added plants production", zap.Int("amount", output.Amount))

		case shared.ResourceEnergyProduction:
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceEnergyProduction: output.Amount,
			})
			log.Info("⚡ Added energy production", zap.Int("amount", output.Amount))

		case shared.ResourceHeatProduction:
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceHeatProduction: output.Amount,
			})
			log.Info("🔥 Added heat production", zap.Int("amount", output.Amount))

		case shared.ResourceTR:
			p.Resources().UpdateTerraformRating(output.Amount)
			log.Info("🌍 Added terraform rating", zap.Int("amount", output.Amount))

		// Tile placements - append to queue for user selection
		case shared.ResourceCityPlacement, shared.ResourceGreeneryPlacement, shared.ResourceOceanPlacement:
			// Map resource type to tile type string
			var tileType string
			switch output.ResourceType {
			case shared.ResourceCityPlacement:
				tileType = "city"
			case shared.ResourceGreeneryPlacement:
				tileType = "greenery"
			case shared.ResourceOceanPlacement:
				tileType = "ocean"
			}

			// Build array of tile types to append (for multiple placements)
			tileTypes := make([]string, output.Amount)
			for i := 0; i < output.Amount; i++ {
				tileTypes[i] = tileType
			}

			// Atomically append to queue (thread-safe)
			if err := g.AppendToPendingTileSelectionQueue(ctx, p.ID(), tileTypes, cardName); err != nil {
				return fmt.Errorf("failed to append to pending tile selection queue: %w", err)
			}

			log.Info("🏗️ Added tile placements to queue",
				zap.String("tile_type", tileType),
				zap.Int("count", output.Amount),
				zap.String("source", cardName))

		default:
			log.Warn("⚠️  Unhandled output type in card behavior",
				zap.String("type", string(output.ResourceType)))
		}
	}

	return nil
}
