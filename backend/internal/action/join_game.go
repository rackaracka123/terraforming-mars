package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/player"

	"go.uber.org/zap"
)

// JoinGameAction handles the business logic for players joining games
type JoinGameAction struct {
	gameRepo   game.Repository
	playerRepo player.Repository
	sessionMgr session.SessionManager
	logger     *zap.Logger
}

// NewJoinGameAction creates a new join game action
func NewJoinGameAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgr session.SessionManager,
) *JoinGameAction {
	return &JoinGameAction{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
		sessionMgr: sessionMgr,
		logger:     logger.Get(),
	}
}

// Execute performs the join game action
func (a *JoinGameAction) Execute(ctx context.Context, gameID string, playerName string) (string, error) {
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

	// Check max players
	if len(g.PlayerIDs) >= g.Settings.MaxPlayers {
		log.Error("Game is full", zap.Int("max_players", g.Settings.MaxPlayers))
		return "", fmt.Errorf("game is full")
	}

	// 2. Create player via subdomain repository
	newPlayer := player.NewPlayer(playerName)
	err = a.playerRepo.Create(ctx, gameID, newPlayer)
	if err != nil {
		log.Error("Failed to create player", zap.Error(err))
		return "", fmt.Errorf("failed to create player: %w", err)
	}

	log.Info("✅ Player created", zap.String("player_id", newPlayer.ID))

	// 3. Add player to game via repository (event-driven)
	err = a.gameRepo.AddPlayer(ctx, gameID, newPlayer.ID)
	if err != nil {
		log.Error("Failed to add player to game", zap.Error(err))
		return "", fmt.Errorf("failed to add player to game: %w", err)
	}

	log.Info("✅ Player added to game")

	// 4. If first player, set as host
	if len(g.PlayerIDs) == 0 { // Was empty before we added
		err = a.gameRepo.SetHostPlayer(ctx, gameID, newPlayer.ID)
		if err != nil {
			log.Error("Failed to set host player", zap.Error(err))
			// Non-fatal, continue
		} else {
			log.Info("👑 Player set as host")
		}
	}

	// 5. Broadcast state via session manager
	err = a.sessionMgr.Broadcast(gameID)
	if err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Non-fatal, player was created successfully
	}

	log.Info("🎉 Player joined game successfully")
	return newPlayer.ID, nil
}
