# Backend Architecture Flow

This document explains the event-driven backend architecture and how different layers interact when processing game actions.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Aggregate Roots vs Features](#aggregate-roots-vs-features)
3. [Complete Example: Playing a Card](#complete-example-playing-a-card)
4. [Layer Responsibilities](#layer-responsibilities)
5. [Event-Driven Architecture](#event-driven-architecture)
6. [Code Examples](#code-examples)
7. [When to Use Which Layer](#when-to-use-which-layer)

---

## Architecture Overview

The backend follows clean architecture principles with strict separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                     Presentation Layer                       │
│           (WebSocket Handlers, HTTP Endpoints)               │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ↓
┌─────────────────────────────────────────────────────────────┐
│                      Actions Layer                           │
│         (Orchestrate, Validate, Coordinate)                  │
│              internal/actions/                               │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ↓
┌─────────────────────────────────────────────────────────────┐
│                     Session Layer                            │
│        (Runtime game container: Game + Players + Features)   │
│              internal/session/                               │
│   ┌──────────────────────────────────────────────────────┐  │
│   │  Session {                                           │  │
│   │    Game (aggregate root)                             │  │
│   │    Players map[playerID]*Player (aggregate roots)    │  │
│   │    ParametersService, BoardService, CardService...   │  │
│   │  }                                                    │  │
│   └──────────────────────────────────────────────────────┘  │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        ↓                             ↓
┌───────────────────┐     ┌───────────────────────┐
│ Services Layer    │     │ Features Layer        │
│ (High-level coord)│     │ (Domain-specific ops) │
│ internal/service/ │     │ internal/features/    │
└─────────┬─────────┘     └──────────┬────────────┘
          │                          │
          └────────────┬─────────────┘
                       │
                       ↓
┌─────────────────────────────────────────────────────────────┐
│                   Repositories Layer                         │
│        (Granular state updates + Event publishing)           │
│         features/*/repository.go                             │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ↓
┌─────────────────────────────────────────────────────────────┐
│                      Event System                            │
│         (EventBus publishes domain events)                   │
│              internal/events/                                │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        ↓                             ↓
┌───────────────────┐     ┌───────────────────────┐
│ Player TR         │     │ Card Passive Effects  │
│ Subscriptions     │     │ (CardEffectSubscriber)│
│ (On global param  │     │ (On various events)   │
│  changes)         │     │                       │
└───────────────────┘     └───────────────────────┘
```

### Key Principle: Event-Driven Separation

**CRITICAL**: Features operate in isolation and have NO knowledge of other domains.

- **Parameters feature** raises temperature → publishes `TemperatureChangedEvent` → DOES NOT update player TR
- **Player TR subscription** listens to `TemperatureChangedEvent` → updates TR automatically
- **Card passive effects** listen to domain events → trigger automatically when conditions are met

---

## Aggregate Roots vs Features

### Critical Architectural Principle

**PLAYER AND GAME ARE AGGREGATE ROOTS, NOT FEATURES.**

This is a fundamental distinction that prevents circular dependencies and maintains clean architecture:

```
internal/
├── player/          ← AGGREGATE ROOT (holds feature service references)
├── game/            ← AGGREGATE ROOT (holds feature service references)
└── features/        ← PURE DOMAIN SERVICES (NO knowledge of Player/Game)
    ├── card/
    ├── tiles/
    ├── parameters/
    ├── resources/
    └── ...
```

### Why Player/Game Are Not Features

**Player** and **Game** are **aggregate roots** that:
1. **Coordinate multiple features** - they hold references to feature services
2. **Represent game entities** - they have identity and lifecycle
3. **Are known by all layers** - actions, services, and features all reference them
4. **Live at** `internal/player` and `internal/game` (NOT in `internal/features/`)

**Features** are **pure domain services** that:
1. **Have ZERO knowledge of Player/Game** - they never import player or game packages
2. **Operate on specific domains** - parameters, tiles, resources, etc.
3. **Are stateless** - they operate via repositories scoped by gameID/playerID
4. **Live at** `internal/features/*`

### Player/Game Hold Feature Service References

```go
// internal/player/model.go
type Player struct {
    // Identity and data
    ID              string
    Name            string
    Cards           []string
    TerraformRating int

    // Feature Services (injected by repository)
    ResourcesService  resources.Service      `json:"-"`
    ProductionService production.Service     `json:"-"`
    TileQueueService  tiles.TileQueueService `json:"-"`
    PlayerTurnService turn.PlayerTurnService `json:"-"`

    // Accessor methods that delegate to feature services
    GetResources() (resources.Resources, error)
    GetProduction() (resources.Production, error)
    GetPassed() (bool, error)
}
```

When you retrieve a Player from the repository, it comes **pre-configured** with its feature services:

```go
// Actions layer
player, err := playerRepo.GetByID(ctx, gameID, playerID)

// Player already has feature services attached!
resources, err := player.GetResources()  // Uses player.ResourcesService
player.ResourcesService.Update(ctx, newResources)
```

### Features Must NOT Import Player/Game

**❌ WRONG - Feature knows about Player:**
```go
// internal/features/card/service.go - WRONG!
package card

import "internal/player"  // ❌ Circular dependency!

func (s *Service) DrawCard(...) error {
    player, err := s.playerRepo.GetByID(...)  // ❌ Feature knows about Player!
    player.Cards = append(player.Cards, card) // ❌ Feature accesses Player directly!
    return s.playerRepo.Update(ctx, player)
}
```

**✅ CORRECT - Feature is pure:**
```go
// internal/features/card/draw_service.go - CORRECT!
package card

// No player or game imports!

type DrawService interface {
    // Pure function - takes current cards, returns updated cards
    DrawCards(ctx context.Context, currentHand []Card, count int) ([]Card, error)
}

func (s *DrawServiceImpl) DrawCards(ctx context.Context, currentHand []Card, count int) ([]Card, error) {
    // Draw from deck repository (card feature's own repo)
    drawnCards, err := s.deckRepo.Draw(ctx, count)
    if err != nil {
        return nil, err
    }

    // Return new hand - caller (Actions) updates player via PlayerRepository
    return append(currentHand, drawnCards...), nil
}
```

### Actions Layer Orchestrates Using Aggregate Roots

The **Actions layer** is responsible for:
1. Getting Player/Game aggregates from repositories
2. Extracting data needed by features
3. Calling feature services with pure data
4. Applying results using Player/Game feature services

**Example - Correct orchestration:**
```go
// internal/actions/play_card.go
func (a *PlayCardAction) Execute(ctx context.Context, gameID, playerID, cardID string, payment Payment) error {
    // 1. Get aggregates (they have feature services attached)
    player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
    game, err := a.gameRepo.GetByID(ctx, gameID)

    // 2. Find card in player's hand (card instance with modifiers already applied)
    var cardToPlay *Card
    for i := range player.Cards {
        if player.Cards[i].ID == cardID {
            cardToPlay = &player.Cards[i]
            break
        }
    }

    // 3. Validate requirements against global parameters (with card modifier leniency)
    globalParams, _ := game.GetGlobalParameters()

    if cardToPlay.Requirements.MinTemperature != nil {
        requiredTemp := *cardToPlay.Requirements.MinTemperature
        // Apply card modifiers that reduce temperature requirement
        for _, mod := range cardToPlay.RequirementModifiers {
            if mod.Type == "temperature" {
                requiredTemp += mod.Adjustment  // e.g., -2 reduces requirement
            }
        }
        if globalParams.Temperature < requiredTemp {
            return errors.New("temperature requirement not met")
        }
    }

    // 4. Calculate final cost using card modifiers (already applied to card instance)
    finalCost := cardToPlay.GetFinalCost()  // Base cost + all cost modifiers

    // 5. Validate payment (steel for building tags, titanium for space tags)
    paymentValue := payment.Credits
    if hasTag(cardToPlay.Tags, "building") {
        paymentValue += payment.Steel * player.SteelValue
    }
    if hasTag(cardToPlay.Tags, "space") {
        paymentValue += payment.Titanium * player.TitaniumValue
    }

    if paymentValue < finalCost {
        return errors.New("insufficient payment")
    }

    // 6. Deduct resources using Player's feature service
    resources, _ := player.GetResources()
    resources.Credits -= payment.Credits
    resources.Steel -= payment.Steel
    resources.Titanium -= payment.Titanium
    player.ResourcesService.Update(ctx, resources)

    // 7. Move card from hand to played cards
    a.playerRepo.RemoveCardFromHand(ctx, gameID, playerID, cardID)
    a.playerRepo.AddPlayedCard(ctx, gameID, playerID, cardToPlay)

    // 8. Process card effects using feature services
    a.processCardEffects(ctx, gameID, playerID, cardToPlay)

    return nil
}
```

### Repository Pattern for Feature Services

Game and Player repositories create and manage per-game feature service instances:

```go
// internal/game/repository.go
type RepositoryImpl struct {
    games    map[string]*Game
    eventBus *events.EventBusImpl

    // Feature repositories (scoped by gameID)
    parametersRepos map[string]parameters.Repository
    boardRepos      map[string]tiles.BoardRepository
    turnOrderRepos  map[string]turn.TurnOrderRepository
}

func (r *RepositoryImpl) Create(ctx context.Context, settings GameSettings) (Game, error) {
    gameID := uuid.New().String()

    // Create per-game feature repositories
    parametersRepo, _ := parameters.NewRepository(gameID, initialParams, r.eventBus)
    r.parametersRepos[gameID] = parametersRepo

    boardRepo := tiles.NewBoardRepository(gameID, initialBoard, r.eventBus)
    r.boardRepos[gameID] = boardRepo

    // Create game with feature services attached
    game := NewGame(gameID, settings)
    game.ParametersService = parameters.NewService(parametersRepo)
    game.BoardService = tiles.NewBoardService(boardRepo)

    return game, nil
}
```

### Key Benefits of This Pattern

**✅ No Circular Dependencies**
- Features don't import player/game
- Player/game can safely import features
- Clean dependency flow: features → player/game → actions

**✅ Pure Feature Services**
- Features are testable in isolation
- No mocking complex aggregates
- Clear inputs and outputs

**✅ Clear Separation of Concerns**
- Features: domain logic
- Player/Game: aggregate coordination
- Actions: orchestration
- Repositories: state management

**✅ Scalability**
- Easy to add new features without touching player/game
- Features can be developed independently
- Clear contracts between layers

---

## Session: The Runtime Game Container

### Overview

**Session** is a runtime aggregate that wraps Player, Game, and feature services for a single active game instance. While Game and Player are domain entities representing game state, Session is the **operational container** that manages the lifecycle and coordination of an active game.

### Session vs Game/Player

**Game and Player** are **domain entities**:
- Represent persistent game state
- Stored in repositories
- Serialized/deserialized for persistence
- Focus on "what the game is"

**Session** is a **runtime aggregate**:
- Wraps Game + all Players + feature services
- Created when game starts, destroyed when game ends
- Never persisted (recreated from Game/Player on server restart)
- Focus on "how the game runs"

### Session Structure

```go
// internal/session/session.go
type Session struct {
    // Core aggregates
    GameID   string
    Game     *game.Game
    Players  map[string]*player.Player  // playerID -> Player

    // Feature service instances (scoped to this game)
    ParametersService parameters.Service
    BoardService      tiles.BoardService
    CardService       card.Service
    TurnOrderService  turn.TurnOrderService

    // Session metadata
    CreatedAt time.Time
    LastActivity time.Time
    HostPlayerID string
}
```

### Session Lifecycle

**1. Creation (Lobby → Session)**
```go
// internal/lobby/service.go
func (s *LobbyService) StartGame(ctx context.Context, gameID string) (*session.Session, error) {
    // Get lobby game
    game, err := s.gameRepo.GetByID(ctx, gameID)

    // Get all players in the game
    players, err := s.playerRepo.ListByGameID(ctx, gameID)

    // Create feature service instances for this game
    parametersService := parameters.NewService(
        parameters.NewRepository(gameID, initialParams, s.eventBus),
    )
    boardService := tiles.NewBoardService(
        tiles.NewBoardRepository(gameID, initialBoard, s.eventBus),
    )

    // Create session
    session := &session.Session{
        GameID:            gameID,
        Game:              game,
        Players:           playersMap,
        ParametersService: parametersService,
        BoardService:      boardService,
        CreatedAt:         time.Now(),
        HostPlayerID:      game.HostPlayerID,
    }

    // Register in session repository
    s.sessionRepo.Add(ctx, session)

    return session, nil
}
```

**2. Active Gameplay**
```go
// Actions access game state via session
func (a *PlayCardAction) Execute(ctx context.Context, gameID, playerID, cardID string, payment Payment) error {
    // Get session (contains everything needed)
    session, err := a.sessionRepo.GetByID(ctx, gameID)

    // Access player from session
    player, exists := session.Players[playerID]
    if !exists {
        return errors.New("player not in game")
    }

    // Access game from session
    globalParams := session.Game.GetGlobalParameters()

    // Use session's feature services
    session.ParametersService.RaiseTemperature(ctx, 2)
    session.BoardService.PlaceTile(ctx, coord, tile)

    return nil
}
```

**3. Cleanup (Game End)**
```go
// When game ends, session is removed
func (s *GameService) EndGame(ctx context.Context, gameID string) error {
    // Final state persisted to repositories
    session, _ := s.sessionRepo.GetByID(ctx, gameID)
    s.gameRepo.Update(ctx, session.Game)
    for _, player := range session.Players {
        s.playerRepo.Update(ctx, gameID, player)
    }

    // Remove session (game state remains in repos)
    s.sessionRepo.Remove(ctx, gameID)
}
```

### Session Repository

```go
// internal/session/repository.go
type Repository interface {
    Add(ctx context.Context, session *Session) error
    GetByID(ctx context.Context, gameID string) (*Session, error)
    Remove(ctx context.Context, gameID string) error
    ListActive(ctx context.Context) ([]*Session, error)
}

type RepositoryImpl struct {
    sessions map[string]*Session  // gameID -> Session
    mu       sync.RWMutex
}
```

### Key Benefits

**✅ Unified Game Access**
- Single object contains all game state and services
- No need to fetch Game, Players, and services separately
- Simplifies action signatures: just pass gameID

**✅ Clean Lifecycle Management**
- Sessions created when lobby starts game
- Sessions destroyed when game ends
- Automatic cleanup of feature service instances

**✅ Scoped Feature Services**
- Each session has its own feature service instances
- Feature repositories scoped to specific gameID
- No cross-game state pollution

**✅ Simplified Actions**
```go
// Before (no sessions):
func (a *Action) Execute(ctx, gameID, playerID string) error {
    game, _ := gameRepo.GetByID(ctx, gameID)
    player, _ := playerRepo.GetByID(ctx, gameID, playerID)
    paramsService := /* need to find/create */
    boardService := /* need to find/create */
    // ...
}

// After (with sessions):
func (a *Action) Execute(ctx, gameID, playerID string) error {
    session, _ := sessionRepo.GetByID(ctx, gameID)
    // Everything is already here!
    player := session.Players[playerID]
    session.ParametersService.RaiseTemperature(...)
}
```

**✅ Server Restart Recovery**
- Sessions are runtime-only (not persisted)
- On restart: load Games and Players from repos
- Recreate Sessions for active games
- Feature services reinitialized from persisted state

### Session vs Service Layer

**Session is NOT a service**:
- Session is a **data structure** (aggregate of aggregates)
- Services are **behavior** (use cases and business logic)
- Actions use SessionRepository to get sessions
- Actions use Services for business operations
- Services access session data, don't replace it

---

## Complete Example: Playing a Card

Let's walk through what happens when a player plays a card that says **"Increase temperature by 2 steps. Gain 3 plants."**

### Step 1: WebSocket Handler (Presentation Layer)

```
Client sends: { type: "play-card", cardId: "capital", payment: {...} }
    ↓
WebSocket Hub receives message
    ↓
Manager routes to PlayCardHandler
    ↓
Handler calls: playCardAction.Execute(ctx, gameID, playerID, cardID, payment, ...)
```

### Step 2: Action Orchestrates (Actions Layer)

```go
// internal/actions/play_card.go
func (a *PlayCardAction) Execute(ctx, gameID, playerID, cardID string, payment Payment) error {
    // STEP 1: Get session (contains game, players, and feature services)
    session, _ := a.sessionRepo.GetByID(ctx, gameID)
    player := session.Players[playerID]
    cardToPlay := findCardInHand(player.Cards, cardID)

    // STEP 2: Validate requirements against global parameters (with modifier leniency)
    globalParams := session.Game.GetGlobalParameters()

    if cardToPlay.Requirements.MinTemperature != nil {
        requiredTemp := *cardToPlay.Requirements.MinTemperature
        // Apply card modifiers that reduce temperature requirement
        for _, mod := range cardToPlay.RequirementModifiers {
            if mod.Type == "temperature" {
                requiredTemp += mod.Adjustment  // e.g., -2 reduces requirement
            }
        }
        if globalParams.Temperature < requiredTemp {
            return errors.New("temperature requirement not met")
        }
    }

    // STEP 3: Validate payment against final cost
    finalCost := cardToPlay.GetFinalCost()  // Cost modifiers already summed

    paymentValue := payment.Credits
    if hasTag(cardToPlay.Tags, "building") {
        paymentValue += payment.Steel * player.SteelValue  // Resource value already set
    }
    if hasTag(cardToPlay.Tags, "space") {
        paymentValue += payment.Titanium * player.TitaniumValue
    }

    if paymentValue < finalCost {
        return errors.New("insufficient payment")
    }

    // STEP 4: Deduct resources via player's resource service
    resources, _ := player.GetResources()
    resources.Credits -= payment.Credits
    resources.Steel -= payment.Steel
    resources.Titanium -= payment.Titanium
    player.ResourcesService.Update(ctx, resources)

    // STEP 5: Move card from hand to played cards
    a.playerRepo.RemoveCardFromHand(ctx, gameID, playerID, cardID)
    a.playerRepo.AddPlayedCard(ctx, gameID, playerID, cardToPlay)

    // STEP 6: Process immediate card effects using session's feature services
    a.processCardEffects(ctx, session, playerID, cardToPlay)

    // STEP 7: Subscribe passive effects (event-driven)
    a.effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, cardID, cardToPlay)

    // STEP 8: Register manual actions
    a.registerManualActions(ctx, gameID, playerID, cardID, cardToPlay)

    // STEP 9: Consume player action
    a.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, newActions)

    // STEP 10: Broadcast game state to all players
    a.sessionManager.Broadcast(gameID)

    return nil
}
```

### Step 3: Process Card Effects (Action orchestrates features)

```go
// internal/actions/play_card.go - processCardEffects()
func (a *PlayCardAction) processCardEffects(ctx context.Context, session *Session, playerID string, card *Card) error {
    player := session.Players[playerID]

    for _, output := range card.Behavior.Outputs {
        switch output.Type {

        // Temperature increase - use session's parameters service
        case resources.ResourceTemperature:
            steps := output.Amount / 2  // Temperature increases in 2°C steps
            if steps > 0 {
                // Call session's parameters feature (NO TR handling here!)
                actualSteps, err := session.ParametersService.RaiseTemperature(ctx, steps)
                // Parameters feature handles ONLY temperature
                // Events will handle TR increase automatically
            }

        // Direct resource gain - update player resources
        case resources.ResourcePlants:
            resources, _ := player.GetResources()
            resources.Plants += output.Amount
            player.ResourcesService.Update(ctx, resources)
        }
    }

    return nil
}
```

### Step 4: Parameters Feature (Domain-specific, NO player knowledge)

```go
// internal/features/parameters/service.go
func (s *Service) RaiseTemperature(ctx context.Context, steps int) (int, error) {
    // ONLY update temperature state via repository
    actualSteps, err := s.repo.IncreaseTemperature(ctx, steps)
    if err != nil {
        return 0, err
    }

    // NO TR updates here!
    // NO player logic here!
    // Repository already published TemperatureChangedEvent

    return actualSteps, nil
}
```

### Step 5: Repository Updates State and Publishes Event

```go
// internal/features/parameters/repository.go
func (r *RepositoryImpl) IncreaseTemperature(ctx context.Context, steps int) (int, error) {
    r.mu.Lock()
    defer r.mu.Unlock()

    currentTemp := r.params.Temperature
    newTemp := currentTemp + (steps * 2)  // Each step = 2°C

    // Cap at maximum (8°C)
    actualSteps := steps
    if newTemp > MaxTemperature {
        newTemp = MaxTemperature
        actualSteps = (MaxTemperature - currentTemp) / 2
    }

    // Update state
    r.params.Temperature = newTemp

    // Publish domain event (NO side effects)
    if r.eventBus != nil && currentTemp != newTemp {
        events.Publish(r.eventBus, events.TemperatureChangedEvent{
            GameID:    r.gameID,
            OldValue:  currentTemp,
            NewValue:  newTemp,
            ChangedBy: "", // Can be set by service if needed
            Timestamp: time.Now(),
        })
    }

    return actualSteps, nil
}
```

### Step 6: Event Subscribers React Automatically

```
TemperatureChangedEvent published to EventBus
    ↓
┌──────────────────────────────────────────────────────────┐
│ EventBus notifies ALL subscribers                        │
└────────────┬─────────────────────────────────────────────┘
             │
     ┌───────┴────────┐
     ↓                ↓
┌─────────────┐  ┌─────────────────────────┐
│ Player TR   │  │ Card Passive Effects    │
│ Subscriber  │  │ (CardEffectSubscriber)  │
└─────────────┘  └─────────────────────────┘
     │                ↓
     │           Checks all player cards
     │           for "when temperature increases"
     │           triggers and applies bonuses
     │
     ↓
Automatically increases player TR:
  - Player who raised temp gets +2 TR (for 2 steps)
  - playerRepo.IncreaseTR(ctx, gameID, playerID, 2)
  - Publishes TerraformRatingChangedEvent
```

### Step 7: SessionManager Broadcasts Complete State

```go
// internal/delivery/websocket/session/manager.go
func (sm *SessionManagerImpl) Broadcast(gameID string) error {
    // Fetch complete game state from repositories
    game, err := sm.gameRepo.GetByID(ctx, gameID)
    players, err := sm.playerRepo.ListByGameID(ctx, gameID)

    // Resolve all card data
    resolvedCards, err := sm.cardRepo.ListCardsByIdMap(ctx, allCardIds)

    // Create personalized DTOs for each player
    for _, player := range players {
        personalizedGameDTO := dto.ToGameDto(game, players, player.ID, resolvedCards, ...)

        // Send complete game state to player via WebSocket Hub
        sm.hub.SendToPlayer(gameID, player.ID, GameUpdatedMessage{
            Game: personalizedGameDTO,
        })
    }

    return nil
}
```

### Step 8: Clients Receive Update

```
All connected clients receive "game-updated" message
    ↓
Frontend React components re-render with new state:
  - Temperature display shows increased value
  - Player TR shows +2 increase
  - Player plants shows +3 increase
  - Card moved from hand to played cards
```

---

## Layer Responsibilities

### Actions Layer (`internal/actions/`)

**Purpose**: Orchestrate complex, multi-step user-initiated operations

**Responsibilities**:
- Validate inputs (turn order, ownership, requirements)
- Coordinate multiple features and services
- Handle payment and resource deductions
- Process card effects by calling appropriate features
- Manage state transitions (move cards, consume actions)
- Trigger broadcasting via SessionManager

**Does NOT**:
- Implement domain logic (delegates to features)
- Directly manipulate low-level state (uses repositories)
- Know about WebSocket internals (uses SessionManager)

**Example**: `PlayCardAction`, `SelectTileAction`, `SkipActionAction`

---

### Services Layer (`internal/service/`)

**Purpose**: High-level business logic coordination (legacy compatibility layer)

**Responsibilities**:
- Coordinate between multiple features
- Handle complex game flows (turn progression, phase transitions)
- Provide backward-compatible interfaces
- Delegate to features for domain-specific operations

**Example**: `GameService`, `PlayerService`

**Note**: New code should prefer Actions and Features over Services when possible.

---

### Features Layer (`internal/features/`)

**Purpose**: Domain-specific operations with clean boundaries

**Key Principle**: **ZERO cross-domain knowledge**

Each feature domain operates in complete isolation:

#### Parameters Feature (`features/parameters/`)
- Manages global parameters (temperature, oxygen, oceans)
- **Does NOT** know about players or TR
- **Does NOT** know about cards or effects
- Publishes events when parameters change

#### Resources Feature (`features/resources/`)
- Manages resource types and production
- Handles payment calculations
- Discount calculations

#### Tiles Feature (`features/tiles/`)
- Manages board state and tile placement
- Validates tile placement rules
- **Does NOT** award TR (publishes events instead)
- **Does NOT** handle adjacency bonuses directly (publishes events)

#### Card Feature (`features/card/`)
- Card registry and lookup
- Card cost and requirement modifiers - **stored on card instances, applied via events**
- Card effect processing - **returns effect results, does NOT update repos**
- Card draw/deck management - **ONLY accesses CardRepository and CardDeckRepository**
- **Does NOT** know about Player or Game
- **Does NOT** import player or game packages
- **Does NOT** have PlayerRepository or GameRepository dependencies
- **NO validation service** - validation happens in Actions layer using card modifiers

**CRITICAL**: Card features manage card data and modifiers. The Actions layer validates by checking card.GetFinalCost() for pricing, card requirements against global parameters, and applies card modifier leniency.

**Pattern**:
```go
// Feature interface
type Service interface {
    RaiseTemperature(ctx context.Context, steps int) (int, error)
}

// Feature implementation
type ServiceImpl struct {
    repo Repository
}

func (s *ServiceImpl) RaiseTemperature(ctx context.Context, steps int) (int, error) {
    // ONLY call repository - NO cross-domain logic
    return s.repo.IncreaseTemperature(ctx, steps)
}
```

---

### Repositories Layer

**Purpose**: Granular state updates and event publishing

**Responsibilities**:
- Store and retrieve domain state
- Provide granular update methods (e.g., `UpdateResources`, `IncreaseTR`)
- Publish domain events via EventBus after state changes
- Return values (not pointers) for immutability

**Does NOT**:
- Contain business logic
- Know about other repositories or domains
- Handle event subscriptions (only publishes)

**Pattern**:
```go
func (r *Repository) IncreaseTemperature(ctx context.Context, steps int) (int, error) {
    // 1. Update state
    r.params.Temperature = newTemp

    // 2. Publish event
    events.Publish(r.eventBus, events.TemperatureChangedEvent{
        GameID: r.gameID,
        OldValue: oldTemp,
        NewValue: newTemp,
    })

    // 3. Return result
    return actualSteps, nil
}
```

---

### Event System (`internal/events/`)

**Purpose**: Decouple domain features from cross-domain side effects

**Components**:

1. **EventBus** - Type-safe publish/subscribe system
2. **Domain Events** - Strongly-typed event definitions
3. **Event Subscribers** - Registered handlers for events

**Event Flow**:
```
Repository publishes event
    ↓
EventBus receives event
    ↓
EventBus notifies all subscribers
    ↓
Subscribers execute handlers
```

**Event Subscription Types**:

#### 1. Player TR Subscriptions (Registered when player joins game)
```go
// When player is added to game:
eventBus.Subscribe(events.TemperatureChangedEvent{}, func(event) {
    if event.GameID == gameID {
        steps := (event.NewValue - event.OldValue) / 2
        playerRepo.IncreaseTR(ctx, gameID, playerID, steps)
    }
})

eventBus.Subscribe(events.OxygenChangedEvent{}, func(event) {
    if event.GameID == gameID {
        steps := event.NewValue - event.OldValue
        playerRepo.IncreaseTR(ctx, gameID, playerID, steps)
    }
})

eventBus.Subscribe(events.OceansChangedEvent{}, func(event) {
    if event.GameID == gameID {
        playerRepo.IncreaseTR(ctx, gameID, playerID, 1) // +1 TR per ocean
    }
})
```

#### 2. Card Passive Effect Subscriptions (Registered when card is played)
```go
// CardEffectSubscriber handles card passive effects
// Example: "When any city is placed, gain 2 MC"
eventBus.Subscribe(events.TilePlacedEvent{}, func(event) {
    if event.TileType == "city" {
        // Check all players' cards for passive effects triggered by city placement
        // Apply bonuses automatically
    }
})
```

---

## Event-Driven Architecture

### Why Event-Driven?

**Problem**: Without events, features would need to know about other domains:
```go
// ❌ BAD: Parameters feature knows about players
func (s *ParametersService) RaiseTemperature(steps int) {
    repo.IncreaseTemperature(steps)
    playerRepo.IncreaseTR(playerID, steps) // Cross-domain coupling!
}
```

**Solution**: Use events to decouple domains:
```go
// ✅ GOOD: Parameters feature is isolated
func (s *ParametersService) RaiseTemperature(steps int) {
    return repo.IncreaseTemperature(steps) // Just updates temperature
}

// Separate subscription handles TR (registered when player joins)
eventBus.Subscribe(TemperatureChangedEvent, func(event) {
    playerRepo.IncreaseTR(gameID, playerID, steps)
})
```

### Event Registration Lifecycle

**1. Game Initialization** (`lobby/service.go` or similar)
```go
func (s *LobbyService) CreateGame(...) {
    // Create game
    game := game.NewGame(...)

    // Initialize event subscriptions for game-level events
    s.subscribeGameEvents(gameID)
}
```

**2. Player Joins Game** (`lobby/service.go`)
```go
func (s *LobbyService) JoinGame(...) {
    // Add player to game
    player := player.NewPlayer(...)

    // Register player's global parameter TR subscriptions
    s.subscribePlayerToGlobalParameters(gameID, playerID)
}

func (s *LobbyService) subscribePlayerToGlobalParameters(gameID, playerID string) {
    // Temperature changes → increase TR
    s.eventBus.Subscribe(events.TemperatureChangedEvent{}, func(event events.TemperatureChangedEvent) {
        if event.GameID != gameID {
            return
        }
        steps := (event.NewValue - event.OldValue) / 2
        if steps > 0 {
            s.playerRepo.IncreaseTR(context.Background(), gameID, playerID, steps)
        }
    })

    // Similar subscriptions for oxygen, oceans, etc.
}
```

**3. Player Plays Card** (`actions/play_card.go`)
```go
func (a *PlayCardAction) Execute(...) {
    // ... play card logic ...

    // Register passive effect subscriptions for this card
    a.effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, cardID, cardObj)
}
```

---

## Code Examples

### ✅ CORRECT: Event-Driven Separation

**Parameters Feature** (NO player knowledge):
```go
// internal/features/parameters/service.go
type Service interface {
    RaiseTemperature(ctx context.Context, steps int) (int, error)
    RaiseOxygen(ctx context.Context, steps int) (int, error)
    PlaceOcean(ctx context.Context) error
}

type ServiceImpl struct {
    repo Repository
}

func (s *ServiceImpl) RaiseTemperature(ctx context.Context, steps int) (int, error) {
    // ONLY delegates to repository - NO cross-domain logic
    actualSteps, err := s.repo.IncreaseTemperature(ctx, steps)
    return actualSteps, err
}
```

**Repository** (publishes event):
```go
// internal/features/parameters/repository.go
func (r *RepositoryImpl) IncreaseTemperature(ctx context.Context, steps int) (int, error) {
    r.mu.Lock()
    defer r.mu.Unlock()

    currentTemp := r.params.Temperature
    newTemp := currentTemp + (steps * 2)

    // Cap at maximum
    if newTemp > MaxTemperature {
        newTemp = MaxTemperature
        actualSteps = (MaxTemperature - currentTemp) / 2
    }

    // Update state
    r.params.Temperature = newTemp

    // Publish event (NO side effects)
    events.Publish(r.eventBus, events.TemperatureChangedEvent{
        GameID:    r.gameID,
        OldValue:  currentTemp,
        NewValue:  newTemp,
        Timestamp: time.Now(),
    })

    return actualSteps, nil
}
```

**Player Subscription** (handles TR increase):
```go
// Registered when player joins game
eventBus.Subscribe(events.TemperatureChangedEvent{}, func(event events.TemperatureChangedEvent) {
    if event.GameID != gameID {
        return // Only react to events for this game
    }

    steps := (event.NewValue - event.OldValue) / 2
    if steps > 0 {
        // Award TR to player who contributed to terraforming
        playerRepo.IncreaseTR(ctx, gameID, playerID, steps)
    }
})
```

**Payment Validation** (event-driven modifiers):
```go
// Actions calculate payment using pre-stored values (no service needed)
func (a *PlayCardAction) Execute(ctx, gameID, playerID, cardID string, payment Payment) error {
    // Get player (has cards with modifiers AND resource values already set)
    player, _ := a.playerRepo.GetByID(ctx, gameID, playerID)
    cardToPlay := findCardInHand(player.Cards, cardID)

    // Calculate final cost - modifiers already summed via events
    finalCost := cardToPlay.GetFinalCost()  // 26 + (-2) = 24

    // Calculate payment value - resource values already set via events
    paymentValue := payment.Credits
    if hasTag(cardToPlay.Tags, "building") {
        paymentValue += payment.Steel * player.SteelValue  // 2 or 3, pre-set via events
    }
    if hasTag(cardToPlay.Tags, "space") {
        paymentValue += payment.Titanium * player.TitaniumValue  // 3 or 4, pre-set via events
    }

    // Simple validation
    if paymentValue < finalCost {
        return errors.New("insufficient payment")
    }

    // Deduct resources and continue...
}
```

---

### ❌ WRONG: Cross-Domain Coupling

**DON'T DO THIS**:
```go
// ❌ Parameters feature knows about players and TR
func (s *ParametersService) RaiseTemperature(ctx context.Context, playerID string, steps int) {
    actualSteps := s.repo.IncreaseTemperature(ctx, steps)

    // ❌ Cross-domain coupling! Parameters shouldn't know about players
    s.playerRepo.IncreaseTR(ctx, gameID, playerID, actualSteps)
}

// ❌ Action directly updates TR instead of using events
func (a *PlayCardAction) processCardEffects(...) {
    actualSteps := a.parametersFeature.RaiseTemperature(ctx, 2)

    // ❌ Manually updating TR instead of using event system
    player.TerraformRating += actualSteps
    playerRepo.Update(ctx, gameID, player)
}

// ❌ Manually checking for passive card effects
func (s *TilesService) PlaceTile(...) {
    // Place tile
    repo.UpdateTileOccupancy(...)

    // ❌ Don't manually loop through cards
    for _, card := range player.Cards {
        if card.TriggersOnTilePlacement {
            applyEffect(card)
        }
    }
}

// ❌ Manually calculating discounts during payment validation
func (a *PlayCardAction) Execute(...) {
    // ❌ Recalculating card cost discounts instead of using GetFinalCost()
    finalCost := card.BaseCost
    for _, playedCard := range player.PlayedCards {
        if playedCard.HasDiscountFor(card.Tags) {
            finalCost -= playedCard.DiscountAmount
        }
    }

    // ❌ Recalculating resource values instead of using player.SteelValue
    steelValue := 2
    for _, playedCard := range player.PlayedCards {
        if playedCard.ID == "steelworks" {
            steelValue = 3
        }
    }
    paymentValue := payment.Credits + (payment.Steel * steelValue)

    if paymentValue < finalCost {
        return errors.New("insufficient payment")
    }
}
```

**Why it's wrong**:
- Violates single responsibility principle
- Creates tight coupling between domains
- Hard to test in isolation
- Difficult to extend with new game mechanics
- Bypasses the event system that handles passive effects
- Redundant calculation - modifiers are already applied via events

---

## When to Use Which Layer

### Use **Actions** when:
- Processing user-initiated operations (play card, select tile, skip action)
- Orchestrating multiple steps with validation
- Coordinating multiple features together
- Managing complex state transitions

**Example**: `PlayCardAction`, `ConvertPlantsAction`, `SelectTileAction`

---

### Use **Services** when:
- Coordinating high-level game flows (turn progression, phase transitions)
- Providing backward-compatible interfaces
- Complex multi-feature coordination that doesn't fit in Actions

**Example**: `GameService.SkipPlayerTurn()`, `PlayerService.OnTileSelected()`

**Note**: Prefer Actions for new code. Services are legacy compatibility.

---

### Use **Features** when:
- Implementing domain-specific operations
- Managing a specific game subsystem (parameters, tiles, resources)
- Need clean separation from other domains

**Example**:
- `parameters.RaiseTemperature()` - ONLY manages temperature
- `tiles.PlaceTile()` - ONLY manages board state
- `resources.CalculatePayment()` - ONLY calculates costs

**Key Rule**: Features NEVER know about other domains. Use events for cross-domain effects.

---

### Use **Events** when:
- Need to react to state changes in other domains
- Implementing passive card effects
- Awarding TR for global parameter changes
- Triggering multiple side effects from a single action

**Example**:
- Temperature increases → Award TR to player (event subscription)
- Tile placed → Trigger "ocean adjacency bonus" cards (event subscription)
- Card played → Subscribe passive effects (event-driven)

---

## Benefits of This Architecture

### ✅ **Clean Separation of Concerns**
- Each layer has a single, well-defined responsibility
- Features are isolated and don't know about other domains
- Easy to understand and reason about code flow

### ✅ **Event-Driven Decoupling**
- Features publish events instead of triggering side effects
- Event subscribers handle cross-domain reactions
- Easy to add new reactions without modifying existing code

### ✅ **Testable in Isolation**
- Can test parameters feature without needing player repositories
- Can test card effects without needing real game state
- Mock event bus for testing event publishing

### ✅ **Extensible for New Mechanics**
- Add new event subscribers without modifying features
- New card effects just subscribe to existing events
- New game mechanics follow established patterns

### ✅ **No Manual Effect Checking**
- Passive card effects trigger automatically via events
- No need to loop through all cards checking for triggers
- Event system handles fan-out efficiently

---

## Living Card Instance Pattern

### Overview

Players hold **Card instances** (not IDs) that maintain their current state with modifiers, simplifying cost calculations and enabling event-driven discount updates.

### Card Model

```go
type Card struct {
    ID        string         `json:"id"`
    BaseCost  int            `json:"baseCost"`      // Base cost
    Modifiers []CostModifier `json:"modifiers"`     // Applied via events
    Tags      []string       `json:"tags"`
    // ... other fields
}

type CostModifier struct {
    SourceCardID string  // Which card provides this modifier
    Amount       int     // -2, +1, etc.
    Tag          string  // "science", "building", etc.
}

func (c *Card) GetFinalCost() int {
    cost := c.BaseCost
    for _, mod := range c.Modifiers {
        cost += mod.Amount
    }
    return cost
}
```

### Player Model

```go
type Player struct {
    Cards       []Card  // Live instances with modifiers
    PlayedCards []Card  // Live instances
}
```

### Event-Driven Modifiers

**When discount card played:**
1. CardPlayedEvent published
2. CardEffectSubscriber processes card's Effects
3. For DiscountEffectByTag: applies modifiers to owner's matching cards
4. Updates via PlayerRepository.UpdateCards()

**When cards drawn:**
1. DrawService creates Card instances from registry
2. Applies current modifiers from owner's played cards
3. Cards enter hand with discounts already applied

### Generalized Effect System

```go
type Effect struct {
    Type   string  // "DiscountEffectByTag", etc.
    Tag    string  // "science", "building"
    Amount int     // -2, +1
    Scope  string  // "self" (owner only), "all_players"
}

type CardBehavior struct {
    Triggers []Trigger
    Outputs  []Output
    Effects  []Effect  // NEW: Passive effects
}

const EffectDiscountByTag = "DiscountEffectByTag"
```

**Card JSON:**
```json
{
    "id": "physics_complex",
    "baseCost": 12,
    "tags": ["science"],
    "behaviors": [{
        "triggers": [{"type": "auto"}],
        "outputs": [{"type": "cards", "amount": 2}],
        "effects": [{
            "type": "DiscountEffectByTag",
            "tag": "science",
            "amount": -2,
            "scope": "self"
        }]
    }]
}
```

**Generic Handler:**
```go
func applyDiscountByTag(gameID string, effect Effect, sourceCardID, ownerPlayerID string) {
    player, _ := playerRepo.GetByID(gameID, ownerPlayerID)  // Scope is "self"

    for i := range player.Cards {
        if hasTag(player.Cards[i].Tags, effect.Tag) {
            player.Cards[i].Modifiers = append(player.Cards[i].Modifiers, CostModifier{
                SourceCardID: sourceCardID,
                Amount:       effect.Amount,
                Tag:          effect.Tag,
            })
        }
    }

    playerRepo.UpdateCards(gameID, ownerPlayerID, player.Cards)
}
```

### Benefits

- **Simplified Actions**: `card.GetFinalCost()` instead of scanning played cards
- **Simplified Payment**: No service needed - actions calculate payment value directly using pre-stored modifiers
- **Better UX**: Selection phases show accurate discounted costs
- **Event-Driven**: Modifiers update automatically when discount cards played
- **Extensible**: Add new effect types (ProductionBonusByTag, ResourceOnTag)
- **Data-Driven**: Card behaviors defined in JSON

### Integration

**Actions orchestrate with Card objects:**
```go
func (a *PlayCardAction) Execute(ctx, gameID, playerID, cardID string) error {
    player, _ := a.playerRepo.GetByID(ctx, gameID, playerID)

    // Find card in hand ([]Card, not []string)
    var cardToPlay *Card
    for i := range player.Cards {
        if player.Cards[i].ID == cardID {
            cardToPlay = &player.Cards[i]
            break
        }
    }

    // Cost already calculated!
    finalCost := cardToPlay.GetFinalCost()

    // Validate and play...
}
```

**Card feature remains pure** - no Player/Game imports, only CardRepository and CardDeckRepository dependencies.

---

## Payment Validation with Discounts

### Overview

When a player plays a card, they must specify how to pay for it. The payment system integrates seamlessly with the Living Card Instance Pattern - since cards already have their modifiers applied, payment validation is straightforward.

**CRITICAL**: There is **NO validation service**. All validation happens in the **Actions layer** using:
- `card.GetFinalCost()` for pricing (cost modifiers already applied to card instance)
- `card.Requirements` compared against global parameters
- `card.RequirementModifiers` for requirement leniency
- Direct payment arithmetic in PlayCardAction

### Payment Flow Architecture

```
1. Player selects card from hand (card instance with modifiers already applied)
   ↓
2. Frontend displays final cost: card.GetFinalCost()
   ↓
3. Player specifies payment method (credits, steel, titanium, heat)
   ↓
4. Client sends play-card message with payment details
   ↓
5. PlayCardAction validates requirements:
   - Compare card.Requirements against session.Game.GetGlobalParameters()
   - Apply card.RequirementModifiers for leniency (e.g., -2 temp requirement)
   ↓
6. PlayCardAction validates payment:
   - Calculate payment value using player's resource values (steel/titanium worth)
   - Validate: paymentValue >= card.GetFinalCost()
   ↓
7. Resources deducted from player via ResourcesService
   ↓
8. Card moved from hand to played cards
   ↓
9. Card effects processed using session's feature services
   ↓
10. Card passive effects subscribed via events (including new discount effects)
```

### WebSocket Message with Payment

```json
{
  "type": "play-card",
  "cardId": "orbital_reflectors",
  "payment": {
    "credits": 14,
    "steel": 5,
    "titanium": 0,
    "heat": 0
  }
}
```

### Player Model with Resource Value Modifiers

**Key Principle**: Resource values (steel/titanium worth) are stored on the Player model and updated via events.

```go
// internal/player/model.go
type Player struct {
    ID              string
    Name            string
    Cards           []Card    // Card instances with cost modifiers
    PlayedCards     []Card

    // Resource values (updated via events)
    SteelValue      int       // Default: 2 MC, can be increased by cards
    TitaniumValue   int       // Default: 3 MC, can be increased by cards
    HeatValue       int       // Default: 1 MC, can be increased by specific cards

    // ... other fields
}
```

### Event-Driven Resource Value Updates

**When a card that increases resource values is played:**

```go
// Example: Player plays "Steelworks" card (steel worth 3 instead of 2)

// 1. CardEffectSubscriber processes the card's effects
for _, effect := range card.Effects {
    if effect.Type == "ResourceValueBonus" {
        // 2. Update player's resource value
        player.SteelValue += effect.Amount  // 2 → 3
        playerRepo.UpdateResourceValues(ctx, gameID, playerID, player)

        // 3. Repository publishes event
        events.Publish(eventBus, ResourceValueChangedEvent{
            GameID:       gameID,
            PlayerID:     playerID,
            ResourceType: "steel",
            OldValue:     2,
            NewValue:     3,
        })
    }
}
```

### Action Layer Validation (NO Service Layer)

**CRITICAL**: All validation happens directly in PlayCardAction - NO validation service exists.

```go
// internal/actions/play_card.go
func (a *PlayCardAction) Execute(ctx context.Context, gameID, playerID, cardID string, payment Payment) error {
    // 1. Get session (contains game, players, feature services)
    session, err := a.sessionRepo.GetByID(ctx, gameID)
    if err != nil {
        return err
    }

    // 2. Get player and card from session
    player, exists := session.Players[playerID]
    if !exists {
        return errors.New("player not in game")
    }

    var cardToPlay *Card
    for i := range player.Cards {
        if player.Cards[i].ID == cardID {
            cardToPlay = &player.Cards[i]
            break
        }
    }
    if cardToPlay == nil {
        return errors.New("card not found in hand")
    }

    // 3. Validate REQUIREMENTS against global parameters (with modifier leniency)
    globalParams := session.Game.GetGlobalParameters()

    if cardToPlay.Requirements.MinTemperature != nil {
        requiredTemp := *cardToPlay.Requirements.MinTemperature
        // Apply requirement modifiers (e.g., -2 reduces temperature requirement)
        for _, mod := range cardToPlay.RequirementModifiers {
            if mod.Type == "temperature" {
                requiredTemp += mod.Adjustment
            }
        }
        if globalParams.Temperature < requiredTemp {
            return fmt.Errorf("temperature requirement not met: need %d, current %d",
                requiredTemp, globalParams.Temperature)
        }
    }

    // Similar checks for oxygen, oceans, etc.
    if cardToPlay.Requirements.MinOxygen != nil {
        requiredOxygen := *cardToPlay.Requirements.MinOxygen
        for _, mod := range cardToPlay.RequirementModifiers {
            if mod.Type == "oxygen" {
                requiredOxygen += mod.Adjustment
            }
        }
        if globalParams.Oxygen < requiredOxygen {
            return errors.New("oxygen requirement not met")
        }
    }

    // 4. Validate PAYMENT - calculate final cost (modifiers already applied!)
    finalCost := cardToPlay.GetFinalCost()
    // Example: baseCost=26, modifiers=[{amount:-2,tag:"science"}] → finalCost=24

    paymentValue := payment.Credits

    // Steel applies only to cards with "building" tag
    if hasTag(cardToPlay.Tags, "building") {
        paymentValue += payment.Steel * player.SteelValue  // Already set via events
    }

    // Titanium applies only to cards with "space" tag
    if hasTag(cardToPlay.Tags, "space") {
        paymentValue += payment.Titanium * player.TitaniumValue  // Already set via events
    }

    // Heat can be used for certain cards
    paymentValue += payment.Heat * player.HeatValue

    if paymentValue < finalCost {
        return fmt.Errorf("insufficient payment: need %d, provided %d", finalCost, paymentValue)
    }

    // 5. Deduct resources via player's resource service
    resources, _ := player.GetResources()
    resources.Credits -= payment.Credits
    resources.Steel -= payment.Steel
    resources.Titanium -= payment.Titanium
    resources.Heat -= payment.Heat
    player.ResourcesService.Update(ctx, resources)

    // 6. Move card from hand to played cards
    a.playerRepo.RemoveCardFromHand(ctx, gameID, playerID, cardID)
    a.playerRepo.AddPlayedCard(ctx, gameID, playerID, cardToPlay)

    // 7. Process immediate effects using session's feature services
    a.processCardEffects(ctx, session, playerID, cardToPlay)

    // 8. Subscribe passive effects (including any NEW discounts this card provides)
    a.effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, cardID, cardToPlay)

    // 9. Broadcast updated state
    a.sessionManager.Broadcast(gameID)

    return nil
}
```

### Payment Type Definition

```go
// internal/features/resources/payment.go
package resources

type Payment struct {
    Credits  int `json:"credits"`
    Steel    int `json:"steel"`
    Titanium int `json:"titanium"`
    Heat     int `json:"heat"`
}
```

**Note**: Resource values (steel worth, titanium worth) are stored directly on the Player model and updated via events, not passed as separate parameters.

### Integration with Card Cost Modifiers

**Two types of modifiers work together (both event-driven):**

1. **Card Cost Modifiers** (stored on card instances)
   - Reduce the card's base cost: `baseCost + modifiers = finalCost`
   - Example: Physics Complex reduces science cards by 2 MC
   - Applied when card is drawn or when discount card is played
   - Stored in `Card.Modifiers[]`
   - Updated via CardEffectSubscriber when discount cards are played

2. **Resource Value Modifiers** (stored on Player model)
   - Increase the value of steel/titanium when paying
   - Example: Steelworks makes steel worth 3 MC instead of 2 MC
   - Stored as `Player.SteelValue`, `Player.TitaniumValue`, etc.
   - Updated via events when resource value bonus cards are played

**Example with both modifier types:**

```go
// Player has:
// - Physics Complex played (science cards -2 MC) → Card modifiers updated
// - Steelworks played (steel worth 3 MC) → Player.SteelValue updated

// Player model after events:
player := Player{
    SteelValue:    3,  // Updated from 2 by Steelworks card effect
    TitaniumValue: 3,  // Default value
    Cards: []Card{
        {
            ID: "orbital_reflectors",
            BaseCost: 26,
            Tags: ["science", "building"],
            Modifiers: [
                {SourceCardID: "physics_complex", Amount: -2, Tag: "science"}
            ]
        }
    }
}

// Payment validation in action:
finalCost := card.GetFinalCost()  // 26 + (-2) = 24 MC

payment := Payment{Credits: 12, Steel: 4}

// Calculate payment value (all values pre-stored!)
paymentValue := payment.Credits                          // 12
if hasTag(card.Tags, "building") {
    paymentValue += payment.Steel * player.SteelValue    // 12 + (4 × 3) = 24
}

canAfford := paymentValue >= finalCost  // 24 >= 24 ✓
```

### Benefits of This Pattern

**✅ Fully Event-Driven**
- Card cost modifiers updated via CardEffectSubscriber events
- Card requirement modifiers updated via CardEffectSubscriber events
- Resource values updated via resource value change events
- No manual scanning or calculation needed
- All state pre-calculated and ready to use

**✅ NO Validation Service Layer**
- All validation done directly in PlayCardAction
- Requirements: Compare `card.Requirements` against `globalParams` with modifier leniency
- Payment: Calculate using `card.GetFinalCost()` and `player.SteelValue/TitaniumValue`
- Simple, direct logic - no intermediate service calls
- Fewer layers = easier to understand and maintain

**✅ Clear Separation of Concerns**
- **Card cost modifiers**: Managed by CardEffectSubscriber, stored on card instances
- **Card requirement modifiers**: Managed by CardEffectSubscriber, stored on card instances
- **Resource value modifiers**: Managed by CardEffectSubscriber, stored on Player
- **Validation logic**: Lives in Actions layer (orchestration, not a separate service)
- **Tag rules**: Simple checks in Actions (building → steel, space → titanium)

**✅ Session-Based Access**
- Session contains Game + Players + feature services
- Single fetch gets all needed state
- Actions access via `session.Players[playerID]` and `session.Game`
- No need to fetch from multiple repositories

**✅ Frontend Simplification**
- Frontend receives cards with cost/requirement modifiers already applied
- Frontend receives player with resource values already set
- Can display final cost: `card.GetFinalCost()`
- Can show requirement status by comparing against global params
- Can calculate payment value using same logic as backend

### Common Patterns

**✅ CORRECT - Event-driven payment validation:**
```go
// Get player (has cards with modifiers AND resource values already set via events)
player, _ := playerRepo.GetByID(ctx, gameID, playerID)

// Get card from hand (has cost modifiers already applied)
cardToPlay := findCardInHand(player.Cards, cardID)

// Calculate final cost (modifiers already summed)
finalCost := cardToPlay.GetFinalCost()

// Calculate payment value (resource values already set)
paymentValue := payment.Credits
if hasTag(cardToPlay.Tags, "building") {
    paymentValue += payment.Steel * player.SteelValue  // Pre-set via events
}
if hasTag(cardToPlay.Tags, "space") {
    paymentValue += payment.Titanium * player.TitaniumValue  // Pre-set via events
}

// Simple validation
canAfford := paymentValue >= finalCost
```

**❌ WRONG - Manually scanning for discounts:**
```go
// Don't do this - modifiers are already applied!
finalCost := card.BaseCost
for _, playedCard := range player.PlayedCards {
    if playedCard.HasDiscountFor(card.Tags) {
        finalCost -= playedCard.DiscountAmount  // ❌ Redundant! Use card.GetFinalCost()
    }
}
```

**❌ WRONG - Manually scanning for resource value bonuses:**
```go
// Don't do this - resource values are already on Player!
steelValue := 2  // Default
for _, playedCard := range player.PlayedCards {
    if playedCard.ID == "steelworks" {
        steelValue = 3  // ❌ Redundant! Use player.SteelValue
    }
}
paymentValue += payment.Steel * steelValue
```

**✅ CORRECT - Update resource values via events:**
```go
// When Steelworks card is played, CardEffectSubscriber processes effect
for _, effect := range card.Effects {
    if effect.Type == "ResourceValueBonus" && effect.ResourceType == "steel" {
        player.SteelValue += effect.Amount
        playerRepo.UpdateResourceValues(ctx, gameID, playerID, player)

        // Repository publishes ResourceValueChangedEvent
        events.Publish(eventBus, ResourceValueChangedEvent{...})
    }
}
```

---

## Summary

**The Flow**:
1. **Action** orchestrates the operation
2. **Feature** performs domain-specific work
3. **Repository** updates state and publishes events
4. **Event subscribers** react to events (TR increases, passive effects)
5. **SessionManager** broadcasts complete state to clients

**Key Principles**:
- Actions orchestrate, don't implement
- Features operate in isolation, publish events
- Repositories update state, publish events
- Event subscriptions handle cross-domain effects
- SessionManager broadcasts complete state

**Remember**: If a feature needs to know about another domain, you're doing it wrong. Use events instead!
