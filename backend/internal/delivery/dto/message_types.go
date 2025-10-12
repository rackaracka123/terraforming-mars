package dto

// MessageType represents different types of WebSocket messages
type MessageType string

const (
	// Existing Client -> Server messages
	MessageTypePlayerConnect MessageType = "player-connect"

	// Existing Server -> Client messages
	MessageTypeGameUpdated            MessageType = "game-updated"
	MessageTypePlayerConnected        MessageType = "player-connected"
	MessageTypePlayerReconnected      MessageType = "player-reconnected"
	MessageTypePlayerDisconnected     MessageType = "player-disconnected"
	MessageTypeError                  MessageType = "error"
	MessageTypeFullState              MessageType = "full-state"
	MessageTypeProductionPhaseStarted MessageType = "production-phase-started"

	// New action-specific message types using composed constants
	// Standard project message types
	MessageTypeActionSellPatents     MessageType = "action.standard-project.sell-patents"
	MessageTypeActionLaunchAsteroid  MessageType = "action.standard-project.launch-asteroid"
	MessageTypeActionBuildPowerPlant MessageType = "action.standard-project.build-power-plant"
	MessageTypeActionBuildAquifer    MessageType = "action.standard-project.build-aquifer"
	MessageTypeActionPlantGreenery   MessageType = "action.standard-project.plant-greenery"
	MessageTypeActionBuildCity       MessageType = "action.standard-project.build-city"

	// Game management message types
	MessageTypeActionStartGame  MessageType = "action.game-management.start-game"
	MessageTypeActionSkipAction MessageType = "action.game-management.skip-action"

	// Tile selection message types
	MessageTypeActionTileSelected MessageType = "action.tile-selection.tile-selected"

	// Card message types
	MessageTypeActionPlayCard           MessageType = "action.card.play-card"
	MessageTypeActionCardAction         MessageType = "action.card.card-action"
	MessageTypeActionSelectStartingCard MessageType = "action.card.select-starting-card"
	MessageTypeActionSelectCards        MessageType = "action.card.select-cards"
	MessageTypeActionCardDrawConfirmed  MessageType = "action.card.card-draw-confirmed"

	// Admin message types (development mode only)
	MessageTypeAdminCommand MessageType = "admin-command"
)
