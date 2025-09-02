package usecase

import (
	"fmt"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/repository"
)

// EventSourcedGameUseCase handles game business logic using event sourcing
type EventSourcedGameUseCase struct {
	eventRepo    *repository.EventRepository
	gameRepo     *repository.EventSourcedGameRepository
	validator    domain.EventValidator
	businessRules []domain.BusinessRule
}

// NewEventSourcedGameUseCase creates a new event-sourced game use case
func NewEventSourcedGameUseCase(eventRepo *repository.EventRepository) *EventSourcedGameUseCase {
	gameRepo := repository.NewEventSourcedGameRepository(eventRepo)
	
	return &EventSourcedGameUseCase{
		eventRepo:     eventRepo,
		gameRepo:      gameRepo,
		validator:     &GameEventValidator{},
		businessRules: GetBusinessRules(),
	}
}

// CreateGame creates a new game
func (uc *EventSourcedGameUseCase) CreateGame(gameID string, settings domain.GameSettings, createdBy string) (*domain.GameState, error) {
	return uc.gameRepo.CreateGame(gameID, settings, createdBy)
}

// GetGame retrieves the current game state
func (uc *EventSourcedGameUseCase) GetGame(gameID string) (*domain.GameState, error) {
	return uc.gameRepo.GetGame(gameID)
}

// JoinGame adds a player to the game
func (uc *EventSourcedGameUseCase) JoinGame(gameID, playerID, playerName string) (*domain.GameState, error) {
	// Get current game state for validation
	currentState, err := uc.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	// Create the event
	factory := domain.NewEventFactory(gameID)
	event, err := factory.CreateEvent(
		domain.EventTypePlayerJoined,
		&playerID,
		domain.PlayerJoinedData{
			PlayerID:   playerID,
			PlayerName: playerName,
			JoinOrder:  len(currentState.Players) + 1,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create player joined event: %w", err)
	}

	// Validate the event
	if err := uc.validateEvent(*event, currentState); err != nil {
		return nil, fmt.Errorf("event validation failed: %w", err)
	}

	// Save the event and return updated state
	return uc.gameRepo.SaveEvent(*event)
}

// SelectCorporation assigns a corporation to a player
func (uc *EventSourcedGameUseCase) SelectCorporation(gameID, playerID, corporationID string) (*domain.GameState, error) {
	// Get current game state for validation
	currentState, err := uc.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	// Validate that the player exists and hasn't already selected a corporation
	var playerFound bool
	for _, player := range currentState.Players {
		if player.ID == playerID {
			playerFound = true
			if player.Corporation != nil {
				return nil, fmt.Errorf("player %s has already selected a corporation", playerID)
			}
			break
		}
	}

	if !playerFound {
		return nil, fmt.Errorf("player %s not found in game %s", playerID, gameID)
	}

	// Create the event
	factory := domain.NewEventFactory(gameID)
	event, err := factory.CreateEvent(
		domain.EventTypeCorporationSelected,
		&playerID,
		domain.CorporationSelectedData{
			PlayerID:      playerID,
			CorporationID: corporationID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create corporation selected event: %w", err)
	}

	// Validate the event
	if err := uc.validateEvent(*event, currentState); err != nil {
		return nil, fmt.Errorf("event validation failed: %w", err)
	}

	// Save the event and return updated state
	return uc.gameRepo.SaveEvent(*event)
}

// RaiseTemperature increases global temperature using heat
func (uc *EventSourcedGameUseCase) RaiseTemperature(gameID, playerID string) (*domain.GameState, error) {
	// Get current game state for validation
	currentState, err := uc.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	// Find the player and validate
	var player *domain.Player
	for i := range currentState.Players {
		if currentState.Players[i].ID == playerID {
			player = &currentState.Players[i]
			break
		}
	}

	if player == nil {
		return nil, fmt.Errorf("player %s not found in game %s", playerID, gameID)
	}

	// Business rule validations
	if currentState.CurrentPlayer != playerID {
		return nil, fmt.Errorf("not player's turn")
	}

	if player.Resources.Heat < 8 {
		return nil, fmt.Errorf("not enough heat (need 8, have %d)", player.Resources.Heat)
	}

	if player.ActionsRemaining <= 0 {
		return nil, fmt.Errorf("no actions remaining")
	}

	if currentState.GlobalParameters.Temperature >= 8 {
		return nil, fmt.Errorf("temperature already at maximum")
	}

	// Create multiple events for this complex action
	factory := domain.NewEventFactory(gameID)
	var events []domain.GameEvent

	// 1. Resource loss event
	resourceLossEvent, err := factory.CreateEvent(
		domain.EventTypeResourcesLost,
		&playerID,
		domain.ResourcesChangedData{
			PlayerID: playerID,
			Changes: domain.ResourcesMap{
				Heat: 8,
			},
			NewTotals: domain.ResourcesMap{
				Credits:  player.Resources.Credits,
				Steel:    player.Resources.Steel,
				Titanium: player.Resources.Titanium,
				Plants:   player.Resources.Plants,
				Energy:   player.Resources.Energy,
				Heat:     player.Resources.Heat - 8,
			},
			Reason: "Convert heat to raise temperature",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource loss event: %w", err)
	}
	events = append(events, *resourceLossEvent)

	// 2. Parameter increase event
	parameterEvent, err := factory.CreateEvent(
		domain.EventTypeParameterIncreased,
		&playerID,
		domain.ParameterIncreasedData{
			PlayerID:     playerID,
			Parameter:    domain.GlobalParamTemperature,
			OldValue:     currentState.GlobalParameters.Temperature,
			NewValue:     currentState.GlobalParameters.Temperature + 2,
			Steps:        1,
			TRIncrease:   1,
			BonusRewards: []domain.CardEffect{},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create parameter increase event: %w", err)
	}
	events = append(events, *parameterEvent)

	// 3. Action performed event
	actionEvent, err := factory.CreateEvent(
		domain.EventTypeActionPerformed,
		&playerID,
		map[string]interface{}{
			"actionType":      "raise_temperature",
			"actionsUsed":     1,
			"actionsRemaining": player.ActionsRemaining - 1,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create action performed event: %w", err)
	}
	events = append(events, *actionEvent)

	// Validate all events
	for _, event := range events {
		if err := uc.validateEvent(event, currentState); err != nil {
			return nil, fmt.Errorf("event validation failed: %w", err)
		}
	}

	// Save all events atomically
	return uc.gameRepo.SaveEvents(gameID, events)
}

// PlayCard plays a project card
func (uc *EventSourcedGameUseCase) PlayCard(gameID, playerID, cardID string, placement *domain.HexCoordinate) (*domain.GameState, error) {
	// Get current game state for validation
	currentState, err := uc.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	// Find the player and validate
	var player *domain.Player
	for i := range currentState.Players {
		if currentState.Players[i].ID == playerID {
			player = &currentState.Players[i]
			break
		}
	}

	if player == nil {
		return nil, fmt.Errorf("player %s not found in game %s", playerID, gameID)
	}

	// Business rule validations
	if currentState.CurrentPlayer != playerID {
		return nil, fmt.Errorf("not player's turn")
	}

	if player.ActionsRemaining <= 0 {
		return nil, fmt.Errorf("no actions remaining")
	}

	// Check if player has the card in hand
	var hasCard bool
	for _, handCardID := range player.Hand {
		if handCardID == cardID {
			hasCard = true
			break
		}
	}

	if !hasCard {
		return nil, fmt.Errorf("player does not have card %s in hand", cardID)
	}

	// TODO: In a real implementation, we would:
	// 1. Look up the card definition
	// 2. Check requirements are met
	// 3. Calculate actual cost with discounts
	// 4. Validate placement if required
	// 5. Apply immediate effects

	// Create the card played event
	factory := domain.NewEventFactory(gameID)
	event, err := factory.CreateEvent(
		domain.EventTypeCardPlayed,
		&playerID,
		domain.CardPlayedData{
			PlayerID: playerID,
			CardID:   cardID,
			Cost:     15, // Placeholder cost
			ResourcesSpent: domain.ResourcesMap{
				Credits: 15,
			},
			Placement:        placement,
			Requirements:     []domain.Requirement{},
			ImmediateEffects: []domain.CardEffect{},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create card played event: %w", err)
	}

	// Validate the event
	if err := uc.validateEvent(*event, currentState); err != nil {
		return nil, fmt.Errorf("event validation failed: %w", err)
	}

	// Save the event and return updated state
	return uc.gameRepo.SaveEvent(*event)
}

// ClaimMilestone claims a milestone for a player
func (uc *EventSourcedGameUseCase) ClaimMilestone(gameID, playerID, milestoneID string) (*domain.GameState, error) {
	// Get current game state for validation
	currentState, err := uc.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	// Find the milestone
	var milestone *domain.Milestone
	for i := range currentState.Milestones {
		if currentState.Milestones[i].ID == milestoneID {
			milestone = &currentState.Milestones[i]
			break
		}
	}

	if milestone == nil {
		return nil, fmt.Errorf("milestone %s not found", milestoneID)
	}

	if milestone.ClaimedBy != nil {
		return nil, fmt.Errorf("milestone %s already claimed by %s", milestoneID, *milestone.ClaimedBy)
	}

	// Count already claimed milestones
	claimedCount := 0
	for _, m := range currentState.Milestones {
		if m.ClaimedBy != nil {
			claimedCount++
		}
	}

	if claimedCount >= 3 {
		return nil, fmt.Errorf("maximum of 3 milestones can be claimed")
	}

	// Find the player and validate
	var player *domain.Player
	for i := range currentState.Players {
		if currentState.Players[i].ID == playerID {
			player = &currentState.Players[i]
			break
		}
	}

	if player == nil {
		return nil, fmt.Errorf("player %s not found", playerID)
	}

	if player.Resources.Credits < milestone.Cost {
		return nil, fmt.Errorf("not enough credits (need %d, have %d)", milestone.Cost, player.Resources.Credits)
	}

	// TODO: Validate milestone requirements are met

	// Create the milestone claimed event
	factory := domain.NewEventFactory(gameID)
	event, err := factory.CreateEvent(
		domain.EventTypeMilestoneClaimed,
		&playerID,
		domain.MilestoneClaimedData{
			PlayerID:     playerID,
			MilestoneID:  milestoneID,
			Cost:         milestone.Cost,
			Requirements: []domain.Requirement{},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create milestone claimed event: %w", err)
	}

	// Validate the event
	if err := uc.validateEvent(*event, currentState); err != nil {
		return nil, fmt.Errorf("event validation failed: %w", err)
	}

	// Save the event and return updated state
	return uc.gameRepo.SaveEvent(*event)
}

// SkipAction passes the current player's turn
func (uc *EventSourcedGameUseCase) SkipAction(gameID, playerID string) (*domain.GameState, error) {
	// Get current game state for validation
	currentState, err := uc.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game state: %w", err)
	}

	// Business rule validations
	if currentState.CurrentPlayer != playerID {
		return nil, fmt.Errorf("not player's turn")
	}

	// Find the next player
	currentIndex := -1
	for i, player := range currentState.Players {
		if player.ID == playerID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return nil, fmt.Errorf("current player not found")
	}

	nextIndex := (currentIndex + 1) % len(currentState.Players)
	nextPlayerID := currentState.Players[nextIndex].ID

	// Create events for passing turn
	factory := domain.NewEventFactory(gameID)
	var events []domain.GameEvent

	// 1. Current turn ended event
	turnEndedEvent, err := factory.CreateEvent(
		domain.EventTypeTurnEnded,
		&playerID,
		map[string]interface{}{
			"playerID": playerID,
			"passed":   true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create turn ended event: %w", err)
	}
	events = append(events, *turnEndedEvent)

	// 2. Next turn started event
	turnStartedEvent, err := factory.CreateEvent(
		domain.EventTypeTurnStarted,
		&nextPlayerID,
		domain.TurnStartedData{
			PlayerID:         nextPlayerID,
			ActionsRemaining: 2, // Standard actions per turn
			TurnIndex:        currentState.Turn + 1,
			Generation:       currentState.Generation,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create turn started event: %w", err)
	}
	events = append(events, *turnStartedEvent)

	// Validate all events
	for _, event := range events {
		if err := uc.validateEvent(event, currentState); err != nil {
			return nil, fmt.Errorf("event validation failed: %w", err)
		}
	}

	// Save all events atomically
	return uc.gameRepo.SaveEvents(gameID, events)
}

// GetGameHistory returns the event history for a game
func (uc *EventSourcedGameUseCase) GetGameHistory(gameID string) ([]domain.GameEvent, error) {
	return uc.gameRepo.GetGameHistory(gameID)
}

// GetGameAtVersion returns the game state at a specific version
func (uc *EventSourcedGameUseCase) GetGameAtVersion(gameID string, version int64) (*domain.GameState, error) {
	return uc.gameRepo.GetGameAtVersion(gameID, version)
}

// ReplayGame replays the game from the beginning
func (uc *EventSourcedGameUseCase) ReplayGame(gameID string) ([]domain.GameState, error) {
	events, err := uc.gameRepo.GetGameHistory(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game history: %w", err)
	}

	var states []domain.GameState
	for i := 1; i <= len(events); i++ {
		state, err := uc.gameRepo.GetGameAtVersion(gameID, int64(i))
		if err != nil {
			return nil, fmt.Errorf("failed to get game state at version %d: %w", i, err)
		}
		states = append(states, *state)
	}

	return states, nil
}

// validateEvent validates an event against the current game state
func (uc *EventSourcedGameUseCase) validateEvent(event domain.GameEvent, currentState *domain.GameState) error {
	// Run event validator
	if err := uc.validator.Validate(event, currentState); err != nil {
		return fmt.Errorf("event validation failed: %w", err)
	}

	// Run business rules
	for _, rule := range uc.businessRules {
		if err := rule.Check(event, currentState); err != nil {
			return fmt.Errorf("business rule '%s' failed: %w", rule.Name(), err)
		}
	}

	return nil
}