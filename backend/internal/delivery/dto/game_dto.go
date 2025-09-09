package dto

import (
	"terraforming-mars-backend/internal/model"
)

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseSetup                 GamePhase = "setup"
	GamePhaseStartingCardSelection GamePhase = "starting_card_selection"
	GamePhaseCorporationSelection  GamePhase = "corporation_selection"
	GamePhaseAction                GamePhase = "action"
	GamePhaseProduction            GamePhase = "production"
	GamePhaseComplete              GamePhase = "complete"
)

// GameStatus represents the current status of the game
type GameStatus string

const (
	GameStatusLobby     GameStatus = "lobby"
	GameStatusActive    GameStatus = "active"
	GameStatusCompleted GameStatus = "completed"
)

// CardType represents different types of cards
type CardType string

const (
	CardTypeEffect      CardType = "effect"
	CardTypeActive      CardType = "active"
	CardTypeEvent       CardType = "event"
	CardTypeCorporation CardType = "corporation"
)

// CardDto represents a card for client consumption
type CardDto struct {
	ID          string   `json:"id" ts:"string"`
	Name        string   `json:"name" ts:"string"`
	Type        CardType `json:"type" ts:"CardType"`
	Cost        int      `json:"cost" ts:"number"`
	Description string   `json:"description" ts:"string"`
}

// GameSettingsDto contains configurable game parameters
type GameSettingsDto struct {
	MaxPlayers int `json:"maxPlayers" ts:"number"`
}

// GlobalParametersDto represents the terraforming progress
type GlobalParametersDto struct {
	Temperature int `json:"temperature" ts:"number"` // Range: -30 to +8°C
	Oxygen      int `json:"oxygen" ts:"number"`      // Range: 0-14%
	Oceans      int `json:"oceans" ts:"number"`      // Range: 0-9
}

// ResourcesDto represents a player's resources
type ResourcesDto struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// ProductionDto represents a player's production values
type ProductionDto struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// PlayerDto represents a player in the game for client consumption
type PlayerDto struct {
	ID               string                 `json:"id" ts:"string"`
	Name             string                 `json:"name" ts:"string"`
	Corporation      string                 `json:"corporation" ts:"string"`
	Cards            []string               `json:"cards" ts:"string[]"`
	Resources        ResourcesDto           `json:"resources" ts:"ResourcesDto"`
	Production       ProductionDto          `json:"production" ts:"ProductionDto"`
	TerraformRating  int                    `json:"terraformRating" ts:"number"`
	IsActive         bool                   `json:"isActive" ts:"boolean"`
	PlayedCards      []string               `json:"playedCards" ts:"string[]"`
	Passed           bool                   `json:"passed" ts:"boolean"`
	AvailableActions int                    `json:"availableActions" ts:"number"`
	VictoryPoints    int                    `json:"victoryPoints" ts:"number"`
	MilestoneIcon    string                 `json:"milestoneIcon" ts:"string"`
	ConnectionStatus model.ConnectionStatus `json:"connectionStatus" ts:"ConnectionStatus"`
}

// OtherPlayerDto represents another player from the viewing player's perspective (limited data)
type OtherPlayerDto struct {
	ID               string                 `json:"id" ts:"string"`
	Name             string                 `json:"name" ts:"string"`
	Corporation      string                 `json:"corporation" ts:"string"`
	HandCardCount    int                    `json:"handCardCount" ts:"number"` // Number of cards in hand (private)
	Resources        ResourcesDto           `json:"resources" ts:"ResourcesDto"`
	Production       ProductionDto          `json:"production" ts:"ProductionDto"`
	TerraformRating  int                    `json:"terraformRating" ts:"number"`
	IsActive         bool                   `json:"isActive" ts:"boolean"`
	PlayedCards      []string               `json:"playedCards" ts:"string[]"` // Played cards are public
	Passed           bool                   `json:"passed" ts:"boolean"`
	AvailableActions int                    `json:"availableActions" ts:"number"`
	VictoryPoints    int                    `json:"victoryPoints" ts:"number"`
	MilestoneIcon    string                 `json:"milestoneIcon" ts:"string"`
	ConnectionStatus model.ConnectionStatus `json:"connectionStatus" ts:"ConnectionStatus"`
}

// GameDto represents a game for client consumption (clean architecture)
type GameDto struct {
	ID               string              `json:"id" ts:"string"`
	Status           GameStatus          `json:"status" ts:"GameStatus"`
	Settings         GameSettingsDto     `json:"settings" ts:"GameSettingsDto"`
	HostPlayerID     string              `json:"hostPlayerId" ts:"string"`
	CurrentPhase     GamePhase           `json:"currentPhase" ts:"GamePhase"`
	GlobalParameters GlobalParametersDto `json:"globalParameters" ts:"GlobalParametersDto"`
	CurrentPlayer    PlayerDto           `json:"currentPlayer" ts:"PlayerDto"`       // Viewing player's full data
	OtherPlayers     []OtherPlayerDto    `json:"otherPlayers" ts:"OtherPlayerDto[]"` // Other players' limited data
	ViewingPlayerID  string              `json:"viewingPlayerId" ts:"string"`        // The player viewing this game state
	CurrentTurn      *string             `json:"currentTurn" ts:"string|null"`       // Whose turn it is (nullable)
	Generation       int                 `json:"generation" ts:"number"`
	RemainingActions int                 `json:"remainingActions" ts:"number"` // Remaining actions in the current turn
}
