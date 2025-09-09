package model

import "time"

// ReadinessType represents the type of readiness check
type ReadinessType string

const (
	ReadinessTypeProductionPhase   ReadinessType = "production-phase"
	ReadinessTypeCardSelection     ReadinessType = "card-selection"
	ReadinessTypeCorporationChoice ReadinessType = "corporation-choice"
)

// ReadinessCheck represents a generic system for tracking player readiness
type ReadinessCheck struct {
	Type        ReadinessType `json:"type" ts:"ReadinessType"`
	ReadyPlayers []string      `json:"readyPlayers" ts:"string[]"`
	IsActive    bool          `json:"isActive" ts:"boolean"`
}

// Game represents a unified game entity containing both metadata and state
type Game struct {
	ID               string           `json:"id" ts:"string"`
	CreatedAt        time.Time        `json:"createdAt" ts:"string"`
	UpdatedAt        time.Time        `json:"updatedAt" ts:"string"`
	Status           GameStatus       `json:"status" ts:"GameStatus"`
	Settings         GameSettings     `json:"settings" ts:"GameSettings"`
	PlayerIDs        []string         `json:"playerIds" ts:"string[]"` // Player IDs in this game
	HostPlayerID     string           `json:"hostPlayerId" ts:"string"`
	CurrentPhase     GamePhase        `json:"currentPhase" ts:"GamePhase"`
	GlobalParameters GlobalParameters `json:"globalParameters" ts:"GlobalParameters"`
	ViewingPlayerID  string           `json:"viewingPlayerId" ts:"string"`  // The player viewing this game state
	CurrentTurn      *string          `json:"currentTurn" ts:"string|null"` // Whose turn it is (nullable)
	Generation       int              `json:"generation" ts:"number"`
	ReadinessCheck   *ReadinessCheck  `json:"readinessCheck,omitempty" ts:"ReadinessCheck | null"`
}

// NewGame creates a new game with the given settings
func NewGame(id string, settings GameSettings) *Game {
	now := time.Now()

	return &Game{
		ID:           id,
		CreatedAt:    now,
		UpdatedAt:    now,
		Status:       GameStatusLobby,
		Settings:     settings,
		PlayerIDs:    make([]string, 0),
		CurrentPhase: GamePhaseWaitingForGameStart,
		GlobalParameters: GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		},
		Generation:     1,
		ReadinessCheck: nil, // No active readiness check initially
	}
}

// DeepCopy creates a deep copy of the Game
func (g *Game) DeepCopy() *Game {
	if g == nil {
		return nil
	}

	// Copy players slice
	playersCopy := make([]Player, len(g.Players))
	for i, player := range g.Players {
		playersCopy[i] = *player.DeepCopy()
	}

	// Copy ReadinessCheck if it exists
	var readinessCheckCopy *ReadinessCheck
	if g.ReadinessCheck != nil {
		readinessCheckCopy = g.ReadinessCheck.DeepCopy()
	}

	return &Game{
		ID:               g.ID,
		CreatedAt:        g.CreatedAt,
		UpdatedAt:        g.UpdatedAt,
		Status:           g.Status,
		Settings:         *g.Settings.DeepCopy(),
		Players:          playersCopy,
		HostPlayerID:     g.HostPlayerID,
		CurrentPhase:     g.CurrentPhase,
		GlobalParameters: *g.GlobalParameters.DeepCopy(),
		CurrentPlayerID:  g.CurrentPlayerID,
		Generation:       g.Generation,
		ReadinessCheck:   readinessCheckCopy,
	}
}

// DeepCopy creates a deep copy of the ReadinessCheck
func (r *ReadinessCheck) DeepCopy() *ReadinessCheck {
	if r == nil {
		return nil
	}

	// Copy ReadyPlayers slice
	readyPlayersCopy := make([]string, len(r.ReadyPlayers))
	copy(readyPlayersCopy, r.ReadyPlayers)

	return &ReadinessCheck{
		Type:         r.Type,
		ReadyPlayers: readyPlayersCopy,
		IsActive:     r.IsActive,
	}
}
