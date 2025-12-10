package turn_management

import (
	baseaction "terraforming-mars-backend/internal/action"
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// SkipActionAction handles the business logic for skipping/passing player turns
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type SkipActionAction struct {
	baseaction.BaseAction
}

// NewSkipActionAction creates a new skip action action
func NewSkipActionAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *SkipActionAction {
	return &SkipActionAction{
		BaseAction: baseaction.NewBaseAction(gameRepo, nil),
	}
}

// Execute performs the skip action
func (a *SkipActionAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "skip_action"))
	log.Info("⏭️ Skipping player turn")

	// 1. Fetch game from repository and validate it's active
	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate it's the player's turn
	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 4. Get all players
	players := g.GetAllPlayers()

	// 5. Find current player and their index
	var currentPlayer *playerPkg.Player
	currentPlayerIndex := -1
	for i, p := range players {
		if p.ID() == playerID {
			currentPlayer = p
			currentPlayerIndex = i
			break
		}
	}

	if currentPlayer == nil {
		log.Error("Current player not found in game")
		return fmt.Errorf("player not found in game")
	}

	// 6. Count active players (not passed)
	activePlayerCount := 0
	for _, p := range players {
		if !p.HasPassed() {
			activePlayerCount++
		}
	}

	// 7. BUSINESS LOGIC: Determine PASS vs SKIP behavior
	// Solo games: skip always means pass (player is done with generation)
	currentTurn := g.CurrentTurn()
	if currentTurn == nil {
		log.Error("No current turn set")
		return fmt.Errorf("no current turn set")
	}
	availableActions := currentTurn.ActionsRemaining()
	isPassing := availableActions == 2 || availableActions == -1 || len(players) == 1
	if isPassing {
		// PASS: Player hasn't done any actions or has unlimited actions
		currentPlayer.SetPassed(true)

		log.Debug("Player PASSED (marked as passed for generation)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", availableActions))

		// If only one active player remains, grant them unlimited actions
		if activePlayerCount == 2 {
			for _, p := range players {
				if !p.HasPassed() && p.ID() != playerID {
					// Set the next player's turn with unlimited actions
					if err := g.SetCurrentTurn(ctx, p.ID(), -1); err != nil {
						log.Error("Failed to grant unlimited actions to last player", zap.Error(err))
						return fmt.Errorf("failed to grant unlimited actions: %w", err)
					}
					log.Info("🏃 Last active player granted unlimited actions due to others passing",
						zap.String("player_id", p.ID()))
				}
			}
		}
	} else {
		// SKIP: Player has done some actions, just advance turn without passing
		log.Debug("Player SKIPPED (turn advanced, not passed)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", availableActions))

		// Consume remaining actions before advancing turn
		consumed := a.ConsumePlayerAction(g, log)
		if !consumed {
			log.Warn("⚠️ Could not consume action during skip (unlimited actions or already at 0)")
		}
	}

	// 8. Refresh player list to reflect status changes
	players = g.GetAllPlayers()

	// 9. BUSINESS LOGIC: Check if all players have finished their actions
	// Since actions are now at game level (only current player has actions),
	// we check if all players have passed
	passedCount := 0
	for _, p := range players {
		if p.HasPassed() {
			passedCount++
		}
	}

	allPlayersFinished := passedCount == len(players)

	log.Debug("Checking generation end condition",
		zap.Int("passed_count", passedCount),
		zap.Int("total_players", len(players)),
		zap.Bool("all_players_finished", allPlayersFinished))

	// 10. If all players finished, trigger production phase
	if allPlayersFinished {
		log.Info("🏭 All players finished their turns - executing production phase",
			zap.String("game_id", gameID),
			zap.Int("generation", g.Generation()),
			zap.Int("passed_players", passedCount))

		// Execute production phase inline
		err = a.executeProductionPhase(ctx, g, players)
		if err != nil {
			log.Error("Failed to execute production phase", zap.Error(err))
			return fmt.Errorf("failed to execute production phase: %w", err)
		}

		log.Info("✅ Production phase completed, new generation started")
		return nil
	}

	// 11. Find next player who hasn't passed
	nextPlayerIndex := (currentPlayerIndex + 1) % len(players)
	for i := 0; i < len(players); i++ {
		nextPlayer := players[nextPlayerIndex]
		if !nextPlayer.HasPassed() {
			break
		}
		nextPlayerIndex = (nextPlayerIndex + 1) % len(players)
	}

	// 12. Update current turn - determine action count based on whether it's the same player
	nextPlayerID := players[nextPlayerIndex].ID()
	nextActions := 2 // Default for new turn

	// If the turn is returning to the same player (after forced action or in solo mode),
	// preserve their remaining action count instead of resetting to 2
	if nextPlayerID == playerID && !isPassing {
		// Player skipped but didn't pass - they should keep their current action count
		// (which was already decremented by ConsumePlayerAction above)
		nextActions = currentTurn.ActionsRemaining()
	}

	err = g.SetCurrentTurn(ctx, nextPlayerID, nextActions)
	if err != nil {
		log.Error("Failed to update current turn", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("✅ Player turn skipped, advanced to next player",
		zap.String("previous_player", playerID),
		zap.String("current_player", nextPlayerID))

	return nil
}

// executeProductionPhase handles the production phase when all players have passed
func (a *SkipActionAction) executeProductionPhase(ctx context.Context, gameInstance *game.Game, players []*playerPkg.Player) error {
	log := a.GetLogger().With(zap.String("game_id", gameInstance.ID()))
	log.Info("🏭 Starting production phase",
		zap.Int("player_count", len(players)),
		zap.Int("generation", gameInstance.Generation()))

	deck := gameInstance.Deck()
	if deck == nil {
		return fmt.Errorf("game deck is nil")
	}

	// 1. For each player: energy→heat, apply production, draw cards
	for _, p := range players {
		// A. Get current resources
		currentResources := p.Resources().Get()

		// B. Convert energy to heat
		energyConverted := currentResources.Energy

		// C. Calculate new resources with production
		production := p.Resources().Production()
		tr := p.Resources().TerraformRating()
		newResources := shared.Resources{
			Credits:  currentResources.Credits + production.Credits + tr,
			Steel:    currentResources.Steel + production.Steel,
			Titanium: currentResources.Titanium + production.Titanium,
			Plants:   currentResources.Plants + production.Plants,
			Energy:   production.Energy, // Reset to production amount
			Heat:     currentResources.Heat + energyConverted + production.Heat,
		}

		// D. Update player resources
		p.Resources().Set(newResources)

		// E. Reset player state for new generation
		p.SetPassed(false)

		// F. Draw 4 cards for production phase selection
		drawnCards := []string{}
		for i := 0; i < 4; i++ {
			cardIDs, err := deck.DrawProjectCards(ctx, 1)
			if err != nil || len(cardIDs) == 0 {
				// Deck might be empty, stop drawing
				log.Debug("⚠️ Deck empty or error drawing card, stopping at card draw",
					zap.Int("cards_drawn", len(drawnCards)),
					zap.Error(err))
				break
			}
			drawnCards = append(drawnCards, cardIDs[0])
		}

		// G. Set production phase data (phase state managed by Game)
		productionPhaseData := &playerPkg.ProductionPhase{
			AvailableCards:    drawnCards,
			SelectionComplete: false,
			BeforeResources:   currentResources, // Before production
			AfterResources:    newResources,     // After production
			EnergyConverted:   energyConverted,
			CreditsIncome:     production.Credits + tr,
		}

		log.Info("📋 Setting production phase data for player",
			zap.String("player_id", p.ID()),
			zap.Int("available_cards", len(drawnCards)))

		err := gameInstance.SetProductionPhase(ctx, p.ID(), productionPhaseData)
		if err != nil {
			log.Error("❌ Failed to set production phase", zap.Error(err))
			return fmt.Errorf("failed to set production phase: %w", err)
		}

		log.Info("✅ Production phase data set successfully",
			zap.String("player_id", p.ID()),
			zap.Int("cards_drawn", len(drawnCards)),
			zap.Int("credits_income", productionPhaseData.CreditsIncome),
			zap.Int("energy_converted", energyConverted))
	}

	// 2. Increment generation
	oldGeneration := gameInstance.Generation()
	if err := gameInstance.AdvanceGeneration(ctx); err != nil {
		return fmt.Errorf("failed to increment generation: %w", err)
	}
	newGeneration := gameInstance.Generation()

	// 3. Set current turn to first player with 2 actions (or -1 for solo)
	if len(players) > 0 {
		firstPlayerID := players[0].ID()
		actionsForNewGeneration := 2
		if len(players) == 1 {
			actionsForNewGeneration = -1 // Unlimited for solo
		}
		if err := gameInstance.SetCurrentTurn(ctx, firstPlayerID, actionsForNewGeneration); err != nil {
			return fmt.Errorf("failed to set current turn: %w", err)
		}
	}

	// 4. Set phase to production and card draw
	log.Info("🔄 Updating game phase to production_and_card_draw",
		zap.String("current_phase", string(gameInstance.CurrentPhase())),
		zap.String("new_phase", string(game.GamePhaseProductionAndCardDraw)))

	err := gameInstance.UpdatePhase(ctx, game.GamePhaseProductionAndCardDraw)
	if err != nil {
		log.Error("❌ Failed to update phase", zap.Error(err))
		return fmt.Errorf("failed to update phase: %w", err)
	}

	log.Info("✅ Game phase updated successfully",
		zap.String("phase", string(gameInstance.CurrentPhase())))

	log.Info("🎉 Production phase complete, generation advanced",
		zap.Int("old_generation", oldGeneration),
		zap.Int("new_generation", newGeneration))

	return nil
}
