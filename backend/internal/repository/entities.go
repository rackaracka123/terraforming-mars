package repository

import (
	"terraforming-mars-backend/internal/model"
	"time"
)

// GameEntity represents a game entity optimized for repository storage
// Contains player IDs instead of full player objects and integrates global parameters
type GameEntity struct {
	ID               string                 `json:"id" ts:"string"`
	CreatedAt        time.Time              `json:"createdAt" ts:"string"`
	UpdatedAt        time.Time              `json:"updatedAt" ts:"string"`
	Status           model.GameStatus       `json:"status" ts:"GameStatus"`
	Settings         model.GameSettings     `json:"settings" ts:"GameSettings"`
	PlayerIDs        []string               `json:"playerIds" ts:"string[]"` // Repository storage: just IDs
	HostPlayerID     string                 `json:"hostPlayerId" ts:"string"`
	CurrentPhase     model.GamePhase        `json:"currentPhase" ts:"GamePhase"`
	GlobalParameters model.GlobalParameters `json:"globalParameters" ts:"GlobalParameters"` // Integrated global parameters
	CurrentPlayerID  string                 `json:"currentPlayerId" ts:"string"`
	Generation       int                    `json:"generation" ts:"number"`
	RemainingActions int                    `json:"remainingActions" ts:"number"`
}

// NewGameEntity creates a new game entity with the given settings
func NewGameEntity(id string, settings model.GameSettings) *GameEntity {
	now := time.Now()

	return &GameEntity{
		ID:           id,
		CreatedAt:    now,
		UpdatedAt:    now,
		Status:       model.GameStatusLobby,
		Settings:     settings,
		PlayerIDs:    make([]string, 0), // Empty list of player IDs
		CurrentPhase: model.GamePhaseWaitingForGameStart,
		GlobalParameters: model.GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		},
		Generation:       1,
		RemainingActions: 0,
	}
}

// ToGame converts GameEntity to Game by populating players from repository
func (ge *GameEntity) ToGame(players []model.Player) *model.Game {
	return &model.Game{
		ID:               ge.ID,
		CreatedAt:        ge.CreatedAt,
		UpdatedAt:        ge.UpdatedAt,
		Status:           ge.Status,
		Settings:         ge.Settings,
		Players:          players, // Populated from repository
		HostPlayerID:     ge.HostPlayerID,
		CurrentPhase:     ge.CurrentPhase,
		GlobalParameters: ge.GlobalParameters,
		CurrentPlayerID:  ge.CurrentPlayerID,
		Generation:       ge.Generation,
		RemainingActions: ge.RemainingActions,
	}
}

// PlayerEntity represents a player entity for repository storage
// This is identical to Player as requested
type PlayerEntity struct {
	ID               string                 `json:"id" ts:"string"`
	Name             string                 `json:"name" ts:"string"`
	Corporation      string                 `json:"corporation" ts:"string"`
	Cards            []string               `json:"cards" ts:"string[]"`
	Resources        model.Resources        `json:"resources" ts:"Resources"`
	Production       model.Production       `json:"production" ts:"Production"`
	TerraformRating  int                    `json:"terraformRating" ts:"number"`
	IsActive         bool                   `json:"isActive" ts:"boolean"`
	PlayedCards      []string               `json:"playedCards" ts:"string[]"`
	Passed           bool                   `json:"passed" ts:"boolean"`
	AvailableActions int                    `json:"availableActions" ts:"number"`
	VictoryPoints    int                    `json:"victoryPoints" ts:"number"`
	MilestoneIcon    string                 `json:"milestoneIcon" ts:"string"`
	ConnectionStatus model.ConnectionStatus `json:"connectionStatus" ts:"ConnectionStatus"`
}

// ToPlayer converts PlayerEntity to Player
func (pe *PlayerEntity) ToPlayer() model.Player {
	return model.Player{
		ID:               pe.ID,
		Name:             pe.Name,
		Corporation:      pe.Corporation,
		Cards:            pe.Cards,
		Resources:        pe.Resources,
		Production:       pe.Production,
		TerraformRating:  pe.TerraformRating,
		IsActive:         pe.IsActive,
		PlayedCards:      pe.PlayedCards,
		Passed:           pe.Passed,
		AvailableActions: pe.AvailableActions,
		VictoryPoints:    pe.VictoryPoints,
		MilestoneIcon:    pe.MilestoneIcon,
		ConnectionStatus: pe.ConnectionStatus,
	}
}

// FromPlayer converts Player to PlayerEntity
func PlayerEntityFromPlayer(p model.Player) PlayerEntity {
	return PlayerEntity{
		ID:               p.ID,
		Name:             p.Name,
		Corporation:      p.Corporation,
		Cards:            p.Cards,
		Resources:        p.Resources,
		Production:       p.Production,
		TerraformRating:  p.TerraformRating,
		IsActive:         p.IsActive,
		PlayedCards:      p.PlayedCards,
		Passed:           p.Passed,
		AvailableActions: p.AvailableActions,
		VictoryPoints:    p.VictoryPoints,
		MilestoneIcon:    p.MilestoneIcon,
		ConnectionStatus: p.ConnectionStatus,
	}
}
