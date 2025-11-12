# Backend Restructuring Plan

## Vision

Restructure the backend from a **layer-based clean architecture** (model/service/repository/delivery) to a **stage-based + action-orchestrated architecture** organized by game flow and isolated mechanics.

## Current Problems

1. **Layer-based organization makes it hard to understand game flow**
   - Code for a single feature is scattered across multiple directories
   - Example: "Launch Asteroid" logic spans service/, repository/, delivery/, model/

2. **Tight coupling between services**
   - Services call other services directly
   - Risk of circular dependencies
   - Hard to test mechanics in isolation

3. **Unclear orchestration**
   - Business logic mixed with coordination logic
   - Hard to see the complete flow for a user action

4. **Difficult to navigate**
   - Finding code requires understanding technical layers, not game concepts
   - New developers need to learn the layer structure before understanding gameplay

## Target Architecture

### High-Level Structure

```
backend/internal/
├── lobby/              # Pre-game phase (game creation, joining, starting)
├── game/               # Active gameplay
│   ├── actions/        # User action orchestration (play card, place tile, etc.)
│   ├── cards/          # Card mechanic (isolated)
│   ├── tiles/          # Tile/board mechanic (isolated)
│   ├── resources/      # Resource mechanic (isolated)
│   ├── parameters/     # Global parameters mechanic (isolated)
│   ├── player/         # Player state mechanic (isolated)
│   └── turn/           # Turn management mechanic (isolated)
├── shared/             # Cross-stage infrastructure (events, websocket, dto)
└── http/               # REST API endpoints
```

### Key Principles

#### 1. Stage-Based Organization

Code organized by **when** it's used in the game lifecycle:

- **Lobby**: Game creation, player joining, host management, start game
- **Game**: Active gameplay after start button is clicked

#### 2. Actions = Orchestration Layer

Each user action is one file that **coordinates** across mechanics:

- `actions/play_card.go` - Orchestrates card validation, payment, effects, tile placement
- `actions/launch_asteroid.go` - Orchestrates payment, temperature increase, TR award
- `actions/place_tile.go` - Orchestrates placement validation, adjacency bonuses

**Actions CAN:**

- Call multiple mechanic services
- Read from multiple repositories
- Coordinate complex flows
- Handle complete user requests

#### 3. Mechanics = Isolated Domain Logic

Each mechanic is **completely self-contained** with its own:

- `service.go` - Business logic operations (top-level)
- `models.go` - Domain models specific to this mechanic
- `repository.go` - Data access for this mechanic's state

**Mechanics CANNOT:**

- Call other mechanic services
- Import other mechanic packages
- Directly orchestrate multi-mechanic flows

**Mechanics CAN:**

- Publish events to event bus
- Subscribe to events from event bus

#### 4. Event-Driven Coordination

For passive effects and cross-mechanic reactions:

- Mechanics publish domain events (TilePlaced, TemperatureChanged, etc.)
- Event bus broadcasts to subscribers
- Effect executor (in actions/) coordinates responses

### Dependency Rules

```
✅ ALLOWED:
   Actions → Mechanics         (orchestration)
   Mechanics → Event Bus       (notifications)
   Event Handlers → Mechanics  (coordination)

❌ FORBIDDEN:
   Mechanics → Mechanics       (creates coupling)
   Mechanics → Actions         (wrong direction)
```

## Example: Launch Asteroid Action

### Current (Layer-Based)

```
Find code across multiple layers:
- model/standard_projects.go - Cost definition
- service/standard_project_service.go - Business logic scattered
- service/player_service.go - TR update
- service/game_service.go - Parameter update
- repository/player_repository.go - State updates
- repository/game_repository.go - State updates
- delivery/websocket/handler/.../launch_asteroid/handler.go - WebSocket routing
```

### Target (Action-Based)

```go
// game/actions/launch_asteroid.go
func (a *LaunchAsteroidAction) Execute(gameID, playerID string) error {
    // 1. Validate and deduct credits
    if !a.resourcesSvc.CanAfford(playerID, Payment{Credits: 14}) {
        return ErrInsufficientCredits
    }
    a.resourcesSvc.DeductCredits(playerID, 14)

    // 2. Increase temperature
    a.parametersSvc.IncreaseTemperature(gameID, 2)

    // 3. Award TR
    a.playerSvc.IncreaseTR(playerID, 1)

    // 4. Broadcast
    a.sessionMgr.Broadcast(gameID)

    return nil
}
```

**Benefits:**

- ✅ Complete action in one file
- ✅ Clear orchestration flow
- ✅ Easy to understand
- ✅ Easy to test (mock mechanic services)

## Example: Isolated Mechanic

### Resources Mechanic

```
game/resources/
├── service.go       # Operations: AddCredits(), DeductPlants(), IncreaseProduction()
├── models.go        # Resources, Production, Payment
├── repository.go    # Player resource state
└── payment.go       # Payment calculation with steel/titanium substitution
```

**Key Points:**

- ✅ All resource logic in one directory
- ✅ No dependencies on cards/, tiles/, or parameters/
- ✅ Can be tested in complete isolation
- ✅ Can publish ResourcesChanged events
- ❌ Cannot call other mechanics directly

## Communication Patterns

### Synchronous (Actions → Mechanics)

```
User Action
    ↓
Action File (orchestration)
    ├→ Cards Service
    ├→ Resources Service
    └→ Tiles Service
```

### Asynchronous (Event-Driven)

```
Action places tile
    ↓
Tiles Service publishes TilePlaced event
    ↓
Event Bus broadcasts
    ↓
Cards passive effects subscribe
    ↓
Effect Executor (in actions/) coordinates
    ↓
Resources Service awards bonuses
```

## Benefits of New Architecture

### For Developers

✅ **Navigate by game concept**: Want to change asteroid mechanics? Go to `actions/launch_asteroid.go`

✅ **Understand flows easily**: One action file shows complete user action flow

✅ **Test in isolation**: Mechanics don't depend on each other, easy to unit test

✅ **Prevent circular dependencies**: Architecture enforces one-way dependencies

✅ **Parallel development**: Different mechanics can be worked on independently

### For Codebase

✅ **Clear boundaries**: Each mechanic owns its domain completely

✅ **Explicit coordination**: All orchestration in visible action files

✅ **Scalable**: New mechanics don't affect existing ones

✅ **Maintainable**: Changes to one mechanic don't break others

## Migration Strategy

### Phase 1: Infrastructure

- Create new directory structure
- Move shared infrastructure (events, websocket, dto)
- Update imports

### Phase 2: Lobby ✅ COMPLETE

- ✅ Extract lobby-specific logic from game_service
- ✅ Create lobby service (internal/lobby/service.go)
- ✅ Update all handlers to use lobby service
- ✅ All tests passing

### Phase 3: Extract Mechanics

For each mechanic (cards, tiles, resources, etc.):

1. Create mechanic folder
2. Extract domain logic into service.go
3. Move models to models.go
4. Move data access to repository.go
5. Remove cross-mechanic dependencies
6. Test in isolation

### Phase 4: Create Actions Layer

For each user action:

1. Create action file
2. Implement orchestration calling mechanics
3. Move WebSocket handler to route to action
4. Test complete flow

### Phase 5: Event-Driven Effects

1. Create effect executor in actions/
2. Update passive effects to publish events
3. Wire event handlers

### Phase 6: Tests & Cleanup

1. Reorganize tests to mirror new structure
2. Delete old layer-based directories
3. Update documentation

## File Organization Examples

### Action File Template

```go
// game/actions/[action_name].go
package actions

type [Action]Action struct {
    // Inject mechanic services
    cardsSvc      *cards.Service
    resourcesSvc  *resources.Service
    tilesSvc      *tiles.Service
    sessionMgr    *websocket.SessionManager
}

func (a *[Action]Action) Execute(...) error {
    // Orchestrate mechanics
    // Handle complete user flow
    // Broadcast result
}
```

### Mechanic Service Template

```go
// game/[mechanic]/service.go
package [mechanic]

type Service struct {
    repo     *Repository
    eventBus *events.Bus
}

// Pure operations - NO orchestration
func (s *Service) Operation1(...) error
func (s *Service) Operation2(...) error
```

## Success Criteria

### Architecture Goals

- ✅ Zero circular dependencies between mechanics
- ✅ One file per user action in actions/
- ✅ Complete mechanic isolation
- ✅ Clear orchestration vs domain logic separation

### Developer Experience

- ✅ Can find any feature in &lt;30 seconds
- ✅ Can understand complete action flow by reading one file
- ✅ Can test mechanics without complex setup
- ✅ Can add new mechanics without touching existing code

### Code Quality

- ✅ All imports compile successfully
- ✅ All tests pass
- ✅ No service-to-service calls (only action-to-service)
- ✅ Event-driven passive effects work correctly

## Estimated Effort

- Phase 1 (Infrastructure): 2-3 hours ✅ DONE
- Phase 2 (Lobby): 3-4 hours ✅ DONE
- Phase 3 (Mechanics): 10-14 hours
- Phase 4 (Actions): 12-16 hours
- Phase 5 (Events): 4-6 hours
- Phase 6 (Tests/Cleanup): 8-11 hours

**Total: 39-54 hours**

## Current Status

✅ **Phase 1 Completed**: New directory structure created and shared infrastructure moved

✅ **Phase 2 Completed**: Lobby service extracted from GameService
   - Created `internal/lobby/service.go` with all pre-game operations
   - GameService now focuses on active gameplay only
   - All HTTP and WebSocket handlers updated
   - All tests passing (cards, events, integration, logger, model, repository, service)
   - Repository design verified (no cross-dependencies, ID-based linking)

✅ **Phase 3 Completed**: Extracted all game mechanics into isolated modules
   - Created Resources mechanic (internal/game/resources/)
   - Created Global Parameters mechanic (internal/game/parameters/)
   - Created Tiles mechanic (internal/game/tiles/)
   - Created Turn mechanic (internal/game/turn/)
   - Created Production mechanic (internal/game/production/)
   - Refactored GameService to use mechanics (reduced ~300 lines)
   - All mechanics have comprehensive test coverage (44 new tests)
   - All tests passing across entire codebase

📋 **Next Steps**:

1. Phase 4: Create actions orchestration layer
2. Phase 5: Implement event-driven effects system
3. Phase 6: Tests & cleanup

## References

- Original architecture: `backend/CLAUDE.md`
- Event system documentation: `backend/docs/EVENT_SYSTEM.md`
- Full planning discussion: See conversation history

---

**Last Updated**: 2025-11-12
**Status**: Phase 3 complete - All game mechanics extracted ✅
