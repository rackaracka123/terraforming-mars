package usecase

import (
	"encoding/json"
	"fmt"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/repository"
	"time"
)

// GameReplayUseCase handles game replay and time-travel functionality
type GameReplayUseCase struct {
	eventRepo *repository.EventRepository
	gameRepo  *repository.EventSourcedGameRepository
}

// NewGameReplayUseCase creates a new game replay use case
func NewGameReplayUseCase(eventRepo *repository.EventRepository) *GameReplayUseCase {
	return &GameReplayUseCase{
		eventRepo: eventRepo,
		gameRepo:  repository.NewEventSourcedGameRepository(eventRepo),
	}
}

// ReplayStep represents a single step in a game replay
type ReplayStep struct {
	Version     int64            `json:"version" ts:"number"`
	Event       domain.GameEvent `json:"event" ts:"GameEvent"`
	GameState   domain.GameState `json:"gameState" ts:"GameState"`
	Duration    time.Duration    `json:"duration" ts:"number"`
	Description string           `json:"description" ts:"string"`
}

// GameReplay represents a complete game replay
type GameReplay struct {
	GameID      string       `json:"gameId" ts:"string"`
	Steps       []ReplayStep `json:"steps" ts:"ReplayStep[]"`
	TotalEvents int          `json:"totalEvents" ts:"number"`
	Duration    time.Duration `json:"duration" ts:"number"`
	StartedAt   time.Time    `json:"startedAt" ts:"string"`
	CompletedAt time.Time    `json:"completedAt" ts:"string"`
}

// TimelineEvent represents an event in the game timeline
type TimelineEvent struct {
	Version     int64                  `json:"version" ts:"number"`
	Event       domain.GameEvent       `json:"event" ts:"GameEvent"`
	Description string                 `json:"description" ts:"string"`
	Timestamp   time.Time              `json:"timestamp" ts:"string"`
	PlayerName  *string                `json:"playerName,omitempty" ts:"string | undefined"`
	Data        map[string]interface{} `json:"data" ts:"Record<string, any>"`
}

// GameTimeline represents the complete timeline of a game
type GameTimeline struct {
	GameID     string          `json:"gameId" ts:"string"`
	Events     []TimelineEvent `json:"events" ts:"TimelineEvent[]"`
	StartTime  time.Time       `json:"startTime" ts:"string"`
	EndTime    *time.Time      `json:"endTime,omitempty" ts:"string | undefined"`
	Duration   time.Duration   `json:"duration" ts:"number"`
	PlayerNames map[string]string `json:"playerNames" ts:"Record<string, string>"`
}

// ReplayOptions configures replay behavior
type ReplayOptions struct {
	StartVersion    *int64        `json:"startVersion,omitempty" ts:"number | undefined"`
	EndVersion      *int64        `json:"endVersion,omitempty" ts:"number | undefined"`
	IncludeState    bool          `json:"includeState" ts:"boolean"`
	StepDelay       time.Duration `json:"stepDelay" ts:"number"`
	FilterEventTypes []domain.GameEventType `json:"filterEventTypes" ts:"GameEventType[]"`
	PlayerFilter    *string       `json:"playerFilter,omitempty" ts:"string | undefined"`
}

// ReplayGame replays the entire game or a portion of it
func (uc *GameReplayUseCase) ReplayGame(gameID string, options ReplayOptions) (*GameReplay, error) {
	startTime := time.Now()
	
	// Get events based on options
	events, err := uc.getFilteredEvents(gameID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for replay: %w", err)
	}

	if len(events) == 0 {
		return &GameReplay{
			GameID:      gameID,
			Steps:       []ReplayStep{},
			TotalEvents: 0,
			Duration:    0,
			StartedAt:   startTime,
			CompletedAt: time.Now(),
		}, nil
	}

	// Initialize empty game state
	currentState := domain.NewEmptyGameState(gameID)
	var steps []ReplayStep

	// Replay each event
	for i, event := range events {
		stepStart := time.Now()
		
		// Apply event to state
		if err := applyEventToState(event, currentState); err != nil {
			return nil, fmt.Errorf("failed to apply event %s during replay: %w", event.ID, err)
		}

		// Create replay step
		step := ReplayStep{
			Version:     event.Version,
			Event:       event,
			Duration:    time.Since(stepStart),
			Description: uc.getEventDescription(event, currentState),
		}

		// Include game state if requested
		if options.IncludeState {
			step.GameState = *currentState
		}

		steps = append(steps, step)

		// Apply step delay if specified
		if options.StepDelay > 0 {
			time.Sleep(options.StepDelay)
		}

		// Progress callback could be added here for long replays
		if i%100 == 0 && i > 0 {
			// Log progress every 100 events
			fmt.Printf("Replay progress: %d/%d events processed\n", i, len(events))
		}
	}

	return &GameReplay{
		GameID:      gameID,
		Steps:       steps,
		TotalEvents: len(events),
		Duration:    time.Since(startTime),
		StartedAt:   startTime,
		CompletedAt: time.Now(),
	}, nil
}

// GetGameTimeline creates a visual timeline of game events
func (uc *GameReplayUseCase) GetGameTimeline(gameID string) (*GameTimeline, error) {
	events, err := uc.eventRepo.GetEvents(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found for game %s", gameID)
	}

	// Build player names map
	playerNames := make(map[string]string)
	for _, event := range events {
		if event.Type == domain.EventTypePlayerJoined {
			var data domain.PlayerJoinedData
			if err := json.Unmarshal(event.Data, &data); err == nil {
				playerNames[data.PlayerID] = data.PlayerName
			}
		}
	}

	// Create timeline events
	var timelineEvents []TimelineEvent
	for _, event := range events {
		timelineEvent := TimelineEvent{
			Version:     event.Version,
			Event:       event,
			Description: uc.getEventDescription(event, nil),
			Timestamp:   event.Timestamp,
			Data:        make(map[string]interface{}),
		}

		// Add player name if available
		if event.PlayerID != nil {
			if name, exists := playerNames[*event.PlayerID]; exists {
				timelineEvent.PlayerName = &name
			}
		}

		// Extract relevant data based on event type
		timelineEvent.Data = uc.extractEventData(event)

		timelineEvents = append(timelineEvents, timelineEvent)
	}

	var endTime *time.Time
	if len(events) > 0 {
		lastEventTime := events[len(events)-1].Timestamp
		endTime = &lastEventTime
	}

	return &GameTimeline{
		GameID:      gameID,
		Events:      timelineEvents,
		StartTime:   events[0].Timestamp,
		EndTime:     endTime,
		Duration:    endTime.Sub(events[0].Timestamp),
		PlayerNames: playerNames,
	}, nil
}

// GetGameStateAtTime returns the game state at a specific timestamp
func (uc *GameReplayUseCase) GetGameStateAtTime(gameID string, timestamp time.Time) (*domain.GameState, error) {
	events, err := uc.eventRepo.GetEvents(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	// Filter events up to the specified timestamp
	var relevantEvents []domain.GameEvent
	for _, event := range events {
		if event.Timestamp.Before(timestamp) || event.Timestamp.Equal(timestamp) {
			relevantEvents = append(relevantEvents, event)
		} else {
			break
		}
	}

	if len(relevantEvents) == 0 {
		return domain.NewEmptyGameState(gameID), nil
	}

	// Reconstruct state from filtered events
	aggregate, err := domain.NewGameAggregate(gameID, relevantEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct state at timestamp: %w", err)
	}

	return aggregate.GetState(), nil
}

// GetEventsByPlayer returns all events for a specific player
func (uc *GameReplayUseCase) GetEventsByPlayer(gameID string, playerID string) ([]domain.GameEvent, error) {
	return uc.eventRepo.GetEventsByPlayer(gameID, playerID)
}

// GetEventsByType returns all events of a specific type
func (uc *GameReplayUseCase) GetEventsByType(gameID string, eventType domain.GameEventType) ([]domain.GameEvent, error) {
	return uc.eventRepo.GetEventsByType(gameID, eventType)
}

// CompareGameStates compares two game states and returns differences
func (uc *GameReplayUseCase) CompareGameStates(gameID string, version1, version2 int64) (*GameStateDiff, error) {
	state1, err := uc.gameRepo.GetGameAtVersion(gameID, version1)
	if err != nil {
		return nil, fmt.Errorf("failed to get state at version %d: %w", version1, err)
	}

	state2, err := uc.gameRepo.GetGameAtVersion(gameID, version2)
	if err != nil {
		return nil, fmt.Errorf("failed to get state at version %d: %w", version2, err)
	}

	return uc.createGameStateDiff(state1, state2, version1, version2), nil
}

// GameStateDiff represents differences between two game states
type GameStateDiff struct {
	FromVersion        int64                    `json:"fromVersion" ts:"number"`
	ToVersion          int64                    `json:"toVersion" ts:"number"`
	PlayerChanges      []PlayerDiff             `json:"playerChanges" ts:"PlayerDiff[]"`
	GlobalParamChanges *GlobalParametersDiff    `json:"globalParamChanges,omitempty" ts:"GlobalParametersDiff | undefined"`
	PhaseChanges       *PhaseChange             `json:"phaseChanges,omitempty" ts:"PhaseChange | undefined"`
	BoardChanges       []BoardChange            `json:"boardChanges" ts:"BoardChange[]"`
	EventsSummary      map[string]int           `json:"eventsSummary" ts:"Record<string, number>"`
}

type PlayerDiff struct {
	PlayerID        string                   `json:"playerId" ts:"string"`
	ResourceChanges *domain.ResourcesMap     `json:"resourceChanges,omitempty" ts:"ResourcesMap | undefined"`
	ProductionChanges *domain.ResourcesMap   `json:"productionChanges,omitempty" ts:"ResourcesMap | undefined"`
	TRChange        int                      `json:"trChange" ts:"number"`
	VPChange        int                      `json:"vpChange" ts:"number"`
	CardsPlayed     []string                 `json:"cardsPlayed" ts:"string[]"`
	TilesPlaced     []domain.HexCoordinate   `json:"tilesPlaced" ts:"HexCoordinate[]"`
}

type GlobalParametersDiff struct {
	TemperatureChange int `json:"temperatureChange" ts:"number"`
	OxygenChange      int `json:"oxygenChange" ts:"number"`
	OceansChange      int `json:"oceansChange" ts:"number"`
}

type PhaseChange struct {
	OldPhase     domain.GamePhase  `json:"oldPhase" ts:"GamePhase"`
	NewPhase     domain.GamePhase  `json:"newPhase" ts:"GamePhase"`
	OldTurnPhase *domain.TurnPhase `json:"oldTurnPhase,omitempty" ts:"TurnPhase | undefined"`
	NewTurnPhase *domain.TurnPhase `json:"newTurnPhase,omitempty" ts:"TurnPhase | undefined"`
}

type BoardChange struct {
	Position  domain.HexCoordinate `json:"position" ts:"HexCoordinate"`
	TileType  domain.TileType      `json:"tileType" ts:"TileType"`
	PlayerID  string               `json:"playerId" ts:"string"`
	Action    string               `json:"action" ts:"string"` // "placed", "removed"
}

// DebugEventStream provides detailed debugging information for events
func (uc *GameReplayUseCase) DebugEventStream(gameID string, startVersion, endVersion int64) (*EventStreamDebug, error) {
	events, err := uc.eventRepo.GetEventsFromVersion(gameID, startVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	// Filter to end version
	var filteredEvents []domain.GameEvent
	for _, event := range events {
		if event.Version <= endVersion {
			filteredEvents = append(filteredEvents, event)
		}
	}

	var debugSteps []EventDebugStep
	currentState := domain.NewEmptyGameState(gameID)

	// Apply events up to start version first
	if startVersion > 1 {
		initialEvents, err := uc.eventRepo.GetEventsToVersion(gameID, startVersion-1)
		if err != nil {
			return nil, fmt.Errorf("failed to get initial events: %w", err)
		}

		for _, event := range initialEvents {
			if err := applyEventToState(event, currentState); err != nil {
				return nil, fmt.Errorf("failed to apply initial event: %w", err)
			}
		}
	}

	// Now process each event in the range with detailed debugging
	for _, event := range filteredEvents {
		beforeState := uc.cloneGameState(currentState)
		
		err := applyEventToState(event, currentState)
		
		debugStep := EventDebugStep{
			Event:        event,
			BeforeState:  beforeState,
			AfterState:   uc.cloneGameState(currentState),
			Applied:      err == nil,
			Error:        nil,
			ChangeSummary: uc.summarizeChanges(beforeState, currentState),
		}

		if err != nil {
			errorStr := err.Error()
			debugStep.Error = &errorStr
		}

		debugSteps = append(debugSteps, debugStep)
	}

	return &EventStreamDebug{
		GameID:     gameID,
		StartVersion: startVersion,
		EndVersion: endVersion,
		Steps:      debugSteps,
		TotalEvents: len(debugSteps),
	}, nil
}

// EventStreamDebug provides detailed debugging information
type EventStreamDebug struct {
	GameID      string            `json:"gameId" ts:"string"`
	StartVersion int64            `json:"startVersion" ts:"number"`
	EndVersion  int64             `json:"endVersion" ts:"number"`
	Steps       []EventDebugStep  `json:"steps" ts:"EventDebugStep[]"`
	TotalEvents int               `json:"totalEvents" ts:"number"`
}

type EventDebugStep struct {
	Event         domain.GameEvent  `json:"event" ts:"GameEvent"`
	BeforeState   *domain.GameState `json:"beforeState" ts:"GameState"`
	AfterState    *domain.GameState `json:"afterState" ts:"GameState"`
	Applied       bool              `json:"applied" ts:"boolean"`
	Error         *string           `json:"error,omitempty" ts:"string | undefined"`
	ChangeSummary string            `json:"changeSummary" ts:"string"`
}

// Helper functions

func (uc *GameReplayUseCase) getFilteredEvents(gameID string, options ReplayOptions) ([]domain.GameEvent, error) {
	allEvents, err := uc.eventRepo.GetEvents(gameID)
	if err != nil {
		return nil, err
	}

	var filteredEvents []domain.GameEvent

	for _, event := range allEvents {
		// Version filter
		if options.StartVersion != nil && event.Version < *options.StartVersion {
			continue
		}
		if options.EndVersion != nil && event.Version > *options.EndVersion {
			continue
		}

		// Event type filter
		if len(options.FilterEventTypes) > 0 {
			typeMatch := false
			for _, filterType := range options.FilterEventTypes {
				if event.Type == filterType {
					typeMatch = true
					break
				}
			}
			if !typeMatch {
				continue
			}
		}

		// Player filter
		if options.PlayerFilter != nil && event.PlayerID != nil && *event.PlayerID != *options.PlayerFilter {
			continue
		}

		filteredEvents = append(filteredEvents, event)
	}

	return filteredEvents, nil
}

func (uc *GameReplayUseCase) getEventDescription(event domain.GameEvent, state *domain.GameState) string {
	switch event.Type {
	case domain.EventTypeGameCreated:
		return "Game created"
	case domain.EventTypePlayerJoined:
		var data domain.PlayerJoinedData
		if json.Unmarshal(event.Data, &data) == nil {
			return fmt.Sprintf("%s joined the game", data.PlayerName)
		}
		return "Player joined"
	case domain.EventTypeCorporationSelected:
		var data domain.CorporationSelectedData
		if json.Unmarshal(event.Data, &data) == nil {
			return fmt.Sprintf("Selected %s corporation", data.CorporationID)
		}
		return "Corporation selected"
	case domain.EventTypeCardPlayed:
		var data domain.CardPlayedData
		if json.Unmarshal(event.Data, &data) == nil {
			return fmt.Sprintf("Played card %s for %d M€", data.CardID, data.Cost)
		}
		return "Card played"
	case domain.EventTypeParameterIncreased:
		var data domain.ParameterIncreasedData
		if json.Unmarshal(event.Data, &data) == nil {
			return fmt.Sprintf("Increased %s by %d steps", data.Parameter, data.Steps)
		}
		return "Global parameter increased"
	case domain.EventTypeMilestoneClaimed:
		var data domain.MilestoneClaimedData
		if json.Unmarshal(event.Data, &data) == nil {
			return fmt.Sprintf("Claimed %s milestone", data.MilestoneID)
		}
		return "Milestone claimed"
	default:
		return fmt.Sprintf("Event: %s", event.Type)
	}
}

func (uc *GameReplayUseCase) extractEventData(event domain.GameEvent) map[string]interface{} {
	data := make(map[string]interface{})
	
	// Common fields
	data["type"] = event.Type
	data["version"] = event.Version
	
	// Type-specific data extraction
	switch event.Type {
	case domain.EventTypeCardPlayed:
		var cardData domain.CardPlayedData
		if json.Unmarshal(event.Data, &cardData) == nil {
			data["cardId"] = cardData.CardID
			data["cost"] = cardData.Cost
		}
	case domain.EventTypeParameterIncreased:
		var paramData domain.ParameterIncreasedData
		if json.Unmarshal(event.Data, &paramData) == nil {
			data["parameter"] = paramData.Parameter
			data["steps"] = paramData.Steps
			data["newValue"] = paramData.NewValue
		}
	case domain.EventTypeMilestoneClaimed:
		var milestoneData domain.MilestoneClaimedData
		if json.Unmarshal(event.Data, &milestoneData) == nil {
			data["milestoneId"] = milestoneData.MilestoneID
			data["cost"] = milestoneData.Cost
		}
	}
	
	return data
}

func (uc *GameReplayUseCase) createGameStateDiff(state1, state2 *domain.GameState, version1, version2 int64) *GameStateDiff {
	diff := &GameStateDiff{
		FromVersion:   version1,
		ToVersion:     version2,
		EventsSummary: make(map[string]int),
	}

	// Compare global parameters
	if state1.GlobalParameters.Temperature != state2.GlobalParameters.Temperature ||
		state1.GlobalParameters.Oxygen != state2.GlobalParameters.Oxygen ||
		state1.GlobalParameters.Oceans != state2.GlobalParameters.Oceans {
		diff.GlobalParamChanges = &GlobalParametersDiff{
			TemperatureChange: state2.GlobalParameters.Temperature - state1.GlobalParameters.Temperature,
			OxygenChange:      state2.GlobalParameters.Oxygen - state1.GlobalParameters.Oxygen,
			OceansChange:      state2.GlobalParameters.Oceans - state1.GlobalParameters.Oceans,
		}
	}

	// Compare phases
	if state1.Phase != state2.Phase || 
		(state1.TurnPhase == nil) != (state2.TurnPhase == nil) ||
		(state1.TurnPhase != nil && state2.TurnPhase != nil && *state1.TurnPhase != *state2.TurnPhase) {
		diff.PhaseChanges = &PhaseChange{
			OldPhase:     state1.Phase,
			NewPhase:     state2.Phase,
			OldTurnPhase: state1.TurnPhase,
			NewTurnPhase: state2.TurnPhase,
		}
	}

	// Compare players
	for i, player2 := range state2.Players {
		if i < len(state1.Players) {
			player1 := state1.Players[i]
			playerDiff := uc.createPlayerDiff(player1, player2)
			if playerDiff != nil {
				diff.PlayerChanges = append(diff.PlayerChanges, *playerDiff)
			}
		}
	}

	return diff
}

func (uc *GameReplayUseCase) createPlayerDiff(player1, player2 domain.Player) *PlayerDiff {
	diff := &PlayerDiff{
		PlayerID: player2.ID,
		TRChange: player2.TerraformRating - player1.TerraformRating,
		VPChange: player2.VictoryPoints - player1.VictoryPoints,
	}

	// Check if there are any actual changes
	hasChanges := false

	// Resource changes
	if player1.Resources != player2.Resources {
		diff.ResourceChanges = &domain.ResourcesMap{
			Credits:  player2.Resources.Credits - player1.Resources.Credits,
			Steel:    player2.Resources.Steel - player1.Resources.Steel,
			Titanium: player2.Resources.Titanium - player1.Resources.Titanium,
			Plants:   player2.Resources.Plants - player1.Resources.Plants,
			Energy:   player2.Resources.Energy - player1.Resources.Energy,
			Heat:     player2.Resources.Heat - player1.Resources.Heat,
		}
		hasChanges = true
	}

	// Production changes
	if player1.Production != player2.Production {
		diff.ProductionChanges = &domain.ResourcesMap{
			Credits:  player2.Production.Credits - player1.Production.Credits,
			Steel:    player2.Production.Steel - player1.Production.Steel,
			Titanium: player2.Production.Titanium - player1.Production.Titanium,
			Plants:   player2.Production.Plants - player1.Production.Plants,
			Energy:   player2.Production.Energy - player1.Production.Energy,
			Heat:     player2.Production.Heat - player1.Production.Heat,
		}
		hasChanges = true
	}

	// Cards played
	for _, card := range player2.PlayedCards {
		found := false
		for _, oldCard := range player1.PlayedCards {
			if oldCard == card {
				found = true
				break
			}
		}
		if !found {
			diff.CardsPlayed = append(diff.CardsPlayed, card)
			hasChanges = true
		}
	}

	// Tiles placed
	for _, tile := range player2.TilePositions {
		found := false
		for _, oldTile := range player1.TilePositions {
			if oldTile == tile {
				found = true
				break
			}
		}
		if !found {
			diff.TilesPlaced = append(diff.TilesPlaced, tile)
			hasChanges = true
		}
	}

	if diff.TRChange != 0 || diff.VPChange != 0 {
		hasChanges = true
	}

	if hasChanges {
		return diff
	}
	return nil
}

func (uc *GameReplayUseCase) cloneGameState(state *domain.GameState) *domain.GameState {
	// In a real implementation, you'd do a proper deep clone
	// For now, return a reference (this is a simplified version)
	return state
}

func (uc *GameReplayUseCase) summarizeChanges(before, after *domain.GameState) string {
	changes := []string{}

	// Check phase changes
	if before.Phase != after.Phase {
		changes = append(changes, fmt.Sprintf("Phase: %s → %s", before.Phase, after.Phase))
	}

	// Check global parameter changes
	if before.GlobalParameters.Temperature != after.GlobalParameters.Temperature {
		changes = append(changes, fmt.Sprintf("Temperature: %d → %d", 
			before.GlobalParameters.Temperature, after.GlobalParameters.Temperature))
	}

	if before.GlobalParameters.Oxygen != after.GlobalParameters.Oxygen {
		changes = append(changes, fmt.Sprintf("Oxygen: %d → %d", 
			before.GlobalParameters.Oxygen, after.GlobalParameters.Oxygen))
	}

	if before.GlobalParameters.Oceans != after.GlobalParameters.Oceans {
		changes = append(changes, fmt.Sprintf("Oceans: %d → %d", 
			before.GlobalParameters.Oceans, after.GlobalParameters.Oceans))
	}

	if len(changes) == 0 {
		return "No major state changes"
	}

	return fmt.Sprintf("%d changes: %v", len(changes), changes)
}

// Helper function to apply event to state (imported from aggregate.go)
func applyEventToState(event domain.GameEvent, state *domain.GameState) error {
	switch event.Type {
	case domain.EventTypeGameCreated:
		return applyGameCreated(event, state)
	case domain.EventTypePlayerJoined:
		return applyPlayerJoined(event, state)
	case domain.EventTypeCorporationSelected:
		return applyCorporationSelected(event, state)
	case domain.EventTypeCardPlayed:
		return applyCardPlayed(event, state)
	case domain.EventTypeTilePlaced:
		return applyTilePlaced(event, state)
	case domain.EventTypeResourcesGained:
		return applyResourcesChanged(event, state, true)
	case domain.EventTypeResourcesLost:
		return applyResourcesChanged(event, state, false)
	case domain.EventTypeProductionChanged:
		return applyProductionChanged(event, state)
	case domain.EventTypeParameterIncreased:
		return applyParameterIncreased(event, state)
	case domain.EventTypeMilestoneClaimed:
		return applyMilestoneClaimed(event, state)
	case domain.EventTypeAwardFunded:
		return applyAwardFunded(event, state)
	case domain.EventTypePhaseChanged:
		return applyPhaseChanged(event, state)
	case domain.EventTypeGenerationStarted:
		return applyGenerationStarted(event, state)
	case domain.EventTypeTurnStarted:
		return applyTurnStarted(event, state)
	case domain.EventTypeVictoryPointsAwarded:
		return applyVictoryPointsAwarded(event, state)
	case domain.EventTypeGameEnded:
		return applyGameEnded(event, state)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}