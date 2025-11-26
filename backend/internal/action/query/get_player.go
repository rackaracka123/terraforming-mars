package query

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/game/player"
)

// GetPlayerAction handles the query for getting a single player
type GetPlayerAction struct {
	action.BaseAction
	gameRepo game.Repository
}

// NewGetPlayerAction creates a new get player query action
func NewGetPlayerAction(
	gameRepo game.Repository,
) *GetPlayerAction {
	return &GetPlayerAction{
		BaseAction: action.NewBaseAction(nil),
		gameRepo:   gameRepo,
	}
}

// Execute performs the get player query
func (a *GetPlayerAction) Execute(ctx context.Context, sess *session.Session, playerID string) (*player.Player, error) {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID)
	log.Info("🔍 Querying player")

	// 1. Validate game exists
	_, err := action.ValidateGameExists(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return nil, err
	}

	// 2. Get player from session
	p, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return nil, err
	}

	log.Info("✅ Player query completed")

	// Return the player
	return p, nil
}
