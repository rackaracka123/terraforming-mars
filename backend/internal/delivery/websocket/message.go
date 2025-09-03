package websocket

import (
	"terraforming-mars-backend/internal/delivery/dto"
)

// Re-export DTO types for backward compatibility and cleaner imports
type MessageType = dto.MessageType
type WebSocketMessage = dto.WebSocketMessage
type PlayerConnectPayload = dto.PlayerConnectPayload
type PlayActionPayload = dto.PlayActionPayload
type GameUpdatedPayload = dto.GameUpdatedPayload
type PlayerConnectedPayload = dto.PlayerConnectedPayload
type ErrorPayload = dto.ErrorPayload
type FullStatePayload = dto.FullStatePayload

// Re-export message type constants
const (
	MessageTypePlayerConnect   = dto.MessageTypePlayerConnect
	MessageTypePlayAction      = dto.MessageTypePlayAction
	MessageTypeGameUpdated     = dto.MessageTypeGameUpdated
	MessageTypePlayerConnected = dto.MessageTypePlayerConnected
	MessageTypeError           = dto.MessageTypeError
	MessageTypeFullState       = dto.MessageTypeFullState
)
