package action

import (
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
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewSkipActionAction creates a new skip action action
func NewSkipActionAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *SkipActionAction {
	return &SkipActionAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the skip action
func (a *SkipActionAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "skip_action"),
	)
	log.Info("⏭️ Skipping player turn")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. BUSINESS LOGIC: Validate game is active
	if g.Status() != game.GameStatusActive {
		log.Warn("Game is not active", zap.String("status", string(g.Status())))
		return fmt.Errorf("game is not active: %s", g.Status())
	}

	// 3. BUSINESS LOGIC: Validate it's the player's turn
	currentTurn := g.CurrentTurn()
	if currentTurn == nil || currentTurn.PlayerID() != playerID {
		var turnPlayerID string
		if currentTurn != nil {
			turnPlayerID = currentTurn.PlayerID()
		}
		log.Warn("Not player's turn",
			zap.String("current_turn_player", turnPlayerID),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("not your turn")
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
		if !p.Turn().Passed() {
			activePlayerCount++
		}
	}

	// 7. BUSINESS LOGIC: Determine PASS vs SKIP behavior
	// Solo games: skip always means pass (player is done with generation)
	availableActions := currentPlayer.Turn().AvailableActions()
	isPassing := availableActions == 2 || availableActions == -1 || len(players) == 1
	if isPassing {
		// PASS: Player hasn't done any actions or has unlimited actions
		currentPlayer.Turn().SetPassed(true)

		log.Debug("Player PASSED (marked as passed for generation)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", availableActions))

		// If only one active player remains, grant them unlimited actions
		if activePlayerCount == 2 {
			for _, p := range players {
				if !p.Turn().Passed() && p.ID() != playerID {
					p.Turn().SetAvailableActions(-1)
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
	}

	// 8. Refresh player list to reflect status changes
	players = g.GetAllPlayers()

	// 9. BUSINESS LOGIC: Check if all players have finished their actions
	allPlayersFinished := true
	passedCount := 0
	playersWithNoActions := 0
	for _, p := range players {
		if p.Turn().Passed() {
			passedCount++
			continue // Skip remaining checks - this player is done
		}
		pActions := p.Turn().AvailableActions()
		if pActions == 0 {
			playersWithNoActions++
		} else if pActions > 0 || pActions == -1 {
			allPlayersFinished = false
		}
	}

	log.Debug("Checking generation end condition",
		zap.Int("passed_count", passedCount),
		zap.Int("players_with_no_actions", playersWithNoActions),
		zap.Int("total_players", len(players)),
		zap.Bool("all_players_finished", allPlayersFinished))

	// 10. If all players finished, trigger production phase
	if allPlayersFinished {
		log.Info("🏭 All players finished their turns - executing production phase",
			zap.String("game_id", gameID),
			zap.Int("generation", g.Generation()),
			zap.Int("passed_players", passedCount),
			zap.Int("players_with_no_actions", playersWithNoActions))

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
		if !nextPlayer.Turn().Passed() {
			break
		}
		nextPlayerIndex = (nextPlayerIndex + 1) % len(players)
	}

	// 12. Update current turn
	nextPlayerID := players[nextPlayerIndex].ID()
	err = g.SetCurrentTurn(ctx, nextPlayerID, []game.ActionType{})
	if err != nil {
		log.Error("Failed to update current turn", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("✅ Player turn skipped, advanced to next player",
		zap.String("previous_player", playerID),
		zap.String("current_player", nextPlayerID))

	// 13. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - currentPlayer.Turn().SetPassed() publishes events
	//    - g.SetCurrentTurn() publishes BroadcastEvent
	//    - (or production phase publishes many events)
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	return nil
}

// executeProductionPhase handles the production phase when all players have passed
func (a *SkipActionAction) executeProductionPhase(ctx context.Context, gameInstance *game.Game, players []*playerPkg.Player) error {
	log := a.logger.With(zap.String("game_id", gameInstance.ID()))
	log.Info("🏭 Starting production phase")

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
		p.Turn().SetPassed(false)

		// Set available actions (2 for normal, -1 for solo)
		availableActions := 2
		if len(players) == 1 {
			availableActions = -1 // Unlimited for solo
		}
		p.Turn().SetAvailableActions(availableActions)

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

		err := gameInstance.SetProductionPhase(ctx, p.ID(), productionPhaseData)
		if err != nil {
			return fmt.Errorf("failed to set production phase: %w", err)
		}

		log.Debug("✅ Production applied for player",
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

	// 3. Set current turn to first player
	if len(players) > 0 {
		firstPlayerID := players[0].ID()
		if err := gameInstance.SetCurrentTurn(ctx, firstPlayerID, []game.ActionType{}); err != nil {
			return fmt.Errorf("failed to set current turn: %w", err)
		}
	}

	// 4. Set phase to production and card draw
	err := gameInstance.UpdatePhase(ctx, game.GamePhaseProductionAndCardDraw)
	if err != nil {
		return fmt.Errorf("failed to update phase: %w", err)
	}

	log.Info("🎉 Production phase complete, generation advanced",
		zap.Int("old_generation", oldGeneration),
		zap.Int("new_generation", newGeneration))

	return nil
}
