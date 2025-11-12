package turn

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// Service handles turn management operations for active games.
//
// Scope: Isolated turn mechanic
//   - Turn rotation and advancing to next player
//   - Pass vs Skip logic based on available actions
//   - Generation end detection (all players finished)
//   - Last active player unlimited actions
//   - Current turn validation
//
// This mechanic is ISOLATED and should NOT:
//   - Call other mechanic services
//   - Handle production phase execution (that's production mechanic)
//   - Manage global parameters or resources
//
// Dependencies:
//   - GameRepository (for reading/updating current turn)
//   - PlayerRepository (for checking player states and updating passed status)
type Service interface {
	// Turn operations
	SkipTurn(ctx context.Context, gameID, playerID string) (generationEnded bool, err error)
	AdvanceToNextPlayer(ctx context.Context, gameID string) error
	ValidateCurrentPlayer(ctx context.Context, gameID, playerID string) error

	// Generation checks
	IsGenerationEnded(ctx context.Context, gameID string) (bool, error)
	GetNextPlayer(ctx context.Context, gameID string) (string, error)
}

// ServiceImpl implements the Turn mechanic service
type ServiceImpl struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
}

// NewService creates a new Turn mechanic service
func NewService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) Service {
	return &ServiceImpl{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
	}
}

// SkipTurn handles a player skipping/passing their turn.
// Returns whether the generation has ended (all players finished).
// Distinguishes between PASS (haven't used actions) and SKIP (used some actions).
func (s *ServiceImpl) SkipTurn(ctx context.Context, gameID, playerID string) (bool, error) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Processing skip turn")

	// Validate player is current player
	if err := s.ValidateCurrentPlayer(ctx, gameID, playerID); err != nil {
		return false, err
	}

	// Get current player to determine pass vs skip
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for skip turn", zap.Error(err))
		return false, fmt.Errorf("failed to get player: %w", err)
	}

	// Get all players to check active count
	allPlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players for skip turn", zap.Error(err))
		return false, fmt.Errorf("failed to list players: %w", err)
	}

	// Count active players (not passed)
	activePlayerCount := 0
	for _, p := range allPlayers {
		if !p.Passed {
			activePlayerCount++
		}
	}

	// PASS vs SKIP logic: PASS if player has full actions (2 or unlimited)
	isPassing := player.AvailableActions == 2 || player.AvailableActions == -1
	if isPassing {
		// PASS: Mark player as passed for generation end check
		if err := s.playerRepo.UpdatePassed(ctx, gameID, playerID, true); err != nil {
			log.Error("Failed to mark player as passed", zap.Error(err))
			return false, fmt.Errorf("failed to update player passed status: %w", err)
		}

		log.Debug("Player PASSED (marked as passed for generation)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", player.AvailableActions))

		// If only one active player remains, grant them unlimited actions
		if activePlayerCount == 2 {
			for _, p := range allPlayers {
				if !p.Passed && p.ID != playerID {
					if err := s.playerRepo.UpdateAvailableActions(ctx, gameID, p.ID, -1); err != nil {
						log.Error("Failed to grant unlimited actions to last active player",
							zap.String("last_player_id", p.ID),
							zap.Error(err))
						return false, fmt.Errorf("failed to update last active player's actions: %w", err)
					}
					log.Info("🏃 Last active player granted unlimited actions due to others passing",
						zap.String("player_id", p.ID))
				}
			}
		}
	} else {
		// SKIP: Player has used some actions, just advance turn without passing
		log.Debug("Player SKIPPED (turn advanced, not passed)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", player.AvailableActions))
	}

	// Check if generation has ended
	generationEnded, err := s.IsGenerationEnded(ctx, gameID)
	if err != nil {
		log.Error("Failed to check generation end", zap.Error(err))
		return false, fmt.Errorf("failed to check generation end: %w", err)
	}

	if generationEnded {
		log.Info("🏭 All players finished their turns - generation ending",
			zap.String("game_id", gameID))
		return true, nil
	}

	// Advance to next player
	if err := s.AdvanceToNextPlayer(ctx, gameID); err != nil {
		log.Error("Failed to advance to next player", zap.Error(err))
		return false, fmt.Errorf("failed to advance to next player: %w", err)
	}

	return false, nil
}

// AdvanceToNextPlayer advances the current turn to the next player who hasn't passed.
func (s *ServiceImpl) AdvanceToNextPlayer(ctx context.Context, gameID string) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Advancing to next player")

	// Get game and current turn
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	if game.CurrentTurn == nil {
		log.Error("Current turn is nil")
		return fmt.Errorf("current turn is not set")
	}

	// Get all players (in order they were added to game)
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players", zap.Error(err))
		return fmt.Errorf("failed to list players: %w", err)
	}

	// Find current player index in players list
	currentPlayerIndex := -1
	for i, player := range players {
		if player.ID == *game.CurrentTurn {
			currentPlayerIndex = i
			break
		}
	}

	if currentPlayerIndex == -1 {
		log.Error("Current player not found in game", zap.String("current_turn", *game.CurrentTurn))
		return fmt.Errorf("current player not found in game")
	}

	// Find next player who hasn't passed
	nextPlayerIndex := (currentPlayerIndex + 1) % len(players)
	for i := 0; i < len(players); i++ {
		nextPlayer := &players[nextPlayerIndex]
		if !nextPlayer.Passed {
			break
		}
		nextPlayerIndex = (nextPlayerIndex + 1) % len(players)
	}

	nextPlayerID := players[nextPlayerIndex].ID

	// Update current turn
	if err := s.gameRepo.UpdateCurrentTurn(ctx, gameID, &nextPlayerID); err != nil {
		log.Error("Failed to update current turn", zap.Error(err))
		return fmt.Errorf("failed to update current turn: %w", err)
	}

	log.Info("Advanced to next player",
		zap.String("previous_player", *game.CurrentTurn),
		zap.String("current_player", nextPlayerID))

	return nil
}

// ValidateCurrentPlayer validates that the specified player is the current player.
func (s *ServiceImpl) ValidateCurrentPlayer(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get game
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for validation", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game is active
	if game.Status != model.GameStatusActive {
		log.Warn("Attempted turn action in non-active game",
			zap.String("current_status", string(game.Status)))
		return fmt.Errorf("game is not active")
	}

	// Validate current turn is set
	if game.CurrentTurn == nil {
		log.Warn("Attempted turn action but current turn is not set")
		return fmt.Errorf("current turn is not set")
	}

	// Validate player is current player
	if *game.CurrentTurn != playerID {
		log.Warn("Non-current player attempted turn action",
			zap.String("current_player", *game.CurrentTurn),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("only the current player can take turn actions")
	}

	return nil
}

// IsGenerationEnded checks if all players have finished their turns (passed or no actions left).
func (s *ServiceImpl) IsGenerationEnded(ctx context.Context, gameID string) (bool, error) {
	log := logger.WithGameContext(gameID, "")

	// Get all players
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players for generation end check", zap.Error(err))
		return false, fmt.Errorf("failed to list players: %w", err)
	}

	// Check if all players have exhausted their actions or formally passed
	allPlayersFinished := true
	passedCount := 0
	playersWithNoActions := 0

	for _, player := range players {
		if player.Passed {
			passedCount++
		} else if player.AvailableActions == 0 {
			playersWithNoActions++
		} else if player.AvailableActions > 0 || player.AvailableActions == -1 {
			// Player still has actions available
			allPlayersFinished = false
		}
	}

	log.Debug("Generation end check",
		zap.Int("passed_count", passedCount),
		zap.Int("players_with_no_actions", playersWithNoActions),
		zap.Int("total_players", len(players)),
		zap.Bool("all_players_finished", allPlayersFinished))

	return allPlayersFinished, nil
}

// GetNextPlayer returns the ID of the next player who should take a turn.
// Returns the next player who hasn't passed.
func (s *ServiceImpl) GetNextPlayer(ctx context.Context, gameID string) (string, error) {
	log := logger.WithGameContext(gameID, "")

	// Get game and current turn
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return "", fmt.Errorf("failed to get game: %w", err)
	}

	if game.CurrentTurn == nil {
		log.Error("Current turn is nil")
		return "", fmt.Errorf("current turn is not set")
	}

	// Get all players
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players", zap.Error(err))
		return "", fmt.Errorf("failed to list players: %w", err)
	}

	// Find current player index in players list
	currentPlayerIndex := -1
	for i, player := range players {
		if player.ID == *game.CurrentTurn {
			currentPlayerIndex = i
			break
		}
	}

	if currentPlayerIndex == -1 {
		log.Error("Current player not found in game")
		return "", fmt.Errorf("current player not found in game")
	}

	// Find next player who hasn't passed
	nextPlayerIndex := (currentPlayerIndex + 1) % len(players)
	for i := 0; i < len(players); i++ {
		nextPlayer := &players[nextPlayerIndex]
		if !nextPlayer.Passed {
			return nextPlayer.ID, nil
		}
		nextPlayerIndex = (nextPlayerIndex + 1) % len(players)
	}

	// If we get here, all players have passed (shouldn't happen in normal flow)
	log.Warn("All players have passed, returning first player")
	return players[0].ID, nil
}
