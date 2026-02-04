package connection

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
)

type KickPlayerAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

func NewKickPlayerAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *KickPlayerAction {
	return &KickPlayerAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

func (a *KickPlayerAction) Execute(ctx context.Context, gameID string, requesterID string, targetPlayerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("requester_id", requesterID),
		zap.String("target_player_id", targetPlayerID),
		zap.String("action", "kick_player"),
	)
	log.Info("👢 Kicking player from lobby")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	if g.Status() != game.GameStatusLobby {
		log.Error("Cannot kick player - game not in lobby phase")
		return fmt.Errorf("cannot kick player: game not in lobby phase")
	}

	if g.HostPlayerID() != requesterID {
		log.Error("Cannot kick player - requester is not host")
		return fmt.Errorf("cannot kick player: only host can kick players")
	}

	if requesterID == targetPlayerID {
		log.Error("Cannot kick player - cannot kick yourself")
		return fmt.Errorf("cannot kick player: cannot kick yourself")
	}

	if err := g.RemovePlayer(ctx, targetPlayerID); err != nil {
		log.Error("Failed to remove player from lobby", zap.Error(err))
		return fmt.Errorf("failed to kick player: %w", err)
	}

	remaining := g.GetAllPlayers()
	if len(remaining) == 0 {
		if err := a.gameRepo.Delete(ctx, gameID); err != nil {
			log.Error("Failed to delete empty game", zap.Error(err))
			return fmt.Errorf("failed to delete empty game: %w", err)
		}
		log.Info("🗑️ Game deleted (no players remaining)")
		return nil
	}

	log.Info("✅ Player kicked from lobby")
	return nil
}
