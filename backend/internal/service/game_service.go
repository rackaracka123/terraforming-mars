package service

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// GameService handles active game operations including turn management, phases, and player actions
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
	gameRepo       repository.GameRepository
	playerRepo     repository.PlayerRepository
	cardRepo       repository.CardRepository
	cardService    CardService
	cardDeckRepo   repository.CardDeckRepository
	boardService   BoardService
	sessionManager session.SessionManager
}

// NewGameService creates a new GameService instance
func NewGameService(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	cardRepo repository.CardRepository,
	cardService CardService,
	cardDeckRepo repository.CardDeckRepository,
	boardService BoardService,
	sessionManager session.SessionManager,
) GameService {
	return &GameServiceImpl{
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		cardRepo:       cardRepo,
		cardService:    cardService,
		cardDeckRepo:   cardDeckRepo,
		boardService:   boardService,
		sessionManager: sessionManager,
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

	// Get current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for skip turn", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game is active
	if game.Status != model.GameStatusActive {
		log.Warn("Attempted to skip turn in non-active game", zap.String("current_status", string(game.Status)))
		return fmt.Errorf("game is not active")
	}

	if game.CurrentTurn == nil {
		log.Warn("Attempted to skip turn but current turn is not set")
		return fmt.Errorf("current turn is not set")
	}

	// Validate requesting player is the current player
	if game.CurrentTurn != nil && *game.CurrentTurn != playerID {
		log.Warn("Non-current player attempted to skip turn",
			zap.String("current_player", *game.CurrentTurn),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("only the current player can skip their turn")
	}

	// Find current player and determine SKIP vs PASS behavior
	currentPlayerIndex := -1
	for i, id := range game.PlayerIDs {
		if id == playerID {
			currentPlayerIndex = i
			break
		}
	}

	if currentPlayerIndex == -1 {
		log.Error("Current player not found in game", zap.String("player_id", playerID))
		return fmt.Errorf("player not found in game")
	}

	gamePlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players for skip turn", zap.Error(err))
	}

	// PASS vs SKIP logic based on available actions
	var currentPlayer *model.Player = nil
	for i := range gamePlayers {
		if gamePlayers[i].ID == playerID {
			currentPlayer = &gamePlayers[i]
		}
	}
	if currentPlayer == nil {
		log.Error("Current player data not found", zap.String("player_id", playerID))
		return fmt.Errorf("player data not found")
	}

	// Check how many players are still active (not passed and have actions)
	activePlayerCount := 0
	for _, player := range gamePlayers {
		if !player.Passed {
			activePlayerCount++
		}
	}

	isPassing := currentPlayer.AvailableActions == 2 || currentPlayer.AvailableActions == -1
	if isPassing {
		// PASS: Player hasn't done any actions or has unlimited actions, mark as passed for generation end check
		err = s.playerRepo.UpdatePassed(ctx, gameID, playerID, true)
		if err != nil {
			log.Error("Failed to mark player as passed", zap.Error(err))
			return fmt.Errorf("failed to update player passed status: %w", err)
		}

		log.Debug("Player PASSED (marked as passed for generation)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", currentPlayer.AvailableActions))

		// If this means only one active player remains, grant them unlimited actions
		if activePlayerCount == 2 {
			for _, player := range gamePlayers {
				if !player.Passed && player.ID != playerID {
					err = s.playerRepo.UpdateAvailableActions(ctx, gameID, player.ID, -1)
					if err != nil {
						log.Error("Failed to grant unlimited actions to last active player", zap.String("player_id", player.ID), zap.Error(err))
						return fmt.Errorf("failed to update last active player's actions: %w", err)
					} else {
						log.Info("🏃 Last active player granted unlimited actions due to others passing", zap.String("player_id", player.ID))
					}
				}
			}
		}

	} else {
		// SKIP: Player has done some actions, just advance turn without passing
		log.Debug("Player SKIPPED (turn advanced, not passed)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", currentPlayer.AvailableActions))
	}

	// List all players again to reflect any status changes
	gamePlayers, err = s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players after skip processing", zap.Error(err))
	}

	// Check if all players have exhausted their actions or formally passed
	allPlayersFinished := true
	passedCount := 0
	playersWithNoActions := 0
	for _, player := range gamePlayers {
		if player.Passed {
			passedCount++
		} else if player.AvailableActions == 0 {
			playersWithNoActions++
		} else if player.AvailableActions > 0 || player.AvailableActions == -1 {
			allPlayersFinished = false
		}
	}

	log.Debug("Checking generation end condition",
		zap.Int("passed_count", passedCount),
		zap.Int("players_with_no_actions", playersWithNoActions),
		zap.Int("total_players", len(gamePlayers)),
		zap.Bool("all_players_finished", allPlayersFinished))

	if allPlayersFinished {
		// All players have either passed or have no actions left - production phase should start
		log.Info("🏭 All players finished their turns - generation ending",
			zap.String("game_id", gameID),
			zap.Int("generation", game.Generation),
			zap.Int("passed_players", passedCount),
			zap.Int("players_with_no_actions", playersWithNoActions))

		_, err = s.executeProductionPhase(ctx, gameID)
		if err != nil {
			log.Error("Failed to execute production phase", zap.Error(err))
			return fmt.Errorf("failed to execute production phase: %w", err)
		}

		// Broadcast updated game state to all players
		if err := s.sessionManager.Broadcast(gameID); err != nil {
			log.Error("Failed to broadcast game state after production phase start", zap.Error(err))
			// Don't fail the skip operation, just log the error
		}

		return nil
	}

	// Find next player who still has actions available
	nextPlayerIndex := (currentPlayerIndex + 1) % len(gamePlayers)

	// Cycle through players to find the next player with available actions
	for i := 0; i < len(gamePlayers); i++ {
		nextPlayer := &gamePlayers[nextPlayerIndex]
		if !nextPlayer.Passed {
			break
		}
		nextPlayerIndex = (nextPlayerIndex + 1) % len(gamePlayers)
	}

	game.CurrentTurn = &gamePlayers[nextPlayerIndex].ID

	// Update game through repository
	if err := s.gameRepo.UpdateCurrentTurn(ctx, game.ID, game.CurrentTurn); err != nil {
		log.Error("Failed to update game after skip turn", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Player turn skipped, advanced to next player",
		zap.String("previous_player", playerID),
		zap.String("current_player", *game.CurrentTurn))

	// Broadcast updated game state to all players
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after skip turn", zap.Error(err))
		// Don't fail the skip operation, just log the error
	}

	return nil
}

// executeProductionPhase updates all players' resources based on their production and advances generation (internal use only)
func (s *GameServiceImpl) executeProductionPhase(ctx context.Context, gameID string) (*model.Game, error) {
	log := logger.WithGameContext(gameID, "")
	log.Debug("Executing production phase via GameService")

	// Get current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for production phase", zap.Error(err))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game is in correct phase
	if game.Status != model.GameStatusActive {
		log.Warn("Attempted to execute production phase in non-active game", zap.String("current_status", string(game.Status)))
		return nil, fmt.Errorf("game is not active")
	}

	gamePlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players for production phase", zap.Error(err))
		return nil, fmt.Errorf("failed to list players: %w", err)
	}

	// Convert energy to heat and apply production to all players using repository methods
	isSolo := len(gamePlayers) == 1
	resetTurns := 2
	if isSolo {
		resetTurns = -1 // Unlimited actions for solo play
	}

	for i := range gamePlayers {
		player := &gamePlayers[i]

		energyConverted := player.Resources.Energy
		oldResources := player.Resources.DeepCopy()
		newResources := model.Resources{
			Credits:  player.Resources.Credits + player.Production.Credits + player.TerraformRating, // TR provides 1 credit per point
			Steel:    player.Resources.Steel + player.Production.Steel,
			Titanium: player.Resources.Titanium + player.Production.Titanium,
			Plants:   player.Resources.Plants + player.Production.Plants,
			Energy:   player.Production.Energy,
			Heat:     player.Resources.Heat + energyConverted + player.Production.Heat,
		}

		// Update player resources through repository (follows clean architecture)
		if err := s.playerRepo.UpdateResources(ctx, gameID, player.ID, newResources); err != nil {
			log.Error("Failed to update player resources during production",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to update player resources: %w", err)
		}

		// Reset player state for next generation
		if err := s.playerRepo.UpdatePassed(ctx, gameID, player.ID, false); err != nil {
			log.Error("Failed to reset player passed status",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to reset player passed status: %w", err)
		}

		if err := s.playerRepo.UpdateAvailableActions(ctx, gameID, player.ID, resetTurns); err != nil {
			log.Error("Failed to reset player available actions",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to reset player available actions: %w", err)
		}

		// Draw 4 cards from deck for production phase card selection
		drawnCards := []string{}
		for i := 0; i < 4; i++ {
			cardID, err := s.cardDeckRepo.Pop(ctx, gameID)
			if err != nil {
				log.Warn("Failed to draw card from deck for production phase",
					zap.String("player_id", player.ID),
					zap.Int("drawn_so_far", i),
					zap.Error(err))
				// If we can't draw more cards, continue with whatever we have
				break
			}
			drawnCards = append(drawnCards, cardID)
		}

		log.Debug("Drew cards for production phase",
			zap.String("player_id", player.ID),
			zap.Strings("drawn_cards", drawnCards),
			zap.Int("cards_drawn", len(drawnCards)))

		// Set their production phase data
		productionPhaseData := model.ProductionPhase{
			AvailableCards:    drawnCards,
			SelectionComplete: false,
			BeforeResources:   oldResources,
			AfterResources:    newResources,
			EnergyConverted:   energyConverted,
			CreditsIncome:     player.Production.Credits + player.TerraformRating,
		}
		if err := s.playerRepo.UpdateProductionPhase(ctx, gameID, player.ID, &productionPhaseData); err != nil {
			log.Error("Failed to set player production phase data",
				zap.String("player_id", player.ID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to set player production phase data: %w", err)
		}

		log.Debug("Applied production to player",
			zap.String("player_id", player.ID),
			zap.Int("energy_converted", energyConverted),
			zap.Int("credits_gained", player.Production.Credits+player.TerraformRating))
	}

	// Update game through repository
	if err := s.gameRepo.UpdateGeneration(ctx, game.ID, game.Generation+1); err != nil {
		log.Error("Failed to update game generation", zap.Error(err))
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	// Update the current turn to the first player for the new generation
	if len(game.PlayerIDs) > 0 {
		if err := s.gameRepo.UpdateCurrentTurn(ctx, game.ID, &game.PlayerIDs[0]); err != nil {
			log.Error("Failed to update current turn for new generation", zap.Error(err))
			return nil, fmt.Errorf("failed to update current turn: %w", err)
		}
	}

	if err := s.gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseProductionAndCardDraw); err != nil {
		log.Error("Failed to update game phase to action", zap.Error(err))
		return nil, fmt.Errorf("failed to update game phase: %w", err)
	}

	game, err = s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game after production phase", zap.Error(err))
		return nil, fmt.Errorf("failed to get updated game: %w", err)
	}

	log.Info("Production phase start",
		zap.String("game_id", gameID),
		zap.Int("generation", game.Generation),
		zap.Int("player_count", len(gamePlayers)))

	return &game, nil
}

// ProcessProductionPhaseReady processes a player's ready acknowledgment and transitions phase when all players are ready
func (s *GameServiceImpl) ProcessProductionPhaseReady(ctx context.Context, gameID string, playerID string) (*model.Game, error) {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Processing production phase ready via GameService")

	// Get current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for production ready", zap.Error(err))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game state
	if game.Status != model.GameStatusActive {
		log.Warn("Attempted to mark ready in non-active game", zap.String("current_status", string(game.Status)))
		return nil, fmt.Errorf("game is not active")
	}

	if game.CurrentPhase != model.GamePhaseProductionAndCardDraw {
		log.Warn("Attempted to mark ready in non-production phase", zap.String("current_phase", string(game.CurrentPhase)))
		return nil, fmt.Errorf("game is not in production phase")
	}

	// Check if all players are ready
	updatedPlayers, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list players to check readiness", zap.Error(err))
		return nil, fmt.Errorf("failed to list players: %w", err)
	}

	allReady := true
	readyCount := 0
	for _, player := range updatedPlayers {
		if player.ProductionPhase.SelectionComplete {
			readyCount++
		} else {
			allReady = false
		}
	}

	log.Debug("Checking production phase readiness",
		zap.Int("ready_count", readyCount),
		zap.Int("total_players", len(updatedPlayers)),
		zap.Bool("all_ready", allReady))

	if allReady {
		// All players are ready - advance to next phase (action phase)
		log.Info("🎯 All players ready for next phase - advancing from production to action",
			zap.String("game_id", gameID),
			zap.Int("generation", game.Generation))

		// Set first player's turn for new generation
		if len(updatedPlayers) > 0 {
			firstPlayerID := updatedPlayers[0].ID
			if err := s.gameRepo.SetCurrentTurn(ctx, gameID, &firstPlayerID); err != nil {
				log.Error("Failed to set current turn for new generation", zap.Error(err))
				return nil, fmt.Errorf("failed to set current turn: %w", err)
			}
		}

		// Reset all player action play counts for the new generation
		if err := s.resetPlayerActionPlayCounts(ctx, gameID); err != nil {
			log.Error("Failed to reset player action play counts", zap.Error(err))
			return nil, fmt.Errorf("failed to reset action play counts: %w", err)
		}

		// Clear production phase data for all players
		for _, player := range updatedPlayers {
			if err := s.playerRepo.UpdateProductionPhase(ctx, gameID, player.ID, nil); err != nil {
				log.Error("Failed to clear player production phase data",
					zap.String("player_id", player.ID),
					zap.Error(err))
				return nil, fmt.Errorf("failed to clear production phase data for player %s: %w", player.ID, err)
			}
		}

		// Advance to action phase
		if err := s.gameRepo.UpdatePhase(ctx, gameID, model.GamePhaseAction); err != nil {
			log.Error("Failed to advance game phase to action", zap.Error(err))
			return nil, fmt.Errorf("failed to advance game phase: %w", err)
		}

		log.Info("Production phase completed, advanced to action phase",
			zap.String("game_id", gameID),
			zap.Int("generation", game.Generation))
	}

	game, err = s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game after production ready", zap.Error(err))
		return nil, fmt.Errorf("failed to get updated game: %w", err)
	}

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
	log := logger.WithGameContext(gameID, "")

	// Get the current game state
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Count actual oceans placed on board (board is source of truth)
	oceansPlaced := 0
	availableHexes := make([]string, 0)

	for _, tile := range game.Board.Tiles {
		if tile.Type == model.ResourceOceanTile {
			if tile.OccupiedBy != nil && tile.OccupiedBy.Type == model.ResourceOceanTile {
				// This ocean space is occupied
				oceansPlaced++
			} else {
				// This ocean space is available for placement
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}
		}
	}

	// Check if we've reached maximum oceans based on actual board state
	if oceansPlaced >= model.MaxOceans {
		log.Debug("🌊 Maximum oceans reached, no more ocean placement allowed",
			zap.Int("oceans_placed", oceansPlaced),
			zap.Int("max_oceans", model.MaxOceans))
		return []string{}, nil
	}

	log.Debug("🌊 Ocean hex availability calculated",
		zap.Int("available_hexes", len(availableHexes)),
		zap.Int("oceans_placed", oceansPlaced),
		zap.Int("max_oceans", model.MaxOceans))

	return availableHexes, nil
}
