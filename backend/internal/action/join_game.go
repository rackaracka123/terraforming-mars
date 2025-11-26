package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/session"
	gameCore "terraforming-mars-backend/internal/session/game/core"

	"go.uber.org/zap"
)

// JoinGameAction handles the business logic for players joining games
// Broadcasting is handled automatically via PlayerJoinedEvent (event-driven architecture)
type JoinGameAction struct {
	BaseAction
	gameRepo       gameCore.Repository
	sessionFactory session.SessionFactory
}

// JoinGameResult contains the result of joining a game
type JoinGameResult struct {
	PlayerID string
	GameDto  dto.GameDto
}

// NewJoinGameAction creates a new join game action
func NewJoinGameAction(
	gameRepo gameCore.Repository,
	sessionFactory session.SessionFactory,
) *JoinGameAction {
	return &JoinGameAction{
		BaseAction:     NewBaseAction(nil), // No sessionMgr (event-driven)
		gameRepo:       gameRepo,
		sessionFactory: sessionFactory,
	}
}

// Execute performs the join game action
// playerID is optional - if empty, a new UUID will be generated
func (a *JoinGameAction) Execute(ctx context.Context, gameID string, playerName string, playerID ...string) (*JoinGameResult, error) {
	var pid string
	if len(playerID) > 0 {
		pid = playerID[0]
	}
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_name", playerName),
	)
	log.Info("🎮 Player joining game")

	// 1. Validate game is in lobby status
	g, err := ValidateLobbyGame(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return nil, err
	}

	// 2. Get or create session
	sess := a.sessionFactory.GetOrCreate(gameID)

	// 3. Check if player with same name already exists (for reconnection/idempotent join)
	existingPlayers := sess.GetAllPlayers()

	// If player with same name exists, return existing playerID (idempotent operation)
	for _, p := range existingPlayers {
		if p.Name() == playerName {
			log.Info("🔄 Player already exists, returning existing ID",
				zap.String("player_id", p.ID()))

			// Return the existing game state
			gameDto := dto.ToGameDtoBasic(*g, dto.GetPaymentConstants())
			return &JoinGameResult{
				PlayerID: p.ID(),
				GameDto:  gameDto,
			}, nil
		}
	}

	// Check max players only for new players
	if len(g.Players) >= g.Settings.MaxPlayers {
		log.Error("Game is full", zap.Int("max_players", g.Settings.MaxPlayers))
		return nil, fmt.Errorf("game is full")
	}

	// 4. Create new player via session
	newPlayer := sess.CreateAndAddPlayer(playerName, pid)

	log.Info("✅ New player created", zap.String("player_id", newPlayer.ID()))

	// 5. Check if this was the first player (after adding to session)
	isFirstPlayer := len(g.Players) == 1

	// 6. Add player to game via repository (event-driven)
	err = a.gameRepo.AddPlayer(ctx, gameID, newPlayer.ID())
	if err != nil {
		log.Error("Failed to add player to game", zap.Error(err))
		return nil, fmt.Errorf("failed to add player to game: %w", err)
	}

	log.Info("✅ Player added to game")

	// 7. If first player, set as host
	if isFirstPlayer {
		err = a.gameRepo.SetHostPlayer(ctx, gameID, newPlayer.ID())
		if err != nil {
			log.Error("Failed to set host player", zap.Error(err))
			// Non-fatal, continue
		} else {
			log.Info("👑 Player set as host")
		}
	}

	// 8. Fetch updated game state
	updatedGame, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game", zap.Error(err))
		return nil, fmt.Errorf("failed to get updated game: %w", err)
	}

	// 8. Convert to DTO
	gameDto := dto.ToGameDtoBasic(*updatedGame, dto.GetPaymentConstants())

	// Note: Broadcasting is now handled automatically via PlayerJoinedEvent
	// gameRepo.AddPlayer() publishes event → SessionManager subscribes → broadcasts automatically

	log.Info("🎉 Player joined game successfully")
	return &JoinGameResult{
		PlayerID: newPlayer.ID(),
		GameDto:  gameDto,
	}, nil
}
