# Session-Based Architecture Migration

## Overview

The backend is undergoing a migration from **global repositories** to **session-scoped repositories**. This migration aims to simplify repository complexity, improve testability, and enable better event-driven architecture.

## Current State: Dual Architecture

The system currently runs **both architectures in parallel**:

- **OLD Architecture**: Global repositories in `/internal/repository/` (still used by most services)
- **NEW Architecture**: Session-scoped repositories in `/internal/session/game/` (used by actions and SessionManager)

---

## Architecture Comparison

### OLD Architecture (Legacy - To Be Removed)

**Location**: `/internal/repository/`

```
repository/
├── game_repository.go          # Global game storage
├── player_repository.go        # Global player storage
├── card_repository.go          # Card definitions (kept)
└── card_deck_repository.go     # Deck management
```

**Characteristics**:
- ❌ Global singleton storage (one instance for all games)
- ✅ Feature-complete (supports all game mechanics)
- ❌ Complex models with many fields
- ❌ Tightly coupled to services
- 🟡 Still used by most services and handlers

### NEW Architecture (Target - In Progress)

**Location**: `/internal/session/game/`

```
session/
├── session_manager.go          # Broadcasts game state via WebSocket
└── game/
    ├── models.go              # Simplified Game entity
    ├── repository.go          # Session-scoped game storage
    ├── player/
    │   ├── models.go          # Simplified Player entity
    │   └── repository.go      # Session-scoped player storage
    ├── card/
    │   ├── models.go          # Card entity
    │   └── repository.go      # Card lookups
    └── deck/
        ├── models.go          # Deck entity
        ├── repository.go      # Deck management
        └── loader.go          # Card loading logic
```

**Characteristics**:
- ✅ Session-scoped storage (data organized by gameID)
- ✅ Event-driven (publishes domain events)
- ✅ Immutable interface (returns values, not pointers)
- ✅ Granular update methods
- ✅ Simplified models focused on core data
- 🟡 Used by actions and SessionManager only

---

## Migration Progress

### ✅ Phase 1: Core Infrastructure (COMPLETE)

**Actions Pattern**
- ✅ `CreateGameAction` - Uses NEW game repository
- ✅ `JoinGameAction` - Uses NEW game + player repositories
- ✅ `StartGameAction` - Uses NEW game + player + deck repositories
- ✅ `SelectStartingCardsAction` - Uses NEW repositories

**Session-Scoped Repositories**
- ✅ `game.Repository` - Core game CRUD operations
- ✅ `player.Repository` - Core player CRUD operations
- ✅ `card.Repository` - Card lookups and management
- ✅ `deck.Repository` - Deck management with card loading

**SessionManager**
- ✅ Injected with NEW repositories
- ✅ Type conversion layer (NEW → OLD for DTO compatibility)
- ✅ WebSocket broadcasting

**HTTP Endpoints**
- ✅ `POST /games` - Uses `CreateGameAction` (NEW)
- ❌ `GET /games/:id` - Still uses OLD `GameService`

**WebSocket Handlers**
- ✅ `ConnectionHandler` - Uses `JoinGameAction` (NEW)
- ✅ `StartGameHandler` - Uses `StartGameAction` (NEW)
- ✅ `SelectStartingCardHandler` - Uses NEW actions

**Services**
- ✅ `CardService` - **MIGRATED** to NEW repositories (commit d53d951)

---

### ⚠️ Phase 2: Model Parity (IN PROGRESS)

**NEW Player Model Missing Fields**:
```go
// Currently stubbed in converters - need to add to player.Player:
PlayedCards               []string
Passed                    bool
AvailableActions          int
VictoryPoints             int
PendingTileSelection      *PendingTileSelection
PendingCardSelection      *PendingCardSelection
TileQueue                 []TilePlacement
```

**Impact**: Type converters currently hardcode empty values, which may cause:
- Incomplete game state in frontend
- Missing UI elements
- Broken game flow

**Next Steps**:
1. Add missing fields to `internal/session/game/player/models.go`
2. Update `player.Repository` methods to handle new fields
3. Remove hardcoded stubs in `session_manager.go` converters
4. Test complete type parity between OLD and NEW models

---

### ❌ Phase 3: Service Layer Migration (NOT STARTED)

**Services Still Using OLD Repositories**:

1. **GameService** (`internal/service/game_service.go`)
   - Most methods use OLD repositories
   - Only `CreateGame` delegates to action
   - Needs: Migrate all methods to use NEW repos or delegate to actions

2. **PlayerService** (`internal/service/player_service.go`)
   - Tile placement logic
   - Resource management
   - Turn progression
   - Needs: Full migration to NEW repositories

3. **StandardProjectService** (`internal/service/standard_project_service.go`)
   - Standard project execution
   - Needs: Migration to NEW repositories

4. **ResourceConversionService** (`internal/service/resource_conversion_service.go`)
   - Heat → Temperature
   - Plants → Greenery
   - Needs: Migration to NEW repositories

5. **BoardService** (`internal/service/board_service.go`)
   - Tile logic and validation
   - Needs: Migration to NEW repositories

6. **TileService** (`internal/service/tile_service.go`)
   - Tile queue management
   - Needs: Migration to NEW repositories

---

### ❌ Phase 4: WebSocket Handler Migration (NOT STARTED)

**Handlers Still Using OLD Services**:

- ❌ `PlayCardHandler` - Uses OLD `CardService` (now migrated, but handler not updated)
- ❌ `TileSelectionHandler` - Uses OLD `PlayerService`
- ❌ Standard project handlers - Use OLD `StandardProjectService`
- ❌ Resource conversion handlers - Use OLD `ResourceConversionService`
- ❌ Other gameplay handlers - All use OLD service layer

**Next Steps**:
1. Convert handlers to use action pattern
2. Remove direct service calls
3. Use SessionManager for all broadcasting

---

### ❌ Phase 5: Complex Features (NOT STARTED)

**Advanced Player Features Not Yet in NEW Models**:
- Tile selection queues
- Card selection modals
- Production phase management
- Victory points calculation
- Player effects system
- Forced action system (corporation abilities)

**Card System Integration**:
- Card effect subscriber (passive effects)
- Card validation with NEW repos
- Event-driven card effects

---

## Type Conversion Layer

The system maintains **backward compatibility** during migration through converters in `session_manager.go`:

### Converter Functions

```go
// Convert NEW session-scoped types → OLD global types → DTOs
gameToModel(g *game.Game) model.Game
playersToModel(players []*player.Player) []model.Player
playerToModel(p *player.Player) model.Player
cardsToModel(cards map[string]card.Card) map[string]model.Card
```

### Why Converters Exist

1. **DTO Layer Compatibility**: Frontend DTOs (`internal/delivery/dto/`) expect OLD `model.*` types
2. **Phased Migration**: Allows gradual transition without breaking existing code
3. **SessionManager Flow**: Retrieves from NEW repos → Converts to OLD types → Creates DTOs

### When Converters Will Be Removed

After all services and handlers migrate to NEW repositories:
1. Update DTO layer to use NEW types directly
2. Remove all converter functions
3. Delete OLD repository implementations

---

## Migration Goals

### End State Vision

**Session-Scoped Repositories Only**:
```
internal/session/game/
├── repository.go          # Game data (session-scoped)
├── player/
│   └── repository.go      # Player data (session-scoped)
├── card/
│   └── repository.go      # Card lookups (session-scoped)
└── deck/
    └── repository.go      # Deck management (session-scoped)
```

**Action Pattern Everywhere**:
- All game operations use action pattern
- Actions use NEW repositories directly
- Actions are testable and reusable

**Simplified Service Layer**:
- Services become thin orchestrators
- Services coordinate actions
- Services handle cross-cutting concerns
- Services call SessionManager for broadcasting

**No Global State**:
- ❌ Delete `internal/repository/game_repository.go`
- ❌ Delete `internal/repository/player_repository.go`
- ❌ Delete `internal/repository/card_deck_repository.go`
- ✅ Keep `internal/repository/card_repository.go` (reference data only)

---

## Current Issue: Game Creation

### Symptoms
Game creation via HTTP endpoint may fail or produce incomplete data.

### Root Causes

1. **Incomplete Type Conversion**
   - NEW `player.Player` lacks fields like `PlayedCards`, `Passed`, etc.
   - Converters stub missing fields with empty values
   - Frontend receives incomplete player data
   - UI may break or display incorrectly

2. **Race Conditions**
   - Player connects → Hub registers
   - JoinGameAction executes → Creates player in NEW repo
   - SessionManager broadcasts → Uses NEW repos
   - Potential: Broadcast happens before player fully set up

3. **DTO Mismatch**
   - Frontend expects certain fields
   - NEW models missing those fields
   - DTOs contain null/empty values
   - Frontend validation fails

### Debugging Steps

1. **Check Backend Logs**:
   ```
   🎮 Creating new game
   ✅ Game created successfully
   🚀 Broadcasting game state
   ```

2. **Inspect HTTP Response**:
   - Look for missing/null fields in game DTO
   - Verify player data completeness

3. **Verify Repository State**:
   - Check NEW `game.Repository` has the game
   - Confirm player data exists in NEW `player.Repository`

4. **Test WebSocket Flow**:
   - Join game via WebSocket after HTTP creation
   - Verify `JoinGameAction` succeeds
   - Check SessionManager broadcasting
   - Confirm frontend receives complete state

5. **Examine Type Converters**:
   - Review `gameToModel()` and `playerToModel()` in `session_manager.go`
   - Identify hardcoded empty values
   - Check for missing field mappings

---

## Migration Strategy

### Immediate Next Steps (Fix Game Creation)

1. **Complete NEW Player Model**:
   - Add all missing fields to `internal/session/game/player/models.go`
   - Update `player.Repository` methods to handle new fields
   - Remove converter stubs in `session_manager.go`

2. **Test Type Parity**:
   - Ensure NEW models have feature parity with OLD models
   - Verify converters produce valid DTOs
   - Test game creation end-to-end

3. **Fix Any Remaining Issues**:
   - Debug specific game creation failures
   - Address race conditions if found
   - Validate frontend integration

### Phase-by-Phase Migration Plan

**Phase 1: Complete Core Models** ✅ (Mostly Done)
- Add missing fields to NEW models
- Ensure type parity with OLD models
- Remove all converter stubs

**Phase 2: Migrate HTTP Endpoints** (Next)
- Migrate `GET /games/:id` to use NEW repos or actions
- Convert other endpoints as needed
- Remove OLD service dependencies from HTTP layer

**Phase 3: Migrate Services** (Sequential)
1. ✅ CardService (already done)
2. GameService → Use NEW repos or delegate to actions
3. PlayerService → Migrate tile placement, resources
4. StandardProjectService → Migrate to NEW repos
5. Other services as needed

**Phase 4: Migrate WebSocket Handlers**
- Convert all gameplay handlers to action pattern
- Remove direct service calls
- Ensure all use SessionManager for broadcasting

**Phase 5: Remove OLD Architecture**
- Delete OLD repository files
- Update all imports to NEW repos
- Remove type converters
- Update tests

**Phase 6: Advanced Features**
- Migrate complex player features (tile queues, selections)
- Migrate card effect system
- Migrate production phase management
- Migrate victory point calculations

---

## Key Design Principles

### Session-Scoped Repository Benefits

1. **Simpler Data Organization**:
   - Data naturally scoped to games
   - No need for global filtering by gameID
   - Clearer data ownership

2. **Better Testability**:
   - Easy to create isolated game sessions for tests
   - No global state pollution between tests
   - Clearer test setup/teardown

3. **Improved Event Handling**:
   - Events naturally scoped to game sessions
   - Easier to track event subscriptions
   - Better event lifecycle management

4. **Reduced Complexity**:
   - Repositories focus on single-game concerns
   - No global state synchronization
   - Simpler concurrency handling

### Action Pattern Benefits

1. **Encapsulated Business Logic**:
   - Each action is self-contained
   - Clear input/output contracts
   - Easy to test in isolation

2. **Reusability**:
   - Actions can be called from HTTP or WebSocket handlers
   - Consistent behavior across entry points

3. **Composability**:
   - Actions can be composed into complex workflows
   - Services orchestrate actions

4. **Testability**:
   - Actions are pure functions of repositories
   - Easy to mock dependencies
   - Clear test scenarios

---

## File Reference

### NEW Architecture Files

**Core**:
- `/internal/session/session_manager.go` - Broadcasts game state
- `/internal/session/game/repository.go` - Game repository
- `/internal/session/game/player/repository.go` - Player repository
- `/internal/session/game/card/repository.go` - Card repository
- `/internal/session/game/deck/repository.go` - Deck repository

**Actions**:
- `/internal/action/create_game.go` - Game creation
- `/internal/action/join_game.go` - Join game
- `/internal/action/start_game.go` - Start game
- `/internal/action/select_starting_cards.go` - Card selection

**Models**:
- `/internal/session/game/models.go` - Game entity
- `/internal/session/game/player/models.go` - Player entity
- `/internal/session/game/card/models.go` - Card entity
- `/internal/session/game/deck/models.go` - Deck entity

### OLD Architecture Files (To Be Removed)

- `/internal/repository/game_repository.go` - OLD game repo
- `/internal/repository/player_repository.go` - OLD player repo
- `/internal/repository/card_deck_repository.go` - OLD deck repo
- `/internal/service/game_service.go` - Still uses OLD repos (mostly)
- `/internal/service/player_service.go` - Uses OLD repos
- `/internal/service/standard_project_service.go` - Uses OLD repos

### Integration Points

- `/cmd/server/main.go` - Dependency injection setup
- `/internal/delivery/http/game_handler.go` - HTTP endpoints
- `/internal/delivery/websocket/registry.go` - WebSocket handler registration

---

## Summary

The backend is mid-migration from global to session-scoped repositories. The system runs both architectures in parallel:

- ✅ **Actions + SessionManager**: Use NEW session-scoped repos
- ❌ **Services + handlers**: Still use OLD global repos (except CardService)
- 🔄 **Type converters**: Bridge NEW ↔ OLD types

**Current blocker**: Incomplete NEW player model causes game creation issues due to missing fields in type conversion.

**Next milestone**: Complete NEW model types to achieve feature parity with OLD models, then systematically migrate services and handlers.
