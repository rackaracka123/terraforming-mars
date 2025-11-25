package query

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/types"
)

// GetPlayerAction handles the query for getting a single player
type GetPlayerAction struct {
	action.BaseAction
	gameRepo game.Repository
}

// NewGetPlayerAction creates a new get player query action
func NewGetPlayerAction(
	gameRepo game.Repository,
	sessionFactory session.SessionFactory,
) *GetPlayerAction {
	return &GetPlayerAction{
		BaseAction: action.NewBaseAction(sessionFactory, nil),
		gameRepo:   gameRepo,
	}
}

// Execute performs the get player query
func (a *GetPlayerAction) Execute(ctx context.Context, gameID, playerID string) (types.Player, error) {
	log := a.InitLogger(gameID, playerID)
	log.Info("🔍 Querying player")

	// 1. Validate game exists
	_, err := action.ValidateGameExists(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return types.Player{}, err
	}

	// 2. Get player from session
	sess := a.GetSessionFactory().Get(gameID)
	if sess == nil {
		log.Error("Game session not found")
		return types.Player{}, err
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return types.Player{}, err
	}

	log.Info("✅ Player query completed")

	// Return the underlying player
	return *player.Player, nil
}
