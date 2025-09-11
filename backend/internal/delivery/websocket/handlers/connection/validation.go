package connection

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Validator handles validation logic for connections
type Validator struct {
	gameService   service.GameService
	playerService service.PlayerService
	parser        *utils.MessageParser
	errorHandler  *utils.ErrorHandler
	logger        *zap.Logger
}

// NewValidator creates a new validator
func NewValidator(
	gameService service.GameService,
	playerService service.PlayerService,
	parser *utils.MessageParser,
	errorHandler *utils.ErrorHandler,
	logger *zap.Logger,
) *Validator {
	return &Validator{
		gameService:   gameService,
		playerService: playerService,
		parser:        parser,
		errorHandler:  errorHandler,
		logger:        logger,
	}
}

// ValidateGameExists validates that a game exists
func (v *Validator) ValidateGameExists(ctx context.Context, connection *core.Connection, gameID string) bool {
	_, err := v.gameService.GetGame(ctx, gameID)
	if err != nil {
		v.logger.Error("Failed to get game for WebSocket connection", zap.Error(err))
		v.errorHandler.SendError(connection, utils.ErrGameNotFound)
		return false
	}
	return true
}

// FindExistingPlayer finds an existing player by name in the game
func (v *Validator) FindExistingPlayer(ctx context.Context, gameID, playerName string) string {
	player, err := v.playerService.GetPlayerByName(ctx, gameID, playerName)
	if err != nil {
		return ""
	}
	return player.ID
}

// ParseAndValidateConnectPayload parses and validates the connection payload
func (v *Validator) ParseAndValidateConnectPayload(connection *core.Connection, message dto.WebSocketMessage, payload *dto.PlayerConnectPayload) bool {
	if err := v.parser.ParsePayload(message.Payload, payload); err != nil {
		v.logger.Error("❌ Failed to parse player connect payload",
			zap.Error(err), zap.String("connection_id", connection.ID))
		v.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return false
	}

	v.logger.Debug("✅ Payload parsed successfully",
		zap.String("game_id", payload.GameID),
		zap.String("player_name", payload.PlayerName))
	return true
}

