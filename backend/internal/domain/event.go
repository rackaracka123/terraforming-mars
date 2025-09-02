package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// GameEventStream represents a stream of events for a game
type GameEventStream struct {
	GameID      string      `json:"gameId" ts:"string"`
	Events      []GameEvent `json:"events" ts:"GameEvent[]"`
	Version     int64       `json:"version" ts:"number"`
	CreatedAt   time.Time   `json:"createdAt" ts:"string"`
	UpdatedAt   time.Time   `json:"updatedAt" ts:"string"`
}

// GameEvent represents a single event that occurred in the game
type GameEvent struct {
	ID           string          `json:"id" ts:"string"`
	GameID       string          `json:"gameId" ts:"string"`
	Type         GameEventType   `json:"type" ts:"GameEventType"`
	Version      int64           `json:"version" ts:"number"`
	PlayerID     *string         `json:"playerId,omitempty" ts:"string | undefined"`
	Timestamp    time.Time       `json:"timestamp" ts:"string"`
	Data         json.RawMessage `json:"data" ts:"any"`
	Metadata     EventMetadata   `json:"metadata" ts:"EventMetadata"`
}

// EventMetadata contains additional information about the event
type EventMetadata struct {
	CorrelationID *string           `json:"correlationId,omitempty" ts:"string | undefined"`
	CausedBy      *string           `json:"causedBy,omitempty" ts:"string | undefined"`
	Tags          []string          `json:"tags" ts:"string[]"`
	UserAgent     *string           `json:"userAgent,omitempty" ts:"string | undefined"`
	IPAddress     *string           `json:"ipAddress,omitempty" ts:"string | undefined"`
	Extra         map[string]string `json:"extra" ts:"Record<string, string>"`
}

// GameEventType represents all possible events that can occur in the game
type GameEventType string

const (
	// Game Lifecycle Events
	EventTypeGameCreated             GameEventType = "game_created"
	EventTypeGameStarted             GameEventType = "game_started"
	EventTypeGameEnded               GameEventType = "game_ended"
	EventTypeGameAbandoned           GameEventType = "game_abandoned"
	
	// Player Events
	EventTypePlayerJoined            GameEventType = "player_joined"
	EventTypePlayerLeft              GameEventType = "player_left"
	EventTypePlayerReconnected       GameEventType = "player_reconnected"
	EventTypeCorporationSelected     GameEventType = "corporation_selected"
	
	// Phase Events
	EventTypePhaseChanged            GameEventType = "phase_changed"
	EventTypeGenerationStarted       GameEventType = "generation_started"
	EventTypeGenerationEnded         GameEventType = "generation_ended"
	EventTypeTurnStarted             GameEventType = "turn_started"
	EventTypeTurnEnded               GameEventType = "turn_ended"
	
	// Action Events
	EventTypeCardPlayed              GameEventType = "card_played"
	EventTypeStandardProjectUsed     GameEventType = "standard_project_used"
	EventTypeTilePlaced              GameEventType = "tile_placed"
	EventTypeResourcesGained         GameEventType = "resources_gained"
	EventTypeResourcesLost           GameEventType = "resources_lost"
	EventTypeProductionChanged       GameEventType = "production_changed"
	EventTypeParameterIncreased      GameEventType = "parameter_increased"
	EventTypeActionPerformed         GameEventType = "action_performed"
	
	// Milestone and Award Events
	EventTypeMilestoneClaimed        GameEventType = "milestone_claimed"
	EventTypeAwardFunded             GameEventType = "award_funded"
	EventTypeAwardRankingUpdated     GameEventType = "award_ranking_updated"
	
	// Resource Conversion Events
	EventTypeHeatConverted           GameEventType = "heat_converted"
	EventTypePlantsConverted         GameEventType = "plants_converted"
	EventTypeEnergyConverted         GameEventType = "energy_converted"
	
	// Card Events
	EventTypeCardsDrawn              GameEventType = "cards_drawn"
	EventTypeCardsDiscarded          GameEventType = "cards_discarded"
	EventTypeCardActivated           GameEventType = "card_activated"
	
	// Victory Point Events
	EventTypeVictoryPointsAwarded    GameEventType = "victory_points_awarded"
	EventTypeFinalScoring            GameEventType = "final_scoring"
)

// Event Data Structures - These are the payloads for each event type

// GameCreatedData contains the initial game setup
type GameCreatedData struct {
	Settings    GameSettings `json:"settings" ts:"GameSettings"`
	CreatedBy   string       `json:"createdBy" ts:"string"`
	MaxPlayers  int          `json:"maxPlayers" ts:"number"`
}

// PlayerJoinedData contains player information
type PlayerJoinedData struct {
	PlayerID   string `json:"playerId" ts:"string"`
	PlayerName string `json:"playerName" ts:"string"`
	JoinOrder  int    `json:"joinOrder" ts:"number"`
}

// CorporationSelectedData contains corporation selection
type CorporationSelectedData struct {
	PlayerID      string `json:"playerId" ts:"string"`
	CorporationID string `json:"corporationId" ts:"string"`
}

// CardPlayedData contains card play information
type CardPlayedData struct {
	PlayerID         string            `json:"playerId" ts:"string"`
	CardID           string            `json:"cardId" ts:"string"`
	Cost             int               `json:"cost" ts:"number"`
	ResourcesSpent   ResourcesMap      `json:"resourcesSpent" ts:"ResourcesMap"`
	Placement        *HexCoordinate    `json:"placement,omitempty" ts:"HexCoordinate | undefined"`
	TargetPlayer     *string           `json:"targetPlayer,omitempty" ts:"string | undefined"`
	Requirements     []Requirement     `json:"requirements" ts:"Requirement[]"`
	ImmediateEffects []CardEffect      `json:"immediateEffects" ts:"CardEffect[]"`
}

// TilePlacedData contains tile placement information
type TilePlacedData struct {
	PlayerID          string          `json:"playerId" ts:"string"`
	TileType          TileType        `json:"tileType" ts:"TileType"`
	Position          HexCoordinate   `json:"position" ts:"HexCoordinate"`
	SpaceBonuses      []ResourceType  `json:"spaceBonuses" ts:"ResourceType[]"`
	AdjacencyBonuses  []ResourceType  `json:"adjacencyBonuses" ts:"ResourceType[]"`
	TriggeredEffects  []CardEffect    `json:"triggeredEffects" ts:"CardEffect[]"`
}

// ResourcesChangedData contains resource change information
type ResourcesChangedData struct {
	PlayerID   string       `json:"playerId" ts:"string"`
	Changes    ResourcesMap `json:"changes" ts:"ResourcesMap"`
	NewTotals  ResourcesMap `json:"newTotals" ts:"ResourcesMap"`
	Reason     string       `json:"reason" ts:"string"`
	SourceCard *string      `json:"sourceCard,omitempty" ts:"string | undefined"`
}

// ProductionChangedData contains production change information
type ProductionChangedData struct {
	PlayerID     string       `json:"playerId" ts:"string"`
	Changes      ResourcesMap `json:"changes" ts:"ResourcesMap"`
	NewTotals    ResourcesMap `json:"newTotals" ts:"ResourcesMap"`
	Reason       string       `json:"reason" ts:"string"`
	SourceCard   *string      `json:"sourceCard,omitempty" ts:"string | undefined"`
}

// ParameterIncreasedData contains global parameter changes
type ParameterIncreasedData struct {
	PlayerID      string      `json:"playerId" ts:"string"`
	Parameter     GlobalParam `json:"parameter" ts:"GlobalParam"`
	OldValue      int         `json:"oldValue" ts:"number"`
	NewValue      int         `json:"newValue" ts:"number"`
	Steps         int         `json:"steps" ts:"number"`
	TRIncrease    int         `json:"trIncrease" ts:"number"`
	BonusRewards  []CardEffect `json:"bonusRewards" ts:"CardEffect[]"`
}

// MilestoneClaimedData contains milestone claim information
type MilestoneClaimedData struct {
	PlayerID     string `json:"playerId" ts:"string"`
	MilestoneID  string `json:"milestoneId" ts:"string"`
	Cost         int    `json:"cost" ts:"number"`
	Requirements []Requirement `json:"requirements" ts:"Requirement[]"`
}

// AwardFundedData contains award funding information
type AwardFundedData struct {
	PlayerID string `json:"playerId" ts:"string"`
	AwardID  string `json:"awardId" ts:"string"`
	Cost     int    `json:"cost" ts:"number"`
	Position int    `json:"position" ts:"number"` // 1st, 2nd, or 3rd award funded
}

// PhaseChangedData contains phase transition information
type PhaseChangedData struct {
	OldPhase     GamePhase  `json:"oldPhase" ts:"GamePhase"`
	NewPhase     GamePhase  `json:"newPhase" ts:"GamePhase"`
	OldTurnPhase *TurnPhase `json:"oldTurnPhase,omitempty" ts:"TurnPhase | undefined"`
	NewTurnPhase *TurnPhase `json:"newTurnPhase,omitempty" ts:"TurnPhase | undefined"`
	Generation   int        `json:"generation" ts:"number"`
	Trigger      string     `json:"trigger" ts:"string"`
}

// GenerationStartedData contains generation start information
type GenerationStartedData struct {
	Generation   int      `json:"generation" ts:"number"`
	PlayerOrder  []string `json:"playerOrder" ts:"string[]"`
	FirstPlayer  string   `json:"firstPlayer" ts:"string"`
}

// TurnStartedData contains turn start information
type TurnStartedData struct {
	PlayerID           string `json:"playerId" ts:"string"`
	ActionsRemaining   int    `json:"actionsRemaining" ts:"number"`
	TurnIndex          int    `json:"turnIndex" ts:"number"`
	Generation         int    `json:"generation" ts:"number"`
}

// VictoryPointsAwardedData contains VP award information
type VictoryPointsAwardedData struct {
	PlayerID    string         `json:"playerId" ts:"string"`
	Points      int            `json:"points" ts:"number"`
	Source      VPSourceType   `json:"source" ts:"VPSourceType"`
	Description string         `json:"description" ts:"string"`
	Details     *string        `json:"details,omitempty" ts:"string | undefined"`
}

// GameEndedData contains game end information
type GameEndedData struct {
	WinnerID      string              `json:"winnerId" ts:"string"`
	FinalScores   map[string]int      `json:"finalScores" ts:"Record<string, number>"`
	EndCondition  string              `json:"endCondition" ts:"string"`
	Duration      time.Duration       `json:"duration" ts:"number"`
	Generations   int                 `json:"generations" ts:"number"`
}

// EventApplier interface defines how events are applied to game state
type EventApplier interface {
	Apply(event GameEvent, state *GameState) error
}

// EventValidator interface defines how events are validated
type EventValidator interface {
	Validate(event GameEvent, state *GameState) error
}

// BusinessRule represents a business rule that must be enforced
type BusinessRule interface {
	Check(event GameEvent, state *GameState) error
	Name() string
	Description() string
}

// EventFactory helps create properly formatted events
type EventFactory struct {
	gameID string
}

// NewEventFactory creates a new event factory for a game
func NewEventFactory(gameID string) *EventFactory {
	return &EventFactory{gameID: gameID}
}

// CreateEvent creates a new game event with proper metadata
func (ef *EventFactory) CreateEvent(eventType GameEventType, playerID *string, data interface{}) (*GameEvent, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	return &GameEvent{
		ID:        generateEventID(),
		GameID:    ef.gameID,
		Type:      eventType,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Data:      dataBytes,
		Metadata: EventMetadata{
			Tags:  []string{},
			Extra: make(map[string]string),
		},
	}, nil
}

// Helper function to generate unique event IDs
func generateEventID() string {
	// In a real implementation, you'd use a proper ID generator
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}