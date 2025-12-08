package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
)

// SelectStartingCardsAction handles the business logic for selecting starting cards and corporation
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type SelectStartingCardsAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	corpProc     *gamecards.CorporationProcessor
	logger       *zap.Logger
}

// NewSelectStartingCardsAction creates a new select starting cards action
func NewSelectStartingCardsAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *SelectStartingCardsAction {
	return &SelectStartingCardsAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		corpProc:     gamecards.NewCorporationProcessor(logger),
		logger:       logger,
	}
}

// Execute performs the select starting cards action
func (a *SelectStartingCardsAction) Execute(ctx context.Context, gameID string, playerID string, cardIDs []string, corporationID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "select_starting_cards"),
		zap.Strings("card_ids", cardIDs),
		zap.String("corporation_id", corporationID),
	)
	log.Info("🃏 Player selecting starting cards and corporation")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. Get player from game
	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 3. BUSINESS LOGIC: Validate selection phase exists (phase state managed by Game)
	selectionPhase := g.GetSelectStartingCardsPhase(playerID)
	if selectionPhase == nil {
		log.Error("Player not in starting card selection phase")
		return fmt.Errorf("not in starting card selection phase")
	}

	// 4. BUSINESS LOGIC: Check if player already has a corporation (selection already complete)
	if player.HasCorporation() {
		log.Error("Starting selection already complete")
		return fmt.Errorf("starting selection already complete")
	}

	// 5. BUSINESS LOGIC: Validate selected cards are in available cards
	availableSet := make(map[string]bool)
	for _, id := range selectionPhase.AvailableCards {
		availableSet[id] = true
	}

	for _, cardID := range cardIDs {
		if !availableSet[cardID] {
			log.Error("Selected card not available", zap.String("card_id", cardID))
			return fmt.Errorf("card %s not available for selection", cardID)
		}
	}

	// 6. BUSINESS LOGIC: Validate corporation is in available corporations
	corpAvailable := false
	for _, corpID := range selectionPhase.AvailableCorporations {
		if corpID == corporationID {
			corpAvailable = true
			break
		}
	}

	if !corpAvailable {
		log.Error("Selected corporation not available")
		return fmt.Errorf("corporation %s not available", corporationID)
	}

	// 7. BUSINESS LOGIC: Calculate cost (3 MC per card)
	cost := len(cardIDs) * 3

	// 8. BUSINESS LOGIC: Fetch corporation card from registry
	corpCard, err := a.cardRegistry.GetByID(corporationID)
	if err != nil {
		log.Error("Failed to fetch corporation card", zap.Error(err))
		return fmt.Errorf("corporation card not found: %s", corporationID)
	}

	// Validate it's actually a corporation card
	if corpCard.Type != gamecards.CardTypeCorporation {
		log.Error("Card is not a corporation",
			zap.String("card_type", string(corpCard.Type)))
		return fmt.Errorf("card %s is not a corporation card", corporationID)
	}

	// 9. BUSINESS LOGIC: Set corporation ID on player
	player.SetCorporationID(corporationID)
	log.Info("✅ Corporation selected", zap.String("corporation_id", corporationID))

	// 10. BUSINESS LOGIC: Apply corporation starting effects (resources and production)
	if err := a.corpProc.ApplyStartingEffects(ctx, corpCard, player, g); err != nil {
		log.Error("Failed to apply corporation starting effects", zap.Error(err))
		return fmt.Errorf("failed to apply corporation starting effects: %w", err)
	}

	// 10a. BUSINESS LOGIC: Apply corporation auto effects (e.g., payment substitutes for Helion)
	if err := a.corpProc.ApplyAutoEffects(ctx, corpCard, player, g); err != nil {
		log.Error("Failed to apply corporation auto effects", zap.Error(err))
		return fmt.Errorf("failed to apply corporation auto effects: %w", err)
	}

	// 10b. BUSINESS LOGIC: Register corporation auto effects for display
	// These are permanent effects like payment substitutes that should show in the effects list
	autoEffects := a.corpProc.GetAutoEffects(corpCard)
	if len(autoEffects) > 0 {
		log.Info("✨ Registering corporation auto effects for display",
			zap.Int("effect_count", len(autoEffects)))

		for _, effect := range autoEffects {
			player.Effects().AddEffect(effect)
			log.Debug("✅ Registered auto effect",
				zap.String("card_id", effect.CardID),
				zap.String("card_name", effect.CardName),
				zap.Int("behavior_index", effect.BehaviorIndex))
		}
	}

	// 10d. BUSINESS LOGIC: Register corporation trigger effects
	// The helper returns CardEffect structs (read-only), we add them to player state (mutation)
	triggerEffects := a.corpProc.GetTriggerEffects(corpCard)
	if len(triggerEffects) > 0 {
		log.Info("⚡ Registering corporation trigger effects",
			zap.Int("effect_count", len(triggerEffects)))

		for _, effect := range triggerEffects {
			player.Effects().AddEffect(effect)
			log.Debug("✅ Registered trigger effect",
				zap.String("card_id", effect.CardID),
				zap.String("card_name", effect.CardName),
				zap.Int("behavior_index", effect.BehaviorIndex))

			// Subscribe trigger effects to relevant events
			subscribePassiveEffectToEvents(ctx, g, player, effect, log)
		}
	}

	// 10e. BUSINESS LOGIC: Register corporation manual actions
	// The helper returns CardAction structs (read-only), we add them to player state (mutation)
	manualActions := a.corpProc.GetManualActions(corpCard)
	if len(manualActions) > 0 {
		log.Info("🎯 Registering corporation manual actions",
			zap.Int("action_count", len(manualActions)))

		for _, action := range manualActions {
			player.Actions().AddAction(action)
			log.Debug("✅ Registered manual action",
				zap.String("card_id", action.CardID),
				zap.String("card_name", action.CardName),
				zap.Int("behavior_index", action.BehaviorIndex))
		}
	}

	// 11. BUSINESS LOGIC: Deduct card selection cost
	resources := player.Resources().Get()
	if resources.Credits < cost {
		log.Error("Insufficient credits after corporation effects",
			zap.Int("cost", cost),
			zap.Int("available", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", cost, resources.Credits)
	}

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: -cost,
	})

	updatedResources := player.Resources().Get()
	log.Info("✅ Card selection cost deducted",
		zap.Int("cost", cost),
		zap.Int("remaining_credits", updatedResources.Credits))

	// 12. BUSINESS LOGIC: Add selected cards to player's hand
	log.Debug("🃏 Adding cards to player hand",
		zap.Strings("card_ids", cardIDs),
		zap.Int("count", len(cardIDs)))

	for _, cardID := range cardIDs {
		player.Hand().AddCard(cardID)
	}

	log.Info("✅ Cards added to hand",
		zap.Strings("card_ids_added", cardIDs),
		zap.Int("card_count", len(cardIDs)))

	// 13. BUSINESS LOGIC: Setup forced first action if corporation requires it
	if err := a.corpProc.SetupForcedFirstAction(ctx, corpCard, g, playerID); err != nil {
		log.Error("Failed to setup forced first action", zap.Error(err))
		return fmt.Errorf("failed to setup forced first action: %w", err)
	}

	// 14. BUSINESS LOGIC: Mark selection as complete (phase state managed by Game)
	if err := g.SetSelectStartingCardsPhase(ctx, playerID, nil); err != nil {
		log.Error("Failed to clear starting cards phase", zap.Error(err))
		return fmt.Errorf("failed to clear starting cards phase: %w", err)
	}

	log.Info("✅ Starting selection marked complete")

	// 15. BUSINESS LOGIC: Check if all players completed selection
	allPlayers := g.GetAllPlayers()
	allComplete := true
	for _, p := range allPlayers {
		if !p.HasCorporation() {
			allComplete = false
			break
		}
	}

	if allComplete {
		log.Info("🎉 All players completed starting selection, advancing to action phase")

		// Advance game phase to Action
		if err := g.UpdatePhase(ctx, game.GamePhaseAction); err != nil {
			log.Error("Failed to transition game phase", zap.Error(err))
			return fmt.Errorf("failed to transition game phase: %w", err)
		}

		// Set current turn to first player from turn order (randomized in start_game)
		turnOrder := g.TurnOrder()
		if len(turnOrder) > 0 {
			firstPlayerID := turnOrder[0]

			// Set available actions based on player count
			availableActions := 2 // Default for multiplayer
			if len(allPlayers) == 1 {
				availableActions = -1 // Unlimited for solo mode
				log.Info("🎮 Solo mode detected - setting unlimited actions")
			}

			// Set current turn with action count
			if err := g.SetCurrentTurn(ctx, firstPlayerID, availableActions); err != nil {
				log.Error("Failed to set current turn", zap.Error(err))
				return fmt.Errorf("failed to set current turn: %w", err)
			}

			log.Info("✅ Set first player turn with actions",
				zap.String("first_player_id", firstPlayerID),
				zap.Int("available_actions", availableActions))
		}
	}

	// 16. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - player.SetCorporationID() publishes events
	//    - player.Resources().Add() publishes ResourcesChangedEvent
	//    - player.Hand().AddCard() publishes CardHandUpdatedEvent
	//    - g.SetSelectStartingCardsPhase() publishes BroadcastEvent
	//    - g.UpdatePhase() publishes GamePhaseChangedEvent + BroadcastEvent
	//    - g.SetCurrentTurn() publishes BroadcastEvent
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("🎉 Starting card selection completed successfully")
	return nil
}
