package card

import (
	"context"

	cardaction "terraforming-mars-backend/internal/action/card"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// PlayCardHandler handles play card requests
type PlayCardHandler struct {
	action      *cardaction.PlayCardAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// Broadcaster interface for explicit broadcasting
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// NewPlayCardHandler creates a new play card handler
func NewPlayCardHandler(action *cardaction.PlayCardAction, broadcaster Broadcaster) *PlayCardHandler {
	return &PlayCardHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *PlayCardHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🃏 Processing play card request")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	payload, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		h.sendError(connection, "Invalid payload format")
		return
	}

	cardID, ok := payload["cardId"].(string)
	if !ok || cardID == "" {
		log.Error("Missing or invalid cardId")
		h.sendError(connection, "Missing cardId")
		return
	}

	payment := cardaction.PaymentRequest{
		Credits:     0,
		Steel:       0,
		Titanium:    0,
		Substitutes: make(map[shared.ResourceType]int),
	}

	if paymentData, ok := payload["payment"].(map[string]interface{}); ok {
		if credits, ok := paymentData["credits"].(float64); ok {
			payment.Credits = int(credits)
		}
		if steel, ok := paymentData["steel"].(float64); ok {
			payment.Steel = int(steel)
		}
		if titanium, ok := paymentData["titanium"].(float64); ok {
			payment.Titanium = int(titanium)
		}
		if substitutesData, ok := paymentData["substitutes"].(map[string]interface{}); ok {
			for resourceTypeStr, amountVal := range substitutesData {
				if amount, ok := amountVal.(float64); ok && amount > 0 {
					resourceType := shared.ResourceType(resourceTypeStr)
					payment.Substitutes[resourceType] = int(amount)
				}
			}
		}
	}

	var choiceIndex *int
	if choiceIndexFloat, ok := payload["choiceIndex"].(float64); ok {
		idx := int(choiceIndexFloat)
		choiceIndex = &idx
	}

	log.Debug("Payment extracted",
		zap.Int("credits", payment.Credits),
		zap.Int("steel", payment.Steel),
		zap.Int("titanium", payment.Titanium),
		zap.Any("substitutes", payment.Substitutes))
	if choiceIndex != nil {
		log.Debug("Choice index extracted", zap.Int("choice_index", *choiceIndex))
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID, cardID, payment, choiceIndex)
	if err != nil {
		log.Error("Failed to execute play card action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Play card action completed successfully")

	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("📡 Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "play-card",
			"success": true,
			"cardId":  cardID,
		},
	}

	connection.Send <- response
}

func (h *PlayCardHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
