# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Digital implementation of Terraforming Mars board game with real-time multiplayer and 3D game view. The game features drag-to-pan 3D Mars visualization, hexagonal tile system, comprehensive card effects engine, and WebSocket multiplayer with a Go backend and React frontend.

## Development Commands

All commands should be run from the **project root directory**. The project now uses a unified Makefile for all development tasks.

### 🚀 Quick Start
```bash
make run         # Run both frontend (3000) and backend (3001) servers
make help        # Show all available commands with descriptions
```

### 🎯 Main Commands
```bash
make frontend    # Start React development server (port 3000)
make backend     # Start Go backend server (port 3001)
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
make lint         # Run all linters (Go fmt + ESLint)
make format       # Format all code (Go + TypeScript)
make generate     # Generate TypeScript types from Go structs
make lint-backend # Go formatting only
make lint-frontend# ESLint only
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
│   ├── cards/             # Card system, validation, and registry
│   ├── delivery/          # Presentation layer
│   │   ├── dto/           # Data Transfer Objects and mappers
│   │   ├── http/          # HTTP handlers, middleware, and routing
│   │   └── websocket/     # WebSocket architecture
│   │       ├── core/      # Hub, connection manager, broadcaster
│   │       └── handler/   # Action-specific message handlers
│   ├── events/            # Event bus and domain event definitions
│   ├── initialization/    # Application setup and card loading
│   ├── logger/            # Structured logging utilities
│   ├── model/             # Domain entities and business objects
│   ├── repository/        # Data access layer with immutable interfaces
│   └── service/           # Application business logic and use cases
├── pkg/typegen/           # TypeScript type generation utilities
├── test/                  # Comprehensive test suite
├── tools/                 # Code generation and development tools
└── docs/swagger/          # Auto-generated API documentation
```

### Full-Stack Communication Flow
1. **Frontend (React)**: UI components with WebSocket client
2. **WebSocket Hub**: Real-time game state synchronization via `gorilla/websocket`
3. **Use Cases**: Game business logic in Go (join game, select corporation, etc.)
4. **Domain Models**: Core game entities with automatic TypeScript generation

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

## Clean Architecture Implementation

The backend follows Clean Architecture principles with strict separation of concerns and dependency inversion.

### Architectural Layers

**Domain Layer** (`internal/model/`)
- **Domain Entities**: Core business objects with identity (Player, Game, GlobalParameters)
- **Value Objects**: Immutable objects defined by their values (Resources, Production)
- **Domain Events**: Represent significant business occurrences (PlayerTRChanged, ResourcesChanged)
- **Defensive Copying**: All entities implement `DeepCopy()` to prevent external mutation
- **No Dependencies**: This layer has no dependencies on other layers

**Application Layer** (`internal/service/`)
- **Use Cases**: Orchestrate business operations and enforce business rules
- **Domain Services**: Handle complex domain logic that spans multiple entities
- **Event Handlers**: Subscribe to and react to domain events
- **Interface Definitions**: Define contracts for infrastructure dependencies
- **Dependency Rule**: Depends only on the Domain layer

**Infrastructure Layer** (`internal/repository/`)
- **Simplified Repositories**: Direct storage of domain models with clean CRUD operations
- **Immutable Getters**: All repository methods return values, not pointers, to maintain immutability
- **Granular Updates**: Specific methods for updating individual fields (UpdateResources, UpdateTerraformRating)
- **Clean Relationships**: Games store PlayerIDs, not embedded Player objects
- **Event Publishing**: Precise events from specific update methods (temperature changed, resources updated)
- **No Entity Classes**: Store domain models directly, eliminating conversion complexity

**Presentation Layer** (`internal/delivery/`)
- **HTTP Endpoints**: Handle REST API requests with middleware and routing
- **WebSocket System**: Sophisticated hub-manager-handler architecture for real-time communication
- **Request/Response Models**: DTOs for external communication with proper mapping
- **Dependency Direction**: Depends on Application layer, not Infrastructure

**Card System Layer** (`internal/cards/`)
- **Card Registry**: Centralized registration and lookup for all game cards
- **Card Validation**: Comprehensive validation system for card plays and requirements
- **Effect Implementation**: Card-specific business logic integrated with game services
- **Modular Design**: Each card type has dedicated handler with consistent interface

**Event System** (`internal/events/`)
- **Event Bus**: Centralized event publishing and subscription system
- **Domain Events**: Consolidated event types (EventTypeGameUpdated, etc.)
- **Event Flow**: Repository operations trigger events → EventBus → WebSocket broadcasting
- **Decoupled Architecture**: Services publish events without knowing about subscribers

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
- **Domain**: Pure business logic with no external dependencies
- **Application**: Coordinates domain operations and defines use cases  
- **Infrastructure**: Handles technical concerns (data, events, external APIs)
- **Presentation**: Manages user interface and external communication

**3. Testability**
- Business logic isolated in Domain and Application layers
- Infrastructure dependencies injected via interfaces
- Easy to mock external dependencies for unit testing

**4. Independence**
- Business rules independent of frameworks, databases, and UI
- Domain entities contain core business logic
- Application services orchestrate domain operations

### Repository Architecture

The backend implements a **Clean Repository Pattern** optimized for real-time game state management:

**Core Design Principles**
- **Direct Model Storage**: Repositories store domain models (`model.Game`, `model.Player`) without entity classes
- **Immutable Interface**: All getters return values, not pointers, preventing external mutation
- **Clean Relationships**: Games reference players via `PlayerIDs []string` instead of embedded objects
- **Granular Updates**: Specific methods for targeted updates enable precise event handling
- **Event Integration**: Repository operations automatically trigger domain events via EventBus

**Service-Repository Coordination**
Services compose data from multiple repositories as needed, maintaining clean separation between business logic and data access while ensuring consistent state management through the event-driven architecture.

### Development Guidelines

**Model and DTO Synchronization**
- Whenever you update model structs in `/internal/model/`, check if corresponding DTOs in `/internal/delivery/dto/` also need updating
- Always run `make generate` after model changes to sync TypeScript types
- Ensure all new fields are properly included in DTO mapping functions in `/internal/delivery/dto/mapper.go`

**Domain Layer**
- Keep entities focused on business invariants
- Use defensive copying to protect entity state
- Implement domain events for significant business occurrences
- No external dependencies or infrastructure concerns

**Application Layer**  
- Orchestrate complex business operations
- Validate business rules before domain operations
- Handle domain events for cross-cutting functionality
- Define interfaces for infrastructure dependencies


**Presentation Layer**
- Use Application services for all business operations
- Never access Infrastructure layer directly  
- Implement proper error handling and validation
- Keep presentation logic separate from business logic

**Card System Integration**
- Use card registry for centralized card management
- Implement card validation before processing effects
- Integrate card actions with existing service layer
- Follow modular design patterns for new card types

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

### Adding New Game Features
1. **Consult game rules**: Check `TERRAFORMING_MARS_RULES.md` for any game rule implications
2. **Define domain entities** in `internal/model/` with proper `ts:` tags
3. **Implement service logic** in `internal/service/`
4. **Add WebSocket handlers** in `internal/delivery/websocket/handler/`
5. **Generate types**: Run `tygo generate` to update frontend types
6. **Frontend integration**: Import generated types and implement UI
7. **Format and lint**: **ALWAYS** run `make format` and `make lint` after completing any feature

### Backend Development Flow
1. Modify Go structs -> Add business logic -> Update handlers
2. Run `tygo generate` for type sync
3. Frontend automatically gets updated TypeScript interfaces

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

### Resource Display Components
When displaying game resources, ALWAYS use existing components instead of creating new ones:

#### Megacredits (MC)
```tsx
import CostDisplay from '../display/CostDisplay.tsx';
<CostDisplay cost={amount} size="medium" />
```
- **Component**: `src/components/ui/display/CostDisplay.tsx`
- **Sizes**: 'small' (24px), 'medium' (32px), 'large' (40px)
- **Asset**: Uses `/assets/resources/megacredit.png` with number overlay
- **Usage**: Cards, resource panels, transaction displays, player boards

#### Production
When displaying production values, use the production asset:
- **Asset**: `/assets/misc/production.png` 
- **Pattern**: Icon background with number overlay (create ProductionDisplay component if needed)

### UI Development Patterns
- **Inspect existing design language**: When updating any UI element in the frontend, other components should ALWAYS be inspected for the design language in the codebase
- **Reuse over creation**: Always check for existing components before creating new ones
- **Consistent styling**: Use established components to maintain visual consistency
- **Asset integration**: Prefer official game assets over text/CSS styling
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
npm run lint           # Check for ESLint errors
```

**Note**: These commands must be run from the respective directories (backend/ and frontend/). Always format both backend and frontend code after any changes, even if you only modified one side, to maintain consistent code quality across the entire codebase.

**Lint Error Policy**: 
- All lint ERRORS must be fixed immediately - no exceptions
- Lint warnings should be addressed when practical
- Never commit code with lint errors
- Run these commands after any significant code changes

**Logging Guidelines**:
- Use emojis in log messages where appropriate to make them more visually distinctive
- Include directional indicators for client/server communication (client→server, server→client)
- Connection logs: 🔗 for connect, ⛓️‍💥 for disconnect
- Broadcasting: 📢 for server broadcasts, 💬 for direct messages
- HTTP requests: 📡 for client requests to server
- Server lifecycle: 🚀 for startup, 🛑 for shutdown, ✅ for completion

## Development Notes

### Backend Development (Go)
- **Clean Architecture**: Always implement new features following the domain -> service -> delivery pattern
- **Type Tags**: Add both `json:` and `ts:` tags to all domain structs for frontend sync
- **WebSocket Events**: Add new action handlers in `internal/delivery/websocket/handler/` and register in the manager
- **Testing**: Use `make test` to run all backend tests
- **API Documentation**: Add Swagger comments to HTTP handlers for auto-generated docs

#### Modern Backend Patterns

**Repository Layer**
- **Immutable Interfaces**: Return values, not pointers, to prevent external state mutation
- **Event Integration**: Repository updates automatically trigger EventBus notifications
- **Clean Relationships**: Use ID references instead of embedded objects for maintainable data flow

**WebSocket Handler Development**
- **Action Handlers**: Create dedicated handlers in `internal/delivery/websocket/handler/`
- **Handler Registration**: Register new handlers in the WebSocket manager for message routing
- **Service Integration**: Use application services for business logic, not direct repository access
- **No Direct SessionManager Usage**: Handlers should call services, which then use SessionManager for broadcasting
- **Event Response**: Let the event system handle state broadcasting to clients

**Card System Development** 
- **Card Registration**: Register new cards in the card registry for centralized management
- **Effect Implementation**: Create card effects that integrate with existing game services
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
- Do not use emojis when building any design. Use assets instead. If you believe there are no assets matching what you need, ASK.
- Always whenever possible use assets from `/frontend/public/assets/`
- Always when working with MC (megacredits) display them using the `/frontend/public/assets/resources/megacredit.png` and add the number inside.
- Always when working with Production display it using the `/frontend/public/assets/misc/production.png` and add the number inside.

## Resource Display Instructions (UPDATED)
- **Megacredits**: ALWAYS use the CostDisplay component instead of raw assets
  ```tsx
  import CostDisplay from '../display/CostDisplay.tsx';
  <CostDisplay cost={amount} size="medium" />
  ```
- **Production**: When displaying production, use ProductionDisplay component
  ```tsx
  import ProductionDisplay from '../display/ProductionDisplay.tsx';
  <ProductionDisplay amount={amount} resourceType="plants" size="medium" />
  ```
- When working with production of lets say plants. You need to add the plant icon inside the production icon, then the number next to it.

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

When encountering timing or state issues, always ask: "What is the proper state flow here?" rather than "How long should I wait?"