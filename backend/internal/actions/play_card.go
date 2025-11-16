package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/turn"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// PlayCardAction handles playing a card from hand
// This action orchestrates:
// - Turn and action validation
// - Card ownership validation
// - Auto-triggered choice validation
// - Card validation (requirements, affordability) - INLINE per ARCHITECTURE_FLOW.md
// - Card payment (deduct resources)
// - Card effects (direct orchestration of feature services)
// - Effect storage in Player.Effects
// - Manual action registration
// - Action consumption
// - State broadcasting
type PlayCardAction struct {
	cardRepo          card.CardRepository
	gameRepo          game.Repository
	playerRepo        player.Repository
	cardDeckRepo      card.CardDeckRepository
	parametersFeature parameters.Service
	turnOrderService  turn.TurnOrderService
	sessionManager    session.SessionManager
}

// NewPlayCardAction creates a new play card action
func NewPlayCardAction(
	cardRepo card.CardRepository,
	gameRepo game.Repository,
	playerRepo player.Repository,
	cardDeckRepo card.CardDeckRepository,
	parametersFeature parameters.Service,
	turnOrderService turn.TurnOrderService,
	sessionManager session.SessionManager,
) *PlayCardAction {
	return &PlayCardAction{
		cardRepo:          cardRepo,
		gameRepo:          gameRepo,
		playerRepo:        playerRepo,
		cardDeckRepo:      cardDeckRepo,
		parametersFeature: parametersFeature,
		turnOrderService:  turnOrderService,
		sessionManager:    sessionManager,
	}
}

// Execute performs the play card action
// Steps:
// 1. Validate turn is player's turn
// 2. Validate player has available actions
// 3. Validate player owns card
// 4. Validate choice selection for auto-triggered choices
// 5. Validate card can be played (via CardManager)
// 6. Play card (via CardManager - handles payment, effects, etc.)
// 7. Process tile queue if card created tiles
// 8. Consume player action (if not infinite)
// 9. Broadcast state
func (a *PlayCardAction) Execute(ctx context.Context, gameID string, playerID string, cardID string, payment *player.CardPayment, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🃏 Executing play card action", zap.String("card_id", cardID))

	// Validate payment is provided
	if payment == nil {
		log.Warn("Payment is required but not provided")
		return fmt.Errorf("payment is required")
	}

	// Validate turn and actions
	currentTurn, err := a.turnOrderService.GetCurrentTurn(ctx)
	if err != nil {
		log.Error("Failed to get current turn", zap.Error(err))
		return fmt.Errorf("failed to get current turn: %w", err)
	}

	if currentTurn == nil {
		log.Warn("No current player turn set")
		return fmt.Errorf("no current player turn set")
	}

	if *currentTurn != playerID {
		log.Warn("Not current player's turn",
			zap.String("current_turn", *currentTurn),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("not current player's turn: current turn is %s", *currentTurn)
	}

	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// TODO: Restore action count validation when turn/action system is reintegrated
	// The new Player model uses Actions slice for card-based actions, not an action count
	// Action counting will be handled by the turn management system when it's restored

	// Check if player owns the card (Cards are now Card instances)
	hasCard := false
	for _, card := range player.Cards {
		if card.ID == cardID {
			hasCard = true
			break
		}
	}
	if !hasCard {
		log.Warn("Player does not own card", zap.String("card_id", cardID))
		return fmt.Errorf("player does not have card %s", cardID)
	}

	// Validate choice selection for cards with AUTO-triggered choices
	// Manual-triggered behaviors (actions) will have their choices resolved when the action is played
	cardObj, err := a.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		log.Error("Failed to get card", zap.Error(err))
		return fmt.Errorf("failed to get card: %w", err)
	}
	if cardObj == nil {
		log.Error("Card not found in repository", zap.String("card_id", cardID))
		return fmt.Errorf("card not found: %s", cardID)
	}

	// Check if any AUTO-triggered behavior has choices
	hasAutoChoices := false
	for _, behavior := range cardObj.Behaviors {
		// Only check behaviors with auto triggers
		hasAutoTrigger := false
		for _, trigger := range behavior.Triggers {
			if trigger.Type == card.ResourceTriggerAuto {
				hasAutoTrigger = true
				break
			}
		}

		// If this is an auto-triggered behavior with choices, validate choiceIndex
		if hasAutoTrigger && len(behavior.Choices) > 0 {
			hasAutoChoices = true
			// Validate that choiceIndex is provided and within valid range
			if choiceIndex == nil {
				log.Warn("Card has auto-triggered choices but no choiceIndex provided")
				return fmt.Errorf("card has auto-triggered choices but no choiceIndex provided")
			}
			if *choiceIndex < 0 || *choiceIndex >= len(behavior.Choices) {
				log.Warn("Invalid choiceIndex",
					zap.Int("choice_index", *choiceIndex),
					zap.Int("max_choices", len(behavior.Choices)))
				return fmt.Errorf("invalid choiceIndex %d: must be between 0 and %d", *choiceIndex, len(behavior.Choices)-1)
			}
			break
		}
	}

	if hasAutoChoices {
		log.Debug("🎯 Card has auto-triggered choices, using choiceIndex", zap.Int("choice_index", *choiceIndex))
	}

	// INLINE VALIDATION per ARCHITECTURE_FLOW.md (no ValidationService)
	// Validate card can be played: requirements + affordability
	if err := a.validateCardPlay(ctx, player, cardObj, payment); err != nil {
		log.Warn("Card cannot be played", zap.Error(err))
		return fmt.Errorf("card cannot be played: %w", err)
	}

	log.Debug("✅ Card validation passed")

	// STEP 1: Apply card cost payment (deduct credits, steel, titanium, substitutes)
	if cardObj.BaseCost > 0 {
		// Player.Resources is a direct field now
		playerResources := player.Resources

		playerResources.Credits -= payment.Credits
		playerResources.Steel -= payment.Steel
		playerResources.Titanium -= payment.Titanium

		// Deduct payment substitutes (e.g., heat for Helion)
		if payment.Substitutes != nil {
			for resourceType, amount := range payment.Substitutes {
				switch resourceType {
				case domain.ResourceHeat:
					playerResources.Heat -= amount
				case domain.ResourceEnergy:
					playerResources.Energy -= amount
				case domain.ResourcePlants:
					playerResources.Plants -= amount
				}
			}
		}

		if err := a.playerRepo.UpdateResources(ctx, gameID, playerID, playerResources); err != nil {
			return fmt.Errorf("failed to update player resources after payment: %w", err)
		}

		log.Debug("💰 Card cost paid", zap.Int("cost", cardObj.BaseCost))
	}

	// STEP 2: Remove card from hand and add to played cards
	if err := a.playerRepo.RemoveCard(ctx, gameID, playerID, cardID); err != nil {
		return fmt.Errorf("failed to remove card from hand: %w", err)
	}
	// Add card instance to played cards (Living Card Instance Pattern)
	if err := a.playerRepo.AddPlayedCard(ctx, gameID, playerID, *cardObj); err != nil {
		return fmt.Errorf("failed to add card to played cards: %w", err)
	}

	log.Debug("🃏 Card moved from hand to played cards")

	// STEP 3: Initialize resource storage if the card has storage capability
	if cardObj.ResourceStorage != nil {
		player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to get player for resource storage init: %w", err)
		}

		if player.ResourceStorage == nil {
			player.ResourceStorage = make(map[string]int)
		}

		player.ResourceStorage[cardID] = cardObj.ResourceStorage.Starting

		if err := a.playerRepo.UpdateResourceStorage(ctx, gameID, playerID, player.ResourceStorage); err != nil {
			return fmt.Errorf("failed to initialize card resource storage: %w", err)
		}

		log.Debug("💾 Card resource storage initialized",
			zap.String("resource_type", string(cardObj.ResourceStorage.Type)),
			zap.Int("starting_amount", cardObj.ResourceStorage.Starting))
	}

	// STEP 4: Process card immediate effects by parsing behaviors
	if err := a.processCardEffects(ctx, gameID, playerID, cardID, cardObj, choiceIndex, cardStorageTarget, log); err != nil {
		return fmt.Errorf("failed to process card effects: %w", err)
	}

	// STEP 5: Extract and store passive card effects in Player.Effects
	playerEffects := a.extractPlayerEffects(cardID, cardObj, log)
	if len(playerEffects) > 0 {
		// Get current player to append new effects
		currentPlayer, err := a.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to get player for effects update: %w", err)
		}

		// Append new effects to existing effects
		updatedEffects := append(currentPlayer.Effects, playerEffects...)
		if err := a.playerRepo.UpdatePlayerEffects(ctx, gameID, playerID, updatedEffects); err != nil {
			return fmt.Errorf("failed to update player effects: %w", err)
		}

		log.Info("✨ Player effects stored",
			zap.Int("new_effects_added", len(playerEffects)),
			zap.Int("total_effects", len(updatedEffects)))
	}

	// STEP 6: Extract and register card manual actions
	if err := a.registerManualActions(ctx, gameID, playerID, cardID, cardObj, log); err != nil {
		return fmt.Errorf("failed to register manual actions: %w", err)
	}

	log.Info("✅ Card played successfully", zap.String("card_id", cardID), zap.String("card_name", cardObj.Name))

	// TODO: Restore action consumption when turn/action system is reintegrated
	// Action consumption will be handled by the turn management system

	log.Info("✅ Play card action completed successfully", zap.String("card_id", cardID))
	return nil
}

// processCardEffects processes immediate card effects by parsing behaviors and calling feature services
func (a *PlayCardAction) processCardEffects(ctx context.Context, gameID, playerID, cardID string, c *card.Card, choiceIndex *int, cardStorageTarget *string, log *zap.Logger) error {
	log.Debug("✨ Processing immediate card effects", zap.String("card_name", c.Name))

	// Get current player for resource/production updates
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for effect processing: %w", err)
	}

	// Player.Resources and Player.Production are direct fields now
	playerResources := player.Resources
	playerProduction := player.Production

	var resourcesChanged, productionChanged bool
	var tilePlacementQueue []string
	var trChange int

	// Process all behaviors to find immediate effects
	for _, behavior := range c.Behaviors {
		// Only process auto triggers WITHOUT conditions (immediate effects when card is played)
		if len(behavior.Triggers) == 0 || behavior.Triggers[0].Type != card.ResourceTriggerAuto || behavior.Triggers[0].Condition != nil {
			continue
		}

		// Aggregate all outputs: behavior.Outputs + choice[choiceIndex].Outputs
		allOutputs := behavior.Outputs

		// If choiceIndex is provided and this behavior has choices, add choice outputs
		if choiceIndex != nil && len(behavior.Choices) > 0 && *choiceIndex < len(behavior.Choices) {
			selectedChoice := behavior.Choices[*choiceIndex]
			allOutputs = append(allOutputs, selectedChoice.Outputs...)
			log.Debug("🎯 Applying choice outputs",
				zap.Int("choice_index", *choiceIndex),
				zap.Int("choice_outputs_count", len(selectedChoice.Outputs)))
		}

		// Process all aggregated outputs
		for _, output := range allOutputs {
			switch output.Type {
			// Resource outputs
			case domain.ResourceCredits:
				playerResources.Credits += output.Amount
				resourcesChanged = true
			case domain.ResourceSteel:
				playerResources.Steel += output.Amount
				resourcesChanged = true
			case domain.ResourceTitanium:
				playerResources.Titanium += output.Amount
				resourcesChanged = true
			case domain.ResourcePlants:
				playerResources.Plants += output.Amount
				resourcesChanged = true
			case domain.ResourceEnergy:
				playerResources.Energy += output.Amount
				resourcesChanged = true
			case domain.ResourceHeat:
				playerResources.Heat += output.Amount
				resourcesChanged = true

			// Production outputs
			case domain.ResourceCreditsProduction:
				playerProduction.Credits += output.Amount
				productionChanged = true
			case domain.ResourceSteelProduction:
				playerProduction.Steel += output.Amount
				productionChanged = true
			case domain.ResourceTitaniumProduction:
				playerProduction.Titanium += output.Amount
				productionChanged = true
			case domain.ResourcePlantsProduction:
				playerProduction.Plants += output.Amount
				productionChanged = true
			case domain.ResourceEnergyProduction:
				playerProduction.Energy += output.Amount
				productionChanged = true
			case domain.ResourceHeatProduction:
				playerProduction.Heat += output.Amount
				productionChanged = true

			// TR output
			case domain.ResourceTR:
				trChange += output.Amount

			// Global parameters - orchestrate via parameters feature
			case domain.ResourceTemperature:
				steps := output.Amount / 2 // Temperature increases in 2°C steps
				if steps > 0 {
					actualSteps, err := a.parametersFeature.RaiseTemperature(ctx, steps)
					if err != nil {
						return fmt.Errorf("failed to raise temperature: %w", err)
					}
					// Award TR for actual steps raised
					if actualSteps > 0 {
						trChange += actualSteps
						log.Info("🌡️ Temperature raised (TR awarded)",
							zap.Int("steps", actualSteps),
							zap.Int("tr_change", actualSteps))
					}
				}

			case domain.ResourceOxygen:
				if output.Amount > 0 {
					actualSteps, err := a.parametersFeature.RaiseOxygen(ctx, output.Amount)
					if err != nil {
						return fmt.Errorf("failed to raise oxygen: %w", err)
					}
					// Award TR for actual steps raised
					if actualSteps > 0 {
						trChange += actualSteps
						log.Info("💨 Oxygen raised (TR awarded)",
							zap.Int("steps", actualSteps),
							zap.Int("tr_change", actualSteps))
					}
				}

			case domain.ResourceOceans:
				for i := 0; i < output.Amount; i++ {
					if err := a.parametersFeature.PlaceOcean(ctx); err != nil {
						log.Warn("Failed to place ocean (may be at max)", zap.Error(err))
						break
					}
					trChange++ // Award TR for ocean placement
				}
				if output.Amount > 0 {
					log.Info("🌊 Oceans placed (TR awarded)",
						zap.Int("count", output.Amount),
						zap.Int("tr_change", output.Amount))
				}

			// Tile placements
			case domain.ResourceCityPlacement:
				for i := 0; i < output.Amount; i++ {
					tilePlacementQueue = append(tilePlacementQueue, "city")
				}
			case domain.ResourceOceanPlacement:
				for i := 0; i < output.Amount; i++ {
					tilePlacementQueue = append(tilePlacementQueue, "ocean")
				}
			case domain.ResourceGreeneryPlacement:
				for i := 0; i < output.Amount; i++ {
					tilePlacementQueue = append(tilePlacementQueue, "greenery")
				}

			// Card storage resources
			case domain.ResourceAnimals, domain.ResourceMicrobes, domain.ResourceFloaters, domain.ResourceScience, domain.ResourceAsteroid:
				if err := a.applyCardStorageResource(ctx, gameID, playerID, cardID, output, cardStorageTarget, log); err != nil {
					return fmt.Errorf("failed to apply card storage resource: %w", err)
				}

			// Card draw/peek effects
			case domain.ResourceCardDraw, domain.ResourceCardPeek, domain.ResourceCardTake, domain.ResourceCardBuy:
				if err := a.processCardDrawEffects(ctx, gameID, playerID, c, log); err != nil {
					return fmt.Errorf("failed to process card draw effects: %w", err)
				}
			}
		}
	}

	// Update resources if changed
	if resourcesChanged {
		if err := a.playerRepo.UpdateResources(ctx, gameID, playerID, playerResources); err != nil {
			return fmt.Errorf("failed to update player resources: %w", err)
		}
		log.Debug("💰 Resources updated")
	}

	// Update production if changed
	if productionChanged {
		if err := a.playerRepo.UpdateProduction(ctx, gameID, playerID, playerProduction); err != nil {
			return fmt.Errorf("failed to update player production: %w", err)
		}
		log.Debug("📈 Production updated")
	}

	// Update TR if changed
	if trChange != 0 {
		newTR := player.TerraformRating + trChange
		if err := a.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR); err != nil {
			return fmt.Errorf("failed to update terraform rating: %w", err)
		}
		log.Info("⭐ Terraform Rating changed",
			zap.Int("change", trChange),
			zap.Int("new_tr", newTR))
	}

	// Create tile placement queue if needed
	if len(tilePlacementQueue) > 0 {
		if err := a.playerRepo.CreateTileQueue(ctx, gameID, playerID, cardID, tilePlacementQueue); err != nil {
			return fmt.Errorf("failed to create tile placement queue: %w", err)
		}
		log.Debug("🏗️ Tile placement queue created", zap.Int("tiles", len(tilePlacementQueue)))
	}

	// Process victory point conditions
	if err := a.processVictoryPoints(ctx, gameID, playerID, c, log); err != nil {
		return fmt.Errorf("failed to process victory points: %w", err)
	}

	log.Debug("✅ Immediate effects processed")
	return nil
}

// registerManualActions extracts and registers manual actions from a card
func (a *PlayCardAction) registerManualActions(ctx context.Context, gameID, playerID, cardID string, c *card.Card, log *zap.Logger) error {
	var manualActions []card.PlayerAction

	for behaviorIndex, behavior := range c.Behaviors {
		// Check if this behavior has manual triggers
		hasManualTrigger := false
		for _, trigger := range behavior.Triggers {
			if trigger.Type == card.ResourceTriggerManual {
				hasManualTrigger = true
				break
			}
		}

		// If behavior has manual triggers, create a PlayerAction
		if hasManualTrigger {
			action := card.PlayerAction{
				CardID:        cardID,
				CardName:      c.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
				PlayCount:     0,
			}
			manualActions = append(manualActions, action)
			log.Debug("🎯 Extracted manual action from card",
				zap.String("card_name", c.Name),
				zap.Int("behavior_index", behaviorIndex))
		}
	}

	// Add manual actions to player if any were found
	if len(manualActions) > 0 {
		// Get current player state to append to existing actions
		currentPlayer, err := a.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to get player for actions update: %w", err)
		}

		// Create new actions list with existing + new card actions
		newActions := make([]card.PlayerAction, len(currentPlayer.Actions)+len(manualActions))
		copy(newActions, currentPlayer.Actions)
		copy(newActions[len(currentPlayer.Actions):], manualActions)

		// Update player with new actions
		if err := a.playerRepo.UpdatePlayerActions(ctx, gameID, playerID, newActions); err != nil {
			return fmt.Errorf("failed to update player manual actions: %w", err)
		}

		log.Info("🎯 Card manual actions registered",
			zap.String("card_name", c.Name),
			zap.Int("actions_count", len(manualActions)))
	}

	return nil
}

// applyCardStorageResource handles adding resources to card storage
func (a *PlayCardAction) applyCardStorageResource(ctx context.Context, gameID, playerID, playedCardID string, output card.ResourceCondition, cardStorageTarget *string, log *zap.Logger) error {
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for card storage update: %w", err)
	}

	if player.ResourceStorage == nil {
		player.ResourceStorage = make(map[string]int)
	}

	// Determine target card based on output.Target
	var targetCardID string
	switch output.Target {
	case card.TargetSelfCard:
		targetCardID = playedCardID
	case card.TargetAnyCard:
		if cardStorageTarget == nil || *cardStorageTarget == "" {
			log.Info("⚠️ No card storage target provided - resources will be lost")
			return nil
		}
		targetCardID = *cardStorageTarget
		// Validate target card exists (PlayedCards are now Card instances)
		found := false
		for _, card := range player.PlayedCards {
			if card.ID == targetCardID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("target card %s not found in player's played cards", targetCardID)
		}
	default:
		return fmt.Errorf("invalid target type for card storage: %s", output.Target)
	}

	// Update resource storage
	player.ResourceStorage[targetCardID] += output.Amount

	if err := a.playerRepo.UpdateResourceStorage(ctx, gameID, playerID, player.ResourceStorage); err != nil {
		return fmt.Errorf("failed to update card resource storage: %w", err)
	}

	log.Debug("💾 Card storage resource applied",
		zap.String("target_card_id", targetCardID),
		zap.Int("amount", output.Amount))

	return nil
}

// processCardDrawEffects handles card draw/peek/take/buy effects
func (a *PlayCardAction) processCardDrawEffects(ctx context.Context, gameID, playerID string, c *card.Card, log *zap.Logger) error {
	var cardDrawAmount, cardPeekAmount, cardTakeAmount, cardBuyAmount int
	cardBuyCost := 3 // Default cost for buying cards

	// Scan for card-draw, card-peek, card-take, card-buy outputs
	for _, behavior := range c.Behaviors {
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == card.ResourceTriggerAuto && behavior.Triggers[0].Condition == nil {
			for _, output := range behavior.Outputs {
				switch output.Type {
				case domain.ResourceCardDraw:
					cardDrawAmount += output.Amount
				case domain.ResourceCardPeek:
					cardPeekAmount += output.Amount
				case domain.ResourceCardTake:
					cardTakeAmount += output.Amount
				case domain.ResourceCardBuy:
					cardBuyAmount += output.Amount
				}
			}
		}
	}

	if cardDrawAmount == 0 && cardPeekAmount == 0 && cardTakeAmount == 0 && cardBuyAmount == 0 {
		return nil
	}

	// Determine scenario and create appropriate PendingCardDrawSelection
	var cardsToShow []string
	var freeTakeCount, maxBuyCount int

	if cardDrawAmount > 0 && cardPeekAmount == 0 && cardTakeAmount == 0 && cardBuyAmount == 0 {
		// Simple card-draw: pop cards and auto-select all
		for i := 0; i < cardDrawAmount; i++ {
			cardID, err := a.cardDeckRepo.Pop(ctx, gameID)
			if err != nil {
				return fmt.Errorf("failed to draw card: %w", err)
			}
			cardsToShow = append(cardsToShow, cardID)
		}
		freeTakeCount = cardDrawAmount
		maxBuyCount = 0
	} else if cardPeekAmount > 0 {
		// Peek-based scenario
		for i := 0; i < cardPeekAmount; i++ {
			cardID, err := a.cardDeckRepo.Pop(ctx, gameID)
			if err != nil {
				return fmt.Errorf("failed to peek card: %w", err)
			}
			cardsToShow = append(cardsToShow, cardID)
		}
		freeTakeCount = cardDrawAmount + cardTakeAmount
		maxBuyCount = cardBuyAmount
	} else {
		return fmt.Errorf("invalid card effect combination: must have either card-draw or card-peek")
	}

	// Create PendingCardDrawSelection
	selection := &player.PendingCardDrawSelection{
		AvailableCards: cardsToShow,
		FreeTakeCount:  freeTakeCount,
		MaxBuyCount:    maxBuyCount,
		CardBuyCost:    cardBuyCost,
		Source:         c.ID,
	}

	if err := a.playerRepo.UpdatePendingCardDrawSelection(ctx, gameID, playerID, selection); err != nil {
		return fmt.Errorf("failed to create pending card draw selection: %w", err)
	}

	log.Info("🃏 Pending card draw selection created",
		zap.Int("available_cards", len(cardsToShow)),
		zap.Int("free_take_count", freeTakeCount),
		zap.Int("max_buy_count", maxBuyCount))

	return nil
}

// processVictoryPoints applies victory point conditions from a card
func (a *PlayCardAction) processVictoryPoints(ctx context.Context, gameID, playerID string, c *card.Card, log *zap.Logger) error {
	if len(c.VPConditions) == 0 {
		return nil
	}

	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for VP update: %w", err)
	}

	var totalVPAwarded int

	for _, vpCondition := range c.VPConditions {
		switch vpCondition.Condition {
		case card.VPConditionFixed:
			totalVPAwarded += vpCondition.Amount
			log.Debug("🏆 Fixed VP condition found", zap.Int("vp_amount", vpCondition.Amount))
		case card.VPConditionOnce, card.VPConditionPer:
			log.Debug("⚠️ VP condition not yet implemented",
				zap.String("condition", string(vpCondition.Condition)))
		}
	}

	if totalVPAwarded > 0 {
		newVictoryPoints := player.VictoryPoints + totalVPAwarded
		if err := a.playerRepo.UpdateVictoryPoints(ctx, gameID, playerID, newVictoryPoints); err != nil {
			return fmt.Errorf("failed to update player victory points: %w", err)
		}
		log.Info("🏆 Victory Points awarded",
			zap.String("card_name", c.Name),
			zap.Int("vp_awarded", totalVPAwarded),
			zap.Int("total_vp", newVictoryPoints))
	}

	return nil
}

// validateCardPlay validates card requirements and affordability inline (per ARCHITECTURE_FLOW.md)
// NO ValidationService - validation happens directly in Actions layer
func (a *PlayCardAction) validateCardPlay(ctx context.Context, player player.Player, cardObj *card.Card, payment *player.CardPayment) error {
	// 1. Validate requirements against global parameters (with modifier leniency)
	globalParams, err := a.parametersFeature.GetGlobalParameters(ctx)
	if err != nil {
		return fmt.Errorf("failed to get global parameters: %w", err)
	}

	// Check all requirements
	if len(cardObj.Requirements) > 0 {
		for _, req := range cardObj.Requirements {
			switch req.Type {
			case card.RequirementTemperature:
				if req.Min != nil {
					requiredTemp := *req.Min
					// TODO: Apply RequirementModifiers for leniency when implemented
					if globalParams.Temperature < requiredTemp {
						return fmt.Errorf("temperature requirement not met: need %d°C, current %d°C", requiredTemp, globalParams.Temperature)
					}
				}
				if req.Max != nil && globalParams.Temperature > *req.Max {
					return fmt.Errorf("temperature too high: need max %d°C, current %d°C", *req.Max, globalParams.Temperature)
				}

			case card.RequirementOxygen:
				if req.Min != nil && globalParams.Oxygen < *req.Min {
					return fmt.Errorf("oxygen requirement not met: need %d%%, current %d%%", *req.Min, globalParams.Oxygen)
				}
				if req.Max != nil && globalParams.Oxygen > *req.Max {
					return fmt.Errorf("oxygen too high: need max %d%%, current %d%%", *req.Max, globalParams.Oxygen)
				}

			case card.RequirementOceans:
				if req.Min != nil && globalParams.Oceans < *req.Min {
					return fmt.Errorf("ocean requirement not met: need %d oceans, current %d", *req.Min, globalParams.Oceans)
				}
				if req.Max != nil && globalParams.Oceans > *req.Max {
					return fmt.Errorf("too many oceans: need max %d, current %d", *req.Max, globalParams.Oceans)
				}

			case card.RequirementTR:
				if req.Min != nil && player.TerraformRating < *req.Min {
					return fmt.Errorf("terraform rating too low: need %d TR, current %d", *req.Min, player.TerraformRating)
				}
				if req.Max != nil && player.TerraformRating > *req.Max {
					return fmt.Errorf("terraform rating too high: need max %d TR, current %d", *req.Max, player.TerraformRating)
				}

			case card.RequirementTags:
				tagCounts := player.CountTags()
				if req.Tag != nil {
					count := tagCounts[*req.Tag]
					if req.Min != nil && count < *req.Min {
						return fmt.Errorf("tag requirement not met: need %d %s tags, have %d", *req.Min, *req.Tag, count)
					}
					if req.Max != nil && count > *req.Max {
						return fmt.Errorf("too many %s tags: need max %d, have %d", *req.Tag, *req.Max, count)
					}
				}

			case card.RequirementProduction:
				// Player.Production is a direct field now
				production := player.Production
				if req.Resource != nil {
					var currentProd int
					switch *req.Resource {
					case domain.ResourceCredits:
						currentProd = production.Credits
					case domain.ResourceSteel:
						currentProd = production.Steel
					case domain.ResourceTitanium:
						currentProd = production.Titanium
					case domain.ResourcePlants:
						currentProd = production.Plants
					case domain.ResourceEnergy:
						currentProd = production.Energy
					case domain.ResourceHeat:
						currentProd = production.Heat
					}
					if req.Min != nil && currentProd < *req.Min {
						return fmt.Errorf("%s production too low: need %d, have %d", *req.Resource, *req.Min, currentProd)
					}
					if req.Max != nil && currentProd > *req.Max {
						return fmt.Errorf("%s production too high: need max %d, have %d", *req.Resource, *req.Max, currentProd)
					}
				}

			case card.RequirementResource:
				playerResources := player.Resources
				if req.Resource != nil {
					var currentRes int
					switch *req.Resource {
					case domain.ResourceCredits:
						currentRes = playerResources.Credits
					case domain.ResourceSteel:
						currentRes = playerResources.Steel
					case domain.ResourceTitanium:
						currentRes = playerResources.Titanium
					case domain.ResourcePlants:
						currentRes = playerResources.Plants
					case domain.ResourceEnergy:
						currentRes = playerResources.Energy
					case domain.ResourceHeat:
						currentRes = playerResources.Heat
					}
					if req.Min != nil && currentRes < *req.Min {
						return fmt.Errorf("%s too low: need %d, have %d", *req.Resource, *req.Min, currentRes)
					}
					if req.Max != nil && currentRes > *req.Max {
						return fmt.Errorf("%s too high: need max %d, have %d", *req.Resource, *req.Max, currentRes)
					}
				}
			}
		}
	}

	// 2. Validate affordability - payment value must cover final cost
	finalCost := cardObj.BaseCost // TODO: Use cardObj.GetFinalCost() when Living Card Instance Pattern is fully implemented

	// Calculate payment value
	paymentValue := payment.Credits

	// Steel applies only to cards with "building" tag (default 2 MC per steel)
	for _, tag := range cardObj.Tags {
		if tag == card.TagBuilding {
			paymentValue += payment.Steel * 2 // TODO: Use player.SteelValue when implemented
			break
		}
	}

	// Titanium applies only to cards with "space" tag (default 3 MC per titanium)
	for _, tag := range cardObj.Tags {
		if tag == card.TagSpace {
			paymentValue += payment.Titanium * 3 // TODO: Use player.TitaniumValue when implemented
			break
		}
	}

	// Add payment substitutes (e.g., heat for Helion corporation)
	if payment.Substitutes != nil {
		for _, amount := range payment.Substitutes {
			paymentValue += amount // 1:1 conversion for substitutes
		}
	}

	if paymentValue < finalCost {
		return fmt.Errorf("insufficient payment: need %d MC, provided %d MC", finalCost, paymentValue)
	}

	return nil
}

// extractPlayerEffects extracts passive effects from a card's behaviors
// Returns both reactive effects (auto triggers with conditions) and static effects (discounts, modifiers)
func (a *PlayCardAction) extractPlayerEffects(cardID string, c *card.Card, log *zap.Logger) []card.PlayerEffect {
	var playerEffects []card.PlayerEffect

	// Check if card has any behaviors
	if len(c.Behaviors) == 0 {
		return playerEffects
	}

	// Extract each behavior that represents a passive effect
	for i, behavior := range c.Behaviors {
		if len(behavior.Triggers) == 0 {
			continue
		}

		trigger := behavior.Triggers[0] // Get first trigger

		// Only process auto triggers (passive effects)
		if trigger.Type != card.ResourceTriggerAuto {
			log.Debug("Behavior trigger is not auto, skipping",
				zap.String("card_name", c.Name),
				zap.String("trigger_type", string(trigger.Type)))
			continue
		}

		// Store all auto-triggered behaviors as player effects
		// This includes both:
		// - Reactive effects (auto with condition): triggered by events
		// - Static effects (auto without condition): applied on-demand (e.g., discounts)
		playerEffects = append(playerEffects, card.PlayerEffect{
			CardID:        cardID,
			CardName:      c.Name,
			BehaviorIndex: i,
			Behavior:      behavior,
		})

		if trigger.Condition != nil {
			log.Debug("✅ Reactive effect extracted",
				zap.String("card_name", c.Name),
				zap.String("trigger_type", string(trigger.Condition.Type)))
		} else {
			log.Debug("✅ Static effect extracted (discount/modifier)",
				zap.String("card_name", c.Name),
				zap.Int("behavior_index", i))
		}
	}

	return playerEffects
}
