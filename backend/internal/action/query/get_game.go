package query

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	gamePackage "terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/card"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/game/player"

	"go.uber.org/zap"
)

// GetGameAction handles the query for getting a single game
type GetGameAction struct {
	action.BaseAction
	gameRepo game.Repository
	cardRepo card.Repository
}

// NewGetGameAction creates a new get game query action
func NewGetGameAction(
	gameRepo game.Repository,
	cardRepo card.Repository,
) *GetGameAction {
	return &GetGameAction{
		BaseAction: action.NewBaseAction(nil),
		gameRepo:   gameRepo,
		cardRepo:   cardRepo,
	}
}

// GameQueryResult contains the full game data for queries
type GameQueryResult struct {
	Game          *gamePackage.Game
	Players       []*player.Player
	ResolvedCards map[string]card.Card
}

// Execute performs the get game query
func (a *GetGameAction) Execute(ctx context.Context, sess *session.Session, playerID string) (*GameQueryResult, error) {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID)
	log.Info("🔍 Querying game")

	// 1. Validate game exists
	gameEntity, err := action.ValidateGameExists(ctx, a.gameRepo, gameID, log)
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

	// 2. Get all players from session
	players := sess.GetAllPlayers()

	// 3. Collect all card IDs that need resolution
	allCardIds := make(map[string]struct{})
	for _, p := range players {
		if p.Corp().HasCorporation() {
			corp := p.Corp().Card()
			allCardIds[corp.ID] = struct{}{}
		}
		for _, cardID := range p.Hand().PlayedCards() {
			allCardIds[cardID] = struct{}{}
		}
		for _, cardID := range p.Hand().Cards() {
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
