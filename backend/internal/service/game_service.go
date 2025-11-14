package service

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/features/production"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/features/turn"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// GameService handles active game operations including turn management, phases, and player actions.
//
// Scope: Active gameplay operations ONLY (game.Status = "active")
//   - Turn management and rotation
//   - Production phase execution
//   - Global parameter reads
//   - Player reconnection handling
//
// Boundary: Pre-game operations (game creation, joining, starting) are handled by lobby.Service
type GameService interface {
	// Get game by ID
	GetGame(ctx context.Context, gameID string) (model.Game, error)

	// Skip a player's turn (advance to next player)
	SkipPlayerTurn(ctx context.Context, gameID string, playerID string) error

	// Get global parameters (read-only access)
	GetGlobalParameters(ctx context.Context, gameID string) (model.GlobalParameters, error)

	// Process production phase ready acknowledgment from client
	ProcessProductionPhaseReady(ctx context.Context, gameID string, playerID string) (*model.Game, error)

	// Handle player reconnection - updates connection status and sends complete game state
	PlayerReconnected(ctx context.Context, gameID string, playerID string) error
}

// GameServiceImpl implements GameService interface
type GameServiceImpl struct {
	gameRepo          game.Repository
	playerRepo        player.Repository
	cardRepo          game.CardRepository
	cardService       CardService
	cardDeckRepo      game.CardDeckRepository
	boardService      BoardService
	sessionManager    session.SessionManager
	turnFeature       turn.Service
	productionFeature production.Service
	tilesFeature      tiles.Service
}

// NewGameService creates a new GameService instance
func NewGameService(
	gameRepo game.Repository,
	playerRepo player.Repository,
	cardRepo game.CardRepository,
	cardService CardService,
	cardDeckRepo game.CardDeckRepository,
	boardService BoardService,
	sessionManager session.SessionManager,
	turnFeature turn.Service,
	productionFeature production.Service,
	tilesFeature tiles.Service,
) GameService {
	return &GameServiceImpl{
		gameRepo:          gameRepo,
		playerRepo:        playerRepo,
		cardRepo:          cardRepo,
		cardService:       cardService,
		cardDeckRepo:      cardDeckRepo,
		boardService:      boardService,
		sessionManager:    sessionManager,
		turnFeature:       turnFeature,
		productionFeature: productionFeature,
		tilesFeature:      tilesFeature,
	}
}

// GetGame retrieves a game by ID
func (s *GameServiceImpl) GetGame(ctx context.Context, gameID string) (model.Game, error) {
	return s.gameRepo.GetByID(ctx, gameID)
}

// GetGameForPlayer gets a game prepared for a specific player's perspective
func (s *GameServiceImpl) GetGameForPlayer(ctx context.Context, gameID string, playerID string) (model.Game, error) {
	// Get the game data
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return model.Game{}, err
	}

	// Get the players separately (clean architecture pattern)
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return model.Game{}, err
	}

	// Create a copy of the game to modify
	gameCopy := game

	// Note: CurrentPlayer and OtherPlayers are legacy fields
	// In clean architecture, frontend should call player repo directly
	// For backward compatibility, we'll skip these fields for now

	// The frontend should use PlayerIds to fetch players when needed
	otherPlayersCap := len(players) - 1
	if otherPlayersCap < 0 {
		otherPlayersCap = 0
	}
	// Clean architecture: Frontend should fetch players separately using PlayerIds
	// No need for OtherPlayers or CurrentPlayer in this layer

	return gameCopy, nil
}

// SkipPlayerTurnResult contains the result of skipping a player's turn
type SkipPlayerTurnResult struct {
	AllPlayersPassed bool
	Game             *model.Game
}

func (s *GameServiceImpl) SkipPlayerTurn(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Skipping player turn via GameService")

	// Delegate to turn mechanic
	generationEnded, err := s.turnFeature.SkipTurn(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to skip turn", zap.Error(err))
		return fmt.Errorf("failed to skip turn: %w", err)
	}

	// If generation ended, execute production phase
	if generationEnded {
		log.Info("🏭 Generation ended - executing production phase")

		if err := s.productionFeature.ExecuteProductionPhase(ctx, gameID); err != nil {
			log.Error("Failed to execute production phase", zap.Error(err))
			return fmt.Errorf("failed to execute production phase: %w", err)
		}
	}

	// Broadcast updated game state to all players
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after skip turn", zap.Error(err))
		// Don't fail the skip operation, just log the error
	}

	return nil
}

// ProcessProductionPhaseReady processes a player's ready acknowledgment and transitions phase when all players are ready
func (s *GameServiceImpl) ProcessProductionPhaseReady(ctx context.Context, gameID string, playerID string) (*model.Game, error) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Processing production phase ready via GameService")

	// Delegate to production mechanic
	allReady, err := s.productionFeature.ProcessPlayerReady(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to process player ready", zap.Error(err))
		return nil, fmt.Errorf("failed to process player ready: %w", err)
	}

	// If all players are ready, additional game service cleanup
	if allReady {
		// Reset all player action play counts for the new generation
		if err := s.resetPlayerActionPlayCounts(ctx, gameID); err != nil {
			log.Error("Failed to reset player action play counts", zap.Error(err))
			return nil, fmt.Errorf("failed to reset action play counts: %w", err)
		}

		// Clear production phase data for all players
		updatedPlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
		if err != nil {
			log.Error("Failed to list players for cleanup", zap.Error(err))
			return nil, fmt.Errorf("failed to list players: %w", err)
		}

		for _, player := range updatedPlayers {
			if err := s.playerRepo.UpdateProductionPhase(ctx, gameID, player.ID, nil); err != nil {
				log.Error("Failed to clear player production phase data",
					zap.String("player_id", player.ID),
					zap.Error(err))
				return nil, fmt.Errorf("failed to clear production phase data for player %s: %w", player.ID, err)
			}
		}

		log.Info("Production phase completed, advanced to action phase", zap.String("game_id", gameID))
	}

	// Get updated game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game after production ready", zap.Error(err))
		return nil, fmt.Errorf("failed to get updated game: %w", err)
	}

	// Broadcast updated state
	err = s.sessionManager.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast game state after production ready", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return &game, nil
}

// GetGlobalParameters gets current global parameters
func (s *GameServiceImpl) GetGlobalParameters(ctx context.Context, gameID string) (model.GlobalParameters, error) {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return model.GlobalParameters{}, err
	}
	return game.GlobalParameters, nil
}

// PlayerReconnected handles player reconnection by updating connection status and sending complete game state
func (s *GameServiceImpl) PlayerReconnected(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithContext().With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
	)

	log.Info("🔄 Processing player reconnection")

	// Update player connection status
	err := s.playerRepo.UpdateConnectionStatus(ctx, gameID, playerID, true)
	if err != nil {
		log.Error("Failed to update player connection status", zap.Error(err))
		// Continue with reconnection even if status update fails
	}

	// Use the session manager to broadcast the complete game state to all players
	err = s.sessionManager.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast game state for reconnection", zap.Error(err))
		return fmt.Errorf("failed to broadcast game state: %w", err)
	}

	log.Info("✅ Player reconnection completed successfully")
	return nil
}

// getPlayerName is a helper method to find player name by ID
func (s *GameServiceImpl) getPlayerName(players []model.Player, playerID string) string {
	for _, player := range players {
		if player.ID == playerID {
			return player.Name
		}
	}
	return "Unknown" // Fallback if player not found
}

// resetPlayerActionPlayCounts resets all player action play counts to 0 for a new generation
func (s *GameServiceImpl) resetPlayerActionPlayCounts(ctx context.Context, gameID string) error {
	log := logger.WithGameContext(gameID, "")
	log.Debug("🔄 Resetting player action play counts for new generation")

	// Get all players in the game
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to list players: %w", err)
	}

	// Reset play counts for each player
	for _, player := range players {
		if len(player.Actions) == 0 {
			// Player has no actions, skip
			continue
		}

		// Create a copy of the actions with reset play counts
		resetActions := make([]model.PlayerAction, len(player.Actions))
		for i, action := range player.Actions {
			resetActions[i] = *action.DeepCopy()
			resetActions[i].PlayCount = 0
		}

		// Update the player's actions
		if err := s.playerRepo.UpdatePlayerActions(ctx, gameID, player.ID, resetActions); err != nil {
			log.Error("Failed to reset action play counts for player",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return fmt.Errorf("failed to reset action play counts for player %s: %w", player.ID, err)
		}

		log.Debug("✅ Reset action play counts for player",
			zap.String("player_id", player.ID),
			zap.Int("actions_count", len(resetActions)))
	}

	log.Info("🔄 Successfully reset all player action play counts for new generation",
		zap.Int("players_updated", len(players)))

	return nil
}

// CalculateAvailableOceanHexes returns a list of hex coordinate strings that are available for ocean placement
func (s *GameServiceImpl) CalculateAvailableOceanHexes(ctx context.Context, gameID string) ([]string, error) {
	// Delegate to tiles mechanic
	return s.tilesFeature.CalculateAvailableHexes(ctx, gameID, "", "ocean")
}
