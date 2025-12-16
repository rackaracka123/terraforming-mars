# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Digital implementation of Terraforming Mars board game with real-time multiplayer and 3D game view. Features: drag-to-pan 3D Mars visualization, hexagonal tile system, comprehensive card effects engine, WebSocket multiplayer with Go backend and React frontend.

## Quick Start

```bash
make run         # Run both frontend (3000) and backend (3001) with hot reload
make help        # Show all available commands
```

## Essential Commands

### Development
```bash
make frontend    # React dev server (port 3000)
make backend     # Go backend with Air hot reload (port 3001)
make dev-setup   # Set up environment (go mod tidy + npm install)
```

### Testing
```bash
make test         # Run all backend tests
make test-verbose # Verbose test output
make test-coverage# Generate coverage report (backend/coverage.html)
make test-quick   # Fast test suite for iteration
```

**IMPORTANT**: Test files go in `test/` directory (e.g., `test/middleware/validator_test.go` tests `internal/middleware/validator.go`).

### Code Quality
```bash
make lint         # Run all linters (Go fmt + oxlint)
make format       # Format all code (Go + TypeScript)
make generate     # Generate TypeScript types from Go structs
```

**CRITICAL**: Always run `make format` and `make lint` after completing any feature. Fix all lint ERRORS immediately - no exceptions.

### Build
```bash
make build        # Build both frontend and backend
make clean        # Clean build artifacts
```

## Backend Architecture

The Go backend follows Clean Architecture with an action-based pattern where each business operation is a focused, single-responsibility action (~100-200 lines).

### Directory Structure

```
backend/
├── cmd/server/            # Application entry point with dependency injection
├── internal/
│   ├── action/            # Business logic actions (ONLY place for state mutation)
│   │   ├── base.go        # BaseAction with common dependencies
│   │   ├── query/         # Read-only HTTP GET operations
│   │   └── admin/         # Admin operations for testing/debugging
│   ├── game/              # Game state repository and domain types
│   │   ├── game.go        # Core Game type with all game state
│   │   ├── player/        # Player entity and components
│   │   ├── board/         # Board and Tile types
│   │   ├── deck/          # Deck management
│   │   ├── cards/         # Card effect helpers (NO state mutation)
│   │   ├── shared/        # Shared types (Resources, HexPosition, etc.)
│   │   └── global_parameters/  # GlobalParameters with terraforming state
│   ├── cards/             # Card data (registry, JSON loading, validation)
│   ├── delivery/          # Presentation layer
│   │   ├── dto/           # Data Transfer Objects and mappers
│   │   ├── http/          # HTTP handlers (delegate to actions)
│   │   └── websocket/     # WebSocket hub, handlers, broadcaster
│   ├── events/            # Event bus and domain events
│   └── logger/            # Structured logging
├── test/                  # Comprehensive test suite
└── tools/                 # Code generation utilities
```

### Core Principles

**Action Pattern**
- Each action performs ONE operation (join game, play card, raise temperature)
- Actions extend `BaseAction` with injected dependencies (GameRepository, CardRegistry, logger)
- Execute method with clear inputs/outputs and error handling
- HTTP and WebSocket handlers both delegate to the same actions

```go
type MyAction struct {
    BaseAction  // gameRepo, cardRegistry, logger
}

func (a *MyAction) Execute(ctx context.Context, params...) (*Result, error) {
    // 1. Validate inputs
    // 2. Fetch game from GameRepository
    // 3. Call game state methods (Game publishes events automatically)
    // 4. Return result (Broadcaster receives events and broadcasts)
}
```

**Game as State Repository**
- Single Game type contains all state (Players, Board, Deck, GlobalParameters)
- Private fields with public accessor methods enforce encapsulation
- State mutation methods publish domain events to EventBus
- GameRepository manages collection of active games
- No separate player/board/deck repositories - all accessed via Game

**Event-Driven Architecture**
- Actions update Game → Game publishes events → Broadcaster sends to clients
- Passive card effects subscribe to domain events automatically
- No manual polling or effect checking in services
- **Core Rule**: Services do ONLY what the action says. Effects trigger via events.

**State Mutation Rule**
- **ONLY actions** in `/internal/action/` may mutate game state
- All other packages provide helpers, parse data, or subscribe to events
- Actions call game state methods: `player.Resources().AddCredits()`, `game.GlobalParameters().IncreaseTemperature()`
- `/internal/game/cards/` provides helper functions to interpret card behaviors, but does NOT apply them
- Actions use helpers to understand WHAT to do, then actions apply the changes

**Card System Architecture**
- **`/internal/cards/`**: Card data outside game context (registry, JSON loading, validation)
- **`/internal/game/cards/`**: Card effect helpers for game context (behavior parsing, NO state mutation)
- **Card behaviors** defined in JSON (`terraforming_mars_cards.json`) with triggers, inputs, outputs
- **Actions** read card behaviors and apply effects to game state
- **90%+ of cards** added via JSON only, no Go code required

**Broadcaster**
Subscribes to BroadcastEvent and sends personalized game state to WebSocket clients:
```go
// In internal/delivery/websocket/broadcaster.go
func (b *Broadcaster) OnBroadcastEvent(event BroadcastEvent) {
    // Fetch game state and send personalized DTO to each player
}
```

### Communication Flow

```
Frontend → WebSocket Hub → Manager.RouteMessage() → Action Handler
    ↓
Action.Execute() → Game State Updates (game methods)
    ↓
EventBus (domain events) → Broadcaster → WebSocket Broadcast
    ↓
All Clients Receive Personalized Game State
```

### WebSocket Messages

**Inbound (Client → Server)**
- `join-game`, `player-reconnect`, `select-corporation`
- `raise-temperature`, `skip-action`, `start-game`

**Outbound (Server → Client)**
- `game-updated` (primary event with complete game state)
- `player-connected`, `player-reconnected`, `player-disconnected`

## Frontend Architecture

### Tech Stack
- **React + TypeScript**: UI components with generated types from Go
- **React Three Fiber**: 3D Mars visualization with custom pan controls
- **WebSocket Client**: Real-time game state synchronization
- **Tailwind CSS v4**: Utility-first styling with custom theme

### Type Safety Bridge
- Go structs with `ts:` tags generate TypeScript interfaces
- Run `make generate` after any Go type changes
- Import types from `src/types/generated/api-types.ts`

```go
// Backend
type Player struct {
    ID       string `json:"id" ts:"string"`
    Credits  int    `json:"credits" ts:"number"`
}

// Frontend (auto-generated)
interface Player {
    id: string;
    credits: number;
}
```

### 3D Rendering System
- **Game3DView.tsx**: Main Three.js canvas
- **HexGrid.tsx**: Hexagonal tile system (cube coordinates: q, r, s where q+r+s=0)
- **PanControls.tsx**: Custom mouse/touch controls (pan + zoom, no orbit)
- **BackgroundCelestials**: Parallax space environment

### UI Component Standards

**GameIcon Component (PRIMARY)**

**CRITICAL**: ALWAYS use GameIcon for ANY game icon display. NEVER use direct `<img src="/assets/...">` tags.

```tsx
import GameIcon from '../ui/display/GameIcon.tsx';

// Basic icons
<GameIcon iconType="steel" size="medium" />
<GameIcon iconType="space" size="small" />

// With amounts
<GameIcon iconType="credits" amount={25} size="large" />  // Number inside icon
<GameIcon iconType="steel" amount={5} size="medium" />    // Number in corner

// Production (automatic brown background)
<GameIcon iconType="energy-production" amount={3} size="small" />
```

**Sizes**: 'small' (24px), 'medium' (32px), 'large' (40px)
**Types**: All ResourceType, CardTag, tiles, global parameters
**Icon Paths**: Centralized in `src/utils/iconStore.ts`

**Tailwind CSS v4 Styling**

**CRITICAL**: CSS Modules (`.module.css`) are DEPRECATED. Use only Tailwind utilities.

- Configuration in `/frontend/src/index.css` (`@theme {}` block)
- Custom colors: `space-black`, `space-blue-500`, `error-red`
- Custom utilities: `font-orbitron`, `shadow-glow`, `backdrop-blur-space`
- Use arbitrary values: `bg-[rgba(10,20,40,0.95)]`
- Global animations go in index.css as `@keyframes` blocks

## Development Guide

### Adding New Game Features

**CRITICAL**: Always check `TERRAFORMING_MARS_RULES.md` first for any task involving game mechanics, rules, or card effects.

1. **Define domain types** in `internal/game/` or subpackages with `json:` and `ts:` tags
2. **Create action** in `internal/action/` extending BaseAction
3. **Update Game methods** if new state access methods needed
4. **Wire handlers** (HTTP or WebSocket) to delegate to action
5. **Generate types**: Run `make generate`
6. **Frontend**: Import generated types, implement UI
7. **Format and lint**: Run `make format` and `make lint`

### Adding Card Effects (JSON-Driven)

**Most cards require ONLY JSON edits, no Go code!**

1. **Immediate effects** (auto trigger):
   ```json
   {"behaviors": [{"triggers": [{"type": "auto"}],
     "outputs": [{"type": "steel-production", "amount": 2}]}]}
   ```
   PlayCardAction applies automatically when card is played.

2. **Manual actions** (blue cards):
   ```json
   {"behaviors": [{"triggers": [{"type": "manual"}],
     "inputs": [{"type": "energy", "amount": 4}],
     "outputs": [{"type": "steel", "amount": 2}]}]}
   ```
   Registered to Player.Actions(), executed via UseCardAction.

3. **Passive effects** (conditional trigger):
   ```json
   {"behaviors": [{"triggers": [{"type": "auto", "condition": {"type": "city-placed"}}],
     "outputs": [{"type": "credits", "amount": 2}]}]}
   ```
   CardEffectSubscriber listens for events, applies outputs automatically.

See `CARD_SYSTEM.md` for complete card architecture documentation.

### Backend Development Patterns

**Action Development**
- ONE operation per action (~100-200 lines)
- Extend BaseAction with explicit dependencies
- Implement `Execute()` with clear parameters
- Design for idempotency when possible

**Game State Usage**
- Access game state via GameRepository: `game, err := a.gameRepo.Get(gameID)`
- Call Game methods: `game.GlobalParameters().IncreaseTemperature(ctx, steps)`
- Access players: `player := game.GetPlayer(playerID)`
- Game methods publish events automatically

**Handler Development**
- HTTP: Parse request → Call action → Map to DTO → Respond
- WebSocket: Parse message → Call action → Events trigger broadcasts
- Always delegate business logic to actions

### Frontend Development Patterns

- **Generated Types**: Always use types from `src/types/generated/api-types.ts`
- **GameIcon First**: Use GameIcon component for all icon displays
- **No Emojis**: Use GameIcon or assets instead
- **Design Consistency**: Inspect existing components for design patterns
- **Mock Data**: Abstract from UI for easier refactoring
- **No Defaults**: Fail explicitly if expected data is missing
- **Promise Handling**: Use `void <function>()` to discard promises in event handlers

### Testing & Debugging

**Backend Tests**
```bash
make test              # All tests
go test -json          # Easier to parse output
go test -json -v       # Verbose JSON output
```

**Frontend Debugging with Playwright**

When asked to "debug frontend", use Playwright MCP to interactively debug:
1. Ensure backend and frontend are running
2. Navigate to `http://localhost:3000`
3. Use MCP tools: snapshot, click, type, screenshot, evaluate
4. Test user flows, inspect state, monitor WebSocket updates

**Playwright waits**: Max 1 second (everything runs locally)

### Type and DTO Synchronization

When updating types in `/internal/game/`:
1. Check if DTOs in `/internal/delivery/dto/` need updates
2. Update mapper functions in `/internal/delivery/dto/mapper.go`
3. Run `make generate` to sync TypeScript types

### Code Quality Standards

**State Management Rules**

**CRITICAL**: Timeouts and delays ARE NOT solutions to bad state management.

❌ **Bad Approaches:**
- `setTimeout()` to wait for state updates
- `sleep()` in tests for timing issues
- Arbitrary retry loops
- Polling instead of event-driven updates

✅ **Correct Approaches:**
- Proper event listeners and callbacks
- Promise/async-await patterns
- Deterministic state machines
- Atomic operations and transaction boundaries
- Proper synchronization (channels, mutexes)

**Logging Guidelines**
- Use emojis for visual distinction
- 🔗 connect, ⛓️‍💥 disconnect
- 📢 broadcast, 💬 direct message
- 📡 HTTP requests
- 🚀 startup, 🛑 shutdown, ✅ completion

## Current Implementation Status

### Working Features
- Real-time WebSocket multiplayer with Go backend
- 3D game view with hexagonal Mars board
- Clean architecture with action-based pattern
- Automatic type generation (Go → TypeScript)
- Resource management and global parameters
- Corporation selection with synchronization
- Waiting room system with lobby management
- Game state persistence and reconnection (localStorage)

### Waiting Room System
- Games start in `lobby` status, transition to `active`
- Host controls (first player): Start game button
- Shareable join links with URL parameters
- Real-time player join updates
- UI adapts: Resource bar hidden in lobby

### Game State Persistence
- localStorage: gameId, playerId, playerName
- Page reload support with automatic reconnection
- State recovery: API call → WebSocket reconnect
- Fallback to landing page if reconnection fails

### Missing Features
- Tile placement logic and adjacency bonuses
- Advanced turn phases and action state machine
- Victory condition checking
- Milestones and awards tracking
- Advanced card effects with complex state interactions

## Important Notes

### Development Workflow
- Both servers run with hot reload (`make run`)
- Type generation: Go changes → `make generate` → React implementation
- State flow: All changes originate from Go backend via WebSocket
- No client-side game logic (prevents desync)

### Test Creation
- Always write tests for new backend features
- Use `git mv` when moving files checked into git

### Hex Coordinate System
Cube coordinates (q, r, s) where q + r + s = 0
Utilities in `HexMath` class for conversions and neighbors

### Energy/Power Reference
When working with energy, it's referenced as `power.png` in assets

## Important Instruction Reminders

- Do what has been asked; nothing more, nothing less
- NEVER create files unless absolutely necessary
- ALWAYS prefer editing existing files over creating new ones
- NEVER proactively create documentation files (only when explicitly requested)
- No need to be backwards compatible
- Write tests for new backend features
- **NEVER use deprecated code or comments** - Remove deprecated fields, functions, and comments entirely. Do not keep them for backwards compatibility. If something is deprecated, delete it completely and update all usages.
