package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/player"

	"go.uber.org/zap"
)

// JoinGameAction handles the business logic for players joining games
// Broadcasting is handled automatically via PlayerJoinedEvent (event-driven architecture)
type JoinGameAction struct {
	gameRepo   game.Repository
	playerRepo player.Repository
	logger     *zap.Logger
}

// NewJoinGameAction creates a new join game action
func NewJoinGameAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
) *JoinGameAction {
	return &JoinGameAction{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
		logger:     logger.Get(),
	}
}

// Execute performs the join game action
// playerID is optional - if empty, a new UUID will be generated
func (a *JoinGameAction) Execute(ctx context.Context, gameID string, playerName string, playerID string) (string, error) {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_name", playerName),
	)
	log.Info("🎮 Player joining game")

	// 1. Validate business rules: game must exist and be in lobby status
	g, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return "", fmt.Errorf("game not found: %w", err)
	}

	if g.Status != game.GameStatusLobby {
		log.Error("Game not in lobby status", zap.String("status", string(g.Status)))
		return "", fmt.Errorf("game not in lobby status, cannot join")
	}

	// 2. Check if player with same name already exists (for reconnection/idempotent join)
	existingPlayers, err := a.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to list existing players", zap.Error(err))
		return "", fmt.Errorf("failed to check existing players: %w", err)
	}

	// If player with same name exists, return existing playerID (idempotent operation)
	for _, p := range existingPlayers {
		if p.Name == playerName {
			log.Info("🔄 Player already exists, returning existing ID",
				zap.String("player_id", p.ID))
			return p.ID, nil
		}
	}

	// Check max players only for new players
	if len(g.PlayerIDs) >= g.Settings.MaxPlayers {
		log.Error("Game is full", zap.Int("max_players", g.Settings.MaxPlayers))
		return "", fmt.Errorf("game is full")
	}

	// 3. Create new player via subdomain repository
	var newPlayer *player.Player
	if playerID != "" {
		// Use provided playerID (for connection setup before event publishing)
		newPlayer = player.NewPlayer(playerName)
		newPlayer.ID = playerID
		log.Debug("Using pre-generated player ID", zap.String("player_id", playerID))
	} else {
		// Generate new playerID
		newPlayer = player.NewPlayer(playerName)
	}

	err = a.playerRepo.Create(ctx, gameID, newPlayer)
	if err != nil {
		log.Error("Failed to create player", zap.Error(err))
		return "", fmt.Errorf("failed to create player: %w", err)
	}

	log.Info("✅ New player created", zap.String("player_id", newPlayer.ID))

	// 4. Check if this will be the first player (before adding to game)
	isFirstPlayer := len(g.PlayerIDs) == 0

	// 5. Add player to game via repository (event-driven)
	err = a.gameRepo.AddPlayer(ctx, gameID, newPlayer.ID)
	if err != nil {
		log.Error("Failed to add player to game", zap.Error(err))
		return "", fmt.Errorf("failed to add player to game: %w", err)
	}

	log.Info("✅ Player added to game")

	// 6. If first player, set as host
	if isFirstPlayer {
		err = a.gameRepo.SetHostPlayer(ctx, gameID, newPlayer.ID)
		if err != nil {
			log.Error("Failed to set host player", zap.Error(err))
			// Non-fatal, continue
		} else {
			log.Info("👑 Player set as host")
		}
	}

	// Note: Broadcasting is now handled automatically via PlayerJoinedEvent
	// gameRepo.AddPlayer() publishes event → SessionManager subscribes → broadcasts automatically

	log.Info("🎉 Player joined game successfully")
	return newPlayer.ID, nil
}
