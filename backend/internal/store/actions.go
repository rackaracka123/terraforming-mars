package store

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
)

// Use DTO action types directly
type ActionType = dto.ActionType

// Extended action types for store-specific operations
const (
	// Internal Store Actions (not exposed to frontend)
	ActionCreateGame       ActionType = "create-game"
	ActionJoinGame         ActionType = "join-game"
	ActionDeleteGame       ActionType = "delete-game"
	ActionUpdateGameStatus ActionType = "update-game-status"
	ActionUpdateGamePhase  ActionType = "update-game-phase"
	ActionSetCurrentTurn   ActionType = "set-current-turn"
	ActionAdvanceTurn      ActionType = "advance-turn"
	ActionUpdateGeneration ActionType = "update-generation"

	// Player Management
	ActionCreatePlayer          ActionType = "create-player"
	ActionRemovePlayer          ActionType = "remove-player"
	ActionLeaveGame             ActionType = "leave-game"
	ActionUpdateResources       ActionType = "update-resources"
	ActionDeductResources       ActionType = "deduct-resources"
	ActionUpdateProduction      ActionType = "update-production"
	ActionUpdateTerraformRating ActionType = "update-terraform-rating"
	ActionConsumeAction         ActionType = "consume-action"
	ActionResetActions          ActionType = "reset-actions"
	ActionSetPassed             ActionType = "set-passed"
	ActionSetCardOptions        ActionType = "set-card-options"

	// New Separated Actions
	ActionPlayerJoin        ActionType = "player-join"         // Game-level player joining
	ActionDealStartingCards ActionType = "deal-starting-cards" // Player-level card dealing
	ActionPlayerTurnEnd     ActionType = "player-turn-end"     // Player-level turn cleanup
	ActionPlayerProduction  ActionType = "player-production"   // Player-level production execution

	// Global Parameters
	ActionUpdateGlobalParams  ActionType = "update-global-params"
	ActionIncreaseTemperature ActionType = "increase-temperature"
	ActionIncreaseOxygen      ActionType = "increase-oxygen"
	ActionPlaceOcean          ActionType = "place-ocean"

	// Production Phase
	ActionExecuteProduction ActionType = "execute-production"

	// Card Management (use DTO types)
	ActionPlayCard            = dto.ActionTypePlayCard
	ActionSelectStartingCards = dto.ActionTypeSelectStartingCard
	ActionSelectCards         = dto.ActionTypeSelectCards

	// Standard Projects (use DTO types)
	ActionSellPatents     = dto.ActionTypeSellPatents
	ActionBuildPowerPlant = dto.ActionTypeBuildPowerPlant
	ActionLaunchAsteroid  = dto.ActionTypeLaunchAsteroid
	ActionBuildAquifer    = dto.ActionTypeBuildAquifer
	ActionPlantGreenery   = dto.ActionTypePlantGreenery
	ActionBuildCity       = dto.ActionTypeBuildCity

	// Game Actions (use DTO types)
	ActionStartGame  = dto.ActionTypeStartGame
	ActionSkipAction = dto.ActionTypeSkipAction
)

// Action represents a state change action
type Action struct {
	Type    ActionType `json:"type"`
	Payload any        `json:"payload"`
	Meta    ActionMeta `json:"meta"`
}

// ActionMeta contains metadata about the action
type ActionMeta struct {
	GameID   string `json:"gameId,omitempty"`
	PlayerID string `json:"playerId,omitempty"`
	Source   string `json:"source,omitempty"` // "websocket", "http", "system"
}

// Minimal internal payload types for store operations that don't have DTO equivalents
type CreateGamePayload struct {
	GameID   string             `json:"gameId"`
	Settings model.GameSettings `json:"settings"`
}

type JoinGamePayload struct {
	GameID     string `json:"gameId"`
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

type SelectStartingCardsPayload struct {
	GameID        string   `json:"gameId"`
	PlayerID      string   `json:"playerId"`
	SelectedCards []string `json:"selectedCards"`
	Cost          int      `json:"cost"`
}

// Additional internal payload types needed by HTTP handlers
type UpdateResourcesPayload struct {
	GameID    string          `json:"gameId"`
	PlayerID  string          `json:"playerId"`
	Resources model.Resources `json:"resources"`
}

// Action Creators

func NewAction(actionType ActionType, payload any, gameID, playerID, source string) Action {
	return Action{
		Type:    actionType,
		Payload: payload,
		Meta: ActionMeta{
			GameID:   gameID,
			PlayerID: playerID,
			Source:   source,
		},
	}
}

// Action Creators

// Internal store operations
func CreateGameAction(gameID string, settings model.GameSettings, source string) Action {
	return NewAction(ActionCreateGame, CreateGamePayload{
		GameID:   gameID,
		Settings: settings,
	}, gameID, "", source)
}

func JoinGameAction(gameID, playerID, playerName, source string) Action {
	return NewAction(ActionJoinGame, JoinGamePayload{
		GameID:     gameID,
		PlayerID:   playerID,
		PlayerName: playerName,
	}, gameID, playerID, source)
}

// DTO-based action creators
func StartGameAction(gameID, playerID, source string) Action {
	return NewAction(ActionStartGame, dto.StartGameAction{
		Type: ActionStartGame,
	}, gameID, playerID, source)
}

func SkipActionAction(gameID, playerID, source string) Action {
	return NewAction(ActionSkipAction, dto.SkipAction{
		Type: ActionSkipAction,
	}, gameID, playerID, source)
}

func UpdateResourcesAction(gameID, playerID string, resources model.Resources, source string) Action {
	return NewAction(ActionUpdateResources, UpdateResourcesPayload{
		GameID:    gameID,
		PlayerID:  playerID,
		Resources: resources,
	}, gameID, playerID, source)
}
