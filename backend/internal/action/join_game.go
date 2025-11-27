package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/game/player"

	"go.uber.org/zap"
)

// JoinGameAction handles players joining games
// New architecture: Uses only GameRepository + logger, events handle broadcasting
type JoinGameAction struct {
	gameRepo game.GameRepository
	eventBus *events.EventBusImpl
	logger   *zap.Logger
}

// JoinGameResult contains the result of joining a game
type JoinGameResult struct {
	PlayerID string
	GameDto  dto.GameDto
}

// NewJoinGameAction creates a new join game action
func NewJoinGameAction(
	gameRepo game.GameRepository,
	eventBus *events.EventBusImpl,
	logger *zap.Logger,
) *JoinGameAction {
	return &JoinGameAction{
		gameRepo: gameRepo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Execute performs the join game action
// playerID is optional - if empty, a new UUID will be generated
func (a *JoinGameAction) Execute(
	ctx context.Context,
	gameID string,
	playerName string,
	playerID ...string,
) (*JoinGameResult, error) {
	var pid string
	if len(playerID) > 0 {
		pid = playerID[0]
	}

	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_name", playerName),
	)
	log.Info("🎮 Player joining game")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return nil, fmt.Errorf("game not found: %w", err)
	}

	// 2. Validate game is in lobby status
	if g.Status() != game.GameStatusLobby {
		log.Warn("Game is not in lobby", zap.String("status", string(g.Status())))
		return nil, fmt.Errorf("game is not in lobby: %s", g.Status())
	}

	// 3. Check if player with same name already exists (idempotent join)
	existingPlayers := g.GetAllPlayers()
	for _, p := range existingPlayers {
		if p.Name() == playerName {
			log.Info("🔄 Player already exists, returning existing ID",
				zap.String("player_id", p.ID()))

			// Return the existing game state
			gameDto := a.createGameDto(g)
			return &JoinGameResult{
				PlayerID: p.ID(),
				GameDto:  gameDto,
			}, nil
		}
	}

	// 4. Check max players only for new players
	maxPlayers := g.Settings().MaxPlayers
	if maxPlayers == 0 {
		maxPlayers = game.DefaultMaxPlayers
	}
	if len(existingPlayers) >= maxPlayers {
		log.Error("Game is full", zap.Int("max_players", maxPlayers))
		return nil, fmt.Errorf("game is full")
	}

	// 5. Create new player
	newPlayer := playerPkg.NewPlayer(a.eventBus, gameID, pid, playerName)
	log.Info("✅ New player created", zap.String("player_id", newPlayer.ID()))

	// 6. Check if this will be the first player (before adding)
	isFirstPlayer := len(existingPlayers) == 0

	// 7. Add player to game (publishes PlayerJoinedEvent)
	err = g.AddPlayer(ctx, newPlayer)
	if err != nil {
		log.Error("Failed to add player to game", zap.Error(err))
		return nil, fmt.Errorf("failed to add player to game: %w", err)
	}

	log.Info("✅ Player added to game")

	// 8. If first player, set as host
	if isFirstPlayer {
		err = g.SetHostPlayerID(ctx, newPlayer.ID())
		if err != nil {
			log.Error("Failed to set host player", zap.Error(err))
			// Non-fatal, continue
		} else {
			log.Info("👑 Player set as host")
		}
	}

	// 9. Convert to DTO
	gameDto := a.createGameDto(g)

	// Note: Broadcasting handled automatically via PlayerJoinedEvent
	// g.AddPlayer() publishes event → SessionManager subscribes → broadcasts

	log.Info("🎉 Player joined game successfully")
	return &JoinGameResult{
		PlayerID: newPlayer.ID(),
		GameDto:  gameDto,
	}, nil
}

// createGameDto creates a basic game DTO for join responses
// TODO: This is a temporary implementation - should use proper DTO mapper
func (a *JoinGameAction) createGameDto(g *game.Game) dto.GameDto {
	// For now, return a minimal DTO with just the essentials
	// This will be replaced with proper DTO mapping once we fully migrate
	return dto.GameDto{
		ID:           g.ID(),
		Status:       dto.GameStatus(g.Status()),
		HostPlayerID: g.HostPlayerID(),
		CurrentPhase: dto.GamePhase(g.CurrentPhase()),
		Generation:   g.Generation(),
		GlobalParameters: dto.GlobalParametersDto{
			Temperature: g.GlobalParameters().Temperature(),
			Oxygen:      g.GlobalParameters().Oxygen(),
			Oceans:      g.GlobalParameters().Oceans(),
		},
		// TODO: Add full player data, board, etc. when DTO mapper is migrated
		PaymentConstants: dto.PaymentConstantsDto{
			SteelValue:    2, // Default steel value
			TitaniumValue: 3, // Default titanium value
		},
	}
}
