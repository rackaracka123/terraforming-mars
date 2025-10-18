# Backend - Terraforming Mars API Server

This document provides guidance for working with the backend API server.

## Overview

Go-based REST and WebSocket API server implementing the Terraforming Mars board game logic. Provides real-time multiplayer game state synchronization and enforces game rules.

## Architecture

### Clean Architecture Layers

The backend follows clean architecture principles with strict separation of concerns:

**Domain Layer** (`internal/model/`)
- Core business entities with identity (Game, Player, GlobalParameters, Card)
- Value objects defined by values (Resources, Production)
- Domain events for significant business occurrences
- Defensive copying via DeepCopy() methods
- Zero external dependencies

**Application Layer** (`internal/service/`)
- Use cases orchestrating business operations
- Domain services for complex multi-entity logic
- Event handlers reacting to domain events
- Interface definitions for infrastructure dependencies
- Depends only on domain layer

**Infrastructure Layer** (`internal/repository/`)
- In-memory storage of domain models
- Immutable getters returning values, not pointers
- Granular update methods for targeted state changes
- Event publishing via EventBus integration
- Clean relationships using ID references

**Presentation Layer** (`internal/delivery/`)
- HTTP endpoints with routing and middleware (`http/`)
- WebSocket real-time communication (`websocket/`)
- DTOs for external communication (`dto/`)
- Request/response mapping
- Depends on application layer only

**Card System** (`internal/cards/`)
- Centralized card registry and lookup
- Card validation for requirements and plays
- Card effect implementations
- Modular effect handlers

**Event System** (`internal/events/`)
- Type-safe event bus for pub/sub
- Domain event definitions
- CardEffectSubscriber for passive card effects
- Event-driven architecture decoupling services from effects

**Session Management** (`internal/delivery/websocket/session/`)
- SessionManager interface for broadcasting game state
- Broadcast(gameID) sends to all players in game
- Send(gameID, playerID) sends to specific player
- Uses repositories directly to avoid circular dependencies

### Directory Structure

```
backend/
├── cmd/                    # Application entry points
│   ├── server/            # Main server with dependency injection
│   └── watch/             # Development file watching
├── internal/              # Private application code
│   ├── cards/             # Card system and registry
│   ├── delivery/          # Presentation layer (HTTP, WebSocket, DTOs)
│   ├── events/            # Event bus and domain events
│   ├── initialization/    # Application bootstrap and card loading
│   ├── logger/            # Structured logging
│   ├── model/             # Domain entities and value objects
│   ├── repository/        # Data access layer
│   └── service/           # Application business logic
├── pkg/                   # Public packages
│   └── typegen/           # TypeScript type generation
├── test/                  # Test suite (mirrors internal/ structure)
├── tools/                 # Development tools (see tools/CLAUDE.md)
├── assets/                # Static game data (CSV, JSON)
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
go test ./test/service/   # Specific package
go test -json ./test/...  # JSON output for parsing
```

**Test Location**: Tests live in `test/` directory, mirroring `internal/` structure. Example: `test/service/player_service_test.go` tests `internal/service/player_service.go`.

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

### Adding New WebSocket Actions

1. **Create handler** in `internal/delivery/websocket/handler/`:
   - Implement ActionHandler interface
   - Validate incoming message
   - Call appropriate service methods
   - Services handle SessionManager broadcasting

2. **Register handler** in WebSocket manager

3. **Add message type** to frontend types if needed

### Implementing Card Effects

**For passive effects** (event-driven):

1. Define behavior in card JSON with triggers and outputs
2. Ensure repository publishes relevant domain event
3. CardEffectSubscriber automatically subscribes on card play
4. No manual service code needed for passive effects

**For immediate effects**:

1. Implement logic in card effect handler
2. Call via CardService when card is played
3. Service updates repositories and broadcasts state

See `docs/EVENT_SYSTEM.md` for complete event system documentation.

### Repository Pattern

- **Immutable Interface**: All getters return values, preventing external mutation
- **Granular Updates**: Specific methods like UpdateResources(), UpdateTerraformRating()
- **Event Integration**: Updates automatically publish domain events
- **Clean Relationships**: Use ID references, not embedded objects

### Service Layer Rules

- **Services do only what the action says**: Don't manually check for passive card effects
- **Call repositories for state changes**: Repositories publish events
- **Use SessionManager for broadcasting**: Broadcast() or Send() after state changes
- **Event system handles passive effects**: CardEffectSubscriber triggers effects automatically

## Data Flow

### WebSocket Message Flow

```
Client → WebSocket Connection → Hub.HandleMessage()
                                       ↓
                                 Manager.RouteMessage()
                                       ↓
                                 ActionHandler.Handle()
                                       ↓
                                 Service Layer
                                       ↓
                           Repository Updates + Events
                                       ↓
                           EventBus → Hub → Broadcaster
                                       ↓
                              All Clients Updated
```

### Game State Synchronization

1. Service performs business logic
2. Repository updates state and publishes events
3. EventBus notifies subscribers (passive effects, etc.)
4. Service calls SessionManager.Broadcast() or Send()
5. SessionManager fetches complete game state from repositories
6. Clients receive full state update via WebSocket

## Type System Integration

### Go to TypeScript

All domain models with `ts:` tags generate TypeScript interfaces:

```go
// backend/internal/model/player.go
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

1. Modify Go structs in `internal/model/`
2. Run `make generate` from project root
3. Check corresponding DTOs in `internal/delivery/dto/`
4. Update DTO mappers in `internal/delivery/dto/mapper.go` if needed
5. Frontend automatically gets updated types

## Important Notes

### Event-Driven Architecture

**CRITICAL**: Services should NOT manually check for passive card effects. The event system handles this automatically:

```go
// ✅ CORRECT
func (s *GameService) RaiseTemperature(...) {
    repo.UpdateTemperature(...)  // Publishes TemperatureChangedEvent
    // CardEffectSubscriber automatically triggers passive effects
}

// ❌ WRONG
func (s *GameService) RaiseTemperature(...) {
    repo.UpdateTemperature(...)
    // Don't manually loop through cards to check effects
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

### Adding a New Domain Entity

1. Create struct in `internal/model/` with `json:` and `ts:` tags
2. Implement DeepCopy() for defensive copying
3. Add repository methods in `internal/repository/`
4. Create service methods in `internal/service/`
5. Add DTOs in `internal/delivery/dto/`
6. Add mappers in `internal/delivery/dto/mapper.go`
7. Run `make generate` to sync TypeScript types

### Adding a New Game Rule

1. Check `TERRAFORMING_MARS_RULES.md` in project root
2. Implement in domain layer (`internal/model/`)
3. Add validation in service layer (`internal/service/`)
4. Create tests in `test/service/`
5. Update relevant WebSocket handlers if needed

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
- **tools/CLAUDE.md**: Card parser tool documentation
- **frontend/CLAUDE.md**: Frontend architecture and patterns
- **docs/EVENT_SYSTEM.md**: Detailed event system documentation
- **TERRAFORMING_MARS_RULES.md**: Complete game rules reference
