package card

import (
	"context"
	"fmt"
	"time"

	baseaction "terraforming-mars-backend/internal/action"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// PlayCardAction handles the business logic for playing a project card from hand
// Card playing involves: validating requirements, calculating costs (with discounts),
// moving card to played cards, applying immediate effects, and deducting payment
type PlayCardAction struct {
	baseaction.BaseAction
}

// NewPlayCardAction creates a new play card action
func NewPlayCardAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *PlayCardAction {
	return &PlayCardAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, cardRegistry, stateRepo),
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
	choiceIndex *int,
	cardStorageTarget *string,
) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("card_id", cardID),
		zap.String("action", "play_card"),
	)
	if choiceIndex != nil {
		log = log.With(zap.Int("choice_index", *choiceIndex))
	}
	if cardStorageTarget != nil {
		log = log.With(zap.String("card_storage_target", *cardStorageTarget))
	}
	log.Info("🃏 Player attempting to play card")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	if err := baseaction.ValidateGamePhase(g, game.GamePhaseAction, log); err != nil {
		return err
	}

	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	if err := baseaction.ValidateActionsRemaining(g, playerID, log); err != nil {
		return err
	}

	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	if !player.Hand().HasCard(cardID) {
		log.Error("Card not in player's hand")
		return fmt.Errorf("card %s not in hand", cardID)
	}

	card, err := a.CardRegistry().GetByID(cardID)
	if err != nil {
		log.Error("Card not found in registry", zap.Error(err))
		return fmt.Errorf("card not found: %w", err)
	}

	log.Debug("Card data retrieved",
		zap.String("card_name", card.Name),
		zap.Int("base_cost", card.Cost))

	if err := validateCardRequirements(card, g, player); err != nil {
		log.Error("Card requirements not met", zap.Error(err))
		return fmt.Errorf("cannot play card: %w", err)
	}

	log.Debug("✅ Card requirements validated")

	calculator := gamecards.NewRequirementModifierCalculator(a.CardRegistry())
	discountAmount := calculator.CalculateCardDiscounts(player, card)
	effectiveCost := card.Cost - discountAmount
	if effectiveCost < 0 {
		effectiveCost = 0
	}

	if discountAmount > 0 {
		log.Debug("Discount applied",
			zap.Int("base_cost", card.Cost),
			zap.Int("discount", discountAmount),
			zap.Int("effective_cost", effectiveCost))
	}

	playerSubstitutes := player.Resources().PaymentSubstitutes()

	allowSteel := hasTag(card, shared.TagBuilding)
	allowTitanium := hasTag(card, shared.TagSpace)

	adjustedPayment := adjustPaymentToEffectiveCost(payment, effectiveCost, allowSteel, allowTitanium, playerSubstitutes)

	cardPayment := gamecards.CardPayment{
		Credits:     adjustedPayment.Credits,
		Steel:       adjustedPayment.Steel,
		Titanium:    adjustedPayment.Titanium,
		Substitutes: adjustedPayment.Substitutes,
	}

	if err := cardPayment.CoversCardCost(effectiveCost, allowSteel, allowTitanium, playerSubstitutes); err != nil {
		log.Error("Payment validation failed", zap.Error(err))
		return err
	}

	totalValue := cardPayment.TotalValue(playerSubstitutes)
	log.Debug("Payment validated",
		zap.Int("effective_cost", effectiveCost),
		zap.Int("payment_value", totalValue),
		zap.Int("credits", adjustedPayment.Credits),
		zap.Int("steel", adjustedPayment.Steel),
		zap.Int("titanium", adjustedPayment.Titanium),
		zap.Any("substitutes", adjustedPayment.Substitutes))

	resources := player.Resources().Get()
	if err := cardPayment.CanAfford(resources); err != nil {
		log.Error("Player can't afford payment", zap.Error(err))
		return err
	}

	if !player.Hand().RemoveCard(cardID) {
		log.Error("Failed to remove card from hand - card not found")
		return fmt.Errorf("failed to remove card from hand: card not found")
	}

	log.Info("✅ Card removed from hand")

	cardTags := make([]string, len(card.Tags))
	for i, tag := range card.Tags {
		cardTags[i] = string(tag)
	}

	player.PlayedCards().AddCard(cardID, card.Name, string(card.Type), cardTags)

	log.Info("✅ Card added to played cards")

	if card.ResourceStorage != nil {
		player.Resources().AddToStorage(cardID, card.ResourceStorage.Starting)
		log.Info("📦 Initialized resource storage",
			zap.String("card_id", cardID),
			zap.String("resource_type", string(card.ResourceStorage.Type)),
			zap.Int("starting_amount", card.ResourceStorage.Starting))
	}

	deductions := map[shared.ResourceType]int{
		shared.ResourceCredit:   -adjustedPayment.Credits,
		shared.ResourceSteel:    -adjustedPayment.Steel,
		shared.ResourceTitanium: -adjustedPayment.Titanium,
	}

	for resourceType, amount := range adjustedPayment.Substitutes {
		deductions[resourceType] = -amount
	}

	player.Resources().Add(deductions)

	log.Info("✅ Payment deducted",
		zap.Int("credits", adjustedPayment.Credits),
		zap.Int("steel", adjustedPayment.Steel),
		zap.Int("titanium", adjustedPayment.Titanium),
		zap.Any("substitutes", adjustedPayment.Substitutes))

	calculatedOutputs, err := a.applyCardBehaviors(ctx, g, card, player, choiceIndex, cardStorageTarget, log)
	if err != nil {
		log.Error("Failed to apply card behaviors", zap.Error(err))
		return fmt.Errorf("failed to apply card behaviors: %w", err)
	}

	a.ConsumePlayerAction(g, log)

	description := fmt.Sprintf("Played %s for %d credits", card.Name, totalValue)
	displayData := baseaction.BuildCardDisplayData(card, game.SourceTypeCardPlay)
	a.WriteStateLogFull(ctx, g, card.Name, game.SourceTypeCardPlay, playerID, description, choiceIndex, calculatedOutputs, displayData)

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
			case shared.ResourceCredit:
				currentAmount = resources.Credits
			case shared.ResourceSteel:
				currentAmount = resources.Steel
			case shared.ResourceTitanium:
				currentAmount = resources.Titanium
			case shared.ResourcePlant:
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
// Returns calculated outputs for logging purposes
func (a *PlayCardAction) applyCardBehaviors(
	ctx context.Context,
	g *game.Game,
	card *gamecards.Card,
	p *player.Player,
	choiceIndex *int,
	cardStorageTarget *string,
	log *zap.Logger,
) ([]game.CalculatedOutput, error) {
	if len(card.Behaviors) == 0 {
		log.Debug("No card behaviors to apply")
		return nil, nil
	}

	log.Info("🎴 Processing card behaviors",
		zap.String("card_id", card.ID),
		zap.Int("behavior_count", len(card.Behaviors)))

	var allCalculatedOutputs []game.CalculatedOutput

	for behaviorIndex, behavior := range card.Behaviors {
		log.Debug("Processing behavior",
			zap.Int("index", behaviorIndex),
			zap.Int("trigger_count", len(behavior.Triggers)))

		// Apply auto-trigger behaviors immediately
		if gamecards.HasAutoTrigger(behavior) {
			// Extract inputs and outputs, incorporating choice if present
			_, outputs := behavior.ExtractInputsOutputs(choiceIndex)

			log.Info("✨ Found auto-trigger behavior, applying outputs immediately",
				zap.Int("output_count", len(outputs)))

			// Use BehaviorApplier for consistent output handling
			applier := gamecards.NewBehaviorApplier(p, g, card.Name, log).
				WithSourceCardID(card.ID).
				WithCardRegistry(a.CardRegistry())
			if cardStorageTarget != nil {
				applier = applier.WithTargetCardID(*cardStorageTarget)
			}
			calculatedOutputs, err := applier.ApplyOutputsAndGetCalculated(ctx, outputs)
			if err != nil {
				return nil, fmt.Errorf("failed to apply auto behavior %d outputs: %w", behaviorIndex, err)
			}
			allCalculatedOutputs = append(allCalculatedOutputs, calculatedOutputs...)

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

				events.Publish(g.EventBus(), events.PlayerEffectsChangedEvent{
					GameID:    g.ID(),
					PlayerID:  p.ID(),
					Timestamp: time.Now(),
				})
			}
		}

		// Register manual-trigger behaviors as player actions
		if gamecards.HasManualTrigger(behavior) {
			log.Info("🎯 Found manual-trigger behavior, registering as player action")

			p.Actions().AddAction(player.CardAction{
				CardID:                  card.ID,
				CardName:                card.Name,
				BehaviorIndex:           behaviorIndex,
				Behavior:                behavior,
				TimesUsedThisTurn:       0,
				TimesUsedThisGeneration: 0,
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

			events.Publish(g.EventBus(), events.PlayerEffectsChangedEvent{
				GameID:    g.ID(),
				PlayerID:  p.ID(),
				Timestamp: time.Now(),
			})

			// Subscribe passive effects to relevant events
			baseaction.SubscribePassiveEffectToEvents(ctx, g, p, effect, log)
		}
	}

	log.Info("✅ All card behaviors processed successfully")
	return allCalculatedOutputs, nil
}

func adjustPaymentToEffectiveCost(
	payment PaymentRequest,
	effectiveCost int,
	allowSteel bool,
	allowTitanium bool,
	playerSubstitutes []shared.PaymentSubstitute,
) PaymentRequest {
	if effectiveCost <= 0 {
		return PaymentRequest{}
	}

	steelRate := 2
	titaniumRate := 3
	for _, sub := range playerSubstitutes {
		if sub.ResourceType == shared.ResourceSteel {
			steelRate = sub.ConversionRate
		}
		if sub.ResourceType == shared.ResourceTitanium {
			titaniumRate = sub.ConversionRate
		}
	}

	nonCreditValue := 0
	if allowSteel {
		nonCreditValue += payment.Steel * steelRate
	}
	if allowTitanium {
		nonCreditValue += payment.Titanium * titaniumRate
	}

	for resourceType, amount := range payment.Substitutes {
		for _, sub := range playerSubstitutes {
			if sub.ResourceType == resourceType {
				nonCreditValue += amount * sub.ConversionRate
				break
			}
		}
	}

	if nonCreditValue >= effectiveCost {
		return PaymentRequest{
			Credits:     0,
			Steel:       payment.Steel,
			Titanium:    payment.Titanium,
			Substitutes: payment.Substitutes,
		}
	}

	creditsNeeded := effectiveCost - nonCreditValue
	if creditsNeeded > payment.Credits {
		creditsNeeded = payment.Credits
	}

	return PaymentRequest{
		Credits:     creditsNeeded,
		Steel:       payment.Steel,
		Titanium:    payment.Titanium,
		Substitutes: payment.Substitutes,
	}
}
