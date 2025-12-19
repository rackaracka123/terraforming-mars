package admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	baseaction "terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
)

// SetCorporationAction handles the admin action to set a player's corporation
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type SetCorporationAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	corpProc     *gamecards.CorporationProcessor
	logger       *zap.Logger
}

// NewSetCorporationAction creates a new set corporation admin action
func NewSetCorporationAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *SetCorporationAction {
	return &SetCorporationAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		corpProc:     gamecards.NewCorporationProcessor(logger),
		logger:       logger,
	}
}

// Execute performs the set corporation admin action
func (a *SetCorporationAction) Execute(ctx context.Context, gameID string, playerID string, corporationID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "admin_set_corporation"),
		zap.String("corporation_id", corporationID),
	)
	log.Info("🏢 Admin: Setting player corporation")

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

	// 3. Clear old corporation effects if player had one
	oldCorpID := player.CorporationID()
	if oldCorpID != "" {
		log.Info("🧹 Clearing old corporation effects", zap.String("old_corporation_id", oldCorpID))

		// Clear effects from old corporation
		player.Effects().RemoveEffectsByCardID(oldCorpID)

		// Clear actions from old corporation
		player.Actions().RemoveActionsByCardID(oldCorpID)

		// Clear card storage from old corporation
		player.Resources().RemoveCardStorage(oldCorpID)

		// Clear payment substitutes (corporation-specific like Helion's heat)
		player.Resources().ClearPaymentSubstitutes()

		// Clear value modifiers (corporation-specific like Phobolog's titanium bonus)
		player.Resources().ClearValueModifiers()

		log.Info("✅ Old corporation effects cleared")
	}

	// 4. Fetch corporation card from registry
	corpCard, err := a.cardRegistry.GetByID(corporationID)
	if err != nil {
		log.Error("Failed to fetch corporation card", zap.Error(err))
		return fmt.Errorf("corporation card not found: %s", corporationID)
	}

	// Validate it's actually a corporation card
	if corpCard.Type != gamecards.CardTypeCorporation {
		log.Error("Card is not a corporation", zap.String("card_type", string(corpCard.Type)))
		return fmt.Errorf("card %s is not a corporation card", corporationID)
	}

	// 5. Set corporation ID on player
	player.SetCorporationID(corporationID)
	log.Info("✅ Corporation ID set", zap.String("corporation_name", corpCard.Name))

	// 6. Apply corporation starting effects (resources, production)
	if err := a.corpProc.ApplyStartingEffects(ctx, corpCard, player, g); err != nil {
		log.Error("Failed to apply corporation starting effects", zap.Error(err))
		return fmt.Errorf("failed to apply corporation starting effects: %w", err)
	}

	// 7. Apply corporation auto effects (e.g., payment substitutes for Helion)
	if err := a.corpProc.ApplyAutoEffects(ctx, corpCard, player, g); err != nil {
		log.Error("Failed to apply corporation auto effects", zap.Error(err))
		return fmt.Errorf("failed to apply corporation auto effects: %w", err)
	}

	// 8. Register corporation auto effects for display
	autoEffects := a.corpProc.GetAutoEffects(corpCard)
	for _, effect := range autoEffects {
		player.Effects().AddEffect(effect)
		log.Debug("✅ Registered auto effect",
			zap.String("card_name", effect.CardName),
			zap.Int("behavior_index", effect.BehaviorIndex))
	}

	// Note: RequirementModifier recalculation removed - discounts are now calculated on-demand during EntityState calculation

	// 9. Register corporation trigger effects and subscribe to events
	triggerEffects := a.corpProc.GetTriggerEffects(corpCard)
	for _, effect := range triggerEffects {
		player.Effects().AddEffect(effect)
		log.Debug("✅ Registered trigger effect",
			zap.String("card_name", effect.CardName),
			zap.Int("behavior_index", effect.BehaviorIndex))

		// Subscribe trigger effects to relevant events
		baseaction.SubscribePassiveEffectToEvents(ctx, g, player, effect, log)
	}

	// 10. Register corporation manual actions
	manualActions := a.corpProc.GetManualActions(corpCard)
	for _, action := range manualActions {
		player.Actions().AddAction(action)
		log.Debug("✅ Registered manual action",
			zap.String("card_name", action.CardName),
			zap.Int("behavior_index", action.BehaviorIndex))
	}

	// 11. Setup forced first action if corporation requires it
	if err := a.corpProc.SetupForcedFirstAction(ctx, corpCard, g, playerID); err != nil {
		log.Error("Failed to setup forced first action", zap.Error(err))
		return fmt.Errorf("failed to setup forced first action: %w", err)
	}

	log.Info("✅ Admin set corporation completed with all effects applied",
		zap.String("corporation_name", corpCard.Name))
	return nil
}
