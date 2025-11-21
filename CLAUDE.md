# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Digital implementation of Terraforming Mars board game with real-time multiplayer and 3D game view. The game features drag-to-pan 3D Mars visualization, hexagonal tile system, comprehensive card effects engine, and WebSocket multiplayer with a Go backend and React frontend.

## Development Commands

All commands should be run from the **project root directory**. The project now uses a unified Makefile for all development tasks.

### 🚀 Quick Start
```bash
make run         # Run both frontend (3000) and backend (3001) servers with hot reload
make help        # Show all available commands with descriptions
```

### 🎯 Main Commands
```bash
make frontend    # Start React development server with hot reload (port 3000)
make backend     # Start Go backend server with hot reload via Air (port 3001)
```

### 🧪 Testing
```bash
make test         # Run all tests (backend only - frontend has no tests)
make test-verbose # Run backend tests with verbose output
make test-coverage# Generate test coverage report (backend/coverage.html)
make test-quick   # Fast test suite for development iteration
make test-watch   # Watch Go files and run tests on changes (requires entr)
```

**IMPORTANT**: Test files should always be created in the test directory (e.g., `test/middleware/validator_test.go` tests `internal/middleware/validator.go`).

### 🔧 Code Quality
```bash
make lint         # Run all linters (Go fmt + oxlint)
make format       # Format all code (Go + TypeScript)
make generate     # Generate TypeScript types from Go structs
make lint-backend # Go formatting only
make lint-frontend# oxlint only
```

### 🏗️ Build & Deploy
```bash
make build        # Build production binaries for both frontend and backend
make build-backend# Build Go server binary (backend/bin/server)
make build-frontend# Build React production bundle (frontend/dist/)
make clean        # Clean all build artifacts
```

### 🧰 Development Helpers
```bash
make dev-setup    # Set up development environment (go mod tidy + npm install)
```

### 🔄 Type Generation
```bash
make generate                # Generate TypeScript types from Go structs (recommended)
cd backend && tygo generate  # Alternative direct command
```

### Legacy Commands (deprecated)
These commands are no longer needed but mentioned for reference:
- ~~`npm run backend`~~ → Use `make backend`
- ~~`npm run frontend`~~ → Use `make frontend`
- ~~`cd backend && make test`~~ → Use `make test`


## Core Architecture

### Clean Architecture Backend (Go)
The Go backend follows clean architecture principles with clear separation of concerns:

```
backend/
├── cmd/                    # Application entry points
│   ├── server/            # Main server application with dependency injection
│   └── watch/             # Development file watching utility
├── internal/
│   ├── action/            # Action pattern - single-responsibility business logic
│   │   ├── base.go        # BaseAction with common dependencies
│   │   ├── join_game.go   # ~100-200 lines per action
│   │   ├── play_card.go   # Each action focused on ONE operation
│   │   ├── query/         # Query actions for HTTP read operations
│   │   │   ├── get_game.go
│   │   │   ├── list_games.go
│   │   │   └── get_player.go
│   │   └── admin/         # Admin actions for game management
│   │       ├── set_resources.go
│   │       └── set_global_parameters.go
│   ├── session/           # Session subdomain repositories
│   │   ├── session_manager.go  # Unified broadcast interface (2 methods)
│   │   ├── game/          # Game subdomain repository
│   │   ├── player/        # Player subdomain repository
│   │   ├── card/          # Card subdomain repository
│   │   ├── board/         # Board subdomain repository
│   │   ├── deck/          # Deck subdomain repository
│   │   └── types/         # Domain type definitions (Player, Game, Card, etc.)
│   ├── cards/             # Card system, validation, and registry
│   ├── delivery/          # Presentation layer
│   │   ├── dto/           # Data Transfer Objects and mappers
│   │   ├── http/          # HTTP handlers (delegate to actions)
│   │   └── websocket/     # WebSocket architecture
│   │       ├── core/      # Hub, connection manager, broadcaster
│   │       └── handler/   # WebSocket handlers (delegate to actions)
│   ├── events/            # Event bus and domain event definitions
│   ├── initialization/    # Application setup and card loading
│   └── logger/            # Structured logging utilities
├── pkg/typegen/           # TypeScript type generation utilities
├── test/                  # Comprehensive test suite
├── tools/                 # Code generation and development tools
└── docs/swagger/          # Auto-generated API documentation
```

### Full-Stack Communication Flow
1. **Frontend (React)**: UI components with WebSocket client
2. **WebSocket Hub**: Real-time game state synchronization via `gorilla/websocket`
3. **Action Layer**: Single-responsibility actions execute business logic (join game, play card, etc.)
4. **Session Repositories**: Focused data access per subdomain with immutable interfaces and event publishing
5. **Domain Types**: Core game entities with automatic TypeScript generation

### Type Safety Bridge
Go structs automatically generate TypeScript interfaces via custom type generator:
- **Go Domain**: Structs with `ts:` tags define frontend types
- **Code Generation**: `tygo generate` creates TypeScript interfaces
- **Frontend Import**: React components use generated types for full type safety

### 3D Rendering System
- **Game3DView.tsx**: Main Three.js Canvas with React Three Fiber
- **HexGrid.tsx**: Hexagonal tile system for Mars board (42 hexes currently)
- **PanControls.tsx**: Custom mouse/touch controls (pan + zoom, no orbit rotation)
- **BackgroundCelestials**: Parallax layers for space environment

## Action-Based Architecture

The backend uses an **action pattern** where each business operation is a focused, single-responsibility action (~100-200 lines).

### Core Principles

**Single Responsibility**
- Each action performs ONE specific operation (join game, play card, raise temperature)
- Actions are small, focused, and easy to understand
- Clear separation between different business operations

**Explicit Dependencies**
- All dependencies injected via `BaseAction` struct
- Actions declare exactly what they need (repositories, session manager, logger)
- No hidden dependencies or global state

**Type Safety**
- Actions use typed interfaces for repositories
- Return explicit result types with proper error handling
- Integration with event system for reactive behavior

### BaseAction Pattern

All actions extend `BaseAction` which provides common dependencies:

```go
type BaseAction struct {
    gameRepo   session.GameRepository     // Game subdomain repository
    playerRepo session.PlayerRepository    // Player subdomain repository
    sessionMgr session.SessionManager      // Broadcast interface
    logger     *zap.Logger                 // Structured logging
}

// Helper methods available to all actions
func (b *BaseAction) BroadcastGameState(gameID string)
func (b *BaseAction) SendToPlayer(gameID, playerID string)
func (b *BaseAction) GetGameRepo() session.GameRepository
func (b *BaseAction) GetPlayerRepo() session.PlayerRepository
```

### Action Types

**Main Actions** (`internal/action/*.go`)
- Business operations that modify game state
- Examples: `JoinGameAction`, `PlayCardAction`, `ConvertHeatToTemperatureAction`
- Execute game rules and coordinate repository updates
- Trigger events via session repositories

**Query Actions** (`internal/action/query/*.go`)
- Read-only operations for HTTP GET endpoints
- Examples: `GetGameAction`, `ListGamesAction`, `GetPlayerAction`
- Compose data from multiple repositories
- Return complete, personalized views of game state

**Admin Actions** (`internal/action/admin/*.go`)
- Administrative operations for game management
- Examples: `SetResourcesAction`, `SetGlobalParametersAction`, `AdminSelectTilesAction`
- Used for testing, debugging, and game setup

### Action Structure Example

```go
type JoinGameAction struct {
    BaseAction  // Embed common dependencies
}

func (a *JoinGameAction) Execute(ctx context.Context, gameID, playerName string) (*JoinGameResult, error) {
    // 1. Validate game state
    game, err := ValidateLobbyGame(ctx, a.gameRepo, gameID, log)
    if err != nil {
        return nil, err
    }

    // 2. Check if player already exists (idempotency)
    existingPlayers, _ := a.playerRepo.ListByGameID(ctx, gameID)
    for _, p := range existingPlayers {
        if p.Name == playerName {
            return &JoinGameResult{PlayerID: p.ID}, nil
        }
    }

    // 3. Create new player
    newPlayer := player.NewPlayer(playerName)
    err = a.playerRepo.Create(ctx, gameID, newPlayer)

    // 4. Add player to game (triggers PlayerJoinedEvent)
    err = a.gameRepo.AddPlayer(ctx, gameID, newPlayer.ID)

    // 5. Return result (SessionManager handles broadcasting)
    return &JoinGameResult{PlayerID: newPlayer.ID}, nil
}
```

### Benefits

- **Clarity**: ~100-200 lines per action vs 600-1,400 lines per service
- **Testability**: Easy to unit test with mock repositories
- **Reusability**: HTTP and WebSocket handlers both use the same actions
- **Maintainability**: Clear dependencies, single responsibility, explicit interfaces
- **Event Integration**: Actions work seamlessly with event-driven passive effects

## Session Subdomain Repositories

The backend organizes repositories by **subdomain** rather than global singletons. Each subdomain has a focused interface and implementation.

### Subdomain Structure

**Game Subdomain** (`internal/session/game/`)
- Game state, global parameters, phase management
- Methods: `Create`, `GetByID`, `AddPlayer`, `UpdateTemperature`, `UpdatePhase`
- Events: Publishes `TemperatureChanged`, `OxygenChanged`, `GamePhaseChanged`

**Player Subdomain** (`internal/session/player/`)
- Player resources, production, terraform rating, cards
- Methods: `Create`, `GetByID`, `UpdateResources`, `UpdateProduction`, `AddCard`
- Events: Publishes `ResourcesChanged`, `ProductionChanged`, `TerraformRatingChanged`

**Card Subdomain** (`internal/session/card/`)
- Card data, card deck management, card lookups
- Methods: `GetByID`, `ListCardsByIdMap`, `DrawCards`, `ShuffleDeck`
- Events: Publishes `CardDrawn`, `DeckShuffled`

**Board Subdomain** (`internal/session/board/`)
- Tile placement, board state, hex occupancy
- Methods: `GetBoard`, `UpdateTileOccupancy`, `GetAdjacentTiles`
- Events: Publishes `TilePlaced`

**Deck Subdomain** (`internal/session/deck/`)
- Project card deck, corporation deck
- Methods: `Initialize`, `Draw`, `Shuffle`, `Discard`

### Immutable Interface Pattern

All repositories return **values, not pointers** to prevent external mutation:

```go
// ✅ CORRECT: Returns value, preventing external mutation
func (r *PlayerRepository) GetByID(ctx context.Context, gameID, playerID string) (*Player, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    player := r.players[gameID][playerID]
    // Return a copy, not a pointer to internal state
    return player.DeepCopy(), nil
}

// ✅ CORRECT: Specific update method publishes precise event
func (r *PlayerRepository) UpdateResources(ctx context.Context, gameID, playerID string, resources Resources) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    player := r.players[gameID][playerID]
    player.Resources = resources

    // Publish specific event
    r.eventBus.Publish(events.ResourcesChangedEvent{
        GameID:   gameID,
        PlayerID: playerID,
        Resources: resources,
    })

    return nil
}
```

### Granular Updates

Repositories provide **specific methods** for each type of update, enabling precise event handling:

```go
// Game Repository
UpdateTemperature(gameID string, temp int)        // → TemperatureChangedEvent
UpdateOxygen(gameID string, oxygen int)           // → OxygenChangedEvent
UpdateOceans(gameID string, oceans int)           // → OceansChangedEvent
AddPlayer(gameID, playerID string)                // → PlayerJoinedEvent

// Player Repository
UpdateResources(gameID, playerID, resources)      // → ResourcesChangedEvent
UpdateProduction(gameID, playerID, production)    // → ProductionChangedEvent
UpdateTerraformRating(gameID, playerID, tr int)   // → TerraformRatingChangedEvent
AddCard(gameID, playerID, cardID string)          // → CardAddedEvent
```

### SessionManager Interface

The `SessionManager` provides a **simplified 2-method interface** for broadcasting:

```go
type SessionManager interface {
    Broadcast(gameID string) error                     // Send to all players in game
    Send(gameID string, playerID string) error         // Send to specific player
}
```

**How it works:**
1. Actions update repositories with granular methods
2. Repositories publish domain events to EventBus
3. SessionManager subscribes to events and broadcasts state
4. All connected clients receive personalized game state

**Internal Implementation:**
- Both methods fetch complete game state from repositories
- Generate personalized DTOs for each target player
- Send via WebSocket Hub to connected clients
- Handle disconnected clients gracefully

## Clean Architecture Implementation

The backend follows Clean Architecture principles with strict separation of concerns and dependency inversion.

### Architectural Layers

**Domain Layer** (`internal/session/types/`)
- **Domain Entities**: Core business objects with identity (Player, Game, GlobalParameters, Card)
- **Value Objects**: Immutable objects defined by their values (Resources, Production, Tile)
- **Domain Events**: Represent significant business occurrences (defined in `internal/events/`)
- **Defensive Copying**: All entities implement `DeepCopy()` to prevent external mutation
- **No Dependencies**: This layer has no dependencies on other layers

**Application Layer** (`internal/action/`)
- **Actions**: Single-responsibility operations that execute business logic (~100-200 lines each)
- **BaseAction**: Common dependencies injected (repositories, session manager, logger)
- **Query Actions**: Read-only operations for composing data from multiple repositories
- **Admin Actions**: Administrative operations for game management
- **Dependency Rule**: Actions depend on domain types and session repository interfaces

**Infrastructure Layer** (`internal/session/*/`)
- **Subdomain Repositories**: Focused repositories per domain (game, player, card, board, deck)
- **Immutable Interfaces**: All getters return values, not pointers, to maintain immutability
- **Granular Updates**: Specific methods for each type of update enable precise event handling
- **Clean Relationships**: Games reference PlayerIDs, not embedded Player objects
- **Event Publishing**: Repository methods automatically trigger domain events via EventBus
- **SessionManager**: Unified 2-method interface for broadcasting game state

**Presentation Layer** (`internal/delivery/`)
- **HTTP Handlers**: Delegate to actions for business logic, return DTOs
- **WebSocket Handlers**: Delegate to actions for WebSocket message processing
- **Request/Response Models**: DTOs for external communication with proper mapping
- **Dependency Direction**: Handlers depend on actions, not repositories directly

**Card System Layer** (`internal/cards/`)
- **Card Registry**: Centralized registration and lookup for all game cards
- **Card Validation**: Comprehensive validation system for card plays and requirements
- **Effect Implementation**: Card-specific business logic integrated with game services
- **Modular Design**: Each card type has dedicated handler with consistent interface

**Event System** (`internal/events/`)
- **Event Bus**: Type-safe event publishing and subscription system
- **Domain Events**: Game and player events (TemperatureChanged, TilePlaced, ResourcesChanged, etc.)
- **CardEffectSubscriber**: Manages passive card effect subscriptions to domain events
- **Event-Driven Effects**: Card passive effects trigger automatically via events, not manual polling
- **Decoupled Architecture**: Services execute actions, repositories publish events, effects subscribe
- **See**: `backend/docs/EVENT_SYSTEM.md` for comprehensive documentation

**Session Management Layer** (`internal/delivery/websocket/session/`)
- **SessionManager Interface**: Simplified to exactly 2 methods: `Broadcast(gameID)` and `Send(gameID, playerID)`
- **Complete State Broadcasting**: Both methods send full game state with all data to relevant players
- **Repository Integration**: Uses repositories directly (GameRepo, PlayerRepo, CardRepo) to avoid circular dependencies
- **Service Integration Pattern**: Services update repositories first, then use SessionManager for broadcasting

### Clean Architecture Principles

**1. Dependency Inversion**
- High-level modules (Application) don't depend on low-level modules (Infrastructure)
- Both depend on abstractions (interfaces)
- Infrastructure implements interfaces defined in Application layer

**2. Separation of Concerns**
- **Domain**: Pure business logic with no external dependencies (types in `session/types/`)
- **Application**: Actions execute single-responsibility operations (in `action/`)
- **Infrastructure**: Session repositories handle data access and events (in `session/*/`)
- **Presentation**: HTTP/WebSocket handlers delegate to actions (in `delivery/`)

**3. Testability**
- Business logic isolated in actions with explicit dependencies
- Session repositories implement interfaces for easy mocking
- Actions can be unit tested with mock repositories
- Event system allows testing of reactive behaviors

**4. Independence**
- Business rules encapsulated in actions, independent of frameworks
- Domain types contain core business logic without external dependencies
- Actions coordinate domain operations without tight coupling to infrastructure
- Session repositories handle data access without leaking implementation details

### Development Guidelines

**Type and DTO Synchronization**
- Whenever you update type structs in `/internal/session/types/`, check if corresponding DTOs in `/internal/delivery/dto/` also need updating
- Always run `make generate` after type changes to sync TypeScript types
- Ensure all new fields are properly included in DTO mapping functions in `/internal/delivery/dto/mapper.go`

**Domain Layer (Session Types)**
- Keep entities focused on business invariants
- Use defensive copying to protect entity state (implement `DeepCopy()` methods)
- Define types in `/internal/session/types/` for domain entities
- No external dependencies or infrastructure concerns

**Action Layer Development**
- **Single Responsibility**: Each action performs ONE operation (~100-200 lines)
- **Extend BaseAction**: Use common dependencies (gameRepo, playerRepo, sessionMgr, logger)
- **Explicit Dependencies**: Inject all dependencies, avoid global state
- **Execute Method**: Implement `Execute()` with clear input parameters and return types
- **Error Handling**: Return explicit errors with context
- **Idempotency**: Design actions to be safely retried when possible
- **Example Pattern**:
  ```go
  type MyAction struct {
      BaseAction
  }

  func (a *MyAction) Execute(ctx context.Context, params...) (*Result, error) {
      // 1. Validate inputs
      // 2. Call session repositories
      // 3. Return result (broadcasting handled by SessionManager)
  }
  ```

**Session Repository Usage**
- Use **subdomain-specific repositories**: gameRepo, playerRepo, cardRepo, boardRepo
- Call **granular update methods**: `UpdateResources()`, `UpdateTemperature()`, etc.
- Let repositories **publish events automatically**
- Never mutate returned values - repositories return copies
- Compose data from multiple repositories as needed in actions

**Presentation Layer**
- **Delegate to actions**: Handlers should call actions, not repositories directly
- **HTTP Handlers**: Parse request → Call action → Map to DTO → Respond
- **WebSocket Handlers**: Parse message → Call action → SessionManager broadcasts
- Implement proper error handling and validation
- Keep presentation logic separate from business logic

**Card System Integration**
- Use card registry for centralized card management
- Implement card validation before processing effects
- Integrate card actions with existing service layer
- Follow modular design patterns for new card types

**Event-Driven Effect System** (📖 See `backend/docs/EVENT_SYSTEM.md`)
- **Core Principle**: Services do ONLY what the action says. Passive effects trigger via events.
- **Repositories Publish Events**: When state changes (tile placed, temperature raised, etc.)
- **CardEffectSubscriber Listens**: Subscribes card passive effects to relevant domain events
- **Automatic Triggering**: Effects fire when events match trigger conditions (no manual polling)
- **Target Filtering**: Effects respect TargetSelfPlayer vs TargetAnyPlayer constraints
- **Example Flow**: Player places tile → GameRepository publishes TilePlacedEvent → CardEffectSubscriber triggers ocean adjacency bonus → Player gains credits

**When Implementing Game Actions:**
```go
// ✅ CORRECT: Service does only the action
func (s *PlayerService) PlaceTile(...) {
    // 1. Update game state
    gameRepo.UpdateTileOccupancy(...)

    // 2. Award immediate bonuses (from board, not cards)
    s.awardTilePlacementBonuses(...)

    // 3. Done! CardEffectSubscriber handles passive card effects via events
}

// ❌ WRONG: Service manually checks for card effects
func (s *PlayerService) PlaceTile(...) {
    gameRepo.UpdateTileOccupancy(...)

    // ❌ Don't do this - let events handle it
    for _, card := range player.Cards {
        if card.TriggersOnTilePlacement {
            applyEffect(...)
        }
    }
}
```

## Game State Flow

### WebSocket Event Architecture

**Modern Handler System**
The backend uses a sophisticated action handler system for WebSocket messages:

```
Client Message -> Hub.HandleMessage() -> Manager.RouteMessage() -> ActionHandler.Handle()
                                                                        ↓
                                                              Service Layer (Business Logic)
                                                                        ↓
                                                              Repository Updates + Events
                                                                        ↓
                                                              EventBus -> Hub -> Broadcaster
                                                                        ↓
                                                              All Clients Receive Updates
```

**Handler Registration**
Each action type has a dedicated handler in `internal/delivery/websocket/handler/`:
- `JoinGameHandler`: Player joining game sessions
- `StartGameHandler`: Host starting games from lobby
- `SelectCorporationHandler`: Corporation selection logic
- `RaiseTemperatureHandler`: Global parameter modifications
- `SkipActionHandler`: Turn progression and phase management

**Message Flow Architecture**
1. **WebSocket Connection**: Client establishes connection -> Hub registers client
2. **Message Reception**: Hub.HandleMessage() receives raw WebSocket message
3. **Action Routing**: Manager.RouteMessage() identifies action type and routes to handler
4. **Handler Processing**: Dedicated ActionHandler validates message and calls services
5. **Business Logic**: Service layer executes domain operations via repositories
6. **Session Broadcasting**: Service calls SessionManager.Broadcast() or Send() to notify players
7. **State Distribution**: SessionManager retrieves complete game state and sends to relevant clients
8. **Frontend Updates**: React components receive state changes and re-render UI

## Type System Overview

### Go Domain Entities (Backend)
- **GameState**: Root state with players, parameters, deck, game settings
- **Player**: Resources, production, corporation, terraform rating, played cards
- **Corporation**: Asymmetric player powers and starting conditions
- **GlobalParameters**: Temperature (-30 to +8°C), Oxygen (0-14%), Oceans (0-9)
- **GamePhase**: Current game phase (setup, corporation_selection, action, production, etc.)

### TypeScript Generation
Go structs use `ts:` tags to specify TypeScript types:
```go
type Player struct {
    ID       string `json:"id" ts:"string"`
    Credits  int    `json:"credits" ts:"number"`
    IsActive bool   `json:"isActive" ts:"boolean"`
}
```

## Terraforming Mars Game Rules Reference

**CRITICAL**: For ANY task that involves Terraforming Mars game mechanics, rules, card effects, or gameplay logic, you MUST consult `TERRAFORMING_MARS_RULES.md` first. This includes:
- Implementing game rules and logic
- Validating game state transitions  
- Creating card effects and interactions
- Designing UI components for game elements
- Debugging game behavior
- Adding new features that interact with existing rules
- Answering questions about game mechanics
- Any feature that even SLIGHTLY touches game rules

The `TERRAFORMING_MARS_RULES.md` file contains the complete, authoritative rulebook reference structured for AI consumption.

## Key Development Patterns

### Adding New Card Effects (Event-Driven)

**For cards with passive effects** (e.g., "Gain 2 MC when any city is placed"):

1. **Define behavior in card JSON:**
   ```json
   {
     "behaviors": [{
       "triggers": [{"type": "auto", "condition": {"type": "city-placed"}}],
       "outputs": [{"type": "credits", "amount": 2, "target": "any-player"}]
     }]
   }
   ```

2. **Ensure repository publishes event:**
   - Check that the relevant repository (e.g., `GameRepository.UpdateTileOccupancy`) publishes the domain event
   - Usually already implemented for common events (TilePlaced, TemperatureChanged, etc.)

3. **CardEffectSubscriber handles subscription automatically:**
   - When card is played, `CardService.OnPlayCard()` calls `effectSubscriber.SubscribeCardEffects()`
   - No additional service code needed!

4. **Test the effect:**
   ```go
   // Play card with passive effect
   cardService.OnPlayCard(ctx, gameID, playerID, cardID, nil, nil)

   // Trigger the event (e.g., place a city)
   gameRepo.UpdateTileOccupancy(ctx, gameID, coord, cityTile, &playerID)

   // Verify effect applied
   player, _ := playerRepo.GetByID(ctx, gameID, playerID)
   assert.Equal(t, expectedCredits, player.Resources.Credits)
   ```

**See `backend/docs/EVENT_SYSTEM.md` for complete documentation.**

### Adding New Game Features
1. **Consult game rules**: Check `TERRAFORMING_MARS_RULES.md` for any game rule implications
2. **Define domain types** in `internal/session/types/` with proper `ts:` tags
3. **Create action** in `internal/action/` extending `BaseAction` with `Execute()` method
4. **Update session repositories** if new data access methods are needed
5. **Wire up handlers**: HTTP or WebSocket handlers delegate to new action
6. **Generate types**: Run `make generate` to update frontend types
7. **Frontend integration**: Import generated types and implement UI
8. **Format and lint**: **ALWAYS** run `make format` and `make lint` after completing any feature

### Backend Development Flow
1. **Define/update types** in `internal/session/types/` with `ts:` tags
2. **Create action** in `internal/action/` extending BaseAction
3. **Wire handler** to call action (HTTP or WebSocket)
4. **Run type generation**: `make generate` syncs TypeScript types
5. **Frontend**: Import generated types and implement UI

### 3D Scene Modifications
- HexGrid positions calculated via hex-to-pixel coordinate conversion
- Mars visual state driven by GameState.globalParameters
- Custom materials respond to terraforming progress (color changes)

## Important Implementation Details

### Hex Coordinate System
Uses cube coordinates (q, r, s) where q + r + s = 0. Utilities in `HexMath` class handle conversions and neighbor calculations for tile-based game mechanics.

### Multiplayer State Synchronization
Game state is authoritative on Go backend. All clients receive full state updates via WebSocket 'game-updated' events. No client-side game logic to prevent desync.

### WebSocket Message System

**Inbound Message Types (Client → Server)**
- `join-game`: Player joins or creates a game session
- `player-reconnect`: Existing player reconnects to game session
- `select-corporation`: Choose starting corporation during setup
- `raise-temperature`: Spend heat to increase global temperature parameter
- `skip-action`: Pass current turn and advance game phase
- `start-game`: Host transitions game from lobby to active status

**Outbound Event Types (Server → Client)**
- `game-updated`: Complete game state synchronization (primary event)
- `player-connected`: Notification when new player joins
- `player-reconnected`: Notification when existing player reconnects  
- `player-disconnected`: Real-time connection status updates

**Event-Driven Broadcasting**
The system uses consolidated event types for efficient state synchronization:
- **Primary Event**: `EventTypeGameUpdated` carries complete game state
- **Event Flow**: Service Action → Repository Update → EventBus → Hub → Broadcast
- **State Consistency**: All clients receive identical state snapshots
- **Connection Management**: Hub tracks client connections and handles disconnections gracefully

### Go Struct Tags for Type Generation
Use both `json:` and `ts:` tags on all domain structs:
```go
type Resource struct {
    Amount int `json:"amount" ts:"number"`
    Production int `json:"production" ts:"number"`
}
```

## Current Implementation Status

### Working Systems
- **Real-time WebSocket multiplayer** with Go backend
- **3D game view** with hexagonal Mars board (React Three Fiber)
- **Clean architecture backend** with clear separation of concerns
- **Automatic type generation** from Go structs to TypeScript
- **Resource management** and global parameter tracking
- **Corporation selection** with WebSocket synchronization
- **Custom pan controls** for 3D Mars view (no orbital rotation)
- **Waiting room system** with lobby phase management

### Waiting Room System
- **Game Status Management**: Games start in `GameStatusLobby` and transition to `GameStatusActive` when started
- **Host Controls**: First player to create/join becomes the host (`game.hostPlayerId`)
- **Start Game Button**: Only visible to the host, triggers `start-game` WebSocket action
- **Shareable Join Links**: Generate URLs like `https://domain/join?code={gameId}` with copy functionality
- **URL Parameter Handling**: JoinGamePage automatically validates and uses `?code` parameter
- **Real-time Updates**: Players see new joins instantly via WebSocket `game-updated` events
- **UI Adaptation**: Bottom resource bar and cards are hidden during lobby phase
- **Mars Background**: 3D Mars view remains visible with translucent overlay for better contrast

### Game State Persistence & Reconnection
- **localStorage Storage**: Game data automatically saved after create/join with `gameId`, `playerId`, `playerName`
- **Page Reload Support**: GameInterface checks localStorage when route state is missing
- **Automatic Reconnection**: Fetches current game state from server and reconnects WebSocket
- **State Recovery Flow**: API call → WebSocket reconnect → Full state restoration
- **Fallback Logic**: Redirects to landing page if reconnection fails or data is invalid
- **Seamless Experience**: Players can reload page without losing game session
- **Error Handling**: Invalid/expired game data is cleaned up automatically
- **Unified Connection Behavior**: Page refresh and close/reopen tab both use the same reconnection flow

#### Game Phase Transitions
1. **Creation**: Game starts in `lobby` status with first player as host
2. **Joining**: Additional players join via game ID or shareable link
3. **Starting**: Host clicks "Start Game" → triggers `start-game` action
4. **Transition**: Backend changes status to `active` and phase to `starting_card_selection`
5. **Active Game**: Resource bars and cards become visible, game logic begins

### Backend Architecture Complete
- **Domain models** with comprehensive game entities
- **Use case layer** for game business logic
- **WebSocket hub** for real-time communication
- **HTTP API** with Swagger documentation
- **In-memory repository** for fast game state access

### Frontend Ready for Extension
- **Generated TypeScript types** ensure backend/frontend sync
- **3D rendering system** using Three.js and React Three Fiber
- **Component architecture** for modular game UI development

### Key Missing Pieces
- **Tile placement** logic and adjacency bonuses
- **Advanced turn phases** and complex action state machine
- **Victory condition** checking and game end detection
- **Milestones and awards** tracking and validation
- **Advanced card effects** requiring complex game state interactions

## UI Component Standards

### Icon Display - GameIcon Component (PRIMARY)
**CRITICAL**: ALWAYS use the GameIcon component for displaying ANY game icon (resources, tags, tiles, global parameters, etc.). NEVER use direct `<img>` tags with asset imports.

#### GameIcon Usage
```tsx
import GameIcon from '../ui/display/GameIcon.tsx';

// Basic resource icon
<GameIcon iconType="steel" size="medium" />

// Credits with amount (number inside icon)
<GameIcon iconType="credits" amount={25} size="large" />

// Production resource (automatic brown background)
<GameIcon iconType="energy-production" amount={3} size="small" />

// Card tags
<GameIcon iconType="space" size="medium" />

// Tiles and global parameters
<GameIcon iconType="ocean-tile" size="small" />
<GameIcon iconType="temperature" size="medium" />
```

**Component**: `src/components/ui/display/GameIcon.tsx`
**Sizes**: 'small' (24px), 'medium' (32px), 'large' (40px)
**Supported Types**: All ResourceType, CardTag, tiles, global parameters, and special icons

**Key Features**:
- Automatic production background for "-production" suffix
- Special number overlay for megacredits (inside icon)
- Centralized icon path management via `iconStore.ts`
- Consistent sizing across all icon types
- Attack indicator support with red glow animation

### Legacy Display Components
These components are kept for backward compatibility but GameIcon should be preferred for new code:

#### CostDisplay (for megacredits only)
```tsx
import CostDisplay from '../display/CostDisplay.tsx';
<CostDisplay cost={amount} size="medium" />
```
Use when you specifically need the CostDisplay wrapper styling.

### UI Development Patterns
- **GameIcon First**: ALWAYS use GameIcon component for any icon display - never use `<img src="/assets/...">` directly
- **Inspect existing design language**: When updating any UI element in the frontend, other components should ALWAYS be inspected for the design language in the codebase
- **Reuse over creation**: Always check for existing components before creating new ones
- **Consistent styling**: Use established components to maintain visual consistency
- **Centralized icons**: All icon paths are managed in `src/utils/iconStore.ts`
- **Responsive sizing**: Components should support multiple sizes for different contexts

## Code Quality Requirements

**CRITICAL**: Always run these commands after completing any task involving code changes:

### Backend Formatting
```bash
cd backend
make format            # Format Go code with gofmt
```

### Frontend Formatting
```bash
cd frontend
npm run format:write   # Format code with Prettier
npm run lint           # Check for oxlint errors
```

**Note**: These commands must be run from the respective directories (backend/ and frontend/). Always format both backend and frontend code after any changes, even if you only modified one side, to maintain consistent code quality across the entire codebase.

**Lint Error Policy**:
- All lint ERRORS must be fixed immediately - no exceptions
- Lint warnings should be addressed when practical
- Never commit code with lint errors
- Run these commands after any significant code changes

### Tailwind CSS v4 Styling Architecture

**CRITICAL**: This project uses Tailwind CSS v4 with CSS-based configuration. Traditional CSS Modules and custom `.module.css` files are **DEPRECATED** and should NEVER be used.

#### Tailwind v4 Configuration
- **Configuration File**: `/frontend/src/index.css` contains the `@theme {}` block
- **No tailwind.config.js**: The JavaScript config file is **IGNORED** by Tailwind v4
- **Import Syntax**: Use `@import "tailwindcss";` instead of `@tailwind` directives

#### Custom Theme Utilities
The project defines custom utilities in `index.css` under `@theme {}`:

**Colors**:
- `space-black`: #0a0a0f
- `space-black-darker`: #050509
- `space-black-light`: #141420
- `space-blue-500/600/900`: rgba(30, 60, 150, ...) variants
- `error-red`: #ff6b6b

**Typography**:
- `font-orbitron`: Orbitron font family (use for titles and headings)
- `text-shadow-glow-strong`: Blue glow text shadow
- `tracking-wider-2xl`: Extra wide letter spacing

**Effects**:
- `shadow-glow/glow-sm/glow-lg`: Blue glow box shadows
- `backdrop-blur-space/space-light`: Backdrop blur utilities

#### Styling Guidelines
1. **NEVER create CSS Module files** (`.module.css`)
2. **Use Tailwind utilities** with arbitrary values: `bg-[rgba(10,20,40,0.95)]`
3. **Use custom theme classes**: `font-orbitron`, `shadow-glow`, `border-space-blue-500`
4. **Add new theme values** to `@theme {}` in index.css if needed
5. **Global animations** go in index.css as `@keyframes` blocks

#### Migration from CSS Modules
When converting existing CSS Module components:
1. Create a `.bak` backup of the original component
2. Remove the CSS module import
3. Convert all className references to Tailwind utilities
4. Add any animations to index.css
5. Delete the `.module.css` file
6. Run `npm run format:write`

**Logging Guidelines**:
- Use emojis in log messages where appropriate to make them more visually distinctive
- Include directional indicators for client/server communication (client→server, server→client)
- Connection logs: 🔗 for connect, ⛓️‍💥 for disconnect
- Broadcasting: 📢 for server broadcasts, 💬 for direct messages
- HTTP requests: 📡 for client requests to server
- Server lifecycle: 🚀 for startup, 🛑 for shutdown, ✅ for completion

## Development Notes

### Backend Development (Go)
- **Action-Based Architecture**: Always implement new features as focused actions in `internal/action/`
- **Type Tags**: Add both `json:` and `ts:` tags to all domain type structs for frontend sync
- **Session Repositories**: Use subdomain repositories (game, player, card, board, deck)
- **Testing**: Use `make test` to run all backend tests
- **API Documentation**: Add Swagger comments to HTTP handlers for auto-generated docs

#### Modern Backend Patterns

**Action Development**
- **Single Responsibility**: Each action performs ONE operation (~100-200 lines)
- **Extend BaseAction**: Inherit common dependencies (repositories, session manager, logger)
- **Execute Method**: Implement clear `Execute()` method with explicit parameters
- **Idempotency**: Design actions to be safely retried when possible
- **Error Handling**: Return explicit errors with proper context

**Session Repository Pattern**
- **Subdomain Focus**: Use focused repositories per domain (game, player, card, board, deck)
- **Immutable Interfaces**: Return values, not pointers, to prevent external state mutation
- **Granular Updates**: Use specific methods (`UpdateResources`, `UpdateTemperature`) for precise events
- **Event Publishing**: Repository operations automatically trigger EventBus notifications
- **Clean Relationships**: Use ID references instead of embedded objects

**HTTP Handler Development**
- **Delegate to Actions**: Handlers parse requests and call actions for business logic
- **DTO Mapping**: Convert action results to DTOs for responses
- **Error Handling**: Map action errors to appropriate HTTP status codes
- **Pattern**: Parse request → Call action → Map to DTO → Respond

**WebSocket Handler Development**
- **Delegate to Actions**: Handlers parse messages and call actions
- **No Direct SessionManager**: Let actions handle business logic, SessionManager broadcasts state
- **Event Response**: Let the event system handle state broadcasting to clients
- **Pattern**: Parse message → Call action → Action updates repos → Events trigger broadcasts

**Card System Development**
- **Card Registration**: Register new cards in the card registry for centralized management
- **Effect Implementation**: Card effects integrate with event system for passive triggers
- **Validation**: Implement comprehensive validation for card requirements and effects
- **Modular Design**: Follow established patterns for consistent card behavior

#### Test Debugging
- **JSON Output**: Use `go test -json` for easier to parse test output when debugging
- **Verbose with JSON**: Use `go test -json -v` for detailed test output in JSON format
- **Specific Package**: `cd backend && go test -json ./test/service/` for focused testing

### Frontend Development (React)
- **Generated Types**: Always use types from `src/types/generated/api-types.ts`
- **3D Rendering**: Uses React Three Fiber - modify scenes in `Game3DView.tsx`
- **WebSocket Client**: Game state updates come via WebSocket, no local game state
- **Component Architecture**: Follow existing patterns for new game UI components
- **Promise Handling**: Use `void <function>()` to explicitly discard promises in event handlers to avoid IDE warnings

### Full-Stack Development
- **Both servers** must be running for full functionality (`make run`)
- **Hot Reload**: Both frontend (Vite) and backend (Air) automatically reload on file changes for rapid development
- **Type Generation**: Run `make generate` after Go struct changes
- **State Flow**: All game state changes originate from Go backend via WebSocket
- **Development Workflow**: Go changes -> generate types -> React implementation
- When creating mock data, abstract it from the UI to enable easier refactoring later
- NEVER set default values - if you expect something, fail explicitly if it's missing

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.
No need to be backwards compatible.

## UI Design Guidelines
- **No Emojis**: Do not use emojis when building any design. Use GameIcon component or assets instead.
- **GameIcon First**: ALWAYS use GameIcon component for displaying icons - NEVER use direct `<img src="/assets/...">` tags
- **Centralized Icons**: All icon paths are managed in `src/utils/iconStore.ts`
- **Asset Location**: Game assets are in `/frontend/public/assets/` but accessed via GameIcon component

## Icon Display Instructions (UPDATED)

**CRITICAL RULE**: NEVER use `<img src="/assets/...">` for game icons. ALWAYS use the GameIcon component.

### Basic Icons
```tsx
import GameIcon from '../ui/display/GameIcon.tsx';

// Resources
<GameIcon iconType="steel" size="medium" />
<GameIcon iconType="plants" size="small" />
<GameIcon iconType="heat" size="large" />

// Card Tags
<GameIcon iconType="space" size="medium" />
<GameIcon iconType="science" size="small" />
<GameIcon iconType="building" size="medium" />

// Global Parameters & Tiles
<GameIcon iconType="temperature" size="medium" />
<GameIcon iconType="oxygen" size="small" />
<GameIcon iconType="ocean-tile" size="medium" />
```

### Icons with Amounts
```tsx
// Megacredits (number displays INSIDE icon)
<GameIcon iconType="credits" amount={25} size="medium" />

// Other resources (number displays in corner if > 1)
<GameIcon iconType="steel" amount={5} size="medium" />
```

### Production Resources
```tsx
// Automatic brown production background when using "-production" suffix
<GameIcon iconType="energy-production" amount={3} size="small" />
<GameIcon iconType="plants-production" amount={2} size="medium" />
<GameIcon iconType="credits-production" amount={5} size="large" />
```

### Adding New Icons to iconStore
If you need to use an icon that's not yet in the centralized system:

1. Add the icon path to the appropriate category in `src/utils/iconStore.ts`:
```tsx
export const RESOURCE_ICONS: { [key: string]: string } = {
  // ... existing icons
  newResource: "/assets/resources/new-resource.png",
};

// Or for tags:
export const TAG_ICONS: { [key: string]: string } = {
  // ... existing icons
  newTag: "/assets/tags/new-tag.png",
};

// Or for special icons:
export const SPECIAL_ICONS: { [key: string]: string } = {
  // ... existing icons
  newIcon: "/assets/misc/new-icon.png",
};
```

2. Use the icon via GameIcon:
```tsx
<GameIcon iconType="newResource" size="medium" />
```

### Legacy Components (Avoid in New Code)
- **CostDisplay**: Use `<GameIcon iconType="credits" amount={X} />` instead
- **Direct asset imports**: Use GameIcon instead of `<img src="/assets/resources/...">`

## UI Components
- **CorporationCard**: Use for displaying corporation options in selection screens
  ```tsx
  import CorporationCard from '../cards/CorporationCard.tsx';
  <CorporationCard corporation={corp} isSelected={selected} onSelect={handler} />
  ```
- When working with energy, its refrenced using power.png
- Use playwright to test UI components.
- **Local Development**: Everything runs locally, so playwright waits only need to be 1 second max.
- Whenever you create a new feature in backend. Write a test for it.
- Whenever you move something that is checked into git. use git mv

## Frontend Debugging with Playwright

**CRITICAL**: When the user asks to "debug frontend", you must launch a Playwright MCP session to interactively debug the application:

### Debugging Protocol
1. **Preparation**: Make sure backend and frontend are running
2. **Launch Playwright**: Use the Playwright MCP server to navigate to `http://localhost:3000` (Playwright config automatically starts frontend via webServer)
3. **Interactive Debugging**: Use Playwright MCP tools to:
   - Navigate through the application
   - Interact with UI elements (click, type, etc.)
   - Take snapshots to inspect page state
   - Capture screenshots for documentation
   - Examine console messages and errors
   - Test user flows and game mechanics

### Playwright MCP Tools Available
- `mcp__playwright__browser_navigate`: Navigate to URLs
- `mcp__playwright__browser_snapshot`: Capture page accessibility snapshot
- `mcp__playwright__browser_click`: Click on UI elements
- `mcp__playwright__browser_type`: Type into form fields
- `mcp__playwright__browser_take_screenshot`: Capture visual screenshots
- `mcp__playwright__browser_evaluate`: Execute JavaScript in browser context

### Debugging Use Cases
- **UI Issues**: Inspect component rendering and layout problems
- **State Problems**: Use the Debug panel to examine real-time game state
- **User Flow Testing**: Navigate through game creation, joining, and gameplay
- **WebSocket Debugging**: Monitor real-time game state updates
- **Performance Issues**: Identify rendering bottlenecks or slow interactions
- **Visual Regressions**: Compare screenshots across different states

**Important**: This is different from writing Playwright tests. When debugging, you should actively use the MCP server to interact with the live application and provide real-time insights about its behavior.

## Code Quality and Architecture Principles

### State Management Rules

**CRITICAL**: Timeouts and temporary fixes ARE NOT SOLUTIONS TO BAD STATE MANAGEMENT.

- **Race Conditions**: Fix the root cause, don't add delays
- **State Synchronization Issues**: Implement proper event handling and state flow
- **Timing Problems**: Design deterministic state transitions
- **Async Coordination**: Use proper synchronization primitives, not arbitrary waits

**Examples of BAD approaches:**
- Adding `setTimeout()` to wait for state updates
- Using `sleep()` in tests to "fix" timing issues  
- Arbitrary retry loops without understanding why they're needed
- Polling instead of proper event-driven updates

**Correct approaches:**
- Implement proper event listeners and callbacks
- Use Promise/async-await patterns correctly
- Design predictable state machines
- Create atomic operations and proper transaction boundaries
- Use proper synchronization (channels, mutexes, etc.) when needed
- No mocks outside of tests.
- Do not close frontend after debugging