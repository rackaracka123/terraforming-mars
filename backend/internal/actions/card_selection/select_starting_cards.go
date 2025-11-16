package card_selection

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// SelectStartingCardsAction handles selection of starting cards and corporation
// This action orchestrates:
// - Card and corporation validation and selection
// - Checking if all players are ready
// - Game phase advancement to action phase when all players ready
type SelectStartingCardsAction struct {
	playerRepo     player.Repository
	cardRepo       card.CardRepository
	gameRepo       game.Repository
	sessionManager session.SessionManager
}

// NewSelectStartingCardsAction creates a new select starting cards action
func NewSelectStartingCardsAction(
	playerRepo player.Repository,
	cardRepo card.CardRepository,
	gameRepo game.Repository,
	sessionManager session.SessionManager,
) *SelectStartingCardsAction {
	return &SelectStartingCardsAction{
		playerRepo:     playerRepo,
		cardRepo:       cardRepo,
		gameRepo:       gameRepo,
		sessionManager: sessionManager,
	}
}

// Execute performs the select starting cards action
// Steps:
// 1. Validate selected cards and corporation
// 2. Update player with selected cards and corporation
// 3. Extract and store corporation effects
// 4. Check if all players have completed selection
// 5. If all ready: advance to action phase
// 6. Broadcast state
func (a *SelectStartingCardsAction) Execute(ctx context.Context, gameID string, playerID string, cardIDs []string, corporationID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🃏 Executing select starting cards action",
		zap.Int("card_count", len(cardIDs)),
		zap.String("corporation_id", corporationID))

	// Get player to access selection phase data
	playerData, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Validate player has a starting card selection phase
	if playerData.SelectStartingCardsPhase == nil {
		return fmt.Errorf("player does not have a starting card selection phase")
	}

	// Check if selection already complete
	if playerData.SelectStartingCardsPhase.SelectionComplete {
		log.Warn("Player already completed starting card selection")
		return fmt.Errorf("starting card selection already complete")
	}

	availableCards := playerData.SelectStartingCardsPhase.AvailableCards
	availableCorps := playerData.SelectStartingCardsPhase.AvailableCorporations

	// Validate selected cards are in available cards
	for _, cardID := range cardIDs {
		found := false
		for _, availID := range availableCards {
			if cardID == availID {
				found = true
				break
			}
		}
		if !found {
			log.Error("Selected card not in available cards",
				zap.String("card_id", cardID))
			return fmt.Errorf("card %s not in available cards", cardID)
		}
	}

	// Validate corporation is in available corporations
	corpFound := false
	for _, availCorpID := range availableCorps {
		if corporationID == availCorpID {
			corpFound = true
			break
		}
	}
	if !corpFound {
		log.Error("Selected corporation not in available corporations",
			zap.String("corporation_id", corporationID))
		return fmt.Errorf("corporation %s not in available corporations", corporationID)
	}

	// Fetch card data for selected cards
	selectedCards := make([]player.Card, 0, len(cardIDs))
	for _, cardID := range cardIDs {
		cardData, err := a.cardRepo.GetCardByID(ctx, cardID)
		if err != nil {
			log.Error("Failed to fetch card data",
				zap.String("card_id", cardID),
				zap.Error(err))
			return fmt.Errorf("failed to fetch card %s: %w", cardID, err)
		}
		selectedCards = append(selectedCards, player.Card(*cardData))
	}

	log.Debug("Fetched card data for selected cards", zap.Int("count", len(selectedCards)))

	// Fetch corporation card data
	corpCard, err := a.cardRepo.GetCardByID(ctx, corporationID)
	if err != nil {
		log.Error("Failed to fetch corporation card data",
			zap.String("corporation_id", corporationID),
			zap.Error(err))
		return fmt.Errorf("failed to fetch corporation %s: %w", corporationID, err)
	}

	log.Debug("Fetched corporation card data", zap.String("corporation_name", corpCard.Name))

	// Add selected cards to player's hand
	if len(selectedCards) > 0 {
		if err := a.playerRepo.AddCards(ctx, gameID, playerID, selectedCards); err != nil {
			log.Error("Failed to add selected cards to player hand", zap.Error(err))
			return fmt.Errorf("failed to add cards to hand: %w", err)
		}
		log.Debug("Added selected cards to player hand", zap.Int("count", len(selectedCards)))
	}

	// Set player's corporation
	if err := a.playerRepo.UpdateCorporation(ctx, gameID, playerID, player.Card(*corpCard)); err != nil {
		log.Error("Failed to set player corporation", zap.Error(err))
		return fmt.Errorf("failed to set corporation: %w", err)
	}
	log.Debug("Set player corporation", zap.String("corporation_name", corpCard.Name))

	// Apply corporation starting resources and production
	if corpCard.StartingResources != nil || corpCard.StartingProduction != nil {
		// Get current player state to add bonuses
		currentPlayer, err := a.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to get player for starting bonuses: %w", err)
		}

		// Apply starting resources
		if corpCard.StartingResources != nil {
			newResources := currentPlayer.Resources
			for resourceType, amount := range *corpCard.StartingResources {
				switch resourceType {
				case domain.ResourceTypeCredits:
					newResources.Credits += amount
				case domain.ResourceTypeSteel:
					newResources.Steel += amount
				case domain.ResourceTypeTitanium:
					newResources.Titanium += amount
				case domain.ResourceTypePlants:
					newResources.Plants += amount
				case domain.ResourceTypeEnergy:
					newResources.Energy += amount
				case domain.ResourceTypeHeat:
					newResources.Heat += amount
				}
			}
			if err := a.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
				return fmt.Errorf("failed to apply starting resources: %w", err)
			}
			log.Info("✅ Applied corporation starting resources",
				zap.Int("credits", newResources.Credits),
				zap.Int("steel", newResources.Steel),
				zap.Int("titanium", newResources.Titanium))
		}

		// Apply starting production
		if corpCard.StartingProduction != nil {
			newProduction := currentPlayer.Production
			for resourceType, amount := range *corpCard.StartingProduction {
				switch resourceType {
				case domain.ResourceTypeCredits:
					newProduction.Credits += amount
				case domain.ResourceTypeSteel:
					newProduction.Steel += amount
				case domain.ResourceTypeTitanium:
					newProduction.Titanium += amount
				case domain.ResourceTypePlants:
					newProduction.Plants += amount
				case domain.ResourceTypeEnergy:
					newProduction.Energy += amount
				case domain.ResourceTypeHeat:
					newProduction.Heat += amount
				}
			}
			if err := a.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
				return fmt.Errorf("failed to apply starting production: %w", err)
			}
			log.Info("✅ Applied corporation starting production",
				zap.Int("credits_production", newProduction.Credits),
				zap.Int("steel_production", newProduction.Steel),
				zap.Int("energy_production", newProduction.Energy))
		}
	}

	// Deduct card purchase cost (3 MC per card)
	cardCost := len(cardIDs) * 3
	if cardCost > 0 {
		// Get current player resources
		currentPlayer, err := a.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to get player for card cost deduction: %w", err)
		}

		// Validate player has enough credits
		if currentPlayer.Resources.Credits < cardCost {
			return fmt.Errorf("insufficient credits: need %d MC, have %d MC", cardCost, currentPlayer.Resources.Credits)
		}

		// Deduct credits
		newResources := currentPlayer.Resources
		newResources.Credits -= cardCost

		if err := a.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
			return fmt.Errorf("failed to deduct card cost: %w", err)
		}

		log.Info("💰 Deducted card purchase cost",
			zap.Int("cards_purchased", len(cardIDs)),
			zap.Int("cost", cardCost),
			zap.Int("remaining_credits", newResources.Credits))
	}

	// Mark selection as complete
	if err := a.playerRepo.SetStartingCardsSelectionComplete(ctx, gameID, playerID); err != nil {
		log.Error("Failed to mark starting card selection as complete", zap.Error(err))
		return fmt.Errorf("failed to mark selection complete: %w", err)
	}

	// Clear the selection phase now that selection is complete
	if err := a.playerRepo.ClearStartingCardsPhase(ctx, gameID, playerID); err != nil {
		log.Error("Failed to clear starting card selection phase", zap.Error(err))
		return fmt.Errorf("failed to clear selection phase: %w", err)
	}

	log.Info("✅ Player completed starting card selection",
		zap.String("corporation_name", corpCard.Name),
		zap.Int("cards_selected", len(selectedCards)))

	// Extract and store corporation passive effects
	playerEffects := a.extractPlayerEffects(corporationID, corpCard, log)
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

		log.Info("✨ Corporation effects stored",
			zap.Int("new_effects_added", len(playerEffects)),
			zap.Int("total_effects", len(updatedEffects)))
	}

	log.Debug("🃏 Player completed starting card selection", zap.Strings("card_ids", cardIDs))

	// Check if all players have completed their starting card selection
	if a.isAllPlayersSelectionComplete(ctx, gameID, log) {
		log.Info("✅ All players completed starting card selection, advancing to action phase")

		// Get current game state to validate phase transition
		gameState, err := a.gameRepo.GetByID(ctx, gameID)
		if err != nil {
			log.Error("Failed to get game for phase advancement", zap.Error(err))
			return fmt.Errorf("failed to get game: %w", err)
		}

		// Validate current phase before transition
		if gameState.CurrentPhase != game.GamePhaseStartingCardSelection {
			log.Warn("Game is not in starting card selection phase, skipping phase transition",
				zap.String("current_phase", string(gameState.CurrentPhase)))
		} else if gameState.Status != game.GameStatusActive {
			log.Warn("Game is not active, skipping phase transition",
				zap.String("current_status", string(gameState.Status)))
		} else {
			// Advance to action phase
			if err := a.gameRepo.UpdatePhase(ctx, gameID, game.GamePhaseAction); err != nil {
				log.Error("Failed to update game phase", zap.Error(err))
				return fmt.Errorf("failed to update game phase: %w", err)
			}

			// Set first player as current player when entering action phase
			if len(gameState.PlayerIDs) > 0 {
				firstPlayerID := gameState.PlayerIDs[0]
				if err := a.gameRepo.SetCurrentPlayer(ctx, gameID, firstPlayerID); err != nil {
					log.Error("Failed to set initial current player", zap.Error(err))
					return fmt.Errorf("failed to set initial current player: %w", err)
				}
				log.Info("🎯 Set initial current player", zap.String("player_id", firstPlayerID))
			}

			log.Info("🎯 Game phase advanced successfully",
				zap.String("previous_phase", string(game.GamePhaseStartingCardSelection)),
				zap.String("new_phase", string(game.GamePhaseAction)))

			// Note: Forced actions are now triggered via event system (GamePhaseChangedEvent)
		}
	}

	// Broadcast updated game state to all players after successful card selection (and potential phase change)
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after starting card selection", zap.Error(err))
		// Don't fail the card selection operation, just log the error
	}

	log.Info("✅ Select starting cards action completed successfully")

	return nil
}

// isAllPlayersSelectionComplete checks if all players in a game have completed their card selection
func (a *SelectStartingCardsAction) isAllPlayersSelectionComplete(ctx context.Context, gameID string, log *zap.Logger) bool {
	// Get all players in the game
	players, err := a.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players for completion check",
			zap.String("game_id", gameID),
			zap.Error(err))
		return false
	}

	if len(players) == 0 {
		log.Warn("No players found in game",
			zap.String("game_id", gameID))
		return false
	}

	// Check each player's selection completion status
	for _, p := range players {
		if p.SelectStartingCardsPhase == nil {
			// Player doesn't have a selection phase (shouldn't happen)
			log.Debug("Player has no starting card selection phase",
				zap.String("game_id", gameID),
				zap.String("player_id", p.ID))
			return false
		}

		if !p.SelectStartingCardsPhase.SelectionComplete {
			log.Debug("Player has not completed starting card selection",
				zap.String("game_id", gameID),
				zap.String("player_id", p.ID))
			return false
		}
	}

	log.Info("✅ All players completed starting card selection",
		zap.String("game_id", gameID),
		zap.Int("player_count", len(players)))

	return true
}

// extractPlayerEffects extracts passive effects from a card's behaviors
// Returns both reactive effects (auto triggers with conditions) and static effects (discounts, modifiers)
func (a *SelectStartingCardsAction) extractPlayerEffects(cardID string, c *card.Card, log *zap.Logger) []card.PlayerEffect {
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
