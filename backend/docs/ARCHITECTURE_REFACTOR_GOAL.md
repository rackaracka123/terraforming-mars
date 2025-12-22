# Backend Architecture: Action-Driven with Game as State Repository

## Overview

The backend architecture implements clean separation of concerns where **Actions contain all business logic** (validation, orchestration, effect application) and **Game serves as a pure state repository** (getters/setters, event publishing) with fully encapsulated types using private fields and public methods.

This document describes the current architecture, design patterns, and the refactor history that led to the current implementation.

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
│ - Broadcaster subscribes to BroadcastEvent                  │
│ - Fetch game state via GameRepository                       │
│ - Create personalized DTOs                                  │
│ - Broadcast to WebSocket clients                            │
└─────────────────────────────────────────────────────────────┘
```

## What Changed

### Previous Architecture (Before Refactor)

- `internal/session/` package with Session wrapper
- Multiple repositories (game, board, card, deck)
- Actions received Session and multiple repositories
- Mixed responsibility between actions and repositories

### Current Architecture (After Refactor)

- `internal/game/` as root package for all game state
- **Package Structure**:
  - `internal/game/` - Core game types (Game, Card types, PlayerActions, PlayerEffects)
  - `internal/game/shared/` - Simple shared types (Resources, ResourceType, HexPosition, etc.)
  - `internal/game/player/` - Player entity and components (Hand, PlayedCards, Resources, Selection)
  - `internal/game/board/` - Board and Tile types
  - `internal/game/deck/` - Deck management
  - `internal/game/global_parameters/` - GlobalParameters with terraforming constants
- Single GameRepository managing active games
- Game contains: Players, Deck, Board, GlobalParameters, Generation, Phase, CurrentTurn
- Actions receive only GameRepository (via BaseAction)
- Game has zero business logic - pure state container
- EventBus injected into Game, Player, GlobalParameters for event publishing

## EventBus Injection Pattern

The EventBus is injected into domain types during construction to enable event publishing without business logic:

```go
// Game construction
func NewGame(id string, eventBus *events.EventBusImpl) *Game {
    return &Game{
        id:       id,
        eventBus: eventBus,
        players:  make(map[string]*Player),
        // ...
    }
}

// Player construction
func NewPlayer(id, name, gameID string, eventBus *events.EventBusImpl) *Player {
    return &Player{
        id:       id,
        name:     name,
        gameID:   gameID,
        eventBus: eventBus,
        hand:     NewHand(),
        // ...
    }
}

// Usage in state methods
func (g *Game) IncrementGeneration(ctx context.Context) {
    g.mu.Lock()
    oldGen := g.generation
    g.generation++
    newGen := g.generation
    g.mu.Unlock()

    // Publish event AFTER releasing lock
    g.eventBus.Publish(GenerationChangedEvent{
        GameID:   g.id,
        OldValue: oldGen,
        NewValue: newGen,
    })
}
```

**Key Points:**

- EventBus injected as private field in Game, Player, GlobalParameters
- State methods publish events after releasing locks (never while holding)
- No business logic in event publishing - just state notification
- Subscribers (Broadcaster, CardEffectSubscriber) react to events

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
    currentTurn       *Turn  // Track active player and available actions
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

### Turn

```go
type Turn struct {
    mu               sync.RWMutex
    playerID         string
    availableActions []ActionType
    actionsRemaining int
}

// Public methods
func (t *Turn) PlayerID() string
func (t *Turn) AvailableActions() []ActionType
func (t *Turn) CanPerformAction(actionType ActionType) bool
func (t *Turn) DecrementActions()
func (t *Turn) SetActions(actions []ActionType)
```

### Player (in `internal/game/player/` package)

```go
package player

type Player struct {
    // Identity (immutable, private)
    id     string
    name   string
    gameID string

    // Connection status
    connected bool

    // Corporation reference (quick lookup in playedCards)
    corporationID string  // References corporation card ID in playedCards

    // Delegated Components (private, exposed via accessors)
    hand        *Hand
    playedCards *PlayedCards
    resources   *PlayerResources
    selection   *Selection

    // Infrastructure
    eventBus *events.EventBusImpl
}

// Public methods - delegate to components
func (p *Player) ID() string
func (p *Player) Name() string
func (p *Player) IsConnected() bool
func (p *Player) SetConnected(connected bool)
func (p *Player) Hand() *Hand
func (p *Player) PlayedCards() *PlayedCards
func (p *Player) Resources() *PlayerResources
func (p *Player) CorporationID() string
func (p *Player) SetCorporationID(id string)
func (p *Player) HasCorporation() bool
```

**Note on Corporations**: Corporations are treated as special played cards, not a separate component. The `corporationID` field provides quick reference to the player's corporation card, which is stored in `playedCards`. Corporation cards have special fields (`StartingResources`, `StartingProduction`) and can trigger forced actions, but are otherwise handled through the standard card system.

**Note on Actions and Effects**: The `Actions` and `Effects` components (managing available player actions and passive effects) remain in the `game` package, not in the `player` package. This avoids import cycles since these components depend heavily on card behavior types.

### Components (Hand, PlayedCards, Resources, etc.)

```go
package player

type Hand struct {
    mu    sync.RWMutex
    cards []string  // Card IDs only
}

// Simple state operations only
func (h *Hand) Cards() []string
func (h *Hand) Contains(cardID string) bool
func (h *Hand) AddCard(cardID string)
func (h *Hand) RemoveCard(cardID string) error

type PlayedCards struct {
    mu    sync.RWMutex
    cards []string  // ALL played cards including corporation
}

func (pc *PlayedCards) Cards() []string
func (pc *PlayedCards) Contains(cardID string) bool
func (pc *PlayedCards) AddCard(cardID string)  // Used for both corporation and project cards

type PlayerResources struct {
    mu                 sync.RWMutex
    resources          shared.Resources
    production         shared.Production
    terraformRating    int
    victoryPoints      int
    resourceStorage    map[string]int
    paymentSubstitutes []shared.PaymentSubstitute
}

func (r *PlayerResources) Get() shared.Resources
func (r *PlayerResources) Production() shared.Production
func (r *PlayerResources) Add(changes map[shared.ResourceType]int)
func (r *PlayerResources) AddProduction(changes map[shared.ResourceType]int)
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
    // - BroadcastEvent → Broadcaster creates DTO and sends to clients
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

Broadcaster subscribes to `BroadcastEvent`:

```go
// In internal/delivery/websocket/broadcaster.go
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

## Migration History

The refactor was completed using a phased approach with parallel migration directories to ensure the existing system continued working during development.

### Completed Phases

**Phase 1: Foundation** ✅

- Created `internal/game_migration/` package structure
- Implemented encapsulated Game, GlobalParameters, Board, Deck types
- Implemented GameRepository with in-memory storage
- Kept existing `internal/session/` code running in parallel

**Phase 2: Event System** ✅

- Updated EventBus to support BroadcastEvent
- Implemented Broadcaster in delivery layer as event subscriber
- Created DTO generation from gameID + playerID

**Phase 3: Incremental Migration** ✅

- Created `internal/action_migration/` directory
- Migrated representative actions using GameRepository from `internal/game_migration/`
- Validated event flow and broadcasting worked end-to-end

**Phase 4: Complete Migration** ✅

- Migrated all action files to `internal/action_migration/`
- Updated tests to use new architecture
- All migrated actions using `internal/game_migration/`

**Phase 5: Final Cutover** ✅

- Renamed `internal/game_migration/` → `internal/game/`
- Renamed `internal/action_migration/` → `internal/action/`
- Removed `internal/session/` package and old repositories
- Updated all imports across codebase
- Simplified main.go dependency injection

**Phase 6: Cleanup** ⚠️ In Progress

- Tests passing ✅
- Linting clean ✅
- Documentation updates ongoing 🔄

## Success Criteria

### ✅ Completed

- All game state lives in `internal/game/` (verified in codebase)
- Zero business logic in Game/Player/Components (encapsulation enforced)
- All actions use only GameRepository dependency (via BaseAction)
- All types have private fields with public methods (Game, Player, GlobalParameters, etc.)
- Event-driven broadcasting works end-to-end (Broadcaster subscribes to BroadcastEvent)
- All tests passing (confirmed via `make test`)
- No linting errors (confirmed via `make lint`)

### 🔄 In Progress

- Documentation updates (this file being updated now)

### 📋 Known Outstanding Items

**Active TODOs in Codebase:**

- `backend/internal/game/game.go:402` - Implement proper turn order mechanism (currently uses map iteration order which is non-deterministic)
- `backend/internal/action/convert_plants_to_greenery.go:80` - Reimplement card discount effects when card system is migrated
- `backend/internal/action/convert_heat.go:76` - Reimplement card discount effects when card system is migrated
- `backend/internal/delivery/dto/mapper_game.go:238,298` - Implement production phase mapping

**Documentation Status:**

- ✅ CLAUDE.md files updated to reflect `internal/game/` structure
- ✅ Broadcaster terminology clarified (removed SessionManager references)
- 🔄 Test directories still use `test/session/` naming (consider renaming to `test/game/` in future cleanup)
