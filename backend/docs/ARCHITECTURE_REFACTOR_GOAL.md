# Backend Architecture Refactor: Action-Driven with Game as State Repository

## High-Level Goal

Transform the backend architecture to achieve clean separation of concerns where **Actions contain all business logic** (validation, orchestration, effect application) and **Game serves as a pure state repository** (getters/setters, event publishing) with fully encapsulated types using private fields and public methods.

**Note**: This document outlines architectural principles and patterns to follow, not rigid specifications. Exact implementation details (method signatures, field names, EventBus injection, etc.) will be determined during development as we work through the refactor. The key is adhering to the core principles: encapsulation, separation of concerns, and thread-safe state management.

## Core Architectural Principle

```
┌─────────────────────────────────────────────────────────────┐
│ Actions (Business Logic Layer)                              │
│ - Validate game rules and requirements                      │
│ - Orchestrate multi-step operations                         │
│ - Apply card effects and bonuses                            │
│ - Calculate derived values                                  │
└─────────────────────────────────────────────────────────────┘
                            ↓ calls methods
┌─────────────────────────────────────────────────────────────┐
│ Game (State Repository)                                     │
│ - Store game state with private fields                      │
│ - Provide public getters/setters                            │
│ - Publish domain events on mutations                        │
│ - NO business logic                                         │
└─────────────────────────────────────────────────────────────┘
                            ↓ publishes events
┌─────────────────────────────────────────────────────────────┐
│ Delivery Layer (Presentation)                               │
│ - Listen for BroadcastEvent with gameID + playerID         │
│ - Fetch game state via GameRepository                       │
│ - Create personalized DTOs                                  │
│ - Broadcast to WebSocket clients                            │
└─────────────────────────────────────────────────────────────┘
```

## What Changes

### Current Architecture
- `internal/session/` package with Session wrapper
- Multiple repositories (game, board, card, deck)
- Actions receive Session and multiple repositories
- Mixed responsibility between actions and repositories

### Target Architecture
- `internal/game/` as root package for all game state
- Single GameRepository managing active games
- Game contains: Players, Deck, Board, GlobalParameters, Generation, Phase, PendingSelections
- Actions receive only GameRepository
- Game has zero business logic - pure state container

## Type Encapsulation Pattern

All domain types use **private fields with public methods**:

### Game
```go
type Game struct {
    mu                sync.RWMutex
    id                string
    players           map[string]*Player
    globalParameters  *GlobalParameters
    board             *Board
    deck              *Deck
    generation        int
    phase             GamePhase
    pendingSelections map[string]*Selection
}

// Public methods
func (g *Game) GetPlayer(id string) *Player
func (g *Game) GlobalParameters() *GlobalParameters
func (g *Game) Board() *Board
func (g *Game) IncrementGeneration(ctx context.Context)
```

### GlobalParameters
```go
type GlobalParameters struct {
    mu          sync.RWMutex
    temperature int
    oxygen      int
    oceans      int
}

// Public methods with business rule enforcement via events
func (gp *GlobalParameters) IncreaseTemperature(ctx context.Context, steps int)
func (gp *GlobalParameters) IncreaseOxygen(ctx context.Context, steps int)
func (gp *GlobalParameters) Temperature() int
```

### Player
```go
type Player struct {
    mu          sync.RWMutex
    id          string
    name        string
    hand        *Hand
    resources   *Resources
    playedCards *PlayedCards
    corporation *Corporation
    // ... other components
}

// Public methods - delegate to components
func (p *Player) Hand() *Hand
func (p *Player) Resources() *Resources
```

### Components (Hand, Resources, etc.)
```go
type Hand struct {
    mu    sync.RWMutex
    cards map[string]*Card
}

// Simple state operations only
func (h *Hand) GetCard(id string) (*Card, error)
func (h *Hand) RemoveCard(id string) error
func (h *Hand) AddCard(card *Card)
func (h *Hand) Cards() []*Card
```

## Action Pattern

Actions contain ALL business logic and orchestrate state changes:

```go
type StandardProjectAsteroidAction struct {
    gameRepo GameRepository
    logger   *logger.Logger
}

func (a *StandardProjectAsteroidAction) Execute(
    ctx context.Context,
    gameID string,
    playerID string,
) error {
    // 1. Fetch state from repository
    game, err := a.gameRepo.Get(gameID)
    if err != nil {
        return err
    }

    player := game.GetPlayer(playerID)

    // 2. Business logic: Validate (action responsibility)
    if player.Resources().Credits() < 14 {
        return errors.New("insufficient credits")
    }

    // 3. Business logic: Calculate and apply effects (action responsibility)
    costAfterDiscounts := a.calculateCost(player, 14)
    tempIncrease := a.calculateTempBonus(player)

    // 4. Update state via game methods (game publishes events)
    player.Resources().SubtractCredits(costAfterDiscounts)
    game.GlobalParameters().IncreaseTemperature(ctx, tempIncrease)

    // Events automatically trigger:
    // - BroadcastEvent → Delivery layer creates DTO and broadcasts
    // - TemperatureChangedEvent → Passive card effects activate

    return nil
}
```

## GameRepository Interface

Single repository for game collection management:

```go
type GameRepository interface {
    Get(gameID string) (*Game, error)
    Create(game *Game) error
    Delete(gameID string) error
    List() []*Game
    Exists(gameID string) bool
}
```

## Event System

Game state methods publish domain events automatically:

```go
// In GlobalParameters.IncreaseTemperature()
func (gp *GlobalParameters) IncreaseTemperature(ctx context.Context, steps int) {
    // CRITICAL: Capture values while holding lock, publish AFTER releasing
    var oldTemp, newTemp int

    gp.mu.Lock()
    oldTemp = gp.temperature
    gp.temperature += steps
    newTemp = gp.temperature
    gp.mu.Unlock()

    // Publish events AFTER releasing lock to avoid deadlocks
    gp.eventBus.Publish(TemperatureChangedEvent{
        OldValue: oldTemp,
        NewValue: newTemp,
        Steps:    steps,
    })

    gp.eventBus.Publish(BroadcastEvent{
        GameID:    gp.gameID,
        PlayerIDs: nil, // Broadcast to all
    })
}
```

**Thread Safety Pattern**:
1. ✅ Acquire lock
2. ✅ Capture old values
3. ✅ Update state
4. ✅ Capture new values
5. ✅ Release lock
6. ✅ THEN publish events (never while holding lock)

**Why**: Publishing events while holding a lock can cause deadlocks if event handlers try to acquire the same lock.

Delivery layer listens for `BroadcastEvent`:
```go
// In websocket broadcaster
func (b *Broadcaster) OnBroadcastEvent(event BroadcastEvent) {
    game, _ := b.gameRepo.Get(event.GameID)

    if event.PlayerIDs == nil {
        // Broadcast to all players - each gets personalized view
        for _, player := range game.Players() {
            // playerID determines perspective:
            // - "player" field has full data (complete hand, hidden info)
            // - "otherPlayers" array has limited data (hand size only, no cards)
            dto := mapper.ToPersonalizedGameDTO(game, player.ID())
            b.sendToPlayer(event.GameID, player.ID(), dto)
        }
    } else {
        // Send to specific players - each gets personalized view
        for _, playerID := range event.PlayerIDs {
            dto := mapper.ToPersonalizedGameDTO(game, playerID)
            b.sendToPlayer(event.GameID, playerID, dto)
        }
    }
}
```

**Personalized DTO Pattern**:
- Each player receives their own perspective of the game state
- `playerID` parameter determines which player is "you" vs "others"
- **Full data for receiving player**: Complete hand, hidden selections, full resource details
- **Limited data for other players**: OtherPlayer type with hand size, visible resources only
- Example: When Alice receives game state, she sees her full hand but only Bob's hand size

## Benefits

1. **Clear Separation of Concerns**
   - Actions = business logic (easy to test, reason about)
   - Game = state storage (simple, predictable)

2. **Single Source of Truth**
   - Game owns all state
   - No repository layer duplicating storage
   - No synchronization issues between repositories

3. **Encapsulation Enforced**
   - Private fields prevent accidental mutation
   - Public methods control state changes
   - Thread-safety via mutexes at component level

4. **Simplified Dependencies**
   - Actions need only GameRepository (+ logger)
   - No more 4-5 repository dependencies per action
   - Easier dependency injection

5. **Better Testability**
   - Mock single GameRepository instead of multiple repos
   - Test actions in isolation with mock game state
   - Test game state methods independently

6. **Domain-Driven Design**
   - Business logic in actions (use cases)
   - Domain entities encapsulated
   - Clear boundaries between layers

## Implementation Strategy

### Phase 1: Foundation
- Create `internal/game/` package structure
- Implement encapsulated Game, GlobalParameters, Board, Deck types
- Implement GameRepository with in-memory storage
- Keep existing code running in parallel

### Phase 2: Event System
- Update EventBus to support BroadcastEvent
- Implement delivery layer event listener
- Create DTO generation from gameID + playerID

### Phase 3: Incremental Migration
- Migrate 5-10 representative actions as proof of concept
- Update tests for migrated actions
- Validate event flow and broadcasting works

### Phase 4: Complete Migration
- Migrate remaining ~30 action files
- Update all tests
- Remove `internal/session/` package and old repositories

### Phase 5: Cleanup
- Update handlers in delivery layer
- Simplify main.go dependency injection
- Update documentation (CLAUDE.md)
- Run full test suite and linting

## Success Criteria

- ✅ All game state lives in `internal/game/`
- ✅ Zero business logic in Game/Player/Components
- ✅ All actions use only GameRepository dependency
- ✅ All types have private fields with public methods
- ✅ Event-driven broadcasting works end-to-end
- ✅ All tests passing
- ✅ No linting errors
- ✅ Documentation updated
