package game

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// CreateGameHandler handles create game requests using the migrated architecture
type CreateGameHandler struct {
	createGameAction *action.CreateGameAction
	logger           *zap.Logger
}

// NewCreateGameHandler creates a new create game handler for migrated actions
func NewCreateGameHandler(createGameAction *action.CreateGameAction) *CreateGameHandler {
	return &CreateGameHandler{
		createGameAction: createGameAction,
		logger:           logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *CreateGameHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🎮 Processing create game request (migrated)")

	// Parse payload - for now using default settings
	// TODO: Extract settings from payload when frontend sends them
	settings := game.GameSettings{
		MaxPlayers: game.DefaultMaxPlayers,
		CardPacks:  game.DefaultCardPacks(),
	}

	// Extract custom settings if provided
	if payloadMap, ok := message.Payload.(map[string]interface{}); ok {
		if maxPlayers, ok := payloadMap["maxPlayers"].(float64); ok {
			settings.MaxPlayers = int(maxPlayers)
		}
		if cardPacks, ok := payloadMap["cardPacks"].([]interface{}); ok {
			packs := make([]string, len(cardPacks))
			for i, pack := range cardPacks {
				if packStr, ok := pack.(string); ok {
					packs[i] = packStr
				}
			}
			if len(packs) > 0 {
				settings.CardPacks = packs
			}
		}
	}

	log.Debug("Parsed create game settings",
		zap.Int("max_players", settings.MaxPlayers),
		zap.Strings("card_packs", settings.CardPacks))

	// Execute the migrated create game action
	game, err := h.createGameAction.Execute(ctx, settings)
	if err != nil {
		log.Error("Failed to execute create game action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Create game action completed successfully",
		zap.String("game_id", game.ID()))

	// Send simple success response with game ID
	// Frontend will then call playerConnect to join the game
	// and receive full game state via the broadcaster
	response := dto.WebSocketMessage{
		Type: dto.MessageTypeGameUpdated, // Use standard message type
		Payload: map[string]interface{}{
			"gameId":  game.ID(),
			"success": true,
			"message": "Game created successfully. Join with playerConnect.",
		},
	}

	connection.Send <- response

	log.Info("📤 Sent game created response to client")
}

// sendError sends an error message to the client
func (h *CreateGameHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
