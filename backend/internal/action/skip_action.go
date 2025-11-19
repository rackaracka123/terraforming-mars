package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/deck"
	"terraforming-mars-backend/internal/session/game/player"

	"go.uber.org/zap"
)

// SkipActionAction handles the business logic for skipping/passing player turns
type SkipActionAction struct {
	BaseAction
	deckRepo deck.Repository
}

// NewSkipActionAction creates a new skip action action
func NewSkipActionAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	deckRepo deck.Repository,
	sessionMgr session.SessionManager,
) *SkipActionAction {
	return &SkipActionAction{
		BaseAction: NewBaseAction(gameRepo, playerRepo, sessionMgr),
		deckRepo:   deckRepo,
	}
}

// Execute performs the skip action
func (a *SkipActionAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("⏭️ Skipping player turn")

	// 1. Validate game is active
	g, err := ValidateActiveGame(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate it's the player's turn
	if err := ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 3. Get all players
	players, err := GetAllPlayers(ctx, a.playerRepo, gameID, log)
	if err != nil {
		return err
	}

	// 4. Find current player and their index
	var currentPlayer *player.Player
	currentPlayerIndex := -1
	for i, p := range players {
		if p.ID == playerID {
			currentPlayer = p
			currentPlayerIndex = i
			break
		}
	}

	if currentPlayer == nil {
		log.Error("Current player not found in game")
		return fmt.Errorf("player not found in game")
	}

	// 5. Count active players (not passed)
	activePlayerCount := 0
	for _, p := range players {
		if !p.Passed {
			activePlayerCount++
		}
	}

	// 6. Determine PASS vs SKIP behavior
	isPassing := currentPlayer.AvailableActions == 2 || currentPlayer.AvailableActions == -1
	if isPassing {
		// PASS: Player hasn't done any actions or has unlimited actions
		err = a.playerRepo.UpdatePassed(ctx, gameID, playerID, true)
		if err != nil {
			log.Error("Failed to mark player as passed", zap.Error(err))
			return fmt.Errorf("failed to update player passed status: %w", err)
		}

		log.Debug("Player PASSED (marked as passed for generation)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", currentPlayer.AvailableActions))

		// If only one active player remains, grant them unlimited actions
		if activePlayerCount == 2 {
			for _, p := range players {
				if !p.Passed && p.ID != playerID {
					err = a.playerRepo.UpdateAvailableActions(ctx, gameID, p.ID, -1)
					if err != nil {
						log.Error("Failed to grant unlimited actions to last active player",
							zap.String("player_id", p.ID),
							zap.Error(err))
						return fmt.Errorf("failed to update last active player's actions: %w", err)
					}
					log.Info("🏃 Last active player granted unlimited actions due to others passing",
						zap.String("player_id", p.ID))
				}
			}
		}
	} else {
		// SKIP: Player has done some actions, just advance turn without passing
		log.Debug("Player SKIPPED (turn advanced, not passed)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", currentPlayer.AvailableActions))
	}

	// 7. Refresh player list to reflect status changes
	players, err = a.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players after skip processing", zap.Error(err))
		return fmt.Errorf("failed to list players: %w", err)
	}

	// 8. Check if all players have finished their actions
	allPlayersFinished := true
	passedCount := 0
	playersWithNoActions := 0
	for _, p := range players {
		if p.Passed {
			passedCount++
		} else if p.AvailableActions == 0 {
			playersWithNoActions++
		} else if p.AvailableActions > 0 || p.AvailableActions == -1 {
			allPlayersFinished = false
		}
	}

	log.Debug("Checking generation end condition",
		zap.Int("passed_count", passedCount),
		zap.Int("players_with_no_actions", playersWithNoActions),
		zap.Int("total_players", len(players)),
		zap.Bool("all_players_finished", allPlayersFinished))

	// 9. If all players finished, trigger production phase
	if allPlayersFinished {
		log.Info("🏭 All players finished their turns - executing production phase",
			zap.String("game_id", gameID),
			zap.Int("generation", g.Generation),
			zap.Int("passed_players", passedCount),
			zap.Int("players_with_no_actions", playersWithNoActions))

		// Execute production phase inline
		err = a.executeProductionPhase(ctx, gameID, players)
		if err != nil {
			log.Error("Failed to execute production phase", zap.Error(err))
			return fmt.Errorf("failed to execute production phase: %w", err)
		}

		log.Info("✅ Production phase completed, new generation started")
		return nil
	}

	// 10. Find next player who hasn't passed
	nextPlayerIndex := (currentPlayerIndex + 1) % len(players)
	for i := 0; i < len(players); i++ {
		nextPlayer := players[nextPlayerIndex]
		if !nextPlayer.Passed {
			break
		}
		nextPlayerIndex = (nextPlayerIndex + 1) % len(players)
	}

	// 11. Update current turn
	nextPlayerID := players[nextPlayerIndex].ID
	err = a.gameRepo.SetCurrentTurn(ctx, gameID, &nextPlayerID)
	if err != nil {
		log.Error("Failed to update current turn", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("✅ Player turn skipped, advanced to next player",
		zap.String("previous_player", playerID),
		zap.String("current_player", nextPlayerID))

	// 12. Broadcast updated game state
	if err := a.sessionMgr.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after skip turn", zap.Error(err))
		// Non-fatal, don't return error
	}

	return nil
}

// executeProductionPhase handles the production phase when all players have passed
func (a *SkipActionAction) executeProductionPhase(ctx context.Context, gameID string, players []*player.Player) error {
	log := a.logger.With(zap.String("game_id", gameID))
	log.Info("🏭 Starting production phase")

	// 1. For each player: energy→heat, apply production, draw cards
	for _, p := range players {
		// A. Convert energy to heat
		energyConverted := p.Resources.Energy

		// B. Calculate new resources with production
		newResources := model.Resources{
			Credits:  p.Resources.Credits + p.Production.Credits + p.TerraformRating,
			Steel:    p.Resources.Steel + p.Production.Steel,
			Titanium: p.Resources.Titanium + p.Production.Titanium,
			Plants:   p.Resources.Plants + p.Production.Plants,
			Energy:   p.Production.Energy, // Reset to production amount
			Heat:     p.Resources.Heat + energyConverted + p.Production.Heat,
		}

		// C. Update player resources
		err := a.playerRepo.UpdateResources(ctx, gameID, p.ID, newResources)
		if err != nil {
			return fmt.Errorf("failed to update resources for player %s: %w", p.ID, err)
		}

		// D. Reset player state for new generation
		err = a.playerRepo.UpdatePassed(ctx, gameID, p.ID, false)
		if err != nil {
			return fmt.Errorf("failed to reset passed status: %w", err)
		}

		// Set available actions (2 for normal, -1 for solo)
		availableActions := 2
		if len(players) == 1 {
			availableActions = -1 // Unlimited for solo
		}
		err = a.playerRepo.UpdateAvailableActions(ctx, gameID, p.ID, availableActions)
		if err != nil {
			return fmt.Errorf("failed to reset available actions: %w", err)
		}

		// E. Draw 4 cards for production phase selection
		drawnCards := []string{}
		for i := 0; i < 4; i++ {
			cardIDs, err := a.deckRepo.DrawProjectCards(ctx, gameID, 1)
			if err != nil || len(cardIDs) == 0 {
				// Deck might be empty, stop drawing
				log.Debug("⚠️ Deck empty or error drawing card, stopping at card draw",
					zap.Int("cards_drawn", len(drawnCards)),
					zap.Error(err))
				break
			}
			drawnCards = append(drawnCards, cardIDs[0])
		}

		// F. Set production phase data
		productionPhaseData := &model.ProductionPhase{
			AvailableCards:    drawnCards,
			SelectionComplete: false,
			BeforeResources:   p.Resources,  // Before production
			AfterResources:    newResources, // After production
			EnergyConverted:   energyConverted,
			CreditsIncome:     p.Production.Credits + p.TerraformRating,
		}

		err = a.playerRepo.UpdateProductionPhase(ctx, gameID, p.ID, productionPhaseData)
		if err != nil {
			return fmt.Errorf("failed to set production phase: %w", err)
		}

		log.Debug("✅ Production applied for player",
			zap.String("player_id", p.ID),
			zap.Int("cards_drawn", len(drawnCards)),
			zap.Int("credits_income", productionPhaseData.CreditsIncome),
			zap.Int("energy_converted", energyConverted))
	}

	// 2. Get updated game state
	g, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// 3. Increment generation
	newGeneration := g.Generation + 1
	err = a.gameRepo.UpdateGeneration(ctx, gameID, newGeneration)
	if err != nil {
		return fmt.Errorf("failed to increment generation: %w", err)
	}

	// 4. Set current turn to first player
	if len(g.PlayerIDs) > 0 {
		firstPlayerID := g.PlayerIDs[0]
		err = a.gameRepo.SetCurrentTurn(ctx, gameID, &firstPlayerID)
		if err != nil {
			return fmt.Errorf("failed to set current turn: %w", err)
		}
	}

	// 5. Set phase to production and card draw
	err = a.gameRepo.UpdatePhase(ctx, gameID, game.GamePhaseProductionAndCardDraw)
	if err != nil {
		return fmt.Errorf("failed to update phase: %w", err)
	}

	// 6. Broadcast state to all players
	err = a.sessionMgr.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast after production phase", zap.Error(err))
		// Non-fatal, continue
	}

	log.Info("🎉 Production phase complete, generation advanced",
		zap.Int("old_generation", g.Generation),
		zap.Int("new_generation", newGeneration))

	return nil
}
