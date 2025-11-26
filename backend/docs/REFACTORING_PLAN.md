# Backend Session Architecture Refactoring Plan

## Goals
- **Encapsulated domain objects** with private fields and public methods
- **Delegation pattern** - Domain objects delegate to focused subfolder types
- **Navigation path**: `SessionRepository → Session → Game → Player → delegated components`
- **Separation of concerns** - Each component manages its own state (Hand, Resources, Production, etc.)
- **Domain types co-located** with their domain
- **EventBus integration** for event publishing where needed
- **No repository confusion** - Components have clear responsibilities

---

## Core Architecture Pattern

### Encapsulation Principle

**"Domain objects hold their own state and mutate through public functions with private properties and public getters"**

This applies to all domain entities: Player, GameDeck, and Session.

### Hierarchy
```
SessionRepository
  └── manages map[gameID]*Session
        └── Session (private game field)
              └── Game (owns EventBus)
                    ├── GlobalParameters (direct updates)
                    ├── Board (direct updates)
                    ├── Turn (direct updates)
                    └── map[playerID]*Player
                          └── Player (encapsulated domain object)
                                ├── Private fields (resources, production, etc.)
                                ├── Public getters (Resources(), Production(), etc.)
                                └── Public methods (SetResources(), PlayCard(), etc.)
```

### Usage in Actions
```go
func (a *UpdateResourcesAction) Execute(ctx context.Context, playerID string, resources types.Resources) error {
    // Get game from session
    game := a.session.Game()

    // Get player from game
    player, err := game.GetPlayer(playerID)
    if err != nil {
        return err
    }

    // Call domain method (publishes events internally)
    return player.SetResources(ctx, resources)
}
```

---

## Phase 1: Player Encapsulation

### 1.1 Current State Analysis

**Current Player Struct** (24 public fields):
- `ID`, `Name`, `GameID` - Identity fields
- `Corporation`, `CorporationID` - Corporation selection
- `Cards`, `PlayedCards` - Card management
- `Resources`, `Production` - Resource state
- `TerraformRating`, `VictoryPoints` - Scoring
- `Passed`, `AvailableActions` - Turn state
- `IsConnected` - Connection state
- `Effects`, `Actions` - Card effects
- `ProductionPhase`, `SelectStartingCardsPhase` - Phase states
- `PendingTileSelection`, `PendingTileSelectionQueue` - Tile placement
- `PendingCardSelection`, `PendingCardDrawSelection` - Card selection
- `ForcedFirstAction` - Corporation requirements
- `ResourceStorage` - Card-stored resources
- `PaymentSubstitutes`, `RequirementModifiers` - Card effects

**Current Repositories** (8 files to merge):
1. `resource_repository.go` - Resources, production, TR, VP, storage, payment substitutes
2. `action_repository.go` - Passed, available actions, actions, forced first action
3. `hand_repository.go` - Add/remove cards
4. `corporation_repository.go` - Set corporation
5. `effect_repository.go` - Effects, requirement modifiers
6. `selection_repository.go` - Phase states, pending selections
7. `tile_queue_repository.go` - Tile placement queue
8. `connection_repository.go` - Connection status

### 1.2 Target Player Structure (Delegation Pattern)

**Core Principle**: Player delegates to focused component types that don't know about Player.

```go
package player

// Player delegates to focused components in subfolders
type Player struct {
    // Infrastructure (private)
    mu       sync.RWMutex
    eventBus *events.EventBusImpl

    // Identity (private)
    id     string
    name   string
    gameID string

    // Delegated Components (private)
    corporation   *Corporation      // player/corporation package
    hand          *Hand             // player/hand package
    resources     *ResourceManager  // player/resources package
    turnState     *TurnState        // player/turn package
    effects       *Effects          // player/effects package
    phases        *Phases           // player/phases package
    selections    *Selections       // player/selections package
}

// Public interface - Player delegates to components
func (p *Player) ID() string
func (p *Player) Name() string
func (p *Player) GameID() string

// Card operations - delegates to hand component
func (p *Player) Cards() []string
func (p *Player) AddCardToHand(ctx context.Context, cardID string) error
func (p *Player) RemoveCardFromHand(ctx context.Context, cardID string) (bool, error)

// Resource operations - delegates to resources component
func (p *Player) Resources() types.Resources
func (p *Player) SetResources(ctx context.Context, resources types.Resources) error
func (p *Player) AddResources(changes map[types.ResourceType]int)

// Corporation operations - delegates to corporation component
func (p *Player) Corporation() *card.Card
func (p *Player) SetCorporation(ctx context.Context, corporation card.Card) error

// Turn operations - delegates to turn state component
func (p *Player) Passed() bool
func (p *Player) SetPassed(ctx context.Context, passed bool) error
func (p *Player) ConsumeAction() bool

// Component Packages Structure
```

**Delegation Components** (in player/ subfolders):

```
player/
├── player.go              # Main Player with delegation
├── corporation/
│   └── corporation.go     # Corporation management (doesn't know about Player)
├── hand/
│   └── hand.go           # Card hand management (doesn't know about Player)
├── resources/
│   └── resources.go      # Resource & production management
├── turn/
│   └── turn.go           # Turn state (passed, actions, connected)
├── effects/
│   └── effects.go        # Effects, actions, modifiers
├── phases/
│   └── phases.go         # Production/selection phases
└── selections/
    └── selections.go     # Pending selections (cards, tiles)
```

**Example Component (Hand)**:

```go
// player/hand/hand.go
package hand

type Hand struct {
    cards       []string
    playedCards []string
}

func NewHand() *Hand {
    return &Hand{
        cards:       []string{},
        playedCards: []string{},
    }
}

// Hand doesn't know about Player - it's self-contained
func (h *Hand) Cards() []string {
    cardsCopy := make([]string, len(h.cards))
    copy(cardsCopy, h.cards)
    return cardsCopy
}

func (h *Hand) AddCard(cardID string) {
    h.cards = append(h.cards, cardID)
}

func (h *Hand) RemoveCard(cardID string) bool {
    for i, id := range h.cards {
        if id == cardID {
            h.cards = append(h.cards[:i], h.cards[i+1:]...)
            return true
        }
    }
    return false
}

func (h *Hand) PlayCard(cardID string) {
    h.playedCards = append(h.playedCards, cardID)
    h.RemoveCard(cardID)
}

// === Identity Getters (read-only) ===
func (p *Player) ID() string { return p.id }
func (p *Player) Name() string { return p.name }
func (p *Player) GameID() string { return p.gameID }

// === Corporation ===
func (p *Player) Corporation() *card.Card { return p.corporation }
func (p *Player) CorporationID() string { return p.corporationID }
func (p *Player) SetCorporationID(ctx context.Context, corporationID string) error
func (p *Player) SetCorporation(ctx context.Context, corporation card.Card) error

// === Cards ===
func (p *Player) Cards() []string { /* defensive copy */ }
func (p *Player) PlayedCards() []string { /* defensive copy */ }
func (p *Player) AddCardToHand(ctx context.Context, cardID string) error
func (p *Player) RemoveCardFromHand(ctx context.Context, cardID string) (bool, error)
func (p *Player) PlayCard(ctx context.Context, cardID string) error

// === Resources ===
func (p *Player) Resources() types.Resources { return p.resources }
func (p *Player) Production() types.Production { return p.production }
func (p *Player) TerraformRating() int { return p.terraformRating }
func (p *Player) VictoryPoints() int { return p.victoryPoints }
func (p *Player) ResourceStorage() map[string]int { /* defensive copy */ }
func (p *Player) PaymentSubstitutes() []card.PaymentSubstitute { /* defensive copy */ }

func (p *Player) SetResources(ctx context.Context, resources types.Resources) error
func (p *Player) SetProduction(ctx context.Context, production types.Production) error
func (p *Player) SetTerraformRating(ctx context.Context, rating int) error
func (p *Player) SetVictoryPoints(ctx context.Context, victoryPoints int) error
func (p *Player) SetResourceStorage(ctx context.Context, storage map[string]int) error
func (p *Player) SetPaymentSubstitutes(ctx context.Context, substitutes []card.PaymentSubstitute) error

// Existing helper methods (update to use private fields)
func (p *Player) AddResources(changes map[types.ResourceType]int)
func (p *Player) AddProduction(changes map[types.ResourceType]int)
func (p *Player) UpdateTerraformRating(delta int)

// === Turn State ===
func (p *Player) Passed() bool { return p.passed }
func (p *Player) AvailableActions() int { return p.availableActions }
func (p *Player) IsConnected() bool { return p.isConnected }

func (p *Player) SetPassed(ctx context.Context, passed bool) error
func (p *Player) SetAvailableActions(ctx context.Context, actions int) error
func (p *Player) SetConnectionStatus(ctx context.Context, isConnected bool) error
func (p *Player) ConsumeAction() bool

// === Effects ===
func (p *Player) Effects() []card.PlayerEffect { /* defensive copy */ }
func (p *Player) Actions() []PlayerAction { /* defensive copy */ }
func (p *Player) RequirementModifiers() []RequirementModifier { /* defensive copy */ }
func (p *Player) ForcedFirstAction() *ForcedFirstAction { /* defensive copy */ }

func (p *Player) SetEffects(ctx context.Context, effects []card.PlayerEffect) error
func (p *Player) SetActions(ctx context.Context, actions []PlayerAction) error
func (p *Player) SetRequirementModifiers(ctx context.Context, modifiers []RequirementModifier) error
func (p *Player) SetForcedFirstAction(ctx context.Context, action *ForcedFirstAction) error

// === Phase States ===
func (p *Player) SelectStartingCardsPhase() *SelectStartingCardsPhase { /* defensive copy */ }
func (p *Player) ProductionPhase() *ProductionPhase { /* defensive copy */ }
func (p *Player) PendingCardSelection() *PendingCardSelection { /* defensive copy */ }
func (p *Player) PendingCardDrawSelection() *PendingCardDrawSelection { /* defensive copy */ }

func (p *Player) SetStartingCardsSelection(ctx context.Context, cardIDs, corpIDs []string) error
func (p *Player) CompleteStartingSelection(ctx context.Context) error
func (p *Player) CompleteProductionSelection(ctx context.Context) error
func (p *Player) SetStartingCardsPhase(ctx context.Context, phase *SelectStartingCardsPhase) error
func (p *Player) SetProductionPhase(ctx context.Context, phase *ProductionPhase) error
func (p *Player) SetPendingCardSelection(ctx context.Context, selection *PendingCardSelection) error
func (p *Player) ClearPendingCardSelection(ctx context.Context) error
func (p *Player) SetPendingCardDrawSelection(ctx context.Context, selection *PendingCardDrawSelection) error

// === Tile Queue ===
func (p *Player) PendingTileSelection() *PendingTileSelection { /* defensive copy */ }
func (p *Player) PendingTileSelectionQueue() *PendingTileSelectionQueue { /* defensive copy */ }

func (p *Player) CreateTileQueue(ctx context.Context, cardID string, tileTypes []string) error
func (p *Player) GetTileQueue(ctx context.Context) (*PendingTileSelectionQueue, error)
func (p *Player) ProcessNextTile(ctx context.Context) (string, error)
func (p *Player) SetPendingTileSelection(ctx context.Context, selection *PendingTileSelection) error
func (p *Player) ClearPendingTileSelection(ctx context.Context) error
func (p *Player) SetTileQueue(ctx context.Context, queue *PendingTileSelectionQueue) error
func (p *Player) ClearTileQueue(ctx context.Context) error
func (p *Player) QueueTilePlacement(source string, tileTypes []string)

// === Utility ===
func (p *Player) DeepCopy() *Player
func (p *Player) GetStartingSelectionCards() []string
func (p *Player) GetProductionPhaseCards() []string
```

### 1.3 Event Publishing

**Methods That Publish Events:**
1. `SetResources()` → `ResourcesChangedEvent` (batched deltas)
2. `SetTerraformRating()` → `TerraformRatingChangedEvent`
3. `CreateTileQueue()` → `TileQueueCreatedEvent`
4. `ProcessNextTile()` → `TileQueueCreatedEvent` (if more remain)

**Implementation Pattern:**
```go
func (p *Player) SetResources(ctx context.Context, resources types.Resources) error {
    // Calculate delta for event
    delta := calculateDelta(p.resources, resources)

    // Update state
    p.resources = resources

    // Publish event
    if p.eventBus != nil && !delta.IsZero() {
        events.Publish(p.eventBus, events.ResourcesChangedEvent{
            GameID:   p.gameID,
            PlayerID: p.id,
            Delta:    delta,
        })
    }

    return nil
}
```

### 1.4 Migration Steps

**Step 1**: Add eventBus field to Player
**Step 2**: Make all fields private (lowercase)
**Step 3**: Add getter methods (~24 methods)
**Step 4**: Add setter methods (~30 methods)
**Step 5**: Update existing methods for private fields
**Step 6**: Delete 8 repository files
**Step 7**: Remove or simplify player.Repository interface
**Step 8**: Update all actions (~22 files)
**Step 9**: Update DTO mapper
**Step 10**: Update tests

---

## Phase 2: GameDeck Encapsulation

### 2.1 Current GameDeck Struct

**Public Fields:**
- `GameID`, `ProjectCards`, `Corporations`, `DiscardPile`, `RemovedCards`, `PreludeCards`
- `DrawnCardCount`, `ShuffleCount`

**Repository Methods** (in `deck_repository.go`):
- `CreateDeck()`, `DrawProjectCards()`, `DrawCorporations()`, `DiscardCards()`
- `GetAvailableCardCount()`, `GetCardByID()`, `GetAllCards()`, etc.

### 2.2 Target GameDeck Structure

```go
type GameDeck struct {
    // Private fields
    gameID         string
    projectCards   []string
    corporations   []string
    discardPile    []string
    removedCards   []string
    preludeCards   []string
    drawnCardCount int
    shuffleCount   int
}

// Getters
func (d *GameDeck) GameID() string
func (d *GameDeck) ProjectCards() []string  // Defensive copy
func (d *GameDeck) Corporations() []string  // Defensive copy
func (d *GameDeck) DiscardPile() []string   // Defensive copy

// Operations
func (d *GameDeck) Draw(ctx context.Context, count int) ([]string, error)
func (d *GameDeck) DrawCorporations(ctx context.Context, count int) ([]string, error)
func (d *GameDeck) Discard(ctx context.Context, cardIDs []string) error
func (d *GameDeck) Shuffle(ctx context.Context) error
func (d *GameDeck) GetAvailableCardCount() int
```

### 2.3 Migration Steps

**Step 1**: Make fields private
**Step 2**: Add getter methods
**Step 3**: Merge repository methods into GameDeck
**Step 4**: Update deck loader
**Step 5**: Update card repository wrapper
**Step 6**: Delete deck_repository.go (or keep as factory)

---

## Phase 3: Session Encapsulation

### 3.1 Current Session Struct

**Problem**: References old `*types.Game` (import cycle issue)

```go
type Session struct {
    Game     *types.Game  // WRONG - should be *game.Game
    eventBus *events.EventBusImpl
    mu       sync.RWMutex
}
```

### 3.2 Target Session Structure

```go
type Session struct {
    game     *game.Game  // Private, correct type
    eventBus *events.EventBusImpl
    mu       sync.RWMutex
}

// Getter
func (s *Session) Game() *game.Game {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.game
}

// Setter (internal use)
func (s *Session) setGame(g *game.Game) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.game = g
}
```

### 3.3 Migration Steps

**Step 1**: Change field type and make private
**Step 2**: Add Game() getter
**Step 3**: Update all session.Game → session.Game()
**Step 4**: Fix any remaining import issues

---

## Phase 4: Update Action Layer

### 4.1 Current Pattern (Repository-based)

```go
type ConvertHeatToTemperatureAction struct {
    BaseAction
    gameRepo game.Repository
}

func (a *ConvertHeatToTemperatureAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
    player, _ := sess.GetPlayer(playerID)
    player.Resources.Heat -= requiredHeat
    player.TerraformRating++

    err = a.gameRepo.UpdateTemperature(ctx, gameID, newTemperature)
    a.BroadcastGameState(gameID, log)
}
```

### 4.2 Target Pattern (Domain Methods)

```go
type ConvertHeatToTemperatureAction struct {
    BaseAction
}

func (a *ConvertHeatToTemperatureAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
    game := sess.Game()
    player, _ := game.GetPlayer(playerID)

    // Use domain methods (events published internally)
    player.AddResources(map[types.ResourceType]int{
        types.ResourceHeat: -requiredHeat,
    })
    player.UpdateTerraformRating(1)
    game.UpdateTemperature(ctx, 2)

    // Events trigger broadcasts automatically
}
```

### 4.3 Migration Pattern

**For each action:**
1. Remove repository dependencies from constructor
2. Change field access to getter calls: `player.Field` → `player.Field()`
3. Change mutations to method calls: `player.Field = value` → `player.SetField(ctx, value)`
4. Remove manual event publishing (done by domain methods)
5. Remove manual broadcasting (done by event system)

---

## Phase 5: Remove Repository Files

### 5.1 Files to Delete

**Player Repositories** (~890 lines):
- `internal/session/game/player/resource_repository.go`
- `internal/session/game/player/action_repository.go`
- `internal/session/game/player/hand_repository.go`
- `internal/session/game/player/corporation_repository.go`
- `internal/session/game/player/effect_repository.go`
- `internal/session/game/player/connection_repository.go`
- `internal/session/game/player/selection_repository.go`
- `internal/session/game/player/tile_queue_repository.go`
- `internal/session/game/player/repository.go` (interface)

**Deck Repository** (optional):
- `internal/session/game/deck/deck_repository.go` (if fully merged)

### 5.2 Files to Keep

**Game** - Already properly encapsulated
**Board** - Value object with repository pattern works well
**Card** - Immutable value object
**Types** - Value objects and DTOs

---

## Phase 6: Update Documentation

### 6.1 Update CLAUDE.md Files

**Backend CLAUDE.md**:
- Document new domain method patterns
- Remove repository pattern references
- Add examples of encapsulated access

**Frontend CLAUDE.md**:
- No changes needed (TypeScript types auto-generated)

### 6.2 Update Event System Docs

**EVENT_SYSTEM.md** (if exists):
- Document which domain methods publish events
- Update event flow diagrams

---

## Expected Final Structure

```
backend/internal/
├── session/
│   ├── session.go              # Session with private game field
│   ├── session_factory.go      # Creates sessions
│   ├── session_manager.go      # Broadcasting interface
│   ├── types/                  # Shared value objects only
│   │   ├── resources.go        # Resources, Production (value objects)
│   │   ├── parameters.go       # GlobalParameters, GameSettings
│   │   ├── resource_type.go    # ResourceType enum
│   │   ├── card_types.go       # CardTag, CardRequirements
│   │   ├── other_player.go     # Public player view
│   │   └── errors.go           # Domain errors
│   └── game/                   # Domain root
│       ├── game.go             # Game domain (already encapsulated)
│       ├── game_types.go       # GameStatus, GamePhase
│       ├── player/
│       │   ├── player.go       # Player with private fields + public methods
│       │   ├── player_types.go # ProductionPhase, SelectStartingCardsPhase, etc.
│       │   ├── player_action.go # PlayerAction type
│       │   └── player_factory.go
│       ├── card/
│       │   ├── card.go         # Card domain object (immutable)
│       │   ├── card_effects.go
│       │   └── [other card files]
│       ├── board/
│       │   ├── board_repository.go  # Keep for now
│       │   ├── board_types.go       # Board, Tile, HexPosition
│       │   └── [other board files]
│       ├── deck/
│       │   ├── deck_models.go       # GameDeck with private fields
│       │   └── [other deck files]
│       └── [card_manager, card_processor, etc.]
├── action/
│   └── [all actions updated for domain methods]
└── [delivery, events, logger, etc.]
```

---

## Key Properties

✅ **Encapsulated domains**: Private fields, public getters/setters
✅ **No repository confusion**: Player has methods, not PlayerRepo
✅ **Event publishing**: Domain methods publish events internally
✅ **Type safety**: Compiler enforces getter/setter usage
✅ **Clean navigation**: `session.Game().GetPlayer().SetResources()`
✅ **Reduced complexity**: ~890 lines of repository wrapper code removed
✅ **Defensive copying**: Getters return copies to prevent external mutation

---

## Migration Timeline

### Phase 1: Player Encapsulation (2 days)
- Add eventBus field
- Make fields private
- Add ~50 getter/setter methods
- Update existing methods
- Delete repositories

### Phase 2: Update Player Usage (2 days)
- Update ~22 action files
- Update DTO mapper
- Fix tests

### Phase 3: GameDeck Encapsulation (0.5 days)
- Make fields private
- Add methods
- Update usage

### Phase 4: Session Fix (0.5 days)
- Fix Game field type
- Make private with getter
- Update all usage

### Phase 5: Documentation (0.5 days)
- Update CLAUDE.md files
- Update progress docs
- Code quality checks

**Total Estimate: 5-6 days**
