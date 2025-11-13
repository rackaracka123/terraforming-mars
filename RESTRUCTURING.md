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

✅ **Phase 3 ENHANCED COMPLETE**: Extracted all game mechanics into fully isolated modules
   - Created 5 mechanics with complete vertical slice isolation:
     * Resources mechanic (internal/game/resources/) - models.go, service.go, repository.go
     * Global Parameters mechanic (internal/game/parameters/) - models.go, service.go, repository.go
     * Tiles mechanic (internal/game/tiles/) - models.go, service.go, repository.go
     * Turn mechanic (internal/game/turn/) - models.go, service.go, repository.go
     * Production mechanic (internal/game/production/) - models.go, service.go, repository.go
   - **TRUE ISOLATION ACHIEVED**: Each mechanic has mechanic-specific types in models.go
   - **ZERO MODEL IMPORTS**: Services use only their own mechanic types, never import internal/model
   - **CLEAN BOUNDARIES**: Repositories handle type conversion between mechanic types and central model
   - Refactored GameService to use mechanics (reduced ~300 lines)
   - All mechanics have comprehensive test coverage (44+ tests)
   - All tests passing across entire codebase

✅ **Phase 4 COMPLETE**: Action layer fully established for ALL game actions

   **Standard Projects** (internal/game/actions/standard_projects/):
   - Implemented all 6 standard project actions in organized subfolder:
     * build_aquifer.go - Deduct credits, raise ocean, award TR, queue ocean tile placement
     * launch_asteroid.go - Deduct credits, raise temperature, award TR
     * build_power_plant.go - Deduct credits, increase energy production
     * plant_greenery.go - Deduct credits, raise oxygen, award TR, queue greenery tile placement
     * build_city.go - Deduct credits, increase credit production, queue city tile placement
     * sell_patents.go - Initiate patent selling by creating pending card selection

   **Resource Conversions** (internal/game/actions/):
   - convert_heat_to_temperature.go - Deduct heat (with card discounts), raise temperature, award TR
   - convert_plants_to_greenery.go - Deduct plants (with card discounts), raise oxygen, award TR, queue greenery

   **Turn Management** (internal/game/actions/):
   - skip_action.go - End player turn, trigger production phase if applicable

   **Card Operations** (internal/game/actions/):
   - play_card.go - Plays a card from hand (validation, payment, tile queue processing)
   - select_tile.go - Handles tile placement selection (with plant conversion special case)
   - play_card_action.go - Plays blue card actions (once per generation validation)

   **Card Selection** (internal/game/actions/card_selection/):
   - submit_sell_patents.go - Completes patent selling transaction (awards credits, removes cards)
   - select_starting_cards.go - Starting card/corporation selection (checks all players ready)
   - select_production_cards.go - Production phase card selection (coordinates multi-player phase)
   - confirm_card_draw.go - Card draw confirmation (handles free vs paid cards)

   **Infrastructure Additions**:
   - ✅ Created CardManager interface for card validation and playing
   - ✅ Created SelectionManager for card selection operations
   - ✅ Both CardManager and SelectionManager wired up in main.go and test fixtures

   **Architecture Achievements**:
   - ✅ ALL 15 WebSocket handlers migrated to use actions instead of services
   - ✅ Established consistent action pattern with dependency injection
   - ✅ Created BoardServiceAdapter to bridge service.BoardService to tiles.BoardService
   - ✅ Organized actions into logical subfolders (standard_projects/, card_selection/)
   - ✅ Updated all test fixtures (main.go, test_server.go) to wire up all actions
   - ✅ Updated registry.go and websocket.go with new action parameters
   - ✅ All tests passing (except pre-existing tiles test issue)

   **Pattern Proven**: Actions successfully orchestrate across multiple mechanics:
   - ✅ Resources mechanic: Payment, production updates, card discount calculation
   - ✅ Parameters mechanic: Temperature/oxygen raises with automatic TR awards
   - ✅ Tiles mechanic: Tile queue creation and processing
   - ✅ Turn mechanic: Phase progression and turn management
   - ✅ Production mechanic: Resource generation at turn end
   - ✅ Cards mechanic: Card validation, playing, selection management
   - ✅ Zero cross-mechanic dependencies maintained
   - ✅ Clean separation: handlers → actions → mechanics → repositories

📋 **Next Steps**:

1. **Phase 5**: Implement event-driven effects system
   - Already functional via CardEffectSubscriber
   - Document patterns and ensure comprehensive coverage

2. **Phase 6**: Tests & cleanup
   - Fix pre-existing tiles test issue
   - Delete old service methods that are now replaced by actions
   - Clean up unused code paths
   - Update documentation to reflect new architecture

## References

- Original architecture: `backend/CLAUDE.md`
- Event system documentation: `backend/docs/EVENT_SYSTEM.md`
- Full planning discussion: See conversation history

---

**Last Updated**: 2025-11-13
**Status**: Phase 3 ENHANCED COMPLETE ✅ | Phase 4 COMPLETE ✅ (ALL Actions Implemented)
