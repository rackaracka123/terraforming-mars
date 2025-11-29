package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/deck"
	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// Game represents a unified game entity containing all game state
// All fields are private with public methods for access and mutation
type Game struct {
	// Private fields - accessed only via public methods
	mu               sync.RWMutex
	id               string
	createdAt        time.Time
	updatedAt        time.Time
	status           GameStatus
	settings         GameSettings
	hostPlayerID     string
	currentPhase     GamePhase
	globalParameters *global_parameters.GlobalParameters
	currentTurn      *Turn // Tracks active player and available actions (nullable)
	generation       int
	board            *board.Board
	deck             *deck.Deck
	players          map[string]*player.Player
	turnOrder        []string // Ordered list of player IDs for turn sequence
	eventBus         *events.EventBusImpl

	// Player-specific non-card phase state (managed by Game)
	pendingTileSelections      map[string]*player.PendingTileSelection
	pendingTileSelectionQueues map[string]*player.PendingTileSelectionQueue
	forcedFirstActions         map[string]*player.ForcedFirstAction
	productionPhases           map[string]*player.ProductionPhase
	selectStartingCardsPhases  map[string]*player.SelectStartingCardsPhase
}

// NewGame creates a new game with the given settings
func NewGame(
	id string,
	hostPlayerID string,
	settings GameSettings,
	eventBus *events.EventBusImpl,
) *Game {
	now := time.Now()

	// Get initial global parameter values from settings or use defaults
	initTemp := DefaultTemperature
	initOxy := DefaultOxygen
	initOcean := DefaultOceans
	if settings.Temperature != nil {
		initTemp = *settings.Temperature
	}
	if settings.Oxygen != nil {
		initOxy = *settings.Oxygen
	}
	if settings.Oceans != nil {
		initOcean = *settings.Oceans
	}

	return &Game{
		id:               id,
		createdAt:        now,
		updatedAt:        now,
		status:           GameStatusLobby,
		settings:         settings,
		hostPlayerID:     hostPlayerID,
		currentPhase:     GamePhaseWaitingForGameStart,
		globalParameters: global_parameters.NewGlobalParametersWithValues(id, initTemp, initOxy, initOcean, eventBus),
		generation:       1,
		board:            board.NewBoardWithTiles(id, board.GenerateMarsBoard(), eventBus),
		deck:             nil, // Set via SetDeck after deck is created
		players:          make(map[string]*player.Player),
		turnOrder:        []string{}, // Empty until game starts
		eventBus:         eventBus,
		// Initialize non-card phase state maps
		pendingTileSelections:      make(map[string]*player.PendingTileSelection),
		pendingTileSelectionQueues: make(map[string]*player.PendingTileSelectionQueue),
		forcedFirstActions:         make(map[string]*player.ForcedFirstAction),
		productionPhases:           make(map[string]*player.ProductionPhase),
		selectStartingCardsPhases:  make(map[string]*player.SelectStartingCardsPhase),
	}
}

// ================== Basic Getters ==================

// ID returns the game ID
func (g *Game) ID() string {
	// Immutable, no lock needed
	return g.id
}

// CreatedAt returns when the game was created
func (g *Game) CreatedAt() time.Time {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.createdAt
}

// UpdatedAt returns when the game was last updated
func (g *Game) UpdatedAt() time.Time {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.updatedAt
}

// Status returns the current game status
func (g *Game) Status() GameStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.status
}

// Settings returns a copy of the game settings
func (g *Game) Settings() GameSettings {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.settings
}

// HostPlayerID returns the host player ID
func (g *Game) HostPlayerID() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.hostPlayerID
}

// CurrentPhase returns the current game phase
func (g *Game) CurrentPhase() GamePhase {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.currentPhase
}

// Generation returns the current generation number
func (g *Game) Generation() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.generation
}

// TurnOrder returns a copy of the turn order
func (g *Game) TurnOrder() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	// Return a copy to prevent external mutation
	order := make([]string, len(g.turnOrder))
	copy(order, g.turnOrder)
	return order
}

// CurrentTurn returns the current turn information (may be nil)
func (g *Game) CurrentTurn() *Turn {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.currentTurn
}

// ================== Component Accessors ==================

// GlobalParameters returns the global parameters component
func (g *Game) GlobalParameters() *global_parameters.GlobalParameters {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.globalParameters
}

// Board returns the board component
func (g *Game) Board() *board.Board {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.board
}

// Deck returns the deck component
func (g *Game) Deck() *deck.Deck {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.deck
}

// SetDeck sets the deck for this game (called during initialization)
func (g *Game) SetDeck(d *deck.Deck) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.deck = d
	g.updatedAt = time.Now()
}

// ================== Player Management ==================

// GetPlayer returns a player by ID
func (g *Game) GetPlayer(playerID string) (*player.Player, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	p, exists := g.players[playerID]
	if !exists {
		return nil, fmt.Errorf("player %s not found in game %s", playerID, g.id)
	}
	return p, nil
}

// GetAllPlayers returns all players in the game
func (g *Game) GetAllPlayers() []*player.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	players := make([]*player.Player, 0, len(g.players))
	for _, p := range g.players {
		players = append(players, p)
	}
	return players
}

// AddPlayer adds a new player to the game and publishes PlayerJoinedEvent
func (g *Game) AddPlayer(ctx context.Context, p *player.Player) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if _, exists := g.players[p.ID()]; exists {
		g.mu.Unlock()
		return fmt.Errorf("player %s already exists in game %s", p.ID(), g.id)
	}

	g.players[p.ID()] = p
	g.updatedAt = time.Now()
	g.mu.Unlock()

	// Publish PlayerJoinedEvent AFTER releasing lock
	// NOTE: BroadcastEvent is NOT published here - the handler is responsible
	// for publishing it AFTER registering the connection with the Hub
	if g.eventBus != nil {
		events.Publish(g.eventBus, events.PlayerJoinedEvent{
			GameID:   g.id,
			PlayerID: p.ID(),
		})
	}

	return nil
}

// RemovePlayer removes a player from the game
func (g *Game) RemovePlayer(ctx context.Context, playerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if _, exists := g.players[playerID]; !exists {
		g.mu.Unlock()
		return fmt.Errorf("player %s not found in game %s", playerID, g.id)
	}

	delete(g.players, playerID)
	g.updatedAt = time.Now()
	g.mu.Unlock()

	return nil
}

// ================== Game State Mutations with Event Publishing ==================

// UpdateStatus updates the game status and publishes GameStatusChangedEvent
func (g *Game) UpdateStatus(ctx context.Context, newStatus GameStatus) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldStatus GameStatus

	g.mu.Lock()
	oldStatus = g.status
	g.status = newStatus
	g.updatedAt = time.Now()
	g.mu.Unlock()

	// Publish event AFTER releasing lock
	if g.eventBus != nil && oldStatus != newStatus {
		events.Publish(g.eventBus, events.GameStatusChangedEvent{
			GameID:    g.id,
			OldStatus: string(oldStatus),
			NewStatus: string(newStatus),
		})
		// Trigger client broadcast
		events.Publish(g.eventBus, events.BroadcastEvent{
			GameID:    g.id,
			PlayerIDs: nil, // Broadcast to all players
		})
	}

	return nil
}

// UpdatePhase updates the game phase and publishes GamePhaseChangedEvent
func (g *Game) UpdatePhase(ctx context.Context, newPhase GamePhase) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldPhase GamePhase

	g.mu.Lock()
	oldPhase = g.currentPhase
	g.currentPhase = newPhase
	g.updatedAt = time.Now()
	g.mu.Unlock()

	// Publish event AFTER releasing lock
	if g.eventBus != nil && oldPhase != newPhase {
		events.Publish(g.eventBus, events.GamePhaseChangedEvent{
			GameID:   g.id,
			OldPhase: string(oldPhase),
			NewPhase: string(newPhase),
		})
		// Trigger client broadcast
		events.Publish(g.eventBus, events.BroadcastEvent{
			GameID:    g.id,
			PlayerIDs: nil, // Broadcast to all players
		})
	}

	return nil
}

// AdvanceGeneration advances the game to the next generation and publishes GenerationAdvancedEvent
func (g *Game) AdvanceGeneration(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldGeneration, newGeneration int

	g.mu.Lock()
	oldGeneration = g.generation
	g.generation++
	newGeneration = g.generation
	g.updatedAt = time.Now()
	g.mu.Unlock()

	// Publish event AFTER releasing lock
	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GenerationAdvancedEvent{
			GameID:        g.id,
			OldGeneration: oldGeneration,
			NewGeneration: newGeneration,
		})
		// Trigger client broadcast
		events.Publish(g.eventBus, events.BroadcastEvent{
			GameID:    g.id,
			PlayerIDs: nil, // Broadcast to all players
		})
	}

	return nil
}

// SetCurrentTurn sets the current turn to a specific player with a specific action count
func (g *Game) SetCurrentTurn(ctx context.Context, playerID string, actionsRemaining int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	g.currentTurn = NewTurn(playerID, actionsRemaining)
	g.updatedAt = time.Now()
	g.mu.Unlock()

	// Trigger client broadcast for turn change
	if g.eventBus != nil {
		events.Publish(g.eventBus, events.BroadcastEvent{
			GameID:    g.id,
			PlayerIDs: nil, // Broadcast to all players
		})
	}

	return nil
}

// SetTurnOrder sets the turn order for the game
func (g *Game) SetTurnOrder(ctx context.Context, turnOrder []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	// Create a copy to prevent external mutation
	g.turnOrder = make([]string, len(turnOrder))
	copy(g.turnOrder, turnOrder)
	g.updatedAt = time.Now()
	g.mu.Unlock()

	// Trigger client broadcast for turn order update
	if g.eventBus != nil {
		events.Publish(g.eventBus, events.BroadcastEvent{
			GameID:    g.id,
			PlayerIDs: nil, // Broadcast to all players
		})
	}

	return nil
}

// SetHostPlayerID sets the host player ID
func (g *Game) SetHostPlayerID(ctx context.Context, playerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	g.hostPlayerID = playerID
	g.updatedAt = time.Now()
	g.mu.Unlock()

	return nil
}

// ================== Turn Management ==================

// NextPlayer returns the next player ID in turn order based on current turn
// Returns nil if CurrentTurn is nil, turnOrder is empty, or no players exist
func (g *Game) NextPlayer() *string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.currentTurn == nil || len(g.turnOrder) == 0 {
		return nil
	}

	currentPlayerID := g.currentTurn.PlayerID()

	// Find current player index in turn order
	currentIndex := -1
	for i, playerID := range g.turnOrder {
		if playerID == currentPlayerID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		// Current turn player not found in turn order, return first player
		return &g.turnOrder[0]
	}

	// Calculate next player index (wrap around)
	nextIndex := (currentIndex + 1) % len(g.turnOrder)
	return &g.turnOrder[nextIndex]
}

// ================== Player Non-Card Phase State Management ==================

// GetPendingTileSelection returns the pending tile selection for a player
func (g *Game) GetPendingTileSelection(playerID string) *player.PendingTileSelection {
	g.mu.RLock()
	defer g.mu.RUnlock()

	selection, exists := g.pendingTileSelections[playerID]
	if !exists || selection == nil {
		return nil
	}
	// Simple struct, return copy
	selectionCopy := *selection
	return &selectionCopy
}

// SetPendingTileSelection sets the pending tile selection for a player
func (g *Game) SetPendingTileSelection(ctx context.Context, playerID string, selection *player.PendingTileSelection) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if selection == nil {
		delete(g.pendingTileSelections, playerID)
	} else {
		selectionCopy := *selection
		g.pendingTileSelections[playerID] = &selectionCopy
	}
	g.updatedAt = time.Now()
	g.mu.Unlock()

	// Trigger client broadcast to specific player
	if g.eventBus != nil {
		events.Publish(g.eventBus, events.BroadcastEvent{
			GameID:    g.id,
			PlayerIDs: []string{playerID}, // Broadcast only to this player
		})
	}

	return nil
}

// GetPendingTileSelectionQueue returns the tile selection queue for a player
func (g *Game) GetPendingTileSelectionQueue(playerID string) *player.PendingTileSelectionQueue {
	g.mu.RLock()
	defer g.mu.RUnlock()

	queue, exists := g.pendingTileSelectionQueues[playerID]
	if !exists || queue == nil {
		return nil
	}
	// Simple struct, return copy
	queueCopy := *queue
	return &queueCopy
}

// SetPendingTileSelectionQueue sets the tile selection queue for a player
func (g *Game) SetPendingTileSelectionQueue(ctx context.Context, playerID string, queue *player.PendingTileSelectionQueue) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if queue == nil {
		delete(g.pendingTileSelectionQueues, playerID)
	} else {
		queueCopy := *queue
		g.pendingTileSelectionQueues[playerID] = &queueCopy
	}
	g.updatedAt = time.Now()
	g.mu.Unlock()

	// Trigger client broadcast to specific player
	if g.eventBus != nil {
		events.Publish(g.eventBus, events.BroadcastEvent{
			GameID:    g.id,
			PlayerIDs: []string{playerID}, // Broadcast only to this player
		})
	}

	return nil
}

// GetForcedFirstAction returns the forced first action for a player
func (g *Game) GetForcedFirstAction(playerID string) *player.ForcedFirstAction {
	g.mu.RLock()
	defer g.mu.RUnlock()

	action, exists := g.forcedFirstActions[playerID]
	if !exists || action == nil {
		return nil
	}
	// Simple struct, return copy
	actionCopy := *action
	return &actionCopy
}

// SetForcedFirstAction sets the forced first action for a player
func (g *Game) SetForcedFirstAction(ctx context.Context, playerID string, action *player.ForcedFirstAction) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if action == nil {
		delete(g.forcedFirstActions, playerID)
	} else {
		actionCopy := *action
		g.forcedFirstActions[playerID] = &actionCopy
	}
	g.updatedAt = time.Now()
	g.mu.Unlock()

	// Trigger client broadcast to specific player
	if g.eventBus != nil {
		events.Publish(g.eventBus, events.BroadcastEvent{
			GameID:    g.id,
			PlayerIDs: []string{playerID}, // Broadcast only to this player
		})
	}

	return nil
}

// GetProductionPhase returns the production phase state for a player
func (g *Game) GetProductionPhase(playerID string) *player.ProductionPhase {
	g.mu.RLock()
	defer g.mu.RUnlock()

	phase, exists := g.productionPhases[playerID]
	if !exists || phase == nil {
		return nil
	}
	// Return copy to prevent external mutation
	phaseCopy := *phase
	return &phaseCopy
}

// SetProductionPhase sets the production phase state for a player
func (g *Game) SetProductionPhase(ctx context.Context, playerID string, phase *player.ProductionPhase) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if phase == nil {
		delete(g.productionPhases, playerID)
	} else {
		phaseCopy := *phase
		g.productionPhases[playerID] = &phaseCopy
	}
	g.updatedAt = time.Now()
	g.mu.Unlock()

	// Trigger client broadcast to specific player
	if g.eventBus != nil {
		events.Publish(g.eventBus, events.BroadcastEvent{
			GameID:    g.id,
			PlayerIDs: []string{playerID}, // Broadcast only to this player
		})
	}

	return nil
}

// GetSelectStartingCardsPhase returns the select starting cards phase state for a player
func (g *Game) GetSelectStartingCardsPhase(playerID string) *player.SelectStartingCardsPhase {
	g.mu.RLock()
	defer g.mu.RUnlock()

	phase, exists := g.selectStartingCardsPhases[playerID]
	if !exists || phase == nil {
		return nil
	}
	// Return copy to prevent external mutation
	phaseCopy := *phase
	return &phaseCopy
}

// SetSelectStartingCardsPhase sets the select starting cards phase state for a player
func (g *Game) SetSelectStartingCardsPhase(ctx context.Context, playerID string, phase *player.SelectStartingCardsPhase) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if phase == nil {
		delete(g.selectStartingCardsPhases, playerID)
	} else {
		phaseCopy := *phase
		g.selectStartingCardsPhases[playerID] = &phaseCopy
	}
	g.updatedAt = time.Now()
	g.mu.Unlock()

	// Trigger client broadcast to specific player
	if g.eventBus != nil {
		events.Publish(g.eventBus, events.BroadcastEvent{
			GameID:    g.id,
			PlayerIDs: []string{playerID}, // Broadcast only to this player
		})
	}

	return nil
}

// ProcessNextTile pops the next tile from a player's tile queue and creates a PendingTileSelection
// This method converts the queue into an actionable tile selection for the player
func (g *Game) ProcessNextTile(ctx context.Context, playerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	// Get queue for this player
	queue, exists := g.pendingTileSelectionQueues[playerID]
	if !exists || queue == nil || len(queue.Items) == 0 {
		g.mu.Unlock()
		return nil // No queue or empty queue
	}

	// Pop first item
	nextTileType := queue.Items[0]
	remainingItems := queue.Items[1:]
	source := queue.Source

	// Update or clear queue
	if len(remainingItems) > 0 {
		g.pendingTileSelectionQueues[playerID] = &player.PendingTileSelectionQueue{
			Items:  remainingItems,
			Source: source,
		}
	} else {
		delete(g.pendingTileSelectionQueues, playerID)
	}
	g.mu.Unlock()

	// Calculate available hexes for this tile type
	availableHexes := g.calculateAvailableHexesForTile(nextTileType, playerID)

	// Create pending tile selection
	err := g.SetPendingTileSelection(ctx, playerID, &player.PendingTileSelection{
		TileType:       nextTileType,
		AvailableHexes: availableHexes,
		Source:         source,
	})

	return err
}

// calculateAvailableHexesForTile returns a list of valid hex positions for placing a tile
func (g *Game) calculateAvailableHexesForTile(tileType string, playerID string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.board == nil {
		return []string{}
	}

	tiles := g.board.Tiles()
	availableHexes := []string{}

	for _, tile := range tiles {
		// Skip tiles that are already occupied
		if tile.OccupiedBy != nil {
			continue
		}

		switch tileType {
		case "city":
			// Cities can be placed on land tiles (not ocean-space)
			// Cities CANNOT be adjacent to other cities
			if tile.Type != shared.ResourceLandTile {
				continue
			}

			// Check if any neighbor has a city
			hasAdjacentCity := false
			neighbors := tile.Coordinates.GetNeighbors()
			for _, neighborPos := range neighbors {
				for _, neighborTile := range tiles {
					if neighborTile.Coordinates.Equals(neighborPos) {
						if neighborTile.OccupiedBy != nil && neighborTile.OccupiedBy.Type == shared.ResourceCityTile {
							hasAdjacentCity = true
							break
						}
					}
				}
				if hasAdjacentCity {
					break
				}
			}

			if !hasAdjacentCity {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}

		case "greenery":
			// Greeneries can be placed on land tiles (not ocean-space)
			if tile.Type == shared.ResourceLandTile {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}

		case "ocean":
			// Oceans can only be placed on ocean-space tiles
			if tile.Type == shared.ResourceOceanSpace {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}

		default:
			// For unknown tile types, allow placement on any land tile
			if tile.Type == shared.ResourceLandTile {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}
		}
	}

	return availableHexes
}
