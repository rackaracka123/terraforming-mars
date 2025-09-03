package domain

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// GameAggregate represents a game that can be reconstructed from events
type GameAggregate struct {
	state   *GameState
	events  []GameEvent
	version int64
}

// NewGameAggregate creates a new game aggregate from events
func NewGameAggregate(gameID string, events []GameEvent) (*GameAggregate, error) {
	// Sort events by version to ensure proper ordering
	sort.Slice(events, func(i, j int) bool {
		return events[i].Version < events[j].Version
	})

	aggregate := &GameAggregate{
		state:  NewEmptyGameState(gameID),
		events: events,
	}

	// Apply all events to reconstruct the current state
	for _, event := range events {
		if err := aggregate.applyEvent(event); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", event.ID, err)
		}
	}

	return aggregate, nil
}

// GetState returns the current game state
func (ga *GameAggregate) GetState() *GameState {
	return ga.state
}

// GetEvents returns all events
func (ga *GameAggregate) GetEvents() []GameEvent {
	return ga.events
}

// GetVersion returns the current version
func (ga *GameAggregate) GetVersion() int64 {
	return ga.version
}

// ApplyEvent applies a new event and updates the state
func (ga *GameAggregate) ApplyEvent(event GameEvent) error {
	// Set the version for the new event
	event.Version = ga.version + 1
	
	// Apply the event to the state
	if err := ga.applyEvent(event); err != nil {
		return err
	}

	// Add to event history
	ga.events = append(ga.events, event)
	ga.version = event.Version

	return nil
}

// GetStateAtVersion reconstructs the game state at a specific version
func (ga *GameAggregate) GetStateAtVersion(version int64) (*GameState, error) {
	state := NewEmptyGameState(ga.state.ID)
	
	for _, event := range ga.events {
		if event.Version > version {
			break
		}
		if err := ApplyEventToState(event, state); err != nil {
			return nil, fmt.Errorf("failed to apply event %s at version %d: %w", event.ID, version, err)
		}
	}

	return state, nil
}

// applyEvent applies an event to the current state
func (ga *GameAggregate) applyEvent(event GameEvent) error {
	return ApplyEventToState(event, ga.state)
}

// ApplyEventToState applies an event to a given state
func ApplyEventToState(event GameEvent, state *GameState) error {
	switch event.Type {
	case EventTypeGameCreated:
		return applyGameCreated(event, state)
	case EventTypePlayerJoined:
		return applyPlayerJoined(event, state)
	case EventTypeCorporationSelected:
		return applyCorporationSelected(event, state)
	case EventTypeCardPlayed:
		return applyCardPlayed(event, state)
	case EventTypeTilePlaced:
		return applyTilePlaced(event, state)
	case EventTypeResourcesGained:
		return applyResourcesChanged(event, state, true)
	case EventTypeResourcesLost:
		return applyResourcesChanged(event, state, false)
	case EventTypeProductionChanged:
		return applyProductionChanged(event, state)
	case EventTypeParameterIncreased:
		return applyParameterIncreased(event, state)
	case EventTypeMilestoneClaimed:
		return applyMilestoneClaimed(event, state)
	case EventTypeAwardFunded:
		return applyAwardFunded(event, state)
	case EventTypePhaseChanged:
		return applyPhaseChanged(event, state)
	case EventTypeGenerationStarted:
		return applyGenerationStarted(event, state)
	case EventTypeTurnStarted:
		return applyTurnStarted(event, state)
	case EventTypeVictoryPointsAwarded:
		return applyVictoryPointsAwarded(event, state)
	case EventTypeGameEnded:
		return applyGameEnded(event, state)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

// NewEmptyGameState creates an empty game state for the given game ID
func NewEmptyGameState(gameID string) *GameState {
	return &GameState{
		ID:                        gameID,
		Players:                   []Player{},
		Generation:                1,
		Phase:                     GamePhaseSetup,
		GlobalParameters:          GlobalParameters{Temperature: -30, Oxygen: 0, Oceans: 0},
		EndGameConditions:         GetEndGameConditions(),
		Board:                     GetBoardSpaces(),
		Milestones:                GetBaseMilestones(),
		Awards:                    GetBaseAwards(),
		AvailableStandardProjects: GetStandardProjects(),
		Deck:                      []string{},
		DiscardPile:               []string{},
		ActionHistory:             []Action{},
		Events:                    []GameEvent{},
		IsGameEnded:               false,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}
}

// Event application functions

func applyGameCreated(event GameEvent, state *GameState) error {
	var data GameCreatedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	state.GameSettings = data.Settings
	state.SoloMode = data.MaxPlayers == 1
	state.CreatedAt = event.Timestamp
	state.UpdatedAt = event.Timestamp

	return nil
}

func applyPlayerJoined(event GameEvent, state *GameState) error {
	var data PlayerJoinedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	// Check if player already exists
	for _, player := range state.Players {
		if player.ID == data.PlayerID {
			return nil // Player already joined
		}
	}

	newPlayer := Player{
		ID:   data.PlayerID,
		Name: data.PlayerName,
		Resources: ResourcesMap{
			Credits:  20, // Starting megacredits
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		Production: ResourcesMap{
			Credits:  1, // Starting production
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   1,
			Heat:     1,
		},
		TerraformRating:     20, // Starting TR
		VictoryPoints:       0,
		VictoryPointSources: []VictoryPointSource{},
		PlayedCards:         []string{},
		Hand:                []string{},
		AvailableActions:    2,
		Tags:                []Tag{},
		TagCounts:           make(map[Tag]int),
		TilePositions:       []HexCoordinate{},
		TileCounts:          make(map[TileType]int),
		Reserved:            ResourcesMap{},
		ClaimedMilestones:   []string{},
		FundedAwards:        []string{},
		HandLimit:           10,
	}

	state.Players = append(state.Players, newPlayer)

	// Set first player if not set
	if state.CurrentPlayer == "" {
		state.CurrentPlayer = data.PlayerID
	}
	if state.FirstPlayer == "" {
		state.FirstPlayer = data.PlayerID
	}

	state.UpdatedAt = event.Timestamp
	return nil
}

func applyCorporationSelected(event GameEvent, state *GameState) error {
	var data CorporationSelectedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	// Find the player
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			player := &state.Players[i]
			player.Corporation = &data.CorporationID

			// Apply corporation effects (this would be more detailed in practice)
			corporations := GetBaseCorporations()
			for _, corp := range corporations {
				if corp.ID == data.CorporationID {
					player.Resources.Credits = corp.StartingMegaCredits

					if corp.StartingProduction != nil {
						player.Production.Credits += corp.StartingProduction.Credits
						player.Production.Steel += corp.StartingProduction.Steel
						player.Production.Titanium += corp.StartingProduction.Titanium
						player.Production.Plants += corp.StartingProduction.Plants
						player.Production.Energy += corp.StartingProduction.Energy
						player.Production.Heat += corp.StartingProduction.Heat
					}

					if corp.StartingResources != nil {
						player.Resources.Credits += corp.StartingResources.Credits
						player.Resources.Steel += corp.StartingResources.Steel
						player.Resources.Titanium += corp.StartingResources.Titanium
						player.Resources.Plants += corp.StartingResources.Plants
						player.Resources.Energy += corp.StartingResources.Energy
						player.Resources.Heat += corp.StartingResources.Heat
					}

					if corp.StartingTR != nil {
						player.TerraformRating += *corp.StartingTR
					}

					// Add corporation tags
					for _, tag := range corp.Tags {
						player.Tags = append(player.Tags, tag)
						player.TagCounts[tag]++
					}
					break
				}
			}

			state.UpdatedAt = event.Timestamp
			return nil
		}
	}

	return fmt.Errorf("player %s not found", data.PlayerID)
}

func applyCardPlayed(event GameEvent, state *GameState) error {
	var data CardPlayedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	// Find the player
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			player := &state.Players[i]
			
			// Remove card from hand and add to played cards
			for j, cardID := range player.Hand {
				if cardID == data.CardID {
					player.Hand = append(player.Hand[:j], player.Hand[j+1:]...)
					break
				}
			}
			player.PlayedCards = append(player.PlayedCards, data.CardID)

			// Deduct resources spent
			player.Resources.Credits -= data.ResourcesSpent.Credits
			player.Resources.Steel -= data.ResourcesSpent.Steel
			player.Resources.Titanium -= data.ResourcesSpent.Titanium
			player.Resources.Plants -= data.ResourcesSpent.Plants
			player.Resources.Energy -= data.ResourcesSpent.Energy
			player.Resources.Heat -= data.ResourcesSpent.Heat

			// Apply immediate effects (simplified - would be more complex in practice)
			// This would involve applying each effect in data.ImmediateEffects

			state.UpdatedAt = event.Timestamp
			return nil
		}
	}

	return fmt.Errorf("player %s not found", data.PlayerID)
}

func applyTilePlaced(event GameEvent, state *GameState) error {
	var data TilePlacedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	// Find the player
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			player := &state.Players[i]
			
			// Add tile position to player
			player.TilePositions = append(player.TilePositions, data.Position)
			player.TileCounts[data.TileType]++

			// Update board state
			for j := range state.Board {
				if state.Board[j].Position.Q == data.Position.Q &&
					state.Board[j].Position.R == data.Position.R &&
					state.Board[j].Position.S == data.Position.S {
					
					state.Board[j].Tile = &Tile{
						Type:     data.TileType,
						Position: data.Position,
						PlayerID: &data.PlayerID,
						Bonus:    data.SpaceBonuses,
					}
					break
				}
			}

			// Apply space bonuses and adjacency bonuses
			for _, bonus := range data.SpaceBonuses {
				switch bonus {
				case ResourceTypeCredits:
					player.Resources.Credits++
				case ResourceTypeSteel:
					player.Resources.Steel++
				case ResourceTypeTitanium:
					player.Resources.Titanium++
				case ResourceTypePlants:
					player.Resources.Plants++
				case ResourceTypeEnergy:
					player.Resources.Energy++
				case ResourceTypeHeat:
					player.Resources.Heat++
				}
			}

			state.UpdatedAt = event.Timestamp
			return nil
		}
	}

	return fmt.Errorf("player %s not found", data.PlayerID)
}

func applyResourcesChanged(event GameEvent, state *GameState, isGain bool) error {
	var data ResourcesChangedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	// Find the player
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			player := &state.Players[i]
			
			if isGain {
				player.Resources.Credits += data.Changes.Credits
				player.Resources.Steel += data.Changes.Steel
				player.Resources.Titanium += data.Changes.Titanium
				player.Resources.Plants += data.Changes.Plants
				player.Resources.Energy += data.Changes.Energy
				player.Resources.Heat += data.Changes.Heat
			} else {
				player.Resources.Credits -= data.Changes.Credits
				player.Resources.Steel -= data.Changes.Steel
				player.Resources.Titanium -= data.Changes.Titanium
				player.Resources.Plants -= data.Changes.Plants
				player.Resources.Energy -= data.Changes.Energy
				player.Resources.Heat -= data.Changes.Heat
			}

			state.UpdatedAt = event.Timestamp
			return nil
		}
	}

	return fmt.Errorf("player %s not found", data.PlayerID)
}

func applyProductionChanged(event GameEvent, state *GameState) error {
	var data ProductionChangedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	// Find the player
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			player := &state.Players[i]
			
			player.Production.Credits += data.Changes.Credits
			player.Production.Steel += data.Changes.Steel
			player.Production.Titanium += data.Changes.Titanium
			player.Production.Plants += data.Changes.Plants
			player.Production.Energy += data.Changes.Energy
			player.Production.Heat += data.Changes.Heat

			state.UpdatedAt = event.Timestamp
			return nil
		}
	}

	return fmt.Errorf("player %s not found", data.PlayerID)
}

func applyParameterIncreased(event GameEvent, state *GameState) error {
	var data ParameterIncreasedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	// Update global parameter
	switch data.Parameter {
	case GlobalParamTemperature:
		state.GlobalParameters.Temperature = data.NewValue
	case GlobalParamOxygen:
		state.GlobalParameters.Oxygen = data.NewValue
	case GlobalParamOceans:
		state.GlobalParameters.Oceans = data.NewValue
	}

	// Update end game conditions
	for i := range state.EndGameConditions {
		if state.EndGameConditions[i].Parameter == data.Parameter {
			state.EndGameConditions[i].CurrentValue = data.NewValue
			state.EndGameConditions[i].IsCompleted = data.NewValue >= state.EndGameConditions[i].TargetValue
		}
	}

	// Increase player's terraform rating
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			state.Players[i].TerraformRating += data.TRIncrease
			break
		}
	}

	state.UpdatedAt = event.Timestamp
	return nil
}

func applyMilestoneClaimed(event GameEvent, state *GameState) error {
	var data MilestoneClaimedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	// Update milestone as claimed
	for i := range state.Milestones {
		if state.Milestones[i].ID == data.MilestoneID {
			state.Milestones[i].ClaimedBy = &data.PlayerID
			break
		}
	}

	// Add to player's claimed milestones
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			player := &state.Players[i]
			player.ClaimedMilestones = append(player.ClaimedMilestones, data.MilestoneID)
			player.Resources.Credits -= data.Cost
			break
		}
	}

	state.UpdatedAt = event.Timestamp
	return nil
}

func applyAwardFunded(event GameEvent, state *GameState) error {
	var data AwardFundedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	// Update award as funded
	for i := range state.Awards {
		if state.Awards[i].ID == data.AwardID {
			state.Awards[i].FundedBy = &data.PlayerID
			break
		}
	}

	// Add to player's funded awards
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			player := &state.Players[i]
			player.FundedAwards = append(player.FundedAwards, data.AwardID)
			player.Resources.Credits -= data.Cost
			break
		}
	}

	state.UpdatedAt = event.Timestamp
	return nil
}

func applyPhaseChanged(event GameEvent, state *GameState) error {
	var data PhaseChangedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	state.Phase = data.NewPhase
	state.TurnPhase = data.NewTurnPhase
	state.Generation = data.Generation
	state.UpdatedAt = event.Timestamp

	return nil
}

func applyGenerationStarted(event GameEvent, state *GameState) error {
	var data GenerationStartedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	state.Generation = data.Generation
	state.FirstPlayer = data.FirstPlayer
	state.CurrentPlayer = data.PlayerOrder[0]
	state.UpdatedAt = event.Timestamp

	return nil
}

func applyTurnStarted(event GameEvent, state *GameState) error {
	var data TurnStartedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	state.CurrentPlayer = data.PlayerID
	state.Turn = data.TurnIndex

	// Reset player actions
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			state.Players[i].ActionsRemaining = data.ActionsRemaining
			state.Players[i].ActionsTaken = 0
			if state.Players[i].Passed != nil {
				*state.Players[i].Passed = false
			}
			break
		}
	}

	state.UpdatedAt = event.Timestamp
	return nil
}

func applyVictoryPointsAwarded(event GameEvent, state *GameState) error {
	var data VictoryPointsAwardedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	// Find the player and add victory points
	for i := range state.Players {
		if state.Players[i].ID == data.PlayerID {
			player := &state.Players[i]
			player.VictoryPoints += data.Points
			
			vpSource := VictoryPointSource{
				Type:        data.Source,
				Points:      data.Points,
				Description: data.Description,
				Details:     data.Details,
			}
			player.VictoryPointSources = append(player.VictoryPointSources, vpSource)
			break
		}
	}

	state.UpdatedAt = event.Timestamp
	return nil
}

func applyGameEnded(event GameEvent, state *GameState) error {
	var data GameEndedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	state.IsGameEnded = true
	state.WinnerID = &data.WinnerID
	state.Phase = GamePhaseGameEnd
	state.UpdatedAt = event.Timestamp

	return nil
}