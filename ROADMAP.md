# Terraforming Mars Hybrid State & Event Architecture Roadmap

## Architecture Overview

This roadmap enhances the existing Terraforming Mars implementation with a hybrid approach:
- **Single Game State Object** containing all game data for simplicity
- **Parallel Event Stream** for action logging and audit trail (like Steam version)
- **Enhanced Card System** with YAML definitions for maintainability
- **Improved Phase Management** within the unified game state

## Current State Analysis

### Existing Strengths
- Clean architecture with domain/service/delivery separation
- Event bus with worker pool for async processing
- Card validation and effect processing foundation
- WebSocket real-time multiplayer system
- Type generation from Go to TypeScript

### Architecture Gaps
- No action log/audit trail (Steam version has this)
- Card definitions scattered across Go structs (hard to maintain 200+ cards)
- Limited event granularity for debugging
- Phase management could be cleaner

## Target Architecture Components

### 1. Unified Game State Structure

**Single Game Object (Everything Embedded):**
```go
type Game struct {
    // Core game identifiers
    ID                    string        `json:"id"`
    Status                GameStatus    `json:"status"`
    HostPlayerID          string        `json:"hostPlayerId"`

    // All game state in one place
    Players               []Player      `json:"players"`           // Embedded players
    GlobalParameters      GlobalParameters `json:"globalParameters"` // Embedded global state
    CurrentPhase          Phase         `json:"currentPhase"`      // Current game phase
    Deck                  []Card        `json:"deck,omitempty"`    // Remaining cards

    // Game metadata
    CreatedAt             time.Time     `json:"createdAt"`
    LastUpdated           time.Time     `json:"lastUpdated"`
    EventSequence         int           `json:"eventSequence"`    // For event ordering
}

// Enhanced Player with cleaner phase management
type Player struct {
    ID               string      `json:"id"`
    Name             string      `json:"name"`
    Resources        Resources   `json:"resources"`
    Production       Production  `json:"production"`
    TerraformRating  int         `json:"terraformRating"`
    PlayedCards      []Card      `json:"playedCards"`     // Full card objects
    Hand             []Card      `json:"hand,omitempty"`  // Full card objects (private)
    Corporation      *Card       `json:"corporation,omitempty"`

    // Private phase data (only for current player)
    Phase            *Phase      `json:"phase,omitempty"` // Player-specific phase
}

// Parallel Event Stream (for action logging)
type GameEvent struct {
    ID        string    `json:"id"`
    GameID    string    `json:"gameId"`
    Type      string    `json:"type"`        // "card_played", "temperature_raised", etc.
    PlayerID  *string   `json:"playerId,omitempty"`
    Timestamp time.Time `json:"timestamp"`
    Sequence  int       `json:"sequence"`    // Order within game
    Action    string    `json:"action"`      // Human-readable action
    Details   any       `json:"details"`     // Action-specific data
}
```

**Simple Repository Pattern:**
```go
type GameRepository interface {
    GetByID(ctx context.Context, gameID string) (*Game, error)
    Create(ctx context.Context, game *Game) error
    Update(ctx context.Context, game *Game) error
}

type EventRepository interface {
    GetByGameID(ctx context.Context, gameID string) ([]GameEvent, error)
    GetByGameIDSince(ctx context.Context, gameID string, sequence int) ([]GameEvent, error)
    Create(ctx context.Context, event *GameEvent) error
}
```

## Implementation Plan - Incremental Feature Chunks

### 🎯 Feature 1: Basic Event Logging System (Week 1)
**Goal**: Add action logging like Steam version without breaking existing architecture

**Deliverable**: Action log UI showing recent player actions
- ✅ **Testable**: Can see "Player X raised temperature" in action log
- ✅ **Incremental**: Works alongside current game state system

**Implementation:**
```go
// Add to existing Game struct
type Game struct {
    // ... existing fields
    EventSequence int `json:"eventSequence"`
}

// New parallel event tracking
type GameEvent struct {
    ID        string    `json:"id"`
    GameID    string    `json:"gameId"`
    Type      string    `json:"type"`        // "temperature_raised", "card_played"
    PlayerID  *string   `json:"playerId,omitempty"`
    Timestamp time.Time `json:"timestamp"`
    Sequence  int       `json:"sequence"`
    Action    string    `json:"action"`      // "Alice raised temperature +1°C"
    Details   any       `json:"details"`
}
```

**Files to Create/Modify:**
- `internal/model/event.go` - GameEvent struct
- `internal/repository/event_repository.go` - Simple event storage
- `internal/service/event_service.go` - Event creation and logging
- `internal/delivery/dto/event_dto.go` - Event DTOs for frontend
- Modify existing services to log events on actions
- `frontend/src/components/ActionLog.tsx` - Display recent events

**Test Success Criteria:**
- Action log shows when players raise temperature
- Action log shows when players select corporations
- Events persist and display chronologically
- No impact on existing game functionality

### 🎯 Feature 2: Enhanced Game State Structure (Week 2)
**Goal**: Migrate to unified Game object with embedded data

**Deliverable**: Single Game object containing all state, simplified repository
- ✅ **Testable**: Game state loads/saves correctly with embedded players
- ✅ **Incremental**: Maintains compatibility with current WebSocket system

**Implementation:**
```go
// Enhanced unified Game struct
type Game struct {
    // Core identifiers
    ID                    string        `json:"id"`
    Status                GameStatus    `json:"status"`
    HostPlayerID          string        `json:"hostPlayerId"`

    // All game state embedded
    Players               []Player      `json:"players"`
    GlobalParameters      GlobalParameters `json:"globalParameters"`
    CurrentPhase          Phase         `json:"currentPhase"`

    // Event tracking
    EventSequence         int           `json:"eventSequence"`

    // Metadata
    CreatedAt             time.Time     `json:"createdAt"`
    LastUpdated           time.Time     `json:"lastUpdated"`
}
```

**Files to Modify:**
- `internal/model/game.go` - Update Game struct
- `internal/repository/game_repository.go` - Simplify to single game operations
- Update all services to work with embedded data
- Update DTOs and type generation
- Run `make generate` to sync TypeScript types

**Test Success Criteria:**
- All existing functionality works with new structure
- WebSocket updates continue to work
- Game creation/joining/reconnection works
- Frontend receives correct typed data

### 🎯 Feature 3: JSON Card System Foundation (Week 3)
**Goal**: Create 5-10 basic cards from JSON definitions to prove the system

**Deliverable**: Working JSON card loader with basic cards functional
- ✅ **Testable**: Can play Power Plant card loaded from JSON
- ✅ **Incremental**: Coexists with existing Go card structs

**Implementation:**
```json
{
  "id": "power_plant",
  "name": "Power Plant",
  "cost": 11,
  "type": "automated",
  "tags": ["power", "building"],
  "description": "Increase your energy production 1 step.",
  "behaviors": [
    {
      "triggers": [
        {
          "type": "auto"
        }
      ],
      "outputs": [
        {
          "type": "energy-production",
          "amount": 1,
          "target": "self-player"
        }
      ]
    }
  ]
}
```

**Architecture Components:**
- **CardFactory Interface**: Loads JSON definitions and builds executable card objects
- **EffectRegistry Pattern**: Maps JSON behavior types to executable functions
- **CardDefinition Model**: Represents the JSON structure with behaviors, triggers, and outputs
- **Integration Service**: Bridges card system with existing game services

**Key Design Patterns:**
- **Behavior Composition**: Cards are built by composing behaviors from JSON definitions
- **Effect Function Registry**: JSON "type" fields map to executable micro-functions
- **Trigger System**: "auto" vs "manual" triggers determine when behaviors execute
- **Resource Target System**: "self-player", "any", "self-card" define effect scope

**Implementation Strategy:**
- Leverage existing comprehensive JSON card database
- Build effect functions that match existing JSON behavior patterns
- Create card instances that integrate with current game state management
- Ensure seamless interaction with existing validation and event systems

### 🎯 Feature 4: Action Availability System (Week 4)
**Goal**: Backend calculates available actions instead of frontend logic

**Deliverable**: Frontend receives available actions list, no client-side validation
- ✅ **Testable**: Frontend shows only playable cards
- ✅ **Incremental**: Replaces existing card validation gradually

**Implementation:**
```go
type Action struct {
    ID          string `json:"id"`          // "play_card:power_plant"
    Type        ActionType `json:"type"`    // "play_card", "standard_project"
    Description string `json:"description"` // "Play Power Plant (11 MC)"
    Cost        *int   `json:"cost,omitempty"`
}

type AvailabilityService interface {
    GetAvailableActions(ctx context.Context, gameID, playerID string) ([]Action, error)
    CanPlayCard(ctx context.Context, gameID, playerID, cardID string) (bool, string, error)
}
```

**Architecture Components:**
- **AvailabilityService Interface**: Calculates which actions are legal for current game state
- **Action Model**: Represents available player actions with ID, type, description, and cost
- **WebSocket Integration**: Extend game state DTOs to include available actions
- **Frontend Action Hooks**: React integration for backend-provided action lists

**Test Success Criteria:**
- Frontend shows only cards player can afford
- Unavailable cards are grayed out with reason
- Action buttons only appear when valid
- No duplicate validation logic between frontend/backend

### 🎯 Feature 5: Enhanced Phase Management (Week 5)
**Goal**: Clean up phase transitions and private player phases

**Deliverable**: Smooth phase transitions with proper private data handling
- ✅ **Testable**: Corporation selection phase works cleanly
- ✅ **Incremental**: Improves existing phase system

**Implementation:**
```go
type Phase struct {
    Type      PhaseType   `json:"type"`
    Data      any         `json:"data"`
    StartedAt time.Time   `json:"startedAt"`
}

type PhaseType string
const (
    PhaseLobby                 PhaseType = "lobby"
    PhaseCorporationSelection  PhaseType = "corporation_selection"
    PhasePlayerAction         PhaseType = "player_action"
    PhaseProduction           PhaseType = "production"
)
```

**Architecture Improvements:**
- **Enhanced Phase Types**: More granular phase definitions with clear transitions
- **Phase Transition Logic**: State machine approach to phase progression
- **Private Data Handling**: Secure separation of player-specific phase information
- **UI Phase Adaptation**: Dynamic component rendering based on phase state

**Test Success Criteria:**
- Clean phase transitions (lobby → corp selection → action)
- Private phases only show data to relevant player
- Phase-specific UI renders correctly
- Action log shows phase transitions

### 🎯 Feature 6: Advanced Card Effects (Week 6)
**Goal**: Implement complex card interactions and ongoing effects

**Deliverable**: Blue (action) cards with multiple abilities working
- ✅ **Testable**: Research card lets you draw and discard
- ✅ **Incremental**: Builds on JSON card system

**Implementation:**
```json
{
  "id": "research",
  "name": "Research",
  "type": "active",
  "cost": 11,
  "description": "Action: Pay 1 MC: Draw 2 cards, keep 1",
  "behaviors": [
    {
      "triggers": [
        {
          "type": "manual"
        }
      ],
      "inputs": [
        {
          "type": "credits",
          "amount": 1,
          "target": "self-player"
        }
      ],
      "outputs": [
        {
          "type": "card-draw",
          "amount": 2,
          "target": "self-player"
        }
      ],
      "choices": [
        {
          "outputs": [
            {
              "type": "card-keep",
              "amount": 1,
              "target": "self-player"
            }
          ]
        }
      ]
    }
  ]
}
```

**Architecture Extensions:**
- **Multi-Choice Behavior System**: Handle JSON "choices" arrays for complex card actions
- **Action Card Interface**: Support "manual" triggers for player-activated abilities
- **Resource Cost System**: Process "inputs" arrays for action costs
- **Card State Management**: Track card-specific resources and counters

**Test Success Criteria:**
- Blue cards show action buttons
- Card actions consume resources correctly
- Action log shows card ability usage
- Multi-step effects work (draw cards, choose which to keep)

### 🎯 Feature 7: Complete Event Stream UI (Week 7)
**Goal**: Rich action log with filtering and game replay features

**Deliverable**: Comprehensive action log like Steam version
- ✅ **Testable**: Can see full game history and filter by player/action type
- ✅ **Incremental**: Enhances basic event logging from Feature 1

**Implementation:**
- Advanced action log filtering (by player, by action type)
- Event details expansion (click to see full action details)
- Game timeline visualization
- Export game log functionality

**Files to Create:**
- `frontend/src/components/ActionLog/FilterableActionLog.tsx`
- `frontend/src/components/ActionLog/EventDetails.tsx`
- `frontend/src/services/eventFilter.ts`

**Test Success Criteria:**
- Can filter events by player
- Can filter events by type (cards, standard projects, etc.)
- Event details show complete action information
- Action log updates in real-time during game

### 🎯 Feature 8: Performance & Polish (Week 8)
**Goal**: Optimize event processing and add quality-of-life improvements

**Deliverable**: Smooth performance with large event logs, polished UI
- ✅ **Testable**: Action log performs well with 100+ events
- ✅ **Incremental**: Optimizes existing systems without breaking them

**Implementation:**
- Event pagination for large games
- WebSocket event batching
- Action log virtualization for performance
- Event search functionality

**Test Success Criteria:**
- Action log scrolls smoothly with many events
- Real-time updates don't cause lag
- Search finds specific actions quickly
- Game performs well in long sessions

---

## Key Principles for Each Feature:

### ✅ **Incremental**:
Each feature builds on the previous ones without breaking existing functionality

### ✅ **Testable**:
Every feature has clear success criteria and can be tested independently

### ✅ **Valuable**:
Each feature provides immediate user value and can be shipped

### ✅ **Focused**:
One clear goal per feature, avoiding scope creep

### ✅ **Compatible**:
Features work with existing architecture and don't require massive rewrites

### 2. Event-Sourced Card System with Reactive Availability

#### 2.1 Reactive Card Availability Architecture
Cards subscribe to relevant events and self-manage their availability:

```go
// Cards subscribe to relevant events and update their own availability
type Card struct {
    ID           string        `json:"id"`
    Name         string        `json:"name"`
    Cost         int           `json:"cost"`
    Type         CardType      `json:"type"`

    // Availability state (calculated by card itself)
    IsAvailable  bool          `json:"isAvailable"`
    BlockReason  string        `json:"blockReason,omitempty"`

    // Event subscriptions (what this card cares about)
    Subscriptions []EventType  `json:"-"` // Not sent to frontend

    // Composed behavior from definitions
    AvailabilityChecker AvailabilityChecker `json:"-"`
    EffectHandler       EffectHandler       `json:"-"`
}

type CardAvailabilityManager struct {
    cardSubscriptions map[string][]EventType  // cardID -> events it cares about
    availabilityCache map[string]bool         // cardID -> current availability
}

// When any event happens, notify subscribed cards
func (cam *CardAvailabilityManager) HandleEvent(event Event) {
    for cardID, subscriptions := range cam.cardSubscriptions {
        if containsEventType(subscriptions, event.Type) {
            cam.updateCardAvailability(cardID, event)
        }
    }
}
```

#### 2.2 Declarative Card Creation System
200+ cards created from JSON definitions instead of individual Go structs:

**Simple Card Definition:**
```json
{
  "id": "card_001",
  "name": "Power Plant",
  "cost": 11,
  "type": "automated",
  "tags": ["power", "building"],
  "description": "Increase your energy production 1 step. Requires temperature -12°C or warmer.",
  "requirements": {
    "temperature": -12
  },
  "behaviors": [
    {
      "triggers": [
        {
          "type": "auto"
        }
      ],
      "outputs": [
        {
          "type": "energy-production",
          "amount": 1,
          "target": "self-player"
        }
      ]
    }
  ]
}
```

**Complex Card with Multiple Actions:**
```json
{
  "id": "card_234",
  "name": "Business Network",
  "cost": 4,
  "type": "active",
  "tags": ["earth"],
  "description": "Decrease your hand size limit by 1. Action: Draw 1 card OR Pay 3 MC to buy 1 card.",
  "behaviors": [
    {
      "triggers": [
        {
          "type": "auto"
        }
      ],
      "outputs": [
        {
          "type": "hand-size-limit",
          "amount": -1,
          "target": "self-player"
        }
      ]
    },
    {
      "triggers": [
        {
          "type": "manual"
        }
      ],
      "choices": [
        {
          "outputs": [
            {
              "type": "card-draw",
              "amount": 1,
              "target": "self-player"
            }
          ]
        },
        {
          "inputs": [
            {
              "type": "credits",
              "amount": 3,
              "target": "self-player"
            }
          ],
          "outputs": [
            {
              "type": "card-draw",
              "amount": 1,
              "target": "self-player",
              "from": "market"
            }
          ]
        }
      ]
    }
  ]
}
```

#### 2.3 Card Factory System
```go
type CardFactory struct {
    definitions map[string]*CardDefinition
    registry    map[string]*Card
    effectRegistry *EffectRegistry
}

func (cf *CardFactory) LoadCardDefinitions(dir string) error {
    files, err := filepath.Glob(filepath.Join(dir, "*.json"))
    // Load all JSON definitions and build cards
    return cf.buildAllCards()
}

func (cf *CardFactory) BuildCard(def *CardDefinition) (*Card, error) {
    card := &Card{
        ID:            def.ID,
        Name:          def.Name,
        Cost:          def.Cost,
        Type:          def.Type,
        Tags:          def.Tags,
        Subscriptions: def.Subscriptions,
    }

    // Build availability checker from definition
    card.AvailabilityChecker = cf.BuildAvailabilityChecker(def.AvailabilityChecks)

    // Build effect handlers from definition
    card.EffectHandler = cf.BuildEffectHandler(def.Effects)

    return card, nil
}
```

#### 2.4 Effect Composition System
```go
type EffectRegistry struct {
    immediateEffects map[string]ImmediateEffectFunc
    ongoingEffects   map[string]OngoingEffectFunc
    actionEffects    map[string]ActionEffectFunc
}

// Micro-function for production increase
func IncreaseProductionEffect(ctx EffectContext, params EffectParams) error {
    resource := params["resource"].(string)
    amount := params["amount"].(int)

    switch resource {
    case "energy":
        ctx.Player.Production.Energy += amount
    case "megacredits":
        ctx.Player.Production.MegaCredits += amount
    }
    return nil
}
```

#### 2.5 Event-Driven Control Flow
- **Granular Events**: Every card operation generates specific events
- **Micro Functions**: 5-10 line functions for each operation (but not one-liners)
- **Clear Control Flow**: Predictable sequence of operations
- **Event Log**: Complete audit trail of all game actions
- **Reactive Updates**: Cards automatically recalculate availability on state changes

### 3. Phase Management System
- **Public Phases**: Game-wide phases visible to all players
- **Private Phases**: Player-specific phases with sensitive data
- **Type Safety**: Strongly typed phase data structures
- **Event-Driven**: Phase transitions through events

## Implementation Plan

### Phase 1: Core Event & Phase Infrastructure (Week 1)

#### 1.1 Enhanced Event System
**Files to Create/Modify:**
- `internal/events/event.go` - Add granular event types
- `internal/events/store.go` - Event store with audit trail
- `internal/events/query.go` - Event querying interface

**New Event Types:**
```go
// Granular Card Events
EventTypeCardValidated
EventTypeCardCostPaid
EventTypeCardEffectApplied
EventTypeCardReactionTriggered
EventTypeResourceGenerated
EventTypeTileEffectApplied

// Phase Events
EventTypePhaseStarted
EventTypePhaseCompleted
EventTypePhaseTransition
```

**Key Features:**
- Immutable event log storage
- Event sequence numbering
- Event replay capabilities
- Historical game state queries

#### 1.2 Phase System
**Architecture Components:**
- **Phase Model**: Strongly-typed phase data structures with type-specific payloads
- **PhaseService Interface**: Manages phase transitions and validation
- **Phase State Management**: Clean separation between public and private phase data

**Phase Types:**
```go
type PhaseType string
const (
    // Game Initialization (grouped)
    PhaseGameInitialize     PhaseType = "game_initialize"     // Setup + corporation selection

    // Generation Phases (grouped by workflow)
    PhaseProductionAndCardSelection PhaseType = "production_and_card_selection" // Production + card selection
    PhasePlayerAction               PhaseType = "player_action"                 // Main action phase

    // Card Decision Phases
    PhaseCardDecision       PhaseType = "card_decision"      // Private card decisions (Business Contacts, etc.)

    // Game End
    PhaseGameEnd           PhaseType = "game_end"
)
```

**Phase Data Structures:**
```go
// Game Initialization Phase Data
type GameInitializeData struct {
    AvailableCorporations []string `json:"availableCorporations"` // Corporations to choose from
    InitialCards          []string `json:"initialCards"`          // Starting card options
    PreludeCards          []string `json:"preludeCards"`          // Prelude card options (if enabled)
}

// Production + Card Selection Phase Data (combined workflow)
type ProductionAndCardSelectionData struct {
    Step            string   `json:"step"`            // "production" or "card_selection"
    Generation      int      `json:"generation"`      // Current generation number
    AvailableCards  []string `json:"availableCards,omitempty"`  // For card selection step
    MaxSelect       int      `json:"maxSelect,omitempty"`       // For card selection step
    MinSelect       int      `json:"minSelect,omitempty"`       // For card selection step
    CostPerCard     int      `json:"costPerCard,omitempty"`     // For card selection step
}

type PlayerActionData struct {
    ActionsRemaining int      `json:"actionsRemaining"`
    AvailableActions []string `json:"availableActions"`
}
```

#### 1.3 Repository Architecture System
**Files to Create/Modify:**
- `internal/repository/game_repository.go` - Game references only
- `internal/repository/player_repository.go` - Player data storage
- `internal/repository/global_parameters_repository.go` - Global game state
- `internal/repository/phase_repository.go` - Phase state storage
- `internal/repository/event_repository.go` - Event storage and querying
- `internal/service/game_dto_service.go` - DTO composition and building
- `internal/service/availability_service.go` - Action/card availability calculation

**Key Features:**
- **Normalized data storage** - no embedded objects in domain models
- **Separate repositories** for each domain entity
- **DTO composition** - merge data from multiple repositories for responses
- **Reference-based relationships** - Game contains IDs, DTO resolves to objects
- **Private data filtering** per player in DTO builder
- **Backend-calculated availability** (Critical DRY implementation)

#### 1.4 DTO Composition System
**Files to Create:**
- `internal/service/dto_builder_service.go` - Compose DTOs from repositories

```go
type DTOBuilderService interface {
    BuildGameDTO(ctx context.Context, gameID, requestingPlayerID string) (*shared.GameDTO, error)
    BuildPlayerDTO(ctx context.Context, player *Player, isCurrentPlayer bool) (*shared.PlayerDTO, error)
}
```
```

#### 1.4 Action Availability System (DRY Implementation)
**Files to Create:**
- `internal/service/availability_service.go` - Calculate available actions
- `internal/model/action.go` - Action type definitions

**Action Availability Architecture:**
```go
type Action struct {
    ID          string     `json:"id"`          // Unique action identifier - CRITICAL
    Type        ActionType `json:"type"`        // "play_card", "standard_project", etc.
    Description string     `json:"description"` // Human-readable action
    Cost        *int       `json:"cost,omitempty"`        // Action cost
    Requirements []string  `json:"requirements,omitempty"` // Why this action is available
}

// Action ID Format Examples - Actions are NOT limited to cards:
// "play_card:card_123"           - Play specific card from hand
// "standard_project:power_plant" - Build power plant standard project
// "standard_project:asteroid"    - Asteroid standard project
// "standard_project:aquifer"     - Aquifer standard project
// "convert_heat:temperature"     - Convert 8 heat to raise temperature
// "convert_plants:greenery"      - Convert 8 plants to greenery tile

// Played Card Actions (Blue Cards with ongoing abilities):
// Single Action Cards:
// "use_active_card:card_456"         - Use card_456's only action
// "use_active_card:card_789"         - Use card_789's only action

// Multiple Action Cards (need action specifier):
// "use_active_card:card_012:action_a" - Use card_012's first action choice
// "use_active_card:card_012:action_b" - Use card_012's second action choice
// "use_active_card:card_034:convert"  - Use card_034's conversion action
// "use_active_card:card_034:draw"     - Use card_034's card draw action

// Other Actions:
// "use_corporation_ability"      - Use corporation power
// "trade_with_colony:luna"       - Trade with specific colony
// "pass_turn"                    - End current turn
// "claim_milestone:gardener"     - Claim specific milestone
// "fund_award:landlord"          - Fund specific award
```

type AvailabilityService interface {
    GetPlayableCards(ctx context.Context, gameID, playerID string) ([]string, error)
    GetAvailableActions(ctx context.Context, gameID, playerID string) ([]Action, error)
    CanPlayCard(ctx context.Context, gameID, playerID, cardID string) (bool, string, error)
}
```

**Enhanced Card Model for Active Cards:**
```go
type Card struct {
    // Existing fields...
    ID                string             `json:"id"`
    Name              string             `json:"name"`
    Type              CardType           `json:"type"` // "automated", "active", "event"

    // Active Card Action Fields (for blue cards)
    Actions           []CardAction       `json:"actions,omitempty"`           // Multiple possible actions
}

type CardAction struct {
    ID                string             `json:"id"`                          // Unique action identifier within card
    Description       string             `json:"description"`                 // "Convert 1 Steel → 2 MC"
    Cost              *int               `json:"cost,omitempty"`             // Cost to use this action
    Requirements      *CardRequirements  `json:"requirements,omitempty"`     // Requirements for this action
    UsesPerTurn       *int               `json:"usesPerTurn,omitempty"`      // Uses per turn for this action
}

// Example Active Cards (using actual card IDs):
// card_001 (Steelworks):
//   Actions: [{ ID: "convert", Description: "Convert 1 Energy → 1 Steel" }]
//
// card_167 (Research):
//   Actions: [{ ID: "draw", Description: "Draw 2 cards, keep 1", Cost: 1 }]
//
// card_234 (Business Network):
//   Actions: [
//     { ID: "draw", Description: "Draw 1 card", Cost: nil },
//     { ID: "buy", Description: "Buy 1 card for 3 MC", Cost: 3 }
//   ]
//
// card_345 (Power Supply Consortium):
//   Actions: [
//     { ID: "energy", Description: "Gain 1 Energy", Cost: nil },
//     { ID: "power", Description: "Gain 3 MC", Cost: 1 } // Costs 1 Energy
//   ]

// Example Cards with Private Decision Phases:
// card_078 (Business Contacts) - "Look at top 4 cards, buy any for 1 MC each"
// card_134 (Research Development) - "Look at top 2 cards, keep 1"
// card_203 (Invention Contest) - "Look at top 3 cards, keep 1"

**Private Card Decisions (Business Contacts, Research, etc.):**
- Player plays card → Backend creates player-specific phase with card options
- Only current player receives phase with full card details in DTO
- Other players see no phase data (privacy guarantee)
- Player makes selections via action system

**Action System:**
- Backend calculates all available actions
- Frontend receives `availableActions` array with ID, type, description, cost
- User actions send `actionId` to backend
- No game logic in frontend

**Action ID Examples:**
- `"play_card:card_123"`, `"use_active_card:card_456:draw"`, `"standard_project:asteroid"`
```

### Phase 2: Card Factory & Availability System Implementation (Week 2)

#### 2.1 Card Factory System
**Files to Create:**
- `backend/cards/factory/factory.go` - CardFactory and card loading
- `backend/cards/factory/availability.go` - Availability checker builders
- `backend/cards/factory/effects.go` - Effect handler builders
- `backend/cards/definitions/` - JSON card definition directories

**Card Factory Implementation:**
```go
type CardFactory struct {
    definitions map[string]*CardDefinition
    registry    map[string]*Card
    effectRegistry *EffectRegistry
}

func (cf *CardFactory) LoadCardDefinitions(dir string) error {
    // Load all YAML files recursively
    // Validate definitions
    // Build cards with composed behavior
    return cf.buildAllCards()
}
```

#### 2.2 Card Availability Manager
**Files to Create:**
- `internal/service/card_availability_service.go` - Reactive availability system
- `internal/events/card_events.go` - Card-specific event types

**Availability System:**
```go
type CardAvailabilityManager struct {
    cardSubscriptions map[string][]EventType
    availabilityCache map[string]bool
}

// Event-driven availability updates
func (cam *CardAvailabilityManager) HandleEvent(event Event) {
    for cardID, subscriptions := range cam.cardSubscriptions {
        if containsEventType(subscriptions, event.Type) {
            cam.updateCardAvailability(cardID, event)
        }
    }
}
```

#### 2.3 Effect Micro-Functions Registry
**Files to Create:**
- `backend/cards/effects/immediate.go` - Immediate effect micro-functions
- `backend/cards/effects/ongoing.go` - Ongoing effect micro-functions
- `backend/cards/effects/actions.go` - Action effect micro-functions

**Registry Pattern:**
```go
type EffectRegistry struct {
    immediateEffects map[string]ImmediateEffectFunc
    ongoingEffects   map[string]OngoingEffectFunc
    actionEffects    map[string]ActionEffectFunc
}

// Micro-functions for common effects
func IncreaseProductionEffect(ctx EffectContext, params EffectParams) error
func GainResourcesEffect(ctx EffectContext, params EffectParams) error
func RaiseTemperatureEffect(ctx EffectContext, params EffectParams) error
```

#### 2.4 Application Integration
**Files to Modify:**
- `cmd/server/main.go` - Add card factory initialization
- `internal/service/game_service.go` - Integrate availability manager

**Initialization Flow:**
```go
// Application startup
func (app *Application) InitializeCards() error {
    cardFactory := NewCardFactory()
    cardFactory.LoadCardDefinitions("backend/assets/terraforming_mars_cards.json")

    availabilityManager := NewCardAvailabilityManager()
    for _, card := range cardFactory.GetAllCards() {
        availabilityManager.RegisterCard(card)
    }

    return nil
}
```

### Phase 3: Advanced Card Interactions (Week 3)

#### 3.1 Card Reaction System
**Files to Create:**
- `internal/cards/reactions/manager.go` - Reaction subscription system
- `internal/cards/reactions/handlers.go` - Card-specific reaction handlers
- `internal/cards/reactions/triggers.go` - Event trigger definitions

**Key Features:**
- Cards subscribe to relevant events
- Automatic reaction triggering
- Complex card interaction chains

#### 3.2 Complex Effect Chains
- Cascading effects from single actions
- Conditional effects based on game state
- Delayed/triggered effects system

#### 3.3 Historical Requirements
- Cards requiring specific event history
- Event-based requirement validation
- Complex conditional logic

### Phase 4: Frontend Event Processing (Week 4)

#### 4.1 Event Stream Processing
**Files to Create/Modify:**
- `frontend/src/services/eventProcessor.ts` - Event processing logic (NO GAME LOGIC)
- `frontend/src/hooks/useGameEvents.ts` - Event subscription hooks
- `frontend/src/types/events.ts` - Auto-generated event type definitions (READ-ONLY)

**Features:**
- Real-time event stream processing
- Local state updates from events
- Event validation and error handling

**DRY Compliance:**
- ❌ **Frontend does NOT**: Validate card requirements, calculate costs, or determine legal moves
- ❌ **Frontend does NOT**: Check if cards are playable or calculate availability
- ❌ **Frontend does NOT**: Implement any Terraforming Mars game rules
- ✅ **Frontend ONLY**: Displays events, renders UI, handles user input
- ✅ **All game logic**: Remains exclusively in Go backend services
- ✅ **Backend provides**: Complete list of available actions and playable cards

#### 4.2 Phase-Aware UI System
**Files to Create:**
- `frontend/src/components/phases/PhaseRenderer.tsx` - Dynamic phase UI
- `frontend/src/components/phases/private/` - Private phase components
- `frontend/src/components/phases/public/` - Public phase components

**UI Components:**
```tsx
const PhaseRenderer = ({ phase, gameState }) => {
    switch(phase.type) {
        case 'card_selection':
            return <CardSelectionPhase data={phase.data} />
        case 'player_action':
            return <ActionPhase data={phase.data} />
        default:
            return <DefaultGameView />
    }
}
```

#### 4.3 Private Data Handling
- Clear separation of public/private UI
- Secure handling of sensitive data
- Privacy-aware component rendering

### Phase 5: Optimization & Polish (Week 5)

#### 5.1 Performance Optimization
- Event processing performance tuning
- State snapshot system for fast reconnection
- Memory optimization for event storage
- Database indexing for event queries

#### 5.2 Testing & Validation
**Test Files to Create:**
- `test/events/store_test.go` - Event store testing
- `test/cards/microfunction_test.go` - Micro-function unit tests
- `test/phases/transition_test.go` - Phase transition testing
- `test/integration/card_flow_test.go` - End-to-end testing

#### 5.3 Documentation & Tooling
- Event type documentation
- Phase transition diagrams
- Development tooling for event inspection
- Performance monitoring

## Key Technical Decisions

### 1. Hybrid Event System
- **Events for audit trail**: Complete game action history
- **Snapshots for performance**: Fast state reconstruction
- **Best of both worlds**: Auditability + efficiency

### 2. Embedded Private Phases
- **Location**: Private phases in `currentPlayer` object
- **Benefit**: Natural separation of public/private data
- **Simplicity**: No complex data filtering logic needed

### 3. Micro-Function Architecture
- **Size limit**: 5-10 lines per function maximum
- **Single responsibility**: Each function does exactly one thing
- **Clear naming**: Function names describe exact purpose
- **Composability**: Functions combine into larger operations

### 4. Type-Safe Phases
- **Strong typing**: All phase data structures defined
- **Runtime validation**: Validate phase data integrity
- **Type generation**: Automatic TypeScript types from Go

### 5. Event-Driven Control Flow
- **Linear progression**: Predictable sequence of operations
- **Event boundaries**: Clear separation between operations
- **Error isolation**: Failures don't cascade unexpectedly

### 6. **DRY Architecture Principle - CRITICAL**
- **Single Source of Truth**: Game logic exists ONLY in Go backend
- **No Duplicate Logic**: Frontend never implements game rules or validation
- **Shared Types**: TypeScript types auto-generated from Go structs
- **Event Processing**: Frontend processes events but doesn't interpret game rules
- **Validation Boundary**: All game rule validation happens server-side only

**DRY Implementation Strategy:**
- ✅ **Backend**: All game logic, validation, and state management
- ✅ **Frontend**: UI rendering, user input, and event display only
- ✅ **Types**: Single Go definition → auto-generated TypeScript
- ✅ **Events**: Backend publishes, frontend subscribes (no game logic in frontend events)
- ❌ **Never**: Duplicate game rules in TypeScript
- ❌ **Never**: Client-side validation of game mechanics
- ❌ **Never**: Frontend calculating game state transitions

**Violation Prevention:**
- Code reviews must verify no game logic in frontend
- TypeScript types are read-only (generated, not hand-written)
- Frontend components are pure presentation layer
- Any game rule change happens in exactly one Go file

## Expected Benefits

### Maintainability
- **Tiny functions**: Easy to understand and test
- **Clear separation**: Public/private data boundaries
- **Event audit**: Complete action history for debugging

### Extensibility
- **New card types**: Easy to add with micro-function pattern
- **Complex interactions**: Event system handles dependencies
- **Phase flexibility**: New phases integrate cleanly

### Testability
- **Unit testing**: Each micro-function independently testable
- **Integration testing**: Event flow testing
- **Replay testing**: Reproduce bugs from event logs

### Debuggability
- **Event trail**: See exactly what happened and when
- **State reconstruction**: Replay events to debug issues
- **Performance monitoring**: Track event processing times

### User Experience
- **Real-time updates**: Immediate feedback from event system
- **Privacy**: Clean separation of public/private information
- **Responsiveness**: Optimized state updates

## Migration Strategy

### Gradual Implementation
1. **Phase 1**: Add event infrastructure without breaking existing code
2. **Phase 2**: Migrate card system piece by piece
3. **Phase 3**: Add advanced features incrementally
4. **Phase 4**: Frontend updates in parallel
5. **Phase 5**: Performance optimization and cleanup

### Backward Compatibility
- No need to think about it

### Risk Mitigation
- Comprehensive testing at each phase
- Rollback capabilities for each change
- Performance monitoring throughout migration

This roadmap transforms the existing solid foundation into a world-class event-sourced card game engine while maintaining clean architecture principles and ensuring the system remains maintainable and extensible for future development.