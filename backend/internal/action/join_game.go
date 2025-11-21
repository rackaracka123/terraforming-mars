package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// JoinGameAction handles the business logic for players joining games
// Broadcasting is handled automatically via PlayerJoinedEvent (event-driven architecture)
type JoinGameAction struct {
	BaseAction // Embed base (note: no sessionMgr needed, event-driven)
}

// JoinGameResult contains the result of joining a game
type JoinGameResult struct {
	PlayerID string
	GameDto  dto.GameDto
}

// NewJoinGameAction creates a new join game action
func NewJoinGameAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
) *JoinGameAction {
	return &JoinGameAction{
		// Pass nil for sessionMgr since this action uses event-driven broadcasting
		BaseAction: NewBaseAction(gameRepo, playerRepo, nil),
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

	// 2. Check if player with same name already exists (for reconnection/idempotent join)
	existingPlayers, err := GetAllPlayers(ctx, a.playerRepo, gameID, log)
	if err != nil {
		return nil, err
	}

	// If player with same name exists, return existing playerID (idempotent operation)
	for _, p := range existingPlayers {
		if p.Name == playerName {
			log.Info("🔄 Player already exists, returning existing ID",
				zap.String("player_id", p.ID))

			// Return the existing game state
			gameDto := dto.ToGameDtoBasic(convertToModelGame(g), dto.GetPaymentConstants())
			return &JoinGameResult{
				PlayerID: p.ID,
				GameDto:  gameDto,
			}, nil
		}
	}

	// Check max players only for new players
	if len(g.PlayerIDs) >= g.Settings.MaxPlayers {
		log.Error("Game is full", zap.Int("max_players", g.Settings.MaxPlayers))
		return nil, fmt.Errorf("game is full")
	}

	// 3. Create new player via subdomain repository
	var newPlayer *player.Player
	if pid != "" {
		// Use provided playerID (for connection setup before event publishing)
		newPlayer = player.NewPlayer(playerName)
		newPlayer.ID = pid
		log.Debug("Using pre-generated player ID", zap.String("player_id", pid))
	} else {
		// Generate new playerID
		newPlayer = player.NewPlayer(playerName)
	}

	err = a.playerRepo.Create(ctx, gameID, newPlayer)
	if err != nil {
		log.Error("Failed to create player", zap.Error(err))
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	log.Info("✅ New player created", zap.String("player_id", newPlayer.ID))

	// 4. Check if this will be the first player (before adding to game)
	isFirstPlayer := len(g.PlayerIDs) == 0

	// 5. Add player to game via repository (event-driven)
	err = a.gameRepo.AddPlayer(ctx, gameID, newPlayer.ID)
	if err != nil {
		log.Error("Failed to add player to game", zap.Error(err))
		return nil, fmt.Errorf("failed to add player to game: %w", err)
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

	// 7. Fetch updated game state
	updatedGame, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game", zap.Error(err))
		return nil, fmt.Errorf("failed to get updated game: %w", err)
	}

	// 8. Convert to DTO
	gameDto := dto.ToGameDtoBasic(convertToModelGame(updatedGame), dto.GetPaymentConstants())

	// Note: Broadcasting is now handled automatically via PlayerJoinedEvent
	// gameRepo.AddPlayer() publishes event → SessionManager subscribes → broadcasts automatically

	log.Info("🎉 Player joined game successfully")
	return &JoinGameResult{
		PlayerID: newPlayer.ID,
		GameDto:  gameDto,
	}, nil
}

// convertToModelGame converts a game.Game to model.Game
func convertToModelGame(g *game.Game) model.Game {
	return model.Game{
		ID:        g.ID,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
		Status:    model.GameStatus(g.Status),
		Settings: model.GameSettings{
			MaxPlayers:      g.Settings.MaxPlayers,
			Temperature:     g.Settings.Temperature,
			Oxygen:          g.Settings.Oxygen,
			Oceans:          g.Settings.Oceans,
			DevelopmentMode: g.Settings.DevelopmentMode,
			CardPacks:       g.Settings.CardPacks,
		},
		PlayerIDs:        g.PlayerIDs,
		HostPlayerID:     g.HostPlayerID,
		CurrentPhase:     model.GamePhase(g.CurrentPhase),
		GlobalParameters: g.GlobalParameters,
		ViewingPlayerID:  g.ViewingPlayerID,
		CurrentTurn:      g.CurrentTurn,
		Generation:       g.Generation,
		Board:            g.Board,
	}
}
