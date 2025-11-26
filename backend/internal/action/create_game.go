package action

import (
	"context"

	"terraforming-mars-backend/internal/session/game/board"
	game "terraforming-mars-backend/internal/session/game/core"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateGameAction handles the business logic for creating new games
type CreateGameAction struct {
	BaseAction // Embed base (no sessionMgr needed for creation)
	gameRepo   game.Repository
	boardRepo  board.Repository
}

// NewCreateGameAction creates a new create game action
func NewCreateGameAction(gameRepo game.Repository, boardRepo board.Repository) *CreateGameAction {
	return &CreateGameAction{
		BaseAction: NewBaseAction(nil), // No sessionFactory or sessionMgr needed
		gameRepo:   gameRepo,
		boardRepo:  boardRepo,
	}
}

// Execute performs the create game action
func (a *CreateGameAction) Execute(ctx context.Context, settings game.GameSettings) (*game.Game, error) {
	log := a.logger.With(
		zap.Int("max_players", settings.MaxPlayers),
		zap.Strings("card_packs", settings.CardPacks),
	)
	log.Info("🎮 Creating new game")

	// 1. Generate game ID
	gameID := uuid.New().String()

	// 2. Apply default settings
	if settings.MaxPlayers == 0 {
		settings.MaxPlayers = game.DefaultMaxPlayers
	}
	if len(settings.CardPacks) == 0 {
		settings.CardPacks = game.DefaultCardPacks()
	}

	// 3. Create game entity
	newGame := game.NewGame(gameID, settings)

	// 4. Initialize empty board on game entity (actual board stored in board repository)
	newGame.Board = board.Board{Tiles: []board.Tile{}}

	// 5. Store game in repository
	err := a.gameRepo.Create(ctx, newGame)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		return nil, err
	}

	// 6. Generate board with 42 tiles via board repository
	err = a.boardRepo.GenerateBoard(ctx)
	if err != nil {
		log.Error("Failed to generate board", zap.Error(err))
		// Note: Rollback not implemented - game repository doesn't support deletion
		// In practice, board generation failures are rare and indicate system issues
		// Failed games will remain in repository but won't be playable without a board
		// Future improvement: Add game status field to mark as "failed" or implement cleanup
		return nil, err
	}

	log.Info("✅ Game created successfully with board", zap.String("game_id", gameID))
	return newGame, nil
}
