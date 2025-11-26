package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/game/board"
	"terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/game/player/selection"
	"terraforming-mars-backend/internal/session/types"
)

// Game represents a unified game entity containing both metadata and state
type Game struct {
	// Serialized game data (sent to frontend)
	ID               string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Status           types.GameStatus
	Settings         types.GameSettings
	HostPlayerID     string
	CurrentPhase     types.GamePhase
	GlobalParameters types.GlobalParameters
	ViewingPlayerID  string  // The player viewing this game state
	CurrentTurn      *string // Whose turn it is (nullable)
	Generation       int
	Board            board.Board // Game board with tiles and occupancy state

	// Non-serialized runtime state
	mu       sync.RWMutex
	Players  map[string]*player.Player // Player map by ID (single source of truth)
	eventBus *events.EventBusImpl      // Event bus for publishing domain events

	// Player-specific non-card phase state (managed by Game)
	pendingTileSelections      map[string]*player.PendingTileSelection      // playerID -> pending tile selection
	pendingTileSelectionQueues map[string]*player.PendingTileSelectionQueue // playerID -> tile selection queue
	forcedFirstActions         map[string]*player.ForcedFirstAction         // playerID -> forced first action
	productionPhases           map[string]*selection.ProductionPhase        // playerID -> production phase state

	// Infrastructure components
	cardManager CardManager // Card validation and playing
}

// NewGame creates a new game with the given settings, event bus, and card manager
func NewGame(
	id string,
	hostPlayerID string,
	settings types.GameSettings,
	eventBus *events.EventBusImpl,
	cardManager CardManager,
) *Game {
	now := time.Now()

	return &Game{
		ID:           id,
		CreatedAt:    now,
		UpdatedAt:    now,
		Status:       types.GameStatusLobby,
		Settings:     settings,
		HostPlayerID: hostPlayerID,
		CurrentPhase: types.GamePhaseWaitingForGameStart,
		GlobalParameters: types.GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		},
		Generation:  1,
		Board:       board.Board{Tiles: []board.Tile{}}, // Initialize with empty board
		Players:     make(map[string]*player.Player),
		eventBus:    eventBus,
		cardManager: cardManager,
		mu:          sync.RWMutex{},
		// Initialize non-card phase state maps
		pendingTileSelections:      make(map[string]*player.PendingTileSelection),
		pendingTileSelectionQueues: make(map[string]*player.PendingTileSelectionQueue),
		forcedFirstActions:         make(map[string]*player.ForcedFirstAction),
		productionPhases:           make(map[string]*selection.ProductionPhase),
	}
}

// GetPlayer returns a player by ID
func (g *Game) GetPlayer(playerID string) (*player.Player, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	p, exists := g.Players[playerID]
	if !exists {
		return nil, fmt.Errorf("player %s not found in game %s", playerID, g.ID)
	}
	return p, nil
}

// GetAllPlayers returns all players in the game
func (g *Game) GetAllPlayers() []*player.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	players := make([]*player.Player, 0, len(g.Players))
	for _, p := range g.Players {
		players = append(players, p)
	}
	return players
}

// AddPlayer adds a new player to the game
func (g *Game) AddPlayer(p *player.Player) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.Players[p.ID()]; exists {
		return fmt.Errorf("player %s already exists in game %s", p.ID(), g.ID)
	}

	g.Players[p.ID()] = p
	g.UpdatedAt = time.Now()
	return nil
}

// RemovePlayer removes a player from the game
func (g *Game) RemovePlayer(playerID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.Players[playerID]; !exists {
		return fmt.Errorf("player %s not found in game %s", playerID, g.ID)
	}

	delete(g.Players, playerID)
	g.UpdatedAt = time.Now()
	return nil
}

// UpdateStatus updates the game status and publishes event
func (g *Game) UpdateStatus(ctx context.Context, newStatus types.GameStatus) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	oldStatus := g.Status
	g.Status = newStatus
	g.UpdatedAt = time.Now()

	// Publish event
	if g.eventBus != nil && oldStatus != newStatus {
		events.Publish(g.eventBus, events.GameStatusChangedEvent{
			GameID:    g.ID,
			OldStatus: string(oldStatus),
			NewStatus: string(newStatus),
		})
	}

	return nil
}

// UpdatePhase updates the game phase and publishes event
func (g *Game) UpdatePhase(ctx context.Context, newPhase types.GamePhase) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	oldPhase := g.CurrentPhase
	g.CurrentPhase = newPhase
	g.UpdatedAt = time.Now()

	// Publish event
	if g.eventBus != nil && oldPhase != newPhase {
		events.Publish(g.eventBus, events.GamePhaseChangedEvent{
			GameID:   g.ID,
			OldPhase: string(oldPhase),
			NewPhase: string(newPhase),
		})
	}

	return nil
}

// UpdateTemperature updates the temperature and publishes event
func (g *Game) UpdateTemperature(ctx context.Context, newTemp int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	oldTemp := g.GlobalParameters.Temperature
	g.GlobalParameters.Temperature = newTemp
	g.UpdatedAt = time.Now()

	// Publish event
	if g.eventBus != nil && oldTemp != newTemp {
		events.Publish(g.eventBus, events.TemperatureChangedEvent{
			GameID:   g.ID,
			OldValue: oldTemp,
			NewValue: newTemp,
		})
	}

	return nil
}

// UpdateOxygen updates the oxygen level and publishes event
func (g *Game) UpdateOxygen(ctx context.Context, newOxygen int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	oldOxygen := g.GlobalParameters.Oxygen
	g.GlobalParameters.Oxygen = newOxygen
	g.UpdatedAt = time.Now()

	// Publish event
	if g.eventBus != nil && oldOxygen != newOxygen {
		events.Publish(g.eventBus, events.OxygenChangedEvent{
			GameID:   g.ID,
			OldValue: oldOxygen,
			NewValue: newOxygen,
		})
	}

	return nil
}

// UpdateOceans updates the ocean count and publishes event
func (g *Game) UpdateOceans(ctx context.Context, newOceans int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	oldOceans := g.GlobalParameters.Oceans
	g.GlobalParameters.Oceans = newOceans
	g.UpdatedAt = time.Now()

	// Publish event
	if g.eventBus != nil && oldOceans != newOceans {
		events.Publish(g.eventBus, events.OceansChangedEvent{
			GameID:   g.ID,
			OldValue: oldOceans,
			NewValue: newOceans,
		})
	}

	return nil
}

// AdvanceGeneration advances the game to the next generation and publishes event
func (g *Game) AdvanceGeneration(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	oldGeneration := g.Generation
	g.Generation++
	g.UpdatedAt = time.Now()

	// Publish event
	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GenerationAdvancedEvent{
			GameID:        g.ID,
			OldGeneration: oldGeneration,
			NewGeneration: g.Generation,
		})
	}

	return nil
}

// SetCurrentTurn sets the current turn to a specific player
func (g *Game) SetCurrentTurn(ctx context.Context, playerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.CurrentTurn = &playerID
	g.UpdatedAt = time.Now()

	return nil
}

// Next returns the next player ID in turn order based on current turn
// Returns nil if CurrentTurn is nil or no players exist
// TODO: Implement proper turn order mechanism (currently uses map iteration order which is non-deterministic)
func (g *Game) Next() *string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.CurrentTurn == nil || len(g.Players) == 0 {
		return nil
	}

	// Get ordered list of player IDs (for now, simple iteration)
	playerIDs := make([]string, 0, len(g.Players))
	for id := range g.Players {
		playerIDs = append(playerIDs, id)
	}

	if len(playerIDs) == 0 {
		return nil
	}

	// Find current player index
	currentIndex := -1
	for i, playerID := range playerIDs {
		if playerID == *g.CurrentTurn {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		// Current turn player not found, return first player
		return &playerIDs[0]
	}

	// Calculate next player index (wrap around)
	nextIndex := (currentIndex + 1) % len(playerIDs)
	return &playerIDs[nextIndex]
}

// CardManager returns the card manager
func (g *Game) CardManager() CardManager {
	return g.cardManager
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
	defer g.mu.Unlock()

	if selection == nil {
		delete(g.pendingTileSelections, playerID)
	} else {
		selectionCopy := *selection
		g.pendingTileSelections[playerID] = &selectionCopy
	}
	g.UpdatedAt = time.Now()
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
	defer g.mu.Unlock()

	if queue == nil {
		delete(g.pendingTileSelectionQueues, playerID)
	} else {
		queueCopy := *queue
		g.pendingTileSelectionQueues[playerID] = &queueCopy
	}
	g.UpdatedAt = time.Now()
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
	defer g.mu.Unlock()

	if action == nil {
		delete(g.forcedFirstActions, playerID)
	} else {
		actionCopy := *action
		g.forcedFirstActions[playerID] = &actionCopy
	}
	g.UpdatedAt = time.Now()
	return nil
}

// GetProductionPhase returns the production phase state for a player
func (g *Game) GetProductionPhase(playerID string) *selection.ProductionPhase {
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
func (g *Game) SetProductionPhase(ctx context.Context, playerID string, phase *selection.ProductionPhase) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if phase == nil {
		delete(g.productionPhases, playerID)
	} else {
		phaseCopy := *phase
		g.productionPhases[playerID] = &phaseCopy
	}
	g.UpdatedAt = time.Now()
	return nil
}

// ProcessNextTile pops the next tile from a player's tile queue
// Returns the tile type and whether more tiles remain in the queue
func (g *Game) ProcessNextTile(ctx context.Context, playerID string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Get queue for this player
	queue, exists := g.pendingTileSelectionQueues[playerID]
	if !exists || queue == nil || len(queue.Items) == 0 {
		return "", nil // No queue or empty queue
	}

	// Pop first item
	nextTileType := queue.Items[0]
	remainingItems := queue.Items[1:]

	// Update or clear queue
	if len(remainingItems) > 0 {
		g.pendingTileSelectionQueues[playerID] = &player.PendingTileSelectionQueue{
			Items:  remainingItems,
			Source: queue.Source,
		}
	} else {
		delete(g.pendingTileSelectionQueues, playerID)
	}

	g.UpdatedAt = time.Now()
	return nextTileType, nil
}
