package query

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// GetGameAction handles the query for getting a single game
type GetGameAction struct {
	action.BaseAction
	oldPlayerRepo repository.PlayerRepository
	cardRepo      repository.CardRepository
}

// NewGetGameAction creates a new get game query action
func NewGetGameAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	oldPlayerRepo repository.PlayerRepository,
	cardRepo repository.CardRepository,
	sessionMgr session.SessionManager,
) *GetGameAction {
	return &GetGameAction{
		BaseAction:    action.NewBaseAction(gameRepo, playerRepo, sessionMgr),
		oldPlayerRepo: oldPlayerRepo,
		cardRepo:      cardRepo,
	}
}

// GameQueryResult contains the full game data for queries
type GameQueryResult struct {
	Game          *game.Game
	Players       []model.Player
	ResolvedCards map[string]model.Card
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
	players, err := a.oldPlayerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players", zap.Error(err))
		return nil, err
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
