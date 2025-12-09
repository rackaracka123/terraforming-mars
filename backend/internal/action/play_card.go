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
	Credits     int                         `json:"credits"`
	Steel       int                         `json:"steel"`
	Titanium    int                         `json:"titanium"`
	Substitutes map[shared.ResourceType]int `json:"substitutes"`
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
	if err := validateCardRequirements(card, g, player, a.cardRegistry); err != nil {
		log.Error("Card requirements not met", zap.Error(err))
		return fmt.Errorf("cannot play card: %w", err)
	}

	log.Debug("✅ Card requirements validated")

	// 8. BUSINESS LOGIC: Convert payment request to CardPayment for validation
	cardPayment := gamecards.CardPayment{
		Credits:     payment.Credits,
		Steel:       payment.Steel,
		Titanium:    payment.Titanium,
		Substitutes: payment.Substitutes,
	}

	// Get player's payment substitutes (e.g., Helion can use heat as credits)
	playerSubstitutes := player.Resources().PaymentSubstitutes()

	// Check if card allows steel/titanium
	allowSteel := hasTag(card, shared.TagBuilding)
	allowTitanium := hasTag(card, shared.TagSpace)

	// 9. BUSINESS LOGIC: Validate payment covers card cost (including steel/titanium/substitutes)
	if err := cardPayment.CoversCardCost(card.Cost, allowSteel, allowTitanium, playerSubstitutes); err != nil {
		log.Error("Payment validation failed", zap.Error(err))
		return err
	}

	totalValue := cardPayment.TotalValue(playerSubstitutes)
	log.Debug("Payment validated",
		zap.Int("card_cost", card.Cost),
		zap.Int("payment_value", totalValue),
		zap.Int("credits", payment.Credits),
		zap.Int("steel", payment.Steel),
		zap.Int("titanium", payment.Titanium),
		zap.Any("substitutes", payment.Substitutes))

	// 10. BUSINESS LOGIC: Validate player has the resources they're trying to spend
	resources := player.Resources().Get()
	if err := cardPayment.CanAfford(resources); err != nil {
		log.Error("Player can't afford payment", zap.Error(err))
		return err
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
	deductions := map[shared.ResourceType]int{
		shared.ResourceCredits:  -payment.Credits,
		shared.ResourceSteel:    -payment.Steel,
		shared.ResourceTitanium: -payment.Titanium,
	}

	// Also deduct substitute resources (e.g., heat for Helion)
	for resourceType, amount := range payment.Substitutes {
		deductions[resourceType] = -amount
	}

	player.Resources().Add(deductions)

	log.Info("✅ Payment deducted",
		zap.Int("credits", payment.Credits),
		zap.Int("steel", payment.Steel),
		zap.Int("titanium", payment.Titanium),
		zap.Any("substitutes", payment.Substitutes))

	// 14. BUSINESS LOGIC: Apply card immediate effects and register behaviors
	if err := a.applyCardBehaviors(ctx, g, card, player, log); err != nil {
		log.Error("Failed to apply card behaviors", zap.Error(err))
		return fmt.Errorf("failed to apply card behaviors: %w", err)
	}

	// 14a. BUSINESS LOGIC: Recalculate requirement modifiers (card played may have discount effects, hand changed)
	calculator := gamecards.NewRequirementModifierCalculator(a.cardRegistry)
	modifiers := calculator.Calculate(player)
	player.Effects().SetRequirementModifiers(modifiers)
	log.Debug("📊 Recalculated requirement modifiers",
		zap.Int("modifier_count", len(modifiers)))

	log.Info("🎉 Card played successfully",
		zap.String("card_name", card.Name),
		zap.Int("card_cost", card.Cost),
		zap.Int("payment_value", totalValue))

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
func validateCardRequirements(card *gamecards.Card, g *game.Game, player *player.Player, cardRegistry cards.CardRegistry) error {
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
				playedCard, err := cardRegistry.GetByID(playedCardID)
				if err != nil {
					// Skip cards that can't be found (shouldn't happen in normal gameplay)
					continue
				}
				if hasTag(playedCard, *req.Tag) {
					tagCount++
				}
			}

			// Also count corporation tags if player has a corporation
			if corpID := player.CorporationID(); corpID != "" {
				corpCard, err := cardRegistry.GetByID(corpID)
				if err == nil && hasTag(corpCard, *req.Tag) {
					tagCount++
				}
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
			production := player.Resources().Production()
			var currentProduction int

			switch *req.Resource {
			case shared.ResourceCreditsProduction:
				currentProduction = production.Credits
			case shared.ResourceSteelProduction:
				currentProduction = production.Steel
			case shared.ResourceTitaniumProduction:
				currentProduction = production.Titanium
			case shared.ResourcePlantsProduction:
				currentProduction = production.Plants
			case shared.ResourceEnergyProduction:
				currentProduction = production.Energy
			case shared.ResourceHeatProduction:
				currentProduction = production.Heat
			default:
				return fmt.Errorf("invalid production resource type: %s", *req.Resource)
			}

			if req.Min != nil && currentProduction < *req.Min {
				return fmt.Errorf("production requirement not met: need %d %s production, have %d", *req.Min, *req.Resource, currentProduction)
			}
			if req.Max != nil && currentProduction > *req.Max {
				return fmt.Errorf("production requirement not met: max %d %s production, have %d", *req.Max, *req.Resource, currentProduction)
			}

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

		case gamecards.RequirementCities:
			// Count cities owned by the player on the board
			cityCount := 0
			for _, tile := range g.Board().Tiles() {
				if tile.OccupiedBy != nil && tile.OccupiedBy.Type == shared.ResourceCityTile {
					if tile.OwnerID != nil && *tile.OwnerID == player.ID() {
						cityCount++
					}
				}
			}

			if req.Min != nil && cityCount < *req.Min {
				return fmt.Errorf("city requirement not met: need %d cities, have %d", *req.Min, cityCount)
			}
			if req.Max != nil && cityCount > *req.Max {
				return fmt.Errorf("city requirement not met: max %d cities, have %d", *req.Max, cityCount)
			}

		case gamecards.RequirementGreeneries:
			// Count greeneries owned by the player on the board
			greeneryCount := 0
			for _, tile := range g.Board().Tiles() {
				if tile.OccupiedBy != nil && tile.OccupiedBy.Type == shared.ResourceGreeneryTile {
					if tile.OwnerID != nil && *tile.OwnerID == player.ID() {
						greeneryCount++
					}
				}
			}

			if req.Min != nil && greeneryCount < *req.Min {
				return fmt.Errorf("greenery requirement not met: need %d greeneries, have %d", *req.Min, greeneryCount)
			}
			if req.Max != nil && greeneryCount > *req.Max {
				return fmt.Errorf("greenery requirement not met: max %d greeneries, have %d", *req.Max, greeneryCount)
			}

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

			// Use BehaviorApplier for consistent output handling
			applier := gamecards.NewBehaviorApplier(p, g, card.Name, log)
			if err := applier.ApplyOutputs(ctx, behavior.Outputs); err != nil {
				return fmt.Errorf("failed to apply auto behavior %d outputs: %w", behaviorIndex, err)
			}

			// Also register as effect if it has persistent outputs (discount, payment-substitute)
			// These need to show in the effects list for display and for modifier calculations
			if gamecards.HasPersistentEffects(behavior) {
				log.Info("🏷️ Registering auto-trigger behavior with persistent effects",
					zap.String("card_name", card.Name))

				effect := player.CardEffect{
					CardID:        card.ID,
					CardName:      card.Name,
					BehaviorIndex: behaviorIndex,
					Behavior:      behavior,
				}
				p.Effects().AddEffect(effect)
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
