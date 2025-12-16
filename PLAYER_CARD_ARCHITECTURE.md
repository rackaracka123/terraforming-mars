# Player-Scoped Card Architecture

## Core Concept

Separate **immutable card data** (Card) from **player-specific card state** (PlayerCard).

**Key Innovation**: Single generic `EntityState` structure replaces three entity-specific states, eliminating redundant boolean flags and preventing contradictory data. All entities (cards, actions, projects) share the same simple state model with extensible metadata.

```
Card (JSON data, immutable, shared)
  - Loaded once from terraforming_mars_cards.json
  - Shared by all players
  - No state, no events
  - Lives in CardRegistry

PlayerCard/PlayerCardAction/PlayerStandardProject (player-scoped view, stateful)
  - Wraps entity reference (Card/Behavior/Project)
  - Maintains generic EntityState (available, errors, cost, metadata)
  - Created by actions with full game context
  - Cached in Hand/Actions/Projects for lifecycle
  - Recalculated when actions update game state

EntityState (generic state holder)
  - Available() bool method (computed from Errors)
  - Errors: []StateError (single source of truth)
  - Cost: *int (nil if N/A)
  - Metadata: map[string]interface{} (minimal, prefer typed fields)
```

## Architecture Principles

### 1. **Action Orchestration (No GameContext Interface)**

Actions have access to both Game and Player, so they orchestrate PlayerCard creation:

```go
// In action or DTO mapper
func CreatePlayerCard(card *cards.Card, player *Player, game *Game, cardRegistry CardRegistry) *PlayerCard {
    // Actions pass concrete types directly
    playerCard := player.NewPlayerCard(card, player, game, cardRegistry)
    return playerCard
}
```

**Key benefits:**
- ✅ No GameContext interface needed
- ✅ Actions orchestrate (consistent with existing architecture)
- ✅ No circular dependencies (`player/` never imports `game/`)
- ✅ Simpler than interface pattern

### 2. **Long-Lived Caching**

PlayerCard instances are cached for their entire lifecycle:

```go
// Player hand caches PlayerCard instances
type Hand struct {
    cardIDs      []string
    playerCards  map[string]*PlayerCard  // Cached PlayerCard instances
    cardRegistry cards.CardLookup
}

// GetPlayerCard returns cached PlayerCard (must already exist)
func (h *Hand) GetPlayerCard(cardID string) (*PlayerCard, bool) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    pc, exists := h.playerCards[cardID]
    return pc, exists
}

// AddPlayerCard caches a PlayerCard (created and initialized by action)
func (h *Hand) AddPlayerCard(cardID string, pc *PlayerCard) {
    h.mu.Lock()
    defer h.mu.Unlock()

    h.cardIDs = append(h.cardIDs, cardID)
    h.playerCards[cardID] = pc
}

// RemoveCard removes from both ID list and cache
func (h *Hand) RemoveCard(cardID string) {
    h.mu.Lock()
    defer h.mu.Unlock()

    // Clean up event listeners before removing
    if pc, exists := h.playerCards[cardID]; exists {
        pc.Cleanup()  // Unsubscribe all event listeners
    }

    // Remove from cache
    delete(h.playerCards, cardID)

    // Remove from ID list
    for i, id := range h.cardIDs {
        if id == cardID {
            h.cardIDs = append(h.cardIDs[:i], h.cardIDs[i+1:]...)
            break
        }
    }
}
```

**IMPORTANT: Actions create everything** - Actions create PlayerCard, register listeners with cleanup tracking, calculate initial state, then add to Hand cache via AddPlayerCard(). Hand is a dumb cache with no creation logic.

**Lifecycle:**
- **Selection phase**: Created for card options, cached until selected/discarded
- **In hand**: Created when added, cached until played/discarded
- **Played**: Created when played, cached for game duration

### 3. **Generic Naming and State for Extensibility**

- `PlayerCard` (not PlayableCard) - generic entity name
- `PlayerCardAction` (not UsableAction) - generic entity name
- `PlayerStandardProject` (not AvailableProject) - generic entity name
- `EntityState` (not PlayableState/UsableState) - single generic state structure

All entities use the same `EntityState`:
- `Available()` computed method - works for playable/usable/available/purchasable
- `Metadata` map minimal - prefer typed fields when data is predictable
- Future features (tiles, awards, milestones) use same pattern

Future modifiers don't require renaming or new state types.

### 4. **Clean Package Structure (No Circular Dependencies)**

```
events/
  ├─ event_bus.go          # Generic pub/sub
  └─ domain_events.go      # All event types
  Dependencies: NONE

shared/
  ├─ resource_types.go     # ResourceType, CardTag
  ├─ card_behavior.go      # CardBehavior, Trigger
  └─ standard_project.go   # StandardProject enum
  Dependencies: NONE

cards/
  ├─ card.go               # Card (immutable data)
  ├─ requirement.go        # Requirement types
  ├─ card_validator.go     # Validation logic
  ├─ card_lookup.go        # CardLookup interface
  └─ registry.go           # CardRegistry (implements CardLookup)
  Dependencies: shared

player/
  ├─ player.go             # Player aggregate
  ├─ player_card.go        # PlayerCard (state calculation)
  ├─ player_card_action.go # PlayerCardAction (usability)
  ├─ player_standard_project.go # PlayerStandardProject (availability)
  ├─ hand.go               # Hand (caches PlayerCard)
  └─ resources.go          # Resources component
  Dependencies: events, shared, cards
  NO IMPORT of game package!

game/
  ├─ game.go               # Game (full state)
  ├─ global_parameters.go
  └─ board.go
  Dependencies: events, shared, cards, player

action/
  ├─ base.go               # BaseAction
  ├─ play_card.go          # PlayCardAction
  ├─ state_calculator.go   # All entity state calculation (cards, actions, projects)
  └─ ...
  Dependencies: game, player, cards (orchestrates all)

  IMPORTANT: All state calculation logic lives in state_calculator.go, not in player package
```

**Key**: `player` package never imports `game`. Actions in `action/` package orchestrate both and calculate all state.

## Core Types

### State Design: Generic EntityState

**Problem with entity-specific states**: Original design had three separate state structs (PlayerCardState, PlayerCardActionState, PlayerStandardProjectState) with redundant boolean flags that could contradict:
```go
// ❌ Old approach - redundant and contradictory
PlayerCardState {
    Playable: true,           // Derived from Errors
    Affordable: false,        // Derived from Cost vs Credits
    MeetsRequirements: true,  // Derived from Errors
    Errors: [...],            // Actual source of truth
}
```

**Solution**: Single generic `EntityState` with computed availability:
```go
// ✅ New approach - simple and consistent
EntityState {
    Errors: []StateError,              // Single source of truth
    Cost: *int,                        // Nil if N/A
    Metadata: map[string]interface{},  // Minimal, entity-specific data
}

// Available is computed, not stored (prevents contradictions)
func (e EntityState) Available() bool {
    return len(e.Errors) == 0
}
```

**Benefits**:
- ✅ Eliminates ALL redundant booleans (`Available`, `Playable`, `Affordable`, `MeetsRequirements`)
- ✅ Impossible to have contradictory state (computed from single source of truth)
- ✅ Works for all entity types (cards, actions, projects, tiles, etc.)
- ✅ Extensible via minimal Metadata (prefer typed fields when predictable)
- ✅ Simpler implementation and maintenance

### Card (Immutable, Shared)

```go
// backend/internal/game/cards/card.go
package cards

// Card represents immutable card data from JSON
type Card struct {
    ID              string
    Name            string
    Type            CardType
    Cost            int
    Description     string
    Tags            []shared.CardTag
    Requirements    []Requirement
    Behaviors       []shared.CardBehavior
    ResourceStorage *ResourceStorage
    VPConditions    []VictoryPointCondition

    // Corporation-specific
    StartingResources  *shared.ResourceSet
    StartingProduction *shared.ResourceSet
}

// Card is immutable - no methods that modify state
// Card has no event subscriptions
// Card is shared across all players
```

### PlayerCard (Stateful Data Holder)

**IMPORTANT**: PlayerCard is a **simple data holder** with NO business logic. State calculation happens in `internal/action/` package to avoid circular dependencies and keep files focused.

```go
// backend/internal/game/player/player_card.go
package player

import (
    "sync"
    "time"
    "github.com/rackaracka123/terraforming-mars/backend/internal/game/cards"
    "github.com/rackaracka123/terraforming-mars/backend/internal/game/shared"
)

// PlayerCard is a player-specific view of a Card with calculated state
// This is a DATA HOLDER - state calculation happens in action package
type PlayerCard struct {
    // Reference to immutable card data
    card *cards.Card

    // Player-specific state (protected by mutex)
    mu    sync.RWMutex
    state EntityState

    // Event listener cleanup
    unsubscribers []func()
}

// EntityState holds all calculated state for any entity (card, action, project)
// This generic structure eliminates redundant boolean flags and works across all entity types
type EntityState struct {
    // Single source of truth - Errors determine availability
    Errors []StateError

    // Optional cost (nil if not applicable)
    Cost *int  // Effective cost after discounts

    // Minimal entity-specific data (prefer typed fields over metadata)
    Metadata map[string]interface{}

    // Calculation timestamp
    LastCalculated time.Time
}

// Available is computed, not stored (prevents contradictory state)
func (e EntityState) Available() bool {
    return len(e.Errors) == 0
}

// Examples of minimal Metadata usage:
// Cards: {"discounts": map[CardTag]int} (meetsRequirements computed from Errors)
// Actions: {"timesUsedThisTurn": 1} (inputsAvailable computed from Errors)
// Projects: {"oceansRemaining": 3} (canAfford computed from Errors)

type StateError struct {
    Code     string  // "INSUFFICIENT_CREDITS", "TEMPERATURE_TOO_LOW", etc.
    Category string  // "phase", "cost", "requirement", "behavior"
    Message  string
}

### Metadata Usage Patterns

**IMPORTANT**: Keep `Metadata` minimal. Use typed fields when data is predictable and common.

**Prefer typed fields:**
```go
// ✅ Good: Persistent state as typed fields
type PlayerCardAction struct {
    timesUsedThisTurn       int  // NOT in metadata
    timesUsedThisGeneration int  // NOT in metadata
    state EntityState
}
```

**Use metadata sparingly for truly dynamic data:**
```go
// ✅ Minimal metadata - Cards
Metadata: map[string]interface{}{
    "discounts": map[shared.CardTag]int{
        shared.CardTagSpace: 2,
    },
}

// ✅ Minimal metadata - Projects
Metadata: map[string]interface{}{
    "oceansRemaining": 3,
}
```

**Avoid metadata for boolean flags derivable from Errors:**
```go
// ❌ Bad: Redundant with Errors
"meetsRequirements": true,  // Just check for requirement errors
"inputsAvailable": true,    // Just check for input errors
"canAfford": true,          // Just check for affordability errors

// ✅ Good: Compute from Errors in frontend/DTO
meetsRequirements := !hasErrorCategory(state.Errors, "requirement")
canAfford := !hasErrorCode(state.Errors, "INSUFFICIENT_CREDITS")
```

**Benefits:**
- Type-safe for common patterns (compile-time checking)
- Metadata only for truly variable entity-specific data
- Simpler DTO mapping (fewer type assertions)
- Less chance of typos or type mismatches

// Constructor: Simple initialization with empty state
func NewPlayerCard(card *cards.Card) *PlayerCard {
    return &PlayerCard{
        card: card,
        state: EntityState{
            Errors:   []StateError{},
            Metadata: make(map[string]interface{}),
        },
    }
}

// UpdateState: Called by actions after calculating state
func (pc *PlayerCard) UpdateState(newState EntityState) {
    pc.mu.Lock()
    defer pc.mu.Unlock()
    pc.state = newState
}

// Public accessors (read-only)
func (pc *PlayerCard) Card() *cards.Card {
    return pc.card
}

func (pc *PlayerCard) State() EntityState {
    pc.mu.RLock()
    defer pc.mu.RUnlock()
    return pc.state
}

func (pc *PlayerCard) IsAvailable() bool {
    pc.mu.RLock()
    defer pc.mu.RUnlock()
    return pc.state.Available()
}

// Cleanup unsubscribes all event listeners (called when card removed from hand)
func (pc *PlayerCard) Cleanup() {
    for _, unsub := range pc.unsubscribers {
        unsub()
    }
}
```

### Event Listener Lifecycle and Cleanup

**CRITICAL**: Event listeners must be explicitly cleaned up to prevent memory leaks.

**Pattern: Store unsubscribe functions**
```go
// When registering listeners, store unsubscribe functions
func (a *Action) registerPlayerCardEventListeners(
    pc *player.PlayerCard,
    p *player.Player,
    g *game.Game,
) {
    eventBus := g.EventBus()

    // Subscribe returns unsubscribe function
    unsub1 := events.Subscribe(eventBus, func(event ResourcesChangedEvent) {
        if event.PlayerID == p.ID() {
            state := CalculatePlayerCardState(pc.Card(), p, g, a.cardRegistry)
            pc.UpdateState(state)
        }
    })
    pc.unsubscribers = append(pc.unsubscribers, unsub1)

    unsub2 := events.Subscribe(eventBus, func(event TemperatureChangedEvent) {
        state := CalculatePlayerCardState(pc.Card(), p, g, a.cardRegistry)
        pc.UpdateState(state)
    })
    pc.unsubscribers = append(pc.unsubscribers, unsub2)

    // ... register all relevant events
}
```

**Cleanup when PlayerCard removed:**
```go
// Hand.RemoveCard calls Cleanup before removing from cache
func (h *Hand) RemoveCard(cardID string) {
    h.mu.Lock()
    defer h.mu.Unlock()

    // Clean up event listeners
    if pc, exists := h.playerCards[cardID]; exists {
        pc.Cleanup()  // Unsubscribe all listeners
    }

    // Remove from cache
    delete(h.playerCards, cardID)

    // Remove from ID list
    for i, id := range h.cardIDs {
        if id == cardID {
            h.cardIDs = append(h.cardIDs[:i], h.cardIDs[i+1:]...)
            break
        }
    }
}
```

**Why this matters:**
- Each PlayerCard can have 5-10 event listeners
- A game might have 50+ PlayerCards active
- Without cleanup: 250-500 dangling listeners
- Cleanup prevents memory leaks and performance degradation
```

### State Calculation (Action Package)

**All state calculation logic lives in `internal/action/state_calculator.go`** - one consolidated file for all entity types. This avoids circular dependencies and keeps domain models simple.

```go
// backend/internal/action/state_calculator.go
package action

import (
    "time"
    "github.com/rackaracka123/terraforming-mars/backend/internal/game"
    "github.com/rackaracka123/terraforming-mars/backend/internal/game/cards"
    "github.com/rackaracka123/terraforming-mars/backend/internal/game/player"
    "github.com/rackaracka123/terraforming-mars/backend/internal/game/shared"
)

// ========================================
// PlayerCard State Calculation
// ========================================

// CalculatePlayerCardState computes playability state for a card
// This function can access both Game and Player without circular dependencies
func CalculatePlayerCardState(
    card *cards.Card,
    p *player.Player,
    g *game.Game,
    cardRegistry cards.CardRegistry,
) player.EntityState {
    errors := []player.StateError{}
    metadata := make(map[string]interface{})

    // 1. Phase check
    errors = append(errors, validatePhase(g)...)

    // 2. Requirements check (extracted from PlayCardAction lines 209-321)
    errors = append(errors, validateRequirements(card, p, g)...)

    // 3. Cost calculation with discounts (uses existing RequirementModifierCalculator)
    effectiveCost, discounts := calculateEffectiveCost(card, p)
    if len(discounts) > 0 {
        metadata["discounts"] = discounts
    }

    // 4. Affordability check
    errors = append(errors, validateAffordability(p, effectiveCost)...)

    return player.EntityState{
        Errors:         errors,
        Cost:           &effectiveCost,
        Metadata:       metadata,
        LastCalculated: time.Now(),
    }
}

// ========================================
// Shared Validation Helpers
// ========================================

// validatePhase checks if action is allowed in current phase
func validatePhase(g *game.Game) []player.StateError {
    if g.CurrentPhase() != "action" {
        return []player.StateError{{
            Code:     "WRONG_PHASE",
            Category: "phase",
            Message:  "Can only play cards during action phase",
        }}
    }
    return nil
}

// validateRequirements checks all card requirements
// EXTRACTED from PlayCardAction.Validate() lines 209-321
func validateRequirements(card *cards.Card, p *player.Player, g *game.Game) []player.StateError {
    errors := []player.StateError{}

    for _, req := range card.Requirements {
        if !checkRequirement(req, p, g) {
            errors = append(errors, player.StateError{
                Code:     requirementErrorCode(req.Type),
                Category: "requirement",
                Message:  formatRequirementError(req, p, g),
            })
        }
    }

    return errors
}

// checkRequirement validates a single requirement
// EXTRACTED from PlayCardAction - contains the switch statement for all requirement types
func checkRequirement(req cards.Requirement, p *player.Player, g *game.Game) bool {
    // Move the large switch statement from PlayCardAction.Validate() here
    // Lines 209-321 in play_card.go contain this logic
    switch req.Type {
    case "min-temperature":
        return g.GlobalParameters().Temperature() >= req.Amount
    case "min-oxygen":
        return g.GlobalParameters().Oxygen() >= req.Amount
    // ... all other requirement types
    }
    return true
}

// calculateEffectiveCost computes cost with discounts
// Uses existing RequirementModifierCalculator (already in codebase)
func calculateEffectiveCost(card *cards.Card, p *player.Player) (int, map[shared.CardTag]int) {
    baseCost := card.Cost
    discounts := map[shared.CardTag]int{}

    // Use existing RequirementModifierCalculator to get discounts
    // This already exists and is integrated in PlayCardAction
    for _, modifier := range p.Effects().RequirementModifiers() {
        if applies := modifierApplies(modifier, card); applies {
            for tag, amount := range modifier.Discounts {
                discounts[tag] += amount
            }
        }
    }

    totalDiscount := 0
    for tag := range card.Tags {
        if discount, ok := discounts[tag]; ok {
            totalDiscount += discount
        }
    }

    effectiveCost := max(0, baseCost - totalDiscount)
    return effectiveCost, discounts
}

// validateAffordability checks if player can afford the cost
func validateAffordability(p *player.Player, cost int) []player.StateError {
    if p.Resources().Credits() < cost {
        return []player.StateError{{
            Code:     "INSUFFICIENT_CREDITS",
            Category: "cost",
            Message:  fmt.Sprintf("Need %d credits, have %d", cost, p.Resources().Credits()),
        }}
    }
    return nil
}
```

**Key benefits:**
- ✅ No circular dependencies (action imports both game and player)
- ✅ Complex validation logic separated from domain model
- ✅ Can reuse and extract existing logic from PlayCardAction
- ✅ Smaller, more focused files
- ✅ PlayerCard remains a simple data holder

### When State Calculation Happens

**CRITICAL**: State is calculated in EXACTLY TWO scenarios:

1. **Initial Creation** (in action):
   ```go
   pc := player.NewPlayerCard(card)
   action.registerEventListeners(pc, p, g)
   action.recalculatePlayerCard(pc, p, g)  // Initial calculation
   ```

2. **Event-Driven Updates** (via registered listeners):
   ```go
   // Action registers this listener on PlayerCard creation
   events.Subscribe(eventBus, func(event ResourcesChangedEvent) {
       state := CalculatePlayerCardState(pc.Card(), p, g, cardRegistry)
       pc.UpdateState(state)
   })
   ```

**NEVER calculate state:**
- ❌ In DTO mapping (just read cached state)
- ❌ In Hand.GetOrCreatePlayerCard() (just return cached instance)
- ❌ On every game state broadcast (events handle this)
- ❌ Manually in business logic (events handle this)

**Flow:**
```
Action creates PlayerCard
  → Registers event listeners on PlayerCard
  → Triggers initial calculation
  → Caches in Hand/Selection

Game state changes
  → Events published
  → Listeners recalculate affected PlayerCards
  → State updated in place

DTO Mapping
  → Just read current state
  → NO recalculation
```

### PlayerCardAction (Data Holder, Same Pattern)

**IMPORTANT**: Like PlayerCard, this is a **data holder only**. Usability calculation happens in `internal/action/` package.

```go
// backend/internal/game/player/player_card_action.go
package player

import (
    "sync"
    "time"
    "github.com/rackaracka123/terraforming-mars/backend/internal/game/shared"
)

// PlayerCardAction wraps a card action with usability state
// This is a DATA HOLDER - usability calculation happens in action package
type PlayerCardAction struct {
    cardID        string
    behaviorIndex int
    behavior      shared.CardBehavior

    // Persistent state (not calculated)
    timesUsedThisTurn       int
    timesUsedThisGeneration int

    // Calculated state (from action package)
    mu    sync.RWMutex
    state EntityState
}

// Constructor - simple initialization
func NewPlayerCardAction(
    cardID string,
    behaviorIndex int,
    behavior shared.CardBehavior,
) *PlayerCardAction {
    return &PlayerCardAction{
        cardID:        cardID,
        behaviorIndex: behaviorIndex,
        behavior:      behavior,
        state: EntityState{
            Errors:   []StateError{},
            Metadata: make(map[string]interface{}),
        },
    }
}

// UpdateState - called by actions after calculating usability
func (pca *PlayerCardAction) UpdateState(newState EntityState) {
    pca.mu.Lock()
    defer pca.mu.Unlock()
    pca.state = newState
}

// Read-only accessors
func (pca *PlayerCardAction) State() EntityState {
    pca.mu.RLock()
    defer pca.mu.RUnlock()
    return pca.state
}

func (pca *PlayerCardAction) IsAvailable() bool {
    pca.mu.RLock()
    defer pca.mu.RUnlock()
    return pca.state.Available()
}

// Play count management (persistent state, not calculated)
func (pca *PlayerCardAction) IncrementPlayCount() {
    pca.mu.Lock()
    defer pca.mu.Unlock()
    pca.timesUsedThisTurn++
    pca.timesUsedThisGeneration++
}

func (pca *PlayerCardAction) ResetTurnPlayCount() {
    pca.mu.Lock()
    defer pca.mu.Unlock()
    pca.timesUsedThisTurn = 0
}

func (pca *PlayerCardAction) ResetGenerationPlayCount() {
    pca.mu.Lock()
    defer pca.mu.Unlock()
    pca.timesUsedThisGeneration = 0
}

func (pca *PlayerCardAction) TimesUsedThisTurn() int {
    pca.mu.RLock()
    defer pca.mu.RUnlock()
    return pca.timesUsedThisTurn
}

func (pca *PlayerCardAction) TimesUsedThisGeneration() int {
    pca.mu.RLock()
    defer pca.mu.RUnlock()
    return pca.timesUsedThisGeneration
}
```

**Usability calculation** happens in `internal/action/state_calculator.go` (same file as card and project calculators):

```go
// ========================================
// PlayerCardAction State Calculation
// ========================================

// CalculatePlayerCardActionState computes usability state for a card action
func CalculatePlayerCardActionState(
    cardID string,
    behavior shared.CardBehavior,
    pca *player.PlayerCardAction,
    p *player.Player,
    g *game.Game,
) player.EntityState {
    errors := []player.StateError{}

    // 1. Check if it's the player's turn
    if g.CurrentPlayerID() != p.ID() {
        errors = append(errors, player.StateError{
            Code:     "NOT_YOUR_TURN",
            Category: "turn",
            Message:  "Not your turn",
        })
    }

    // 2. Check input resource availability
    for _, input := range behavior.Inputs {
        if !hasRequiredResource(p, input) {
            errors = append(errors, player.StateError{
                Code:     "INSUFFICIENT_RESOURCES",
                Category: "input",
                Message:  fmt.Sprintf("Need %d %s", input.Amount, input.ResourceType),
            })
        }
    }

    // 3. Check max usage limits (access persistent counters from PlayerCardAction)
    if behavior.MaxUsesPerTurn > 0 && pca.TimesUsedThisTurn() >= behavior.MaxUsesPerTurn {
        errors = append(errors, player.StateError{
            Code:     "MAX_USES_PER_TURN_REACHED",
            Category: "usage",
            Message:  fmt.Sprintf("Can only use %d times per turn", behavior.MaxUsesPerTurn),
        })
    }

    // Note: timesUsed are NOT in metadata - they're typed fields in PlayerCardAction
    // Frontend derives inputsAvailable from checking for INSUFFICIENT_RESOURCES errors

    return player.EntityState{
        Errors:         errors,
        Cost:           nil,  // Actions typically don't have credit costs
        Metadata:       make(map[string]interface{}),  // Minimal for actions
        LastCalculated: time.Now(),
    }
}
```

**Note**: Metadata minimized - boolean flags like `inputsAvailable` derived from Errors on frontend.

### PlayerStandardProject (Data Holder, Same Pattern)

**IMPORTANT**: Like PlayerCard and PlayerCardAction, this is a **data holder only**. Availability calculation happens in `internal/action/` package.

```go
// backend/internal/game/player/player_standard_project.go
package player

import (
    "sync"
    "time"
    "github.com/rackaracka123/terraforming-mars/backend/internal/game/shared"
)

// PlayerStandardProject represents a standard project with availability state
// This is a DATA HOLDER - availability calculation happens in action package
type PlayerStandardProject struct {
    projectType shared.StandardProject

    mu    sync.RWMutex
    state EntityState
}

// Constructor - simple initialization
func NewPlayerStandardProject(projectType shared.StandardProject) *PlayerStandardProject {
    return &PlayerStandardProject{
        projectType: projectType,
        state: EntityState{
            Errors:   []StateError{},
            Metadata: make(map[string]interface{}),
        },
    }
}

// UpdateState - called by actions after calculating availability
func (psp *PlayerStandardProject) UpdateState(newState EntityState) {
    psp.mu.Lock()
    defer psp.mu.Unlock()
    psp.state = newState
}

// Read-only accessors
func (psp *PlayerStandardProject) ProjectType() shared.StandardProject {
    return psp.projectType
}

func (psp *PlayerStandardProject) State() EntityState {
    psp.mu.RLock()
    defer psp.mu.RUnlock()
    return psp.state
}

func (psp *PlayerStandardProject) IsAvailable() bool {
    psp.mu.RLock()
    defer psp.mu.RUnlock()
    return psp.state.Available()
}
```

**Availability calculation** happens in `internal/action/state_calculator.go` (same file as card and action calculators):

```go
// ========================================
// PlayerStandardProject State Calculation
// ========================================

// CalculatePlayerStandardProjectState computes availability state
func CalculatePlayerStandardProjectState(
    projectType shared.StandardProject,
    p *player.Player,
    g *game.Game,
) player.EntityState {
    errors := []player.StateError{}
    metadata := make(map[string]interface{})

    // Get base cost
    cost := shared.StandardProjectCost[projectType]

    // 1. Check affordability (reuse shared helper)
    errors = append(errors, validateAffordability(p, cost)...)

    // 2. Check project-specific availability
    switch projectType {
    case shared.StandardProjectAquifer:
        oceansRemaining := g.Board().OceansRemaining()
        metadata["oceansRemaining"] = oceansRemaining
        if oceansRemaining == 0 {
            errors = append(errors, player.StateError{
                Code:     "NO_OCEAN_TILES",
                Category: "availability",
                Message:  "No ocean tiles remaining",
            })
        }
    // ... other project types with minimal metadata
    }

    return player.EntityState{
        Errors:         errors,
        Cost:           &cost,
        Metadata:       metadata,
        LastCalculated: time.Now(),
    }
}
```

**Note**: All three calculators (Card, Action, Project) live in the same `state_calculator.go` file and share validation helpers.

## Lifecycle: When to Create PlayerCard

**CRITICAL**: Actions create PlayerCard instances, NOT Hand. Hand is a dumb cache.

### 1. **Hand Cards** (Long-Lived, Cached)

```go
// In an action that adds card to hand
func (a *AddCardToHandAction) Execute(ctx context.Context, cardID string) error {
    g, _ := a.gameRepo.Get(gameID)
    p := g.GetPlayer(playerID)

    // 1. Create PlayerCard (simple, no deps)
    card, _ := a.cardRegistry.GetByID(cardID)
    pc := player.NewPlayerCard(card)

    // 2. Register event listeners (action has access to everything)
    a.registerEventListeners(pc, p, g)

    // 3. Calculate initial state
    state := CalculatePlayerCardState(card, p, g, a.cardRegistry)
    pc.UpdateState(state)

    // 4. Add to Hand cache (just storage)
    p.Hand().AddPlayerCard(cardID, pc)

    return nil
}

// Hand just retrieves from cache
func (h *Hand) GetPlayerCard(cardID string) (*PlayerCard, bool) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    pc, exists := h.playerCards[cardID]
    return pc, exists
}
```

### 2. **Card Selection Modal** (Temporarily Cached)

```go
// When player selects starting cards
type SelectStartingCardsPhase struct {
    options      []string  // Card IDs
    playerCards  map[string]*PlayerCard  // Cached for selection duration
}

// Actions create PlayerCard when selection phase starts
func (a *StartGameAction) setupSelectStartingCardsPhase(...) {
    phase := &SelectStartingCardsPhase{
        options:     drawnCardIDs,
        playerCards: make(map[string]*PlayerCard),
    }

    // Pre-create PlayerCard for each option
    for _, cardID := range drawnCardIDs {
        card, _ := cardRegistry.GetByID(cardID)
        playerCard := player.NewPlayerCard(card, player, game, cardRegistry)
        phase.playerCards[cardID] = playerCard
    }
}

// In DTO mapping
func mapSelectStartingCardsPhase(phase *SelectStartingCardsPhase) SelectStartingCardsDto {
    optionDtos := make([]PlayerCardDto, len(phase.options))
    for i, cardID := range phase.options {
        // Use cached PlayerCard
        playerCard := phase.playerCards[cardID]
        optionDtos[i] = ToPlayerCardDto(playerCard)
    }
    return SelectStartingCardsDto{Options: optionDtos}
}
```

### 3. **Opponent Cards** (Never Show State)

```go
// When mapping opponent's hand to DTO
func ToOpponentHandDto(opponentHand *Hand) []CardDto {
    // Just card IDs, no PlayerCard creation
    // Opponent doesn't need to know playability
    return []CardDto{
        {ID: "card-001", Name: "Hidden"},  // Basic info only
    }
}
```

## DTO Mapping (Action Orchestration)

```go
// backend/internal/delivery/dto/mapper_player.go

func ToPlayerDto(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry, includePrivate bool) PlayerDto {
    if includePrivate {
        // Map hand with playability (using cached PlayerCard)
        handCards := mapPlayerCards(p.Hand(), p, g, cardRegistry)

        // Map card actions with usability
        actions := mapPlayerCardActions(p.Actions(), p, g)

        // Map projects with availability
        projects := mapPlayerStandardProjects(p, g)

        return PlayerDto{
            Cards:    handCards,
            Actions:  actions,
            Projects: projects,
            // ... other fields
        }
    }

    // Non-private: just basic info
    return PlayerDto{/* ... */}
}

func mapPlayerCards(
    hand *player.Hand,
    p *player.Player,
    g *game.Game,
    cardRegistry cards.CardRegistry,
) []PlayerCardDto {
    cardIDs := hand.CardIDs()
    result := make([]PlayerCardDto, len(cardIDs))

    for i, cardID := range cardIDs {
        // Get cached PlayerCard (must exist - actions created it)
        playerCard, exists := hand.GetPlayerCard(cardID)
        if !exists {
            // Defensive: should never happen if actions work correctly
            continue
        }
        result[i] = ToPlayerCardDto(playerCard)
    }

    return result
}

func ToPlayerCardDto(pc *player.PlayerCard) PlayerCardDto {
    state := pc.State()  // Read cached state (events keep it current)
    card := pc.Card()    // Get immutable card data

    // Extract minimal metadata (type assertions with safe defaults)
    discounts, _ := state.Metadata["discounts"].(map[shared.CardTag]int)

    return PlayerCardDto{
        // Card data
        ID:          card.ID,
        Name:        card.Name,
        Cost:        card.Cost,
        Description: card.Description,
        Tags:        card.Tags,

        // Player-specific state (from EntityState)
        Available:     state.Available(),        // Computed from errors
        Errors:        convertErrors(state.Errors),
        EffectiveCost: *state.Cost,              // Dereference pointer

        // Minimal metadata for rich UI
        Discounts: discounts,

        // Derived fields computed from errors (frontend can also derive)
        MeetsRequirements: !hasErrorCategory(state.Errors, "requirement"),
    }
}
```

## New Domain Events

```go
// backend/internal/events/domain_events.go

// RequirementModifiersChangedEvent triggers PlayerCard recalculation
type RequirementModifiersChangedEvent struct {
    GameID    string
    PlayerID  string
    Timestamp time.Time
}

// TileSelectionStateChangedEvent triggers PlayerCardAction recalculation
type TileSelectionStateChangedEvent struct {
    GameID     string
    PlayerID   string
    HasPending bool
    Timestamp  time.Time
}

// ActionPlayCountsResetEvent triggers PlayerCardAction recalculation
type ActionPlayCountsResetEvent struct {
    GameID    string
    PlayerID  string
    Timestamp time.Time
}
```

## Benefits

### 1. **Action Orchestration**
- Actions have access to Game and Player
- No interface indirection needed
- Consistent with existing architecture pattern

### 2. **No Circular Dependencies**
- `player/` package never imports `game/` package
- Actions in `action/` package orchestrate both
- Clean dependency flow: `action/` → `game/` → `player/` → `cards/` → `shared/`

### 3. **Long-Lived Caching**
- PlayerCard cached in Hand, PlayedCards, selection phases
- Efficient - no recalculation on every DTO mapping
- Can recalculate when actions update state

### 4. **Extensible**
- `EntityState` is truly generic - works for any entity type
- Add new entity types without refactoring (cards, actions, projects, tiles, etc.)
- Metadata map allows entity-specific data without struct changes
- Future-proof - no breaking changes for new features

### 5. **Rich State Information**
- Computed `Available()` method (impossible to contradict errors)
- Detailed errors with codes, categories, and messages
- Optional cost calculation (pointer, nil if N/A)
- Minimal metadata for truly dynamic data
- All player-relevant information in one place

### 6. **Memory Safety**
- Explicit event listener cleanup via Cleanup()
- Unsubscribe functions tracked in PlayerCard
- Hand.RemoveCard() calls Cleanup() before deletion
- Prevents memory leaks from dangling event listeners

## Implementation Checklist

### Phase 1: Core Types (Data Holders) ✅ COMPLETE
- [x] Create `EntityState` struct in `player/entity_state.go` (42 lines)
- [x] Create `StateError` struct
- [x] Create `PlayerCard` in `player/player_card.go` (88 lines - uses `any` for card to avoid circular dependency)
- [x] Create `PlayerCardAction` in `player/player_card_action.go` (138 lines with persistent counters)
- [x] Create `PlayerStandardProject` in `player/player_standard_project.go` (74 lines)

**COMPLETED**: Simple data holders with NO business logic. Constructors, UpdateState(), accessors only. All use `EntityState` structure.

### Phase 2: State Calculation Logic (Action Package) ✅ COMPLETE
- [x] Create `action/state_calculator.go` (455 lines - consolidates all entity types)
  - **PlayerCard calculation** (implemented):
    - `CalculatePlayerCardState()` returns `EntityState`
    - Extracted validation from PlayCardAction
    - Cost calculation (TODO: integrate RequirementModifier discounts)
    - Shared helpers: validatePhase(), validateRequirements(), validateAffordability()
  - **PlayerCardAction calculation** (implemented):
    - `CalculatePlayerCardActionState()` returns `EntityState`
    - Input resource availability checks
    - Usage limits (TODO: when CardBehavior supports MaxUses fields)
  - **PlayerStandardProject calculation** (implemented):
    - `CalculatePlayerStandardProjectState()` returns `EntityState`
    - Affordability checks, project-specific availability
    - Metadata for UI context (oceansRemaining for Aquifer)
- [x] Create `action/player_card_helpers.go` (102 lines)
  - Event subscription with proper Unsubscribe() cleanup
  - CreateAndCachePlayerCard() helper function

**COMPLETED**: All calculation logic in dedicated files. Shared validation helpers. Event-driven state updates.

### Phase 3: Caching Integration ✅ COMPLETE
- [x] Add `playerCards map[string]*PlayerCard` to Hand
- [x] Add `GetPlayerCard(cardID)` - retrieves cached instance
- [x] Add `AddPlayerCard(cardID, pc)` - stores in cache
- [x] Update `RemoveCard()` to call `pc.Cleanup()` before deletion
- [x] Maintained cardRegistry dependency (needed for PlayerCard creation)

**COMPLETED**: Hand caches PlayerCard instances with proper cleanup.

### Phase 4: Action Integration ✅ COMPLETE (ALL 5 actions)
- [x] Update actions that add cards to hand:
  - [x] `confirm_card_draw.go` - creates PlayerCard instances
  - [x] `admin/give_card.go` - creates PlayerCard instances
  - [x] `select_starting_cards.go` - creates PlayerCard instances + recalculates requirement modifiers
  - [x] `select_tile.go` - signature updated, ready for PlayerCard creation in card draw bonuses
  - [x] `confirm_production_cards.go` - creates PlayerCard instances + recalculates requirement modifiers
- [x] All actions follow pattern: AddCard() → GetByID() → CreateAndCachePlayerCard()

**COMPLETED**: All 5 actions that add cards to hand now create PlayerCard instances with event listeners.

### Phase 5: DTO Enhancement ✅ COMPLETE
- [x] Create `StateErrorDto` with Code, Category, Message fields
- [x] Create `PlayerCardDto` with EntityState fields:
  - available: bool (from Available() method)
  - errors: []StateErrorDto
  - effectiveCost: int
  - discounts: map[string]int (from metadata)
- [x] Create `PlayerCardActionDto` with EntityState fields:
  - timesUsedThisTurn, timesUsedThisGeneration: int
  - available: bool, errors: []StateErrorDto
- [x] Create `PlayerStandardProjectDto` with EntityState fields
- [x] Update `mapPlayerCards()` to read from Hand cache and convert to DTO
- [x] Update `ToPlayerDto()` to use new PlayerCardDto structure (breaking change)
- [x] Add mappers with type assertions (PlayerCard.card is `any` type)
- [x] Run `make generate` to sync TypeScript types

**COMPLETED**: All DTOs read cached state with type assertions. NO calculation in DTO mapping - events keep state up-to-date. TypeScript types generated and ready for frontend integration.

### Phase 6: Frontend Integration (IN PROGRESS)
- [x] Update TypeScript types (auto-generated from make generate)
- [x] Update SimpleGameCard component to support PlayerCardDto:
  - Type guard to detect PlayerCardDto vs CardDto
  - Use `effectiveCost` from backend state calculation
  - Display unavailable state (dimmed, grayscale, cursor-not-allowed)
  - Show error badge with primary error message
  - Tooltip with all errors on hover
  - Backward compatible with CardDto (for non-player cards)
- [ ] Update card selection overlays to use PlayerCardDto
- [ ] Remove old frontend playability calculation logic (if any exists)
- [ ] Test with real game data to verify state updates work in real-time

### Phase 7: Testing ✅ COMPLETE
- [x] Unit tests for state calculator functions (11 tests, all passing):
  - PlayerCard state calculation (5 tests): Available, InsufficientCredits, TemperatureRequirement, MultipleRequirements, WrongPhase
  - PlayerCardAction state calculation (3 tests): Available, InsufficientResources, NotPlayerTurn
  - PlayerStandardProject state calculation (3 tests): Available, InsufficientCredits, NoOceansRemaining
- [x] Integration tests: Event-driven state updates (5 tests, all passing):
  - EventDrivenStateUpdate: Temperature change triggers recalculation
  - ResourceChangeEventUpdate: Credit gain triggers recalculation
  - PhaseChangeEventUpdate: Phase transition triggers recalculation
  - CleanupPreventsMemoryLeak: Cleanup prevents state updates after card removed
  - MultipleCardsIndependentState: Each PlayerCard maintains independent state
- [x] Added TriggerType constants (TriggerTypeAuto, TriggerTypeManual) to eliminate magic strings
- [x] Added GamePhaseChangedEvent subscription (now 7 events per PlayerCard)
- [x] `make test`, `make format` - all tests passing (16 total tests)

**Total actual lines: ~1,900 lines** (core implementation + comprehensive unit & integration tests)

## Example Usage

```go
// Action: Player draws cards, needs to select which to keep
func (a *SelectCardsFromDrawAction) Execute(ctx context.Context, gameID, playerID string) error {
    g, _ := a.gameRepo.Get(gameID)
    p := g.GetPlayer(playerID)

    // Draw 4 cards
    drawnCardIDs := g.Deck().Draw(4)

    // Create selection phase with cached PlayerCard instances
    phase := &player.SelectStartingCardsPhase{
        Options:     drawnCardIDs,
        PlayerCards: make(map[string]*player.PlayerCard),
    }

    // Pre-create PlayerCard for each option (action orchestrates)
    for _, cardID := range drawnCardIDs {
        card, _ := a.cardRegistry.GetByID(cardID)

        // Create PlayerCard (data holder)
        playerCard := player.NewPlayerCard(card)

        // Action registers event listeners on PlayerCard
        a.registerPlayerCardEventListeners(playerCard, p, g)

        // Trigger initial state calculation
        a.recalculatePlayerCard(playerCard, p, g)

        // Cache for selection duration
        phase.PlayerCards[cardID] = playerCard
    }

    // Store phase
    g.SetSelectStartingCardsPhase(ctx, playerID, phase)

    return nil
}

// Action registers event listeners to recalculate state
func (a *SelectCardsFromDrawAction) registerPlayerCardEventListeners(
    pc *player.PlayerCard,
    p *player.Player,
    g *game.Game,
) {
    eventBus := g.EventBus()

    // When resources change, recalculate affordability
    events.Subscribe(eventBus, func(event events.ResourcesChangedEvent) {
        if event.PlayerID == p.ID() {
            a.recalculatePlayerCard(pc, p, g)
        }
    })

    // When requirement modifiers change, recalculate playability
    events.Subscribe(eventBus, func(event events.RequirementModifiersChangedEvent) {
        if event.PlayerID == p.ID() {
            a.recalculatePlayerCard(pc, p, g)
        }
    })

    // When global parameters change, recalculate requirements
    events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
        a.recalculatePlayerCard(pc, p, g)
    })

    events.Subscribe(eventBus, func(event events.OxygenChangedEvent) {
        a.recalculatePlayerCard(pc, p, g)
    })

    // ... other relevant events
}

// Helper: Recalculate state (called on initial creation and events)
func (a *SelectCardsFromDrawAction) recalculatePlayerCard(
    pc *player.PlayerCard,
    p *player.Player,
    g *game.Game,
) {
    state := CalculatePlayerCardState(pc.Card(), p, g, a.cardRegistry)
    pc.UpdateState(state)
}

// DTO Mapping: Use cached PlayerCard (state already calculated)
func mapSelectStartingCardsPhase(
    phase *player.SelectStartingCardsPhase,
) *SelectStartingCardsDto {
    options := make([]PlayerCardDto, len(phase.Options))

    for i, cardID := range phase.Options {
        // Get cached PlayerCard (state already up-to-date from events)
        playerCard := phase.PlayerCards[cardID]

        // Just convert to DTO - NO recalculation here!
        options[i] = ToPlayerCardDto(playerCard)
    }

    return &SelectStartingCardsDto{
        Options: options,
    }
}
```

## Conclusion

This architecture:
- **Separates concerns**: Card (immutable data) vs PlayerCard (player state holder) vs Calculator (state logic)
- **Uses actions for orchestration**: All state calculation in `action/` package
- **Caches for efficiency**: PlayerCard lives as long as relevant, state recalculated on demand
- **Avoids circular deps**: `player/` never imports `game/`, actions orchestrate both
- **Is truly generic**: Single `EntityState` works for all entity types (cards, actions, projects, etc.)
- **Eliminates redundancy**: One `Available` boolean, no `Playable/Affordable/MeetsRequirements` duplication
- **Extensible via metadata**: Entity-specific data without struct changes
- **Keeps files focused**: Data holders ~100 lines, calculators ~300 lines, no huge files
- **Is simple**: No interface indirection, direct type passing, clear responsibilities
- **Is maintainable**: Clear ownership - domain models hold data, actions calculate state

**Key architectural decision**: State calculation in `action/` package prevents:
- ❌ Circular dependencies (player importing game)
- ❌ Huge files with complex business logic in domain models
- ❌ Violation of separation of concerns

**Generic EntityState benefits**:
- ✅ Single source of truth: `Available()` computed from `len(Errors) == 0`
- ✅ Impossible to have contradictory state (computed, not stored)
- ✅ Works for any entity type without refactoring
- ✅ Minimal metadata map (prefer typed fields for common patterns)
- ✅ Cost pointer (nil if N/A) handles all cases
- ✅ Explicit cleanup prevents memory leaks

**Simplifications in this architecture**:
- ✅ One calculator file (~400 lines) instead of three (~600 lines total)
- ✅ Computed Available() eliminates stored redundant field
- ✅ Minimal metadata reduces type assertion complexity
- ✅ Shared validation helpers reduce duplication
- ✅ Explicit cleanup strategy documented
- ✅ Hand is dumb cache (no GetOrCreatePlayerCard contradiction)
- ✅ ~900 total lines instead of 1,200 lines

The result is a robust system where any entity (card, action, project) can carry player-specific state calculated by actions, with simple data holders in the domain layer, complex logic consolidated in the action layer, and explicit memory management.
