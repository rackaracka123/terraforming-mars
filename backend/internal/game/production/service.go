package production

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Service handles production phase operations for active games.
//
// Scope: Isolated production phase mechanic
//   - Energy to heat conversion
//   - Production application (resources += production)
//   - TR as income (credits += TR)
//   - Card drawing for next generation
//   - Player state reset (passed, available actions)
//   - Generation advancement
//   - Production phase ready coordination
//
// This mechanic is ISOLATED and should NOT:
//   - Call other mechanic services
//   - Handle turn rotation (that's turn mechanic)
//   - Manage tile placement or global parameters
//
// Dependencies:
//   - GameRepository (for reading/updating generation and phase)
//   - PlayerRepository (for updating resources, production data, player state)
//   - CardDeckRepository (for drawing cards for next generation)
type Service interface {
	// Production phase execution
	ExecuteProductionPhase(ctx context.Context, gameID string) error
	ProcessPlayerReady(ctx context.Context, gameID, playerID string) (allReady bool, err error)

	// Helper operations
	ApplyProduction(ctx context.Context, gameID, playerID string) (oldResources, newResources Resources, energyConverted int, err error)
	DrawProductionCards(ctx context.Context, gameID, playerID string, count int) ([]string, error)
	ResetPlayerForNewGeneration(ctx context.Context, gameID, playerID string, isSolo bool) error
}

// ServiceImpl implements the Production Phase mechanic service
type ServiceImpl struct {
	repo Repository
}

// NewService creates a new Production Phase mechanic service
func NewService(repo Repository) Service {
	return &ServiceImpl{
		repo: repo,
	}
}

// ExecuteProductionPhase executes the production phase for all players.
// Converts energy to heat, applies production, draws cards, resets player state, and advances generation.
func (s *ServiceImpl) ExecuteProductionPhase(ctx context.Context, gameID string) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Executing production phase")

	// Get current game state
	game, err := s.repo.GetGame(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for production phase", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game is active
	if game.Status != GameStatusActive {
		log.Warn("Attempted to execute production phase in non-active game",
			zap.String("current_status", string(game.Status)))
		return fmt.Errorf("game is not active")
	}

	// Get all players
	gamePlayers, err := s.repo.ListPlayers(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players for production phase", zap.Error(err))
		return fmt.Errorf("failed to list players: %w", err)
	}

	// Determine if solo game (affects available actions)
	isSolo := len(gamePlayers) == 1

	// Process production for each player
	for i := range gamePlayers {
		player := &gamePlayers[i]

		// Apply production to player
		oldResources, newResources, energyConverted, err := s.ApplyProduction(ctx, gameID, player.ID)
		if err != nil {
			log.Error("Failed to apply production to player",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return fmt.Errorf("failed to apply production to player %s: %w", player.ID, err)
		}

		// Draw 4 cards for production phase card selection
		drawnCards, err := s.DrawProductionCards(ctx, gameID, player.ID, 4)
		if err != nil {
			log.Warn("Failed to draw cards for production phase",
				zap.String("player_id", player.ID),
				zap.Error(err))
			// Continue with empty cards rather than failing
			drawnCards = []string{}
		}

		log.Debug("Drew cards for production phase",
			zap.String("player_id", player.ID),
			zap.Strings("drawn_cards", drawnCards),
			zap.Int("cards_drawn", len(drawnCards)))

		// Set production phase data
		productionPhaseData := ProductionPhase{
			AvailableCards:    drawnCards,
			SelectionComplete: false,
			BeforeResources:   oldResources,
			AfterResources:    newResources,
			EnergyConverted:   energyConverted,
			CreditsIncome:     player.Production.Credits + player.TerraformRating,
		}

		if err := s.repo.UpdateProductionPhase(ctx, gameID, player.ID, &productionPhaseData); err != nil {
			log.Error("Failed to set player production phase data",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return fmt.Errorf("failed to set player production phase data: %w", err)
		}

		// Reset player state for new generation
		if err := s.ResetPlayerForNewGeneration(ctx, gameID, player.ID, isSolo); err != nil {
			log.Error("Failed to reset player for new generation",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return fmt.Errorf("failed to reset player: %w", err)
		}

		log.Debug("Applied production to player",
			zap.String("player_id", player.ID),
			zap.Int("energy_converted", energyConverted),
			zap.Int("credits_gained", player.Production.Credits+player.TerraformRating))
	}

	// Advance generation
	if err := s.repo.UpdateGeneration(ctx, game.ID, game.Generation+1); err != nil {
		log.Error("Failed to update game generation", zap.Error(err))
		return fmt.Errorf("failed to update game generation: %w", err)
	}

	// Set phase to production and card draw
	if err := s.repo.UpdatePhase(ctx, game.ID, GamePhaseProductionAndCardDraw); err != nil {
		log.Error("Failed to update game phase", zap.Error(err))
		return fmt.Errorf("failed to update game phase: %w", err)
	}

	// Update the current turn to the first player for the new generation
	if len(game.PlayerIDs) > 0 {
		if err := s.repo.UpdateCurrentTurn(ctx, game.ID, &game.PlayerIDs[0]); err != nil {
			log.Error("Failed to update current turn for new generation", zap.Error(err))
			return fmt.Errorf("failed to update current turn: %w", err)
		}
	}

	log.Info("🏭 Production phase started",
		zap.String("game_id", gameID),
		zap.Int("generation", game.Generation+1),
		zap.Int("player_count", len(gamePlayers)))

	return nil
}

// ApplyProduction converts energy to heat and applies production to a player's resources.
// Returns old resources, new resources, and energy converted.
func (s *ServiceImpl) ApplyProduction(ctx context.Context, gameID, playerID string) (Resources, Resources, int, error) {
	log := logger.WithGameContext(gameID, playerID)

	// Get player
	player, err := s.repo.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for production", zap.Error(err))
		return Resources{}, Resources{}, 0, fmt.Errorf("failed to get player: %w", err)
	}

	// Save old resources for production phase display
	oldResources := player.Resources.DeepCopy()

	// Convert energy to heat
	energyConverted := player.Resources.Energy

	// Calculate new resources: resources + production, energy becomes 0, heat += old energy
	newResources := Resources{
		Credits:  player.Resources.Credits + player.Production.Credits + player.TerraformRating, // TR provides 1 credit per point
		Steel:    player.Resources.Steel + player.Production.Steel,
		Titanium: player.Resources.Titanium + player.Production.Titanium,
		Plants:   player.Resources.Plants + player.Production.Plants,
		Energy:   player.Production.Energy, // Energy resets to production value
		Heat:     player.Resources.Heat + energyConverted + player.Production.Heat,
	}

	// Update player resources
	if err := s.repo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
		log.Error("Failed to update player resources during production", zap.Error(err))
		return Resources{}, Resources{}, 0, fmt.Errorf("failed to update player resources: %w", err)
	}

	log.Debug("Applied production to player",
		zap.Int("energy_converted", energyConverted),
		zap.Int("credits_gained", player.Production.Credits+player.TerraformRating))

	return oldResources, newResources, energyConverted, nil
}

// DrawProductionCards draws cards from the deck for production phase card selection.
// Returns the drawn card IDs (may be fewer than requested if deck is exhausted).
func (s *ServiceImpl) DrawProductionCards(ctx context.Context, gameID, playerID string, count int) ([]string, error) {
	log := logger.WithGameContext(gameID, playerID)

	drawnCards := []string{}
	for i := 0; i < count; i++ {
		cardID, err := s.repo.PopCard(ctx, gameID)
		if err != nil {
			log.Warn("Failed to draw card from deck for production phase",
				zap.Int("drawn_so_far", i),
				zap.Error(err))
			// Return whatever cards we managed to draw
			break
		}
		drawnCards = append(drawnCards, cardID)
	}

	return drawnCards, nil
}

// ResetPlayerForNewGeneration resets player state for the new generation.
// Sets passed = false and available actions (2 for multiplayer, -1 for solo).
func (s *ServiceImpl) ResetPlayerForNewGeneration(ctx context.Context, gameID, playerID string, isSolo bool) error {
	log := logger.WithGameContext(gameID, playerID)

	// Reset passed status
	if err := s.repo.UpdatePassed(ctx, gameID, playerID, false); err != nil {
		log.Error("Failed to reset player passed status", zap.Error(err))
		return fmt.Errorf("failed to reset player passed status: %w", err)
	}

	// Reset available actions (2 for multiplayer, -1 for solo)
	resetTurns := 2
	if isSolo {
		resetTurns = -1 // Unlimited actions for solo play
	}

	if err := s.repo.UpdateAvailableActions(ctx, gameID, playerID, resetTurns); err != nil {
		log.Error("Failed to reset player available actions", zap.Error(err))
		return fmt.Errorf("failed to reset player available actions: %w", err)
	}

	return nil
}

// ProcessPlayerReady processes a player marking themselves as ready after production phase.
// Returns whether all players are ready (ready to advance to action phase).
func (s *ServiceImpl) ProcessPlayerReady(ctx context.Context, gameID, playerID string) (bool, error) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Processing production phase ready")

	// Get current game state
	game, err := s.repo.GetGame(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for production ready", zap.Error(err))
		return false, fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game state
	if game.Status != GameStatusActive {
		log.Warn("Attempted to mark ready in non-active game",
			zap.String("current_status", string(game.Status)))
		return false, fmt.Errorf("game is not active")
	}

	if game.CurrentPhase != GamePhaseProductionAndCardDraw {
		log.Warn("Attempted to mark ready in non-production phase",
			zap.String("current_phase", string(game.CurrentPhase)))
		return false, fmt.Errorf("game is not in production phase")
	}

	// Check if all players are ready
	players, err := s.repo.ListPlayers(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players to check readiness", zap.Error(err))
		return false, fmt.Errorf("failed to list players: %w", err)
	}

	allReady := true
	readyCount := 0
	for _, player := range players {
		if player.ProductionPhase.SelectionComplete {
			readyCount++
		} else {
			allReady = false
		}
	}

	log.Debug("Checking production phase readiness",
		zap.Int("ready_count", readyCount),
		zap.Int("total_players", len(players)),
		zap.Bool("all_ready", allReady))

	if allReady {
		log.Info("🎯 All players ready for next phase - advancing from production to action",
			zap.String("game_id", gameID),
			zap.Int("generation", game.Generation))

		// Advance to action phase
		if err := s.repo.UpdatePhase(ctx, gameID, GamePhaseAction); err != nil {
			log.Error("Failed to advance phase to action", zap.Error(err))
			return false, fmt.Errorf("failed to advance phase to action: %w", err)
		}

		// Set first player's turn for new generation
		if len(players) > 0 {
			firstPlayerID := players[0].ID
			if err := s.repo.SetCurrentTurn(ctx, gameID, &firstPlayerID); err != nil {
				log.Error("Failed to set current turn for new generation", zap.Error(err))
				return false, fmt.Errorf("failed to set current turn: %w", err)
			}
		}
	}

	return allReady, nil
}
