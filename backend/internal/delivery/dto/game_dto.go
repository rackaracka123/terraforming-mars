package dto

import (
	"terraforming-mars-backend/internal/model"
)

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseWaitingForGameStart   GamePhase = "waiting_for_game_start"
	GamePhaseStartingCardSelection GamePhase = "starting_card_selection"
	GamePhaseStartGameSelection    GamePhase = "start_game_selection"
	GamePhaseAction                GamePhase = "action"
	GamePhaseProductionAndCardDraw GamePhase = "production_and_card_draw"
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
	CardTypeAutomated   CardType = "automated" // Updated from "effect" to match JSON data
	CardTypeActive      CardType = "active"
	CardTypeEvent       CardType = "event"
	CardTypeCorporation CardType = "corporation"
	CardTypePrelude     CardType = "prelude"
)

// CardTag represents different card categories and attributes
type CardTag string

const (
	TagSpace    CardTag = "space"
	TagEarth    CardTag = "earth"
	TagScience  CardTag = "science"
	TagPower    CardTag = "power"
	TagBuilding CardTag = "building"
	TagMicrobe  CardTag = "microbe"
	TagAnimal   CardTag = "animal"
	TagPlant    CardTag = "plant"
	TagEvent    CardTag = "event"
	TagCity     CardTag = "city"
	TagVenus    CardTag = "venus"
	TagJovian   CardTag = "jovian"
	TagWildlife CardTag = "wildlife"
	TagWild     CardTag = "wild"
)

// ResourceSet represents a collection of resources and their amounts
type ResourceSet struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// CardRequirements defines what conditions must be met to play a card
type CardRequirements struct {
	MinTemperature     *int         `json:"minTemperature,omitempty" ts:"number | undefined"`
	MaxTemperature     *int         `json:"maxTemperature,omitempty" ts:"number | undefined"`
	MinOxygen          *int         `json:"minOxygen,omitempty" ts:"number | undefined"`
	MaxOxygen          *int         `json:"maxOxygen,omitempty" ts:"number | undefined"`
	MinOceans          *int         `json:"minOceans,omitempty" ts:"number | undefined"`
	MaxOceans          *int         `json:"maxOceans,omitempty" ts:"number | undefined"`
	RequiredTags       []CardTag    `json:"requiredTags" ts:"CardTag[]"`
	RequiredProduction *ResourceSet `json:"requiredProduction,omitempty" ts:"ResourceSet | undefined"`
}

// ProductionEffects represents changes to resource production
type ProductionEffects struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// CardDto represents a card for client consumption
type CardDto struct {
	ID                string             `json:"id" ts:"string"`
	Name              string             `json:"name" ts:"string"`
	Type              CardType           `json:"type" ts:"CardType"`
	Cost              int                `json:"cost" ts:"number"`
	Description       string             `json:"description" ts:"string"`
	Tags              []CardTag          `json:"tags" ts:"CardTag[]"`
	Requirements      CardRequirements   `json:"requirements" ts:"CardRequirements"`
	VictoryPoints     int                `json:"victoryPoints" ts:"number"`
	Number            string             `json:"number" ts:"string"`
	ProductionEffects *ProductionEffects `json:"productionEffects,omitempty" ts:"ProductionEffects | undefined"`
}

// CorporationDto represents a corporation for client consumption
type CorporationDto struct {
	ID                 string      `json:"id" ts:"string"`
	Name               string      `json:"name" ts:"string"`
	Description        string      `json:"description" ts:"string"`
	StartingCredits    int         `json:"startingCredits" ts:"number"`
	StartingResources  ResourceSet `json:"startingResources" ts:"ResourceSet"`
	StartingProduction ResourceSet `json:"startingProduction" ts:"ResourceSet"`
	Tags               []CardTag   `json:"tags" ts:"CardTag[]"`
	SpecialEffects     []string    `json:"specialEffects" ts:"string[]"`
	Number             string      `json:"number" ts:"string"`
}

// ProductionPhaseDto represents card selection and production phase state for a player
type ProductionPhaseDto struct {
	AvailableCards    []CardDto `json:"availableCards" ts:"CardDto[]"`  // Cards available for selection
	SelectionComplete bool      `json:"selectionComplete" ts:"boolean"` // Whether player completed card selection
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
	Production       ProductionDto          `json:"resourceProduction" ts:"ProductionDto"`
	TerraformRating  int                    `json:"terraformRating" ts:"number"`
	IsActive         bool                   `json:"isActive" ts:"boolean"`
	IsReady          bool                   `json:"isReady" ts:"boolean"`
	PlayedCards      []string               `json:"playedCards" ts:"string[]"`
	Passed           bool                   `json:"passed" ts:"boolean"`
	AvailableActions int                    `json:"availableActions" ts:"number"`
	VictoryPoints    int                    `json:"victoryPoints" ts:"number"`
	MilestoneIcon    string                 `json:"milestoneIcon" ts:"string"`
	ConnectionStatus model.ConnectionStatus `json:"connectionStatus" ts:"ConnectionStatus"`
	// Card selection state - nullable, exists only during selection phase
	CardSelection *ProductionPhaseDto `json:"production" ts:"ProductionPhaseDto | null"`
	// Starting card selection - available during starting_card_selection phase
	StartingSelection []CardDto `json:"startingSelection" ts:"CardDto[]"`
}

// OtherPlayerDto represents another player from the viewing player's perspective (limited data)
type OtherPlayerDto struct {
	ID               string                 `json:"id" ts:"string"`
	Name             string                 `json:"name" ts:"string"`
	Corporation      string                 `json:"corporation" ts:"string"`
	HandCardCount    int                    `json:"handCardCount" ts:"number"` // Number of cards in hand (private)
	Resources        ResourcesDto           `json:"resources" ts:"ResourcesDto"`
	Production       ProductionDto          `json:"resourceProduction" ts:"ProductionDto"`
	TerraformRating  int                    `json:"terraformRating" ts:"number"`
	IsActive         bool                   `json:"isActive" ts:"boolean"`
	IsReady          bool                   `json:"isReady" ts:"boolean"`
	PlayedCards      []string               `json:"playedCards" ts:"string[]"` // Played cards are public
	Passed           bool                   `json:"passed" ts:"boolean"`
	AvailableActions int                    `json:"availableActions" ts:"number"`
	VictoryPoints    int                    `json:"victoryPoints" ts:"number"`
	MilestoneIcon    string                 `json:"milestoneIcon" ts:"string"`
	ConnectionStatus model.ConnectionStatus `json:"connectionStatus" ts:"ConnectionStatus"`
	// Card selection state - limited visibility for other players
	IsSelectingCards bool `json:"isSelectingCards" ts:"boolean"` // Whether player is currently selecting cards
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
	TurnOrder        []string            `json:"turnOrder" ts:"string[]"`      // Turn order of all players in game
}
