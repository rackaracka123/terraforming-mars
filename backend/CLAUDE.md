Read backend/go.instructions.md

# Backend - Terraforming Mars API Server

This document provides guidance for working with the backend API server.

## Overview

Go-based REST and WebSocket API server implementing the Terraforming Mars board game logic. Provides real-time multiplayer game state synchronization and enforces game rules.

## Go Coding Standards

**IMPORTANT**: This backend follows idiomatic Go practices and community standards. For comprehensive Go coding guidelines, see **[go.instructions.md](./go.instructions.md)**.

Key standards include:

- Follow Effective Go, Go Code Review Comments, and Google's Go Style Guide
- Write simple, clear, and idiomatic Go code
- Use proper naming conventions (mixedCaps, avoid underscores)
- Check all errors immediately
- Keep the happy path left-aligned (minimize indentation, return early)
- Document all exported symbols
- Use `gofmt` and `goimports` for formatting
- **CRITICAL**: Each `.go` file must have exactly ONE `package` declaration

For detailed guidance on naming, error handling, concurrency, API design, testing, and more, consult `go.instructions.md`.

## Server Restart Policy

**CRITICAL**: NEVER restart the backend server yourself. ALWAYS ask the user if you think a restart is needed.

- **Normal Mode**: User runs `make backend` or `make run` with **Air hot reload** - server automatically restarts on code changes
- **Watch Mode Active**: Code changes (Go files, JSON assets) trigger instant automatic reload
- **No Manual Restarts**: You should NEVER execute restart commands
- **If Restart Seems Needed**: Ask user "Should I restart the backend?" (they'll confirm or explain why it's not needed)

The user's development environment handles all server lifecycle management. Your role is to write code, not manage processes.

## Architecture

### Clean Architecture Layers

The backend follows clean architecture principles with strict separation of concerns:

**Domain Layer** (`internal/game/`)

- Core business entities: Game (containing all state), Player, GlobalParameters, Board, Deck
- Subpackages: player/, board/, deck/, shared/, global_parameters/
- Value objects in shared/: Resources, Production, Tile, HexPosition
- Domain events defined in `internal/events/`
- Private fields with public accessor methods
- Zero external dependencies

**Action Layer** (`internal/action/`)

- Single-responsibility actions executing business logic (~100-200 lines each)
- BaseAction provides common dependencies (GameRepository, CardRegistry, logger)
- Main actions modify game state (JoinGameAction, PlayCardAction)
- Query actions for read operations (GetGameAction, ListGamesAction)
- Admin actions for game management (SetResourcesAction)
- Depends on domain types via GameRepository

**Infrastructure Layer** (`internal/game/`)

- GameRepository manages collection of active games
- Game contains all state: Players, Board, Deck, GlobalParameters
- State methods publish events via injected EventBus
- Private fields enforce encapsulation
- No separate subdomain repositories - all accessed via Game

**Presentation Layer** (`internal/delivery/`)

- HTTP endpoints delegate to actions (`http/`)
- WebSocket handlers delegate to actions (`websocket/`)
- DTOs for external communication (`dto/`)
- Request/response mapping
- Depends on action layer, not repositories directly

**Card System** (`internal/cards/`)

- Centralized card registry and lookup
- Card validation for requirements and plays
- Card effect implementations
- Modular effect handlers

**Event System** (`internal/events/`)

- Type-safe event bus for pub/sub
- Domain event definitions (TemperatureChanged, ResourcesChanged, TilePlaced, etc.)
- CardEffectSubscriber for passive card effects
- Event-driven architecture decoupling actions from effects

### Directory Structure

```
backend/
├── cmd/                    # Application entry points
│   ├── server/            # Main server with dependency injection
│   └── watch/             # Development file watching
├── internal/              # Private application code
│   ├── action/            # Action layer - single-responsibility business logic
│   │   ├── base.go        # BaseAction with common dependencies
│   │   ├── query/         # Query actions for reads
│   │   └── admin/         # Admin actions
│   ├── game/              # Game state repository and domain types
│   │   ├── game.go        # Core Game type with all game state
│   │   ├── repository.go  # GameRepository interface and implementation
│   │   ├── player/        # Player entity and components
│   │   ├── board/         # Board and Tile types
│   │   ├── deck/          # Deck management
│   │   ├── shared/        # Shared types (Resources, HexPosition, etc.)
│   │   └── global_parameters/  # GlobalParameters with terraforming state
│   ├── cards/             # Card system and registry
│   ├── delivery/          # Presentation layer (HTTP, WebSocket, DTOs)
│   │   └── websocket/     # Includes Broadcaster for event-driven updates
│   ├── events/            # Event bus and domain events
│   ├── initialization/    # Application bootstrap and card loading
│   └── logger/            # Structured logging
├── pkg/                   # Public packages
│   └── typegen/           # TypeScript type generation
├── test/                  # Test suite (mirrors internal/ structure)
├── tools/                 # DEPRECATED: Card parser tool (being removed)
├── assets/                # Static game data (JSON card definitions)
└── docs/                  # Documentation
    └── swagger/           # Auto-generated API docs
```

## Development Workflow

### Running the Server

```bash
# From project root
make backend              # Hot reload via Air (port 3001)
make run                  # Run both frontend and backend

# Direct commands (from backend/)
go run cmd/server/main.go
air                       # Hot reload with Air
```

### Testing

```bash
# From project root
make test                 # Run all backend tests
make test-verbose         # Detailed test output
make test-coverage        # Generate coverage report
make test-quick           # Fast iteration tests

# From backend/
go test ./test/...        # All tests
go test ./test/action/    # Specific package
go test -json ./test/...  # JSON output for parsing
```

**Test Location**: Tests live in `test/` directory, mirroring `internal/` structure. Example: `test/action/confirm_production_cards_test.go` tests `internal/action/confirm_production_cards.go`.

### Code Quality

```bash
# From project root
make lint-backend         # Run Go formatting
make format               # Format all code

# From backend/
make format               # Run gofmt
go fmt ./...              # Direct formatting
```

### Type Generation

Generate TypeScript types for frontend consumption:

```bash
# From project root
make generate             # Generate types from Go structs

# From backend/
tygo generate             # Direct tygo command
```

Add `ts:` tags to structs for type generation:

```go
type Player struct {
    ID       string `json:"id" ts:"string"`
    Credits  int    `json:"credits" ts:"number"`
}
```

## Key Development Patterns

### Adding New Game Operations

1. **Create action** in `internal/action/`:
   - Extend `BaseAction` struct
   - Implement `Execute()` method with clear parameters
   - Validate inputs and call session repositories
   - Return explicit result type or error

2. **Create WebSocket handler** (if needed) in `internal/delivery/websocket/handler/`:
   - Parse incoming WebSocket message
   - Call the action's Execute() method
   - SessionManager handles broadcasting

3. **Create HTTP handler** (if needed) in `internal/delivery/http/`:
   - Parse HTTP request
   - Call the action's Execute() method
   - Map result to DTO and respond

4. **Add message/request types** to frontend types if needed

### Implementing Card Effects

**For passive effects** (event-driven):

1. Define behavior in card JSON with triggers and outputs
2. Ensure Game state methods publish relevant domain events
3. CardEffectSubscriber automatically subscribes on card play
4. No manual action code needed for passive effects

See `docs/EVENT_SYSTEM.md` for complete event system documentation.

**For immediate effects**:

1. Implement logic in card effect handler
2. Call via action when card is played
3. Action updates Game via state methods
4. Game publishes events, Broadcaster sends updates to clients

### Game Repository Pattern

- **Single Source of Truth**: Game contains all state (Players, Board, Deck, GlobalParameters)
- **Encapsulation**: Private fields with public accessor methods
- **Event Integration**: State methods automatically publish domain events
- **GameRepository**: Manages collection of active Game instances
- **Access Pattern**: `game := gameRepo.Get(gameID)` → `player := game.GetPlayer(playerID)`

### Action Layer Rules

- **Single Responsibility**: Each action performs ONE operation (~100-200 lines)
- **Extend BaseAction**: Use common dependencies (GameRepository, CardRegistry, logger)
- **Actions do only what they say**: Don't manually check for passive card effects
- **Call Game state methods**: Game methods publish events automatically
- **Broadcaster handles WebSocket updates**: Subscribes to BroadcastEvent
- **Event system handles passive effects**: CardEffectSubscriber triggers effects via events

### State Ownership and Encapsulation

The architecture follows clear ownership boundaries for game state:

**Game Repository Owns:**

- Game-wide state (status, phase, generation, current turn)
- Player-specific phase state (ProductionPhase, SelectStartingCardsPhase, PendingCardSelection, PendingCardDrawSelection, PendingTileSelection, PendingTileSelectionQueue, ForcedFirstAction)
- Global parameters (temperature, oxygen, oceans)
- Game configuration and settings

**Player Repository Owns:**

- Player identity (ID, name, gameID)
- Corporation selection
- Cards (hand and played cards)
- Resources, production, terraform rating, victory points
- Turn state (passed, available actions, connection status)
- Player effects, actions, requirement modifiers

**Why Phase State Lives in Game:**

- Phase state is transient - exists only during specific game phases
- Game controls phase transitions and needs atomic access to all players' phase states
- Cleaner separation: Player represents persistent player state, Game manages workflow state
- Simplifies phase transition logic (e.g., checking if all players completed starting selection)

**Access Pattern:**

```go
// ✅ CORRECT: Access phase state via Game
game, _ := gameRepo.Get(gameID)
productionPhase := game.GetProductionPhase(playerID)
game.SetProductionPhase(ctx, playerID, phase)

// ❌ WRONG: Phase state not on Player
player := game.GetPlayer(playerID)
productionPhase := player.ProductionPhase() // This method doesn't exist
```

## Data Flow

### WebSocket Message Flow

```
Client → WebSocket Connection → Hub.HandleMessage()
                                       ↓
                                 Manager.RouteMessage()
                                       ↓
                             WebSocket Handler.Handle()
                                       ↓
                                  Action.Execute()
                                       ↓
                            Game State Updates + Events
                                       ↓
                           EventBus → Broadcaster
                                       ↓
                              All Clients Updated
```

### Game State Synchronization

1. Action performs business logic via Execute() method
2. Game state methods update state and publish events
3. EventBus notifies subscribers (Broadcaster, passive effects, etc.)
4. Broadcaster automatically sends updates on BroadcastEvent
5. Broadcaster fetches complete game state from GameRepository
6. Clients receive personalized state update via WebSocket

## Type System Integration

### Go to TypeScript

All domain types with `ts:` tags generate TypeScript interfaces:

```go
// backend/internal/game/player/player.go
type Player struct {
    ID           string    `json:"id" ts:"string"`
    Credits      int       `json:"credits" ts:"number"`
    Corporation  *Corporation `json:"corporation" ts:"Corporation | null"`
}
```

Generates:

```typescript
// frontend/src/types/generated/api-types.ts
export interface Player {
    id: string;
    credits: number;
    corporation: Corporation | null;
}
```

### Keeping Types in Sync

1. Modify Go structs in `internal/game/` or subpackages
2. Run `make generate` from project root
3. Check corresponding DTOs in `internal/delivery/dto/`
4. Update DTO mappers in `internal/delivery/dto/mapper.go` if needed
5. Frontend automatically gets updated types

## Important Notes

### Event-Driven Architecture

**CRITICAL**: Actions should NOT manually check for passive card effects. The event system handles this automatically:

```go
// ✅ CORRECT
func (a *ConvertHeatToTemperatureAction) Execute(...) {
    game.GlobalParameters().IncreaseTemperature(ctx, steps)  // Publishes TemperatureChangedEvent
    // CardEffectSubscriber automatically triggers passive effects
    return result, nil
}

// ❌ WRONG
func (a *ConvertHeatToTemperatureAction) Execute(...) {
    game.GlobalParameters().IncreaseTemperature(ctx, steps)
    // Don't manually loop through cards to check effects
    for _, card := range player.PlayedCards().Cards() { ... }
    return result, nil
}
```

### State Management

- No timeouts or arbitrary delays
- Implement proper event handling
- Design deterministic state transitions
- Use proper synchronization when needed

### Logging Guidelines

- Use emojis for visual distinction
- Include direction indicators (client→server, server→client)
- 🔗 for connect, ⛓️‍💥 for disconnect
- 📢 for broadcasts, 💬 for direct messages
- 📡 for HTTP requests
- 🚀 for startup, 🛑 for shutdown

## Testing Guidelines

### Test Organization

- Tests mirror `internal/` structure in `test/`
- Use table-driven tests for multiple scenarios
- Mock external dependencies via interfaces
- Test business logic in isolation

### Test Patterns

```go
func TestPlayerService_DoAction(t *testing.T) {
    tests := []struct {
        name     string
        input    Input
        expected Expected
        wantErr  bool
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Common Tasks

### Adding a New Domain Type

1. Create struct in `internal/game/` or subpackage with `json:` and `ts:` tags
2. Add private fields with public accessor methods
3. Add Game methods to access/modify the new type
4. Create action in `internal/action/` extending BaseAction
5. Add DTOs in `internal/delivery/dto/`
6. Add mappers in `internal/delivery/dto/mapper.go`
7. Run `make generate` to sync TypeScript types

### Adding a New Game Rule

1. Check `TERRAFORMING_MARS_RULES.md` in project root
2. Define types in `internal/game/` or subpackages
3. Create action in `internal/action/` with validation logic
4. Update Game methods if new state access needed
5. Create tests in `test/action/`
6. Update relevant HTTP or WebSocket handlers to call action

### Debugging

- Use structured logger for consistent output
- Check `EventBus` for event flow
- Inspect WebSocket messages in browser DevTools
- Use `go test -json` for parseable test output
- Review `docs/swagger/` for API contract

## Dependencies

### Core Libraries

- **gorilla/websocket**: WebSocket communication
- **go-chi/chi**: HTTP routing and middleware
- **swaggo/swag**: API documentation generation
- **tygo**: TypeScript type generation

### Development Tools

- **Air**: Hot reload for development
- **golangci-lint**: Code quality checks

## Related Documentation

- **Project Root CLAUDE.md**: Full-stack architecture and workflows
- **frontend/CLAUDE.md**: Frontend architecture and patterns
- **docs/ARCHITECTURE_REFACTOR_GOAL.md**: Backend architecture and refactor history
- **docs/EVENT_SYSTEM.md**: Event-driven architecture and broadcasting
- **TERRAFORMING_MARS_RULES.md**: Complete game rules reference
- **assets/CLAUDE.md**: Card database documentation (behavior types, output formats)
- **assets/terraforming_mars_cards.json**: Authoritative card definitions (manually edited)
