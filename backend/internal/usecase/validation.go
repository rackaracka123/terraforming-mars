package usecase

import (
	"encoding/json"
	"fmt"
	"terraforming-mars-backend/internal/domain"
)

// GameEventValidator implements event validation
type GameEventValidator struct{}

// Validate validates an event against the current game state
func (v *GameEventValidator) Validate(event domain.GameEvent, state *domain.GameState) error {
	// Basic validation
	if event.GameID == "" {
		return fmt.Errorf("event must have a game ID")
	}

	if event.GameID != state.ID {
		return fmt.Errorf("event game ID %s does not match state game ID %s", event.GameID, state.ID)
	}

	if event.Type == "" {
		return fmt.Errorf("event must have a type")
	}

	if event.Timestamp.IsZero() {
		return fmt.Errorf("event must have a timestamp")
	}

	// Type-specific validation
	switch event.Type {
	case domain.EventTypePlayerJoined:
		return v.validatePlayerJoined(event, state)
	case domain.EventTypeCorporationSelected:
		return v.validateCorporationSelected(event, state)
	case domain.EventTypeCardPlayed:
		return v.validateCardPlayed(event, state)
	case domain.EventTypeMilestoneClaimed:
		return v.validateMilestoneClaimed(event, state)
	case domain.EventTypeParameterIncreased:
		return v.validateParameterIncreased(event, state)
	case domain.EventTypeResourcesLost:
		return v.validateResourcesLost(event, state)
	case domain.EventTypeResourcesGained:
		return v.validateResourcesGained(event, state)
	default:
		// For events we don't have specific validation for, just pass
		return nil
	}
}

func (v *GameEventValidator) validatePlayerJoined(event domain.GameEvent, state *domain.GameState) error {
	var data domain.PlayerJoinedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return fmt.Errorf("invalid player joined data: %w", err)
	}

	if data.PlayerID == "" {
		return fmt.Errorf("player ID cannot be empty")
	}

	if data.PlayerName == "" {
		return fmt.Errorf("player name cannot be empty")
	}

	// Check if player already exists
	for _, player := range state.Players {
		if player.ID == data.PlayerID {
			return fmt.Errorf("player %s already exists in game", data.PlayerID)
		}
	}

	// Check game capacity (assuming max 4 players)
	if len(state.Players) >= 4 {
		return fmt.Errorf("game is full (max 4 players)")
	}

	// Check if game is in correct phase
	if state.Phase != domain.GamePhaseSetup {
		return fmt.Errorf("cannot join game in phase %s", state.Phase)
	}

	return nil
}

func (v *GameEventValidator) validateCorporationSelected(event domain.GameEvent, state *domain.GameState) error {
	var data domain.CorporationSelectedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return fmt.Errorf("invalid corporation selected data: %w", err)
	}

	if data.PlayerID == "" {
		return fmt.Errorf("player ID cannot be empty")
	}

	if data.CorporationID == "" {
		return fmt.Errorf("corporation ID cannot be empty")
	}

	// Check if player exists
	var playerFound bool
	for _, player := range state.Players {
		if player.ID == data.PlayerID {
			playerFound = true
			if player.Corporation != nil {
				return fmt.Errorf("player %s already has a corporation", data.PlayerID)
			}
			break
		}
	}

	if !playerFound {
		return fmt.Errorf("player %s not found", data.PlayerID)
	}

	// Validate corporation exists
	corporations := domain.GetBaseCorporations()
	var corpFound bool
	for _, corp := range corporations {
		if corp.ID == data.CorporationID {
			corpFound = true
			break
		}
	}

	if !corpFound {
		return fmt.Errorf("corporation %s not found", data.CorporationID)
	}

	return nil
}

func (v *GameEventValidator) validateCardPlayed(event domain.GameEvent, state *domain.GameState) error {
	var data domain.CardPlayedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return fmt.Errorf("invalid card played data: %w", err)
	}

	if data.PlayerID == "" {
		return fmt.Errorf("player ID cannot be empty")
	}

	if data.CardID == "" {
		return fmt.Errorf("card ID cannot be empty")
	}

	if data.Cost < 0 {
		return fmt.Errorf("card cost cannot be negative")
	}

	// Check if player exists and it's their turn
	var player *domain.Player
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			player = &state.Players[i]
			break
		}
	}

	if player == nil {
		return fmt.Errorf("player %s not found", data.PlayerID)
	}

	if state.CurrentPlayer != data.PlayerID {
		return fmt.Errorf("not player %s's turn", data.PlayerID)
	}

	if player.ActionsRemaining <= 0 {
		return fmt.Errorf("player %s has no actions remaining", data.PlayerID)
	}

	// Check if player has the card in hand
	var hasCard bool
	for _, cardID := range player.Hand {
		if cardID == data.CardID {
			hasCard = true
			break
		}
	}

	if !hasCard {
		return fmt.Errorf("player %s does not have card %s in hand", data.PlayerID, data.CardID)
	}

	// Validate resources spent
	totalCost := data.ResourcesSpent.Credits + data.ResourcesSpent.Steel + data.ResourcesSpent.Titanium +
		data.ResourcesSpent.Plants + data.ResourcesSpent.Energy + data.ResourcesSpent.Heat

	if totalCost != data.Cost {
		return fmt.Errorf("resources spent (%d) do not match card cost (%d)", totalCost, data.Cost)
	}

	// Check if player has enough resources
	if player.Resources.Credits < data.ResourcesSpent.Credits ||
		player.Resources.Steel < data.ResourcesSpent.Steel ||
		player.Resources.Titanium < data.ResourcesSpent.Titanium ||
		player.Resources.Plants < data.ResourcesSpent.Plants ||
		player.Resources.Energy < data.ResourcesSpent.Energy ||
		player.Resources.Heat < data.ResourcesSpent.Heat {
		return fmt.Errorf("player %s does not have enough resources", data.PlayerID)
	}

	return nil
}

func (v *GameEventValidator) validateMilestoneClaimed(event domain.GameEvent, state *domain.GameState) error {
	var data domain.MilestoneClaimedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return fmt.Errorf("invalid milestone claimed data: %w", err)
	}

	if data.PlayerID == "" {
		return fmt.Errorf("player ID cannot be empty")
	}

	if data.MilestoneID == "" {
		return fmt.Errorf("milestone ID cannot be empty")
	}

	// Check if milestone exists and is unclaimed
	var milestone *domain.Milestone
	for i := range state.Milestones {
		if state.Milestones[i].ID == data.MilestoneID {
			milestone = &state.Milestones[i]
			break
		}
	}

	if milestone == nil {
		return fmt.Errorf("milestone %s not found", data.MilestoneID)
	}

	if milestone.ClaimedBy != nil {
		return fmt.Errorf("milestone %s already claimed by %s", data.MilestoneID, *milestone.ClaimedBy)
	}

	// Check milestone limit (max 3)
	claimedCount := 0
	for _, m := range state.Milestones {
		if m.ClaimedBy != nil {
			claimedCount++
		}
	}

	if claimedCount >= 3 {
		return fmt.Errorf("maximum of 3 milestones can be claimed")
	}

	// Check if player exists and has enough credits
	var player *domain.Player
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			player = &state.Players[i]
			break
		}
	}

	if player == nil {
		return fmt.Errorf("player %s not found", data.PlayerID)
	}

	if player.Resources.Credits < data.Cost {
		return fmt.Errorf("player %s does not have enough credits (%d needed, %d available)", 
			data.PlayerID, data.Cost, player.Resources.Credits)
	}

	return nil
}

func (v *GameEventValidator) validateParameterIncreased(event domain.GameEvent, state *domain.GameState) error {
	var data domain.ParameterIncreasedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return fmt.Errorf("invalid parameter increased data: %w", err)
	}

	if data.PlayerID == "" {
		return fmt.Errorf("player ID cannot be empty")
	}

	if data.Steps <= 0 {
		return fmt.Errorf("steps must be positive")
	}

	// Validate parameter bounds
	switch data.Parameter {
	case domain.GlobalParamTemperature:
		if data.NewValue < -30 || data.NewValue > 8 {
			return fmt.Errorf("temperature must be between -30 and 8, got %d", data.NewValue)
		}
		if data.OldValue != state.GlobalParameters.Temperature {
			return fmt.Errorf("old temperature value %d does not match current %d", 
				data.OldValue, state.GlobalParameters.Temperature)
		}
	case domain.GlobalParamOxygen:
		if data.NewValue < 0 || data.NewValue > 14 {
			return fmt.Errorf("oxygen must be between 0 and 14, got %d", data.NewValue)
		}
		if data.OldValue != state.GlobalParameters.Oxygen {
			return fmt.Errorf("old oxygen value %d does not match current %d", 
				data.OldValue, state.GlobalParameters.Oxygen)
		}
	case domain.GlobalParamOceans:
		if data.NewValue < 0 || data.NewValue > 9 {
			return fmt.Errorf("oceans must be between 0 and 9, got %d", data.NewValue)
		}
		if data.OldValue != state.GlobalParameters.Oceans {
			return fmt.Errorf("old oceans value %d does not match current %d", 
				data.OldValue, state.GlobalParameters.Oceans)
		}
	default:
		return fmt.Errorf("unknown parameter %s", data.Parameter)
	}

	return nil
}

func (v *GameEventValidator) validateResourcesLost(event domain.GameEvent, state *domain.GameState) error {
	var data domain.ResourcesChangedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return fmt.Errorf("invalid resources changed data: %w", err)
	}

	if data.PlayerID == "" {
		return fmt.Errorf("player ID cannot be empty")
	}

	// Find player
	var player *domain.Player
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			player = &state.Players[i]
			break
		}
	}

	if player == nil {
		return fmt.Errorf("player %s not found", data.PlayerID)
	}

	// Check if player has enough resources to lose
	if player.Resources.Credits < data.Changes.Credits ||
		player.Resources.Steel < data.Changes.Steel ||
		player.Resources.Titanium < data.Changes.Titanium ||
		player.Resources.Plants < data.Changes.Plants ||
		player.Resources.Energy < data.Changes.Energy ||
		player.Resources.Heat < data.Changes.Heat {
		return fmt.Errorf("player %s does not have enough resources to lose", data.PlayerID)
	}

	// Validate that new totals are correct
	expectedCredits := player.Resources.Credits - data.Changes.Credits
	expectedSteel := player.Resources.Steel - data.Changes.Steel
	expectedTitanium := player.Resources.Titanium - data.Changes.Titanium
	expectedPlants := player.Resources.Plants - data.Changes.Plants
	expectedEnergy := player.Resources.Energy - data.Changes.Energy
	expectedHeat := player.Resources.Heat - data.Changes.Heat

	if data.NewTotals.Credits != expectedCredits ||
		data.NewTotals.Steel != expectedSteel ||
		data.NewTotals.Titanium != expectedTitanium ||
		data.NewTotals.Plants != expectedPlants ||
		data.NewTotals.Energy != expectedEnergy ||
		data.NewTotals.Heat != expectedHeat {
		return fmt.Errorf("new totals do not match expected values after resource loss")
	}

	return nil
}

func (v *GameEventValidator) validateResourcesGained(event domain.GameEvent, state *domain.GameState) error {
	var data domain.ResourcesChangedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return fmt.Errorf("invalid resources changed data: %w", err)
	}

	if data.PlayerID == "" {
		return fmt.Errorf("player ID cannot be empty")
	}

	// Find player
	var player *domain.Player
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			player = &state.Players[i]
			break
		}
	}

	if player == nil {
		return fmt.Errorf("player %s not found", data.PlayerID)
	}

	// Validate that new totals are correct
	expectedCredits := player.Resources.Credits + data.Changes.Credits
	expectedSteel := player.Resources.Steel + data.Changes.Steel
	expectedTitanium := player.Resources.Titanium + data.Changes.Titanium
	expectedPlants := player.Resources.Plants + data.Changes.Plants
	expectedEnergy := player.Resources.Energy + data.Changes.Energy
	expectedHeat := player.Resources.Heat + data.Changes.Heat

	if data.NewTotals.Credits != expectedCredits ||
		data.NewTotals.Steel != expectedSteel ||
		data.NewTotals.Titanium != expectedTitanium ||
		data.NewTotals.Plants != expectedPlants ||
		data.NewTotals.Energy != expectedEnergy ||
		data.NewTotals.Heat != expectedHeat {
		return fmt.Errorf("new totals do not match expected values after resource gain")
	}

	return nil
}

// Business Rules

// GetBusinessRules returns all business rules for the game
func GetBusinessRules() []domain.BusinessRule {
	return []domain.BusinessRule{
		&TurnOrderRule{},
		&ActionLimitRule{},
		&ResourceLimitRule{},
		&PhaseRule{},
		&GameCapacityRule{},
		&DuplicatePreventionRule{},
	}
}

// TurnOrderRule ensures actions can only be taken on the player's turn
type TurnOrderRule struct{}

func (r *TurnOrderRule) Name() string { return "TurnOrderRule" }
func (r *TurnOrderRule) Description() string { return "Players can only take actions on their turn" }

func (r *TurnOrderRule) Check(event domain.GameEvent, state *domain.GameState) error {
	// Only apply to player action events
	playerActionEvents := []domain.GameEventType{
		domain.EventTypeCardPlayed,
		domain.EventTypeMilestoneClaimed,
		domain.EventTypeAwardFunded,
		domain.EventTypeParameterIncreased,
		domain.EventTypeActionPerformed,
	}

	isPlayerAction := false
	for _, actionType := range playerActionEvents {
		if event.Type == actionType {
			isPlayerAction = true
			break
		}
	}

	if !isPlayerAction {
		return nil
	}

	if event.PlayerID == nil {
		return fmt.Errorf("player action event must have a player ID")
	}

	if state.CurrentPlayer != *event.PlayerID {
		return fmt.Errorf("player %s cannot take action, it's %s's turn", *event.PlayerID, state.CurrentPlayer)
	}

	return nil
}

// ActionLimitRule ensures players don't exceed their action limit
type ActionLimitRule struct{}

func (r *ActionLimitRule) Name() string { return "ActionLimitRule" }
func (r *ActionLimitRule) Description() string { return "Players cannot exceed their action limit per turn" }

func (r *ActionLimitRule) Check(event domain.GameEvent, state *domain.GameState) error {
	// Only apply to actions that consume action points
	actionEvents := []domain.GameEventType{
		domain.EventTypeCardPlayed,
		domain.EventTypeMilestoneClaimed,
		domain.EventTypeAwardFunded,
		domain.EventTypeParameterIncreased,
	}

	isAction := false
	for _, actionType := range actionEvents {
		if event.Type == actionType {
			isAction = true
			break
		}
	}

	if !isAction || event.PlayerID == nil {
		return nil
	}

	// Find the player
	for _, player := range state.Players {
		if player.ID == *event.PlayerID {
			if player.ActionsRemaining <= 0 {
				return fmt.Errorf("player %s has no actions remaining", *event.PlayerID)
			}
			break
		}
	}

	return nil
}

// ResourceLimitRule ensures players have sufficient resources
type ResourceLimitRule struct{}

func (r *ResourceLimitRule) Name() string { return "ResourceLimitRule" }
func (r *ResourceLimitRule) Description() string { return "Players cannot spend resources they don't have" }

func (r *ResourceLimitRule) Check(event domain.GameEvent, state *domain.GameState) error {
	if event.Type != domain.EventTypeResourcesLost {
		return nil
	}

	var data domain.ResourcesChangedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return nil // Skip if data is malformed - validator will catch this
	}

	// Find the player
	for _, player := range state.Players {
		if player.ID == data.PlayerID {
			if player.Resources.Credits < data.Changes.Credits ||
				player.Resources.Steel < data.Changes.Steel ||
				player.Resources.Titanium < data.Changes.Titanium ||
				player.Resources.Plants < data.Changes.Plants ||
				player.Resources.Energy < data.Changes.Energy ||
				player.Resources.Heat < data.Changes.Heat {
				return fmt.Errorf("player %s cannot spend resources they don't have", data.PlayerID)
			}
			break
		}
	}

	return nil
}

// PhaseRule ensures actions are only taken in appropriate phases
type PhaseRule struct{}

func (r *PhaseRule) Name() string { return "PhaseRule" }
func (r *PhaseRule) Description() string { return "Actions can only be taken in appropriate game phases" }

func (r *PhaseRule) Check(event domain.GameEvent, state *domain.GameState) error {
	switch event.Type {
	case domain.EventTypePlayerJoined:
		if state.Phase != domain.GamePhaseSetup {
			return fmt.Errorf("players can only join during setup phase")
		}
	case domain.EventTypeCorporationSelected:
		if state.Phase != domain.GamePhaseCorporationSelection {
			return fmt.Errorf("corporations can only be selected during corporation selection phase")
		}
	case domain.EventTypeCardPlayed, domain.EventTypeMilestoneClaimed, domain.EventTypeAwardFunded:
		if state.Phase != domain.GamePhaseGeneration || 
		   (state.TurnPhase != nil && *state.TurnPhase != domain.TurnPhaseAction) {
			return fmt.Errorf("game actions can only be taken during action phase")
		}
	}

	return nil
}

// GameCapacityRule ensures game doesn't exceed capacity limits
type GameCapacityRule struct{}

func (r *GameCapacityRule) Name() string { return "GameCapacityRule" }
func (r *GameCapacityRule) Description() string { return "Game cannot exceed capacity limits" }

func (r *GameCapacityRule) Check(event domain.GameEvent, state *domain.GameState) error {
	if event.Type == domain.EventTypePlayerJoined {
		if len(state.Players) >= 4 {
			return fmt.Errorf("game is full (maximum 4 players)")
		}
	}

	if event.Type == domain.EventTypeMilestoneClaimed {
		claimedCount := 0
		for _, milestone := range state.Milestones {
			if milestone.ClaimedBy != nil {
				claimedCount++
			}
		}
		if claimedCount >= 3 {
			return fmt.Errorf("maximum 3 milestones can be claimed")
		}
	}

	if event.Type == domain.EventTypeAwardFunded {
		fundedCount := 0
		for _, award := range state.Awards {
			if award.FundedBy != nil {
				fundedCount++
			}
		}
		if fundedCount >= 3 {
			return fmt.Errorf("maximum 3 awards can be funded")
		}
	}

	return nil
}

// DuplicatePreventionRule prevents duplicate actions
type DuplicatePreventionRule struct{}

func (r *DuplicatePreventionRule) Name() string { return "DuplicatePreventionRule" }
func (r *DuplicatePreventionRule) Description() string { return "Prevents duplicate actions" }

func (r *DuplicatePreventionRule) Check(event domain.GameEvent, state *domain.GameState) error {
	if event.Type == domain.EventTypeCorporationSelected && event.PlayerID != nil {
		// Check if player already has a corporation
		for _, player := range state.Players {
			if player.ID == *event.PlayerID && player.Corporation != nil {
				return fmt.Errorf("player %s already has a corporation", *event.PlayerID)
			}
		}
	}

	return nil
}