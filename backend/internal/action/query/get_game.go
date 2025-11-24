package query

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	sessionCard "terraforming-mars-backend/internal/session/card"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// GetGameAction handles the query for getting a single game
type GetGameAction struct {
	action.BaseAction
	cardRepo sessionCard.Repository
}

// NewGetGameAction creates a new get game query action
func NewGetGameAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	cardRepo sessionCard.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *GetGameAction {
	return &GetGameAction{
		BaseAction: action.NewBaseAction(gameRepo, playerRepo, sessionMgrFactory),
		cardRepo:   cardRepo,
	}
}

// GameQueryResult contains the full game data for queries
type GameQueryResult struct {
	Game          *game.Game
	Players       []types.Player
	ResolvedCards map[string]types.Card
}

// Execute performs the get game query
func (a *GetGameAction) Execute(ctx context.Context, gameID, playerID string) (*GameQueryResult, error) {
	log := a.InitLogger(gameID, playerID)
	log.Info("🔍 Querying game")

	// 1. Validate game exists
	gameEntity, err := action.ValidateGameExists(ctx, a.GetGameRepo(), gameID, log)
	if err != nil {
		return nil, err
	}

	// If no playerID specified, return basic game info without players
	if playerID == "" {
		log.Info("✅ Game query completed (basic)")
		return &GameQueryResult{
			Game:          gameEntity,
			Players:       nil,
			ResolvedCards: nil,
		}, nil
	}

	// 2. Get all players
	playerPointers, err := a.GetPlayerRepo().ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players", zap.Error(err))
		return nil, err
	}

	// Convert player pointers to values
	players := make([]types.Player, len(playerPointers))
	for i, playerPtr := range playerPointers {
		players[i] = *playerPtr
	}

	// 3. Collect all card IDs that need resolution
	allCardIds := make(map[string]struct{})
	for _, player := range players {
		if player.Corporation != nil {
			allCardIds[player.Corporation.ID] = struct{}{}
		}
		for _, cardID := range player.PlayedCards {
			allCardIds[cardID] = struct{}{}
		}
		for _, cardID := range player.Cards {
			allCardIds[cardID] = struct{}{}
		}
	}

	// 4. Resolve cards
	resolvedCards, err := a.cardRepo.ListCardsByIdMap(ctx, allCardIds)
	if err != nil {
		log.Error("Failed to resolve cards", zap.Error(err))
		return nil, err
	}

	log.Info("✅ Game query completed (full)",
		zap.Int("player_count", len(players)),
		zap.Int("resolved_cards", len(resolvedCards)))

	return &GameQueryResult{
		Game:          gameEntity,
		Players:       players,
		ResolvedCards: resolvedCards,
	}, nil
}
