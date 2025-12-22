package game

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/events"
	internalgame "terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// ConfirmDemoSetupAction handles a player confirming their demo setup configuration
type ConfirmDemoSetupAction struct {
	gameRepo     internalgame.GameRepository
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewConfirmDemoSetupAction creates a new confirm demo setup action
func NewConfirmDemoSetupAction(
	gameRepo internalgame.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *ConfirmDemoSetupAction {
	return &ConfirmDemoSetupAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// Execute performs the confirm demo setup action
func (a *ConfirmDemoSetupAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
	request *dto.ConfirmDemoSetupRequest,
) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "confirm_demo_setup"),
	)
	log.Info("🎮 Player confirming demo setup")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. Validate game is in DemoSetup phase
	if g.CurrentPhase() != internalgame.GamePhaseDemoSetup {
		log.Warn("Game is not in demo setup phase", zap.String("phase", string(g.CurrentPhase())))
		return fmt.Errorf("game is not in demo setup phase: %s", g.CurrentPhase())
	}

	// 3. Get the player
	p, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. Set corporation if provided
	if request.CorporationID != nil && *request.CorporationID != "" {
		p.SetCorporationID(*request.CorporationID)
		log.Info("✅ Set corporation", zap.String("corporation_id", *request.CorporationID))
	}

	// 5. Add cards to hand with proper PlayerCard caching
	if len(request.CardIDs) > 0 {
		a.addCardsToPlayerHand(request.CardIDs, p, g, log)
		log.Info("✅ Added cards to hand", zap.Int("card_count", len(request.CardIDs)))
	}

	// 6. Set resources
	resources := shared.Resources{
		Credits:  request.Resources.Credits,
		Steel:    request.Resources.Steel,
		Titanium: request.Resources.Titanium,
		Plants:   request.Resources.Plants,
		Energy:   request.Resources.Energy,
		Heat:     request.Resources.Heat,
	}
	p.Resources().Set(resources)
	log.Info("✅ Set resources", zap.Any("resources", resources))

	// 7. Set production
	production := shared.Production{
		Credits:  request.Production.Credits,
		Steel:    request.Production.Steel,
		Titanium: request.Production.Titanium,
		Plants:   request.Production.Plants,
		Energy:   request.Production.Energy,
		Heat:     request.Production.Heat,
	}
	p.Resources().SetProduction(production)
	log.Info("✅ Set production", zap.Any("production", production))

	// 8. Set terraform rating
	p.Resources().SetTerraformRating(request.TerraformRating)
	log.Info("✅ Set terraform rating", zap.Int("rating", request.TerraformRating))

	// 9. If host, set global parameters and generation
	isHost := g.HostPlayerID() == playerID
	if isHost && request.GlobalParameters != nil {
		gp := g.GlobalParameters()
		if err := gp.SetTemperature(ctx, request.GlobalParameters.Temperature); err != nil {
			log.Error("Failed to set temperature", zap.Error(err))
			return fmt.Errorf("failed to set temperature: %w", err)
		}
		if err := gp.SetOxygen(ctx, request.GlobalParameters.Oxygen); err != nil {
			log.Error("Failed to set oxygen", zap.Error(err))
			return fmt.Errorf("failed to set oxygen: %w", err)
		}
		if err := gp.SetOceans(ctx, request.GlobalParameters.Oceans); err != nil {
			log.Error("Failed to set oceans", zap.Error(err))
			return fmt.Errorf("failed to set oceans: %w", err)
		}
		log.Info("✅ Set global parameters",
			zap.Int("temperature", request.GlobalParameters.Temperature),
			zap.Int("oxygen", request.GlobalParameters.Oxygen),
			zap.Int("oceans", request.GlobalParameters.Oceans))
	}

	if isHost && request.Generation != nil {
		if err := g.SetGeneration(ctx, *request.Generation); err != nil {
			log.Error("Failed to set generation", zap.Error(err))
			return fmt.Errorf("failed to set generation: %w", err)
		}
		log.Info("✅ Set generation", zap.Int("generation", *request.Generation))
	}

	// 10. Mark player as having confirmed demo setup
	p.SetDemoSetupConfirmed(true)
	log.Info("✅ Player confirmed demo setup")

	// 11. Check if all players have confirmed
	allConfirmed := true
	for _, pl := range g.GetAllPlayers() {
		if !pl.DemoSetupConfirmed() {
			allConfirmed = false
			break
		}
	}

	// 12. If all players confirmed, transition to Action phase
	if allConfirmed {
		if err := g.UpdatePhase(ctx, internalgame.GamePhaseAction); err != nil {
			log.Error("Failed to update game phase", zap.Error(err))
			return fmt.Errorf("failed to update game phase: %w", err)
		}
		log.Info("🎉 All players confirmed, transitioning to Action phase")
	}

	return nil
}

// addCardsToPlayerHand adds cards to a player's hand and creates PlayerCard instances with state.
func (a *ConfirmDemoSetupAction) addCardsToPlayerHand(
	cardIDs []string,
	p *player.Player,
	g *internalgame.Game,
	log *zap.Logger,
) {
	for _, cardID := range cardIDs {
		// Add card ID to hand (triggers CardHandUpdatedEvent)
		p.Hand().AddCard(cardID)

		// Get card from registry
		card, err := a.cardRegistry.GetByID(cardID)
		if err != nil {
			log.Warn("Failed to get card from registry, skipping PlayerCard creation",
				zap.String("card_id", cardID),
				zap.Error(err))
			continue
		}

		// Create and cache PlayerCard with state and event listeners
		a.createAndCachePlayerCard(card, p, g)
	}
}

// createAndCachePlayerCard creates a PlayerCard with event listeners and initial state.
func (a *ConfirmDemoSetupAction) createAndCachePlayerCard(
	card *gamecards.Card,
	p *player.Player,
	g *internalgame.Game,
) *player.PlayerCard {
	// Create PlayerCard data holder
	pc := player.NewPlayerCard(card)

	// Register event listeners for state recalculation
	a.registerPlayerCardEventListeners(pc, p, g)

	// Calculate initial state
	a.recalculatePlayerCard(pc, p, g)

	// Cache in Hand
	p.Hand().AddPlayerCard(card.ID, pc)

	return pc
}

// registerPlayerCardEventListeners registers event listeners on a PlayerCard.
func (a *ConfirmDemoSetupAction) registerPlayerCardEventListeners(
	pc *player.PlayerCard,
	p *player.Player,
	g *internalgame.Game,
) {
	eventBus := g.EventBus()

	// When player resources change, recalculate affordability
	subID1 := events.Subscribe(eventBus, func(event events.ResourcesChangedEvent) {
		if event.PlayerID == p.ID() {
			a.recalculatePlayerCard(pc, p, g)
		}
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID1) })

	// When temperature changes, recalculate requirements
	subID2 := events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		a.recalculatePlayerCard(pc, p, g)
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID2) })

	// When oxygen changes, recalculate requirements
	subID3 := events.Subscribe(eventBus, func(event events.OxygenChangedEvent) {
		a.recalculatePlayerCard(pc, p, g)
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID3) })

	// When oceans change, recalculate requirements
	subID4 := events.Subscribe(eventBus, func(event events.OceansChangedEvent) {
		a.recalculatePlayerCard(pc, p, g)
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID4) })

	// When player effects change (requirement modifiers), recalculate cost
	subID5 := events.Subscribe(eventBus, func(event events.PlayerEffectsChangedEvent) {
		if event.PlayerID == p.ID() {
			a.recalculatePlayerCard(pc, p, g)
		}
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID5) })

	// When game phase changes, recalculate state
	subID6 := events.Subscribe(eventBus, func(event events.GamePhaseChangedEvent) {
		if event.GameID == g.ID() {
			a.recalculatePlayerCard(pc, p, g)
		}
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID6) })

	// When general game state changes, recalculate availability
	subID7 := events.Subscribe(eventBus, func(event events.GameStateChangedEvent) {
		a.recalculatePlayerCard(pc, p, g)
	})
	pc.AddUnsubscriber(func() { eventBus.Unsubscribe(subID7) })
}

// recalculatePlayerCard recalculates and updates PlayerCard state.
// For demo setup, we use a simplified calculation that marks cards as playable
// based on cost vs available credits.
func (a *ConfirmDemoSetupAction) recalculatePlayerCard(
	pc *player.PlayerCard,
	p *player.Player,
	g *internalgame.Game,
) {
	card, ok := pc.Card().(*gamecards.Card)
	if !ok {
		return
	}

	var errors []player.StateError
	costMap := make(map[string]int)

	// Calculate effective cost with discounts
	calculator := gamecards.NewRequirementModifierCalculator(a.cardRegistry)
	discountAmount := calculator.CalculateCardDiscounts(p, card)
	effectiveCost := card.Cost - discountAmount
	if effectiveCost < 0 {
		effectiveCost = 0
	}
	if effectiveCost > 0 {
		costMap[string(shared.ResourceCredit)] = effectiveCost
	}

	// Check affordability
	credits := p.Resources().Get().Credits
	if credits < effectiveCost {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeInsufficientCredits,
			Category: player.ErrorCategoryCost,
			Message:  fmt.Sprintf("Need %d credits, have %d", effectiveCost, credits),
		})
	}

	state := player.EntityState{
		Errors:         errors,
		Cost:           costMap,
		Metadata:       make(map[string]interface{}),
		LastCalculated: time.Now(),
	}
	pc.UpdateState(state)
}
