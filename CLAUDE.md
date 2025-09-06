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
make tm          # Launch interactive CLI tool for backend testing
```

### 🧪 Testing
```bash
make test         # Run all tests (backend + frontend)
make test-backend # Run Go backend tests only
make test-verbose # Run backend tests with verbose output
make test-coverage# Generate test coverage report (backend/coverage.html)
make test-quick   # Fast test suite for development iteration
```

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

### 🛠️ CLI Tool
```bash
make tm           # Run CLI tool locally
make install-cli  # Install CLI tool globally as 'tm' command
tm                # Run installed CLI from anywhere (after install-cli)
```

### 🧰 Development Helpers
```bash
make dev-setup    # Set up development environment (go mod tidy + npm install)
make test-watch   # Watch Go files and run tests on changes (requires entr)
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
- ~~`cd backend && make test`~~ → Use `make test-backend`


## Core Architecture

### Clean Architecture Backend (Go)
The Go backend follows clean architecture principles with clear separation of concerns:

```
backend/
├── cmd/server/              # Application entry point with dependency injection
├── internal/
│   ├── domain/             # Core business entities (GameState, Player, Corporation)
│   ├── service/            # Application business rules and game logic  
│   ├── repository/         # Data access layer (in-memory game storage)
│   └── delivery/           # HTTP handlers and WebSocket hub
├── pkg/typegen/            # TypeScript type generation utilities
├── tools/                  # Code generation tools
└── docs/swagger/           # Auto-generated API documentation
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
- **Data Access**: Implement data persistence and retrieval
- **Event Publishing**: Emit domain events when data changes occur
- **External Services**: Handle integrations with external systems
- **Dependency Implementation**: Implement interfaces defined in Application layer

**Presentation Layer** (`internal/cards/`, `internal/delivery/`)
- **API Endpoints**: Handle HTTP requests and WebSocket connections
- **Request/Response Models**: DTOs for external communication
- **Card Handlers**: Implement game card effects using Application services
- **Dependency Direction**: Depends on Application layer, not Infrastructure

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

### Event-Driven Architecture Patterns

**Domain Events**
- Published by Infrastructure layer when data changes
- Carry rich payloads with before/after state
- Enable loose coupling between domain concepts
- Support audit trails and eventual consistency

**Event Handlers**
- Implemented in Application services
- React to domain events for cross-cutting concerns
- Enable features like milestone tracking and notifications
- Execute concurrently with proper error handling

**Event Flow**
```
Domain Entity → Repository (publishes event) → Event Bus → Service Handler
```

### Development Guidelines

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

**Infrastructure Layer**
- Implement data persistence with defensive copying
- Publish domain events when data changes
- Handle external service integrations
- Manage technical cross-cutting concerns

**Presentation Layer**
- Use Application services for all business operations
- Never access Infrastructure layer directly
- Implement proper error handling and validation
- Keep presentation logic separate from business logic

Reference existing code implementations for concrete examples of these architectural patterns.

## Game State Flow

### WebSocket Event Architecture
```
Client -> 'join-game' -> GameService.JoinGame -> Broadcasts 'game-updated'
Client -> 'select-corporation' -> GameService.SelectCorporation -> Broadcasts updated state
Client -> 'raise-temperature' -> GameService.RaiseTemperature -> Visual feedback
Client -> 'skip-action' -> GameUseCase.SkipAction -> Turn progression
```

### Backend Request Flow
1. WebSocket message received -> Hub.handleMessage() -> Client.handleMessage()
2. Use case method called (e.g., JoinGame, SelectCorporation)  
3. Domain logic executed -> Game state updated in repository
4. Updated GameState broadcast to all game clients
5. Frontend receives state update -> React components re-render

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

### Resource System
Six resource types: Credits, Steel, Titanium, Plants, Energy, Heat. Heat converts to temperature raises (8 heat = 1 step). Production tracks for sustainable resource generation.

## CLI Tool (`make tm`)

The project includes an interactive CLI tool for backend testing and development:

### Features
- **WebSocket Connection**: Direct connection to backend for real-time testing
- **Numbered Actions**: Interactive action selection (0-9) with skip option (0)
- **Game Management**: Join games, view status, send custom messages
- **Real-time Feedback**: Live updates from game state changes

### Usage
```bash
make tm                    # Launch CLI tool
make install-cli           # Install globally as 'tm' command
tm                        # Run from anywhere (after install)
```

### Available Commands
- `help` - Show all commands
- `connect <game>` - Join or create a game
- `status` - Show connection and game status  
- `actions` - Display numbered action list
- `0-9` - Select action by number (0 = skip)
- `send <type>` - Send raw WebSocket messages
- `quit` - Exit CLI tool

### Example Session
```bash
tm> connect test-game      # Join game
tm> actions               # Show available actions
tm> 1                     # Raise temperature
tm> 0                     # Skip action
tm> quit                  # Exit
```

The CLI tool is particularly useful for:
- Testing game mechanics without frontend
- Debugging WebSocket message flow  
- Rapid iteration on backend features
- Automated testing scenarios

## Key Development Patterns

### Adding New Game Features
1. **Define domain entities** in `internal/domain/` with proper `ts:` tags
2. **Implement service logic** in `internal/service/`
3. **Add WebSocket handlers** in `internal/delivery/websocket/`
4. **Generate types**: Run `tygo generate` to update frontend types
5. **Frontend integration**: Import generated types and implement UI
6. **Format and lint**: **ALWAYS** run `make format` and `make lint` after completing any feature

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

### WebSocket Message Types
Current supported messages:
- `join-game`: Player joins a game session
- `select-corporation`: Choose starting corporation
- `raise-temperature`: Spend heat to increase global temperature
- `skip-action`: Pass current turn
- `start-game`: Host starts the game (transitions from lobby to active status)

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
- **Full card system** implementation in Go backend
- **Tile placement** logic and adjacency bonuses
- **Turn phases** and action state machine
- **Victory condition** checking
- **Milestones and awards** game mechanics

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
- **WebSocket Events**: Add new game actions in `internal/delivery/websocket/game_hub.go`
- **Testing**: Use `go test ./...` to run all backend tests
- **API Documentation**: Add Swagger comments to HTTP handlers for auto-generated docs

### Frontend Development (React)
- **Generated Types**: Always use types from `src/types/generated/api-types.ts`
- **3D Rendering**: Uses React Three Fiber - modify scenes in `Game3DView.tsx`
- **WebSocket Client**: Game state updates come via WebSocket, no local game state
- **Component Architecture**: Follow existing patterns for new game UI components
- **Promise Handling**: Use `void <function>()` to explicitly discard promises in event handlers to avoid IDE warnings

### Full-Stack Development
- **Both servers** must be running for full functionality (`npm start`)
- **Type Generation**: Run `npm run generate-types` after Go struct changes
- **State Flow**: All game state changes originate from Go backend via WebSocket
- **Development Workflow**: Go changes -> generate types -> React implementation
- When creating mock data, abstract it from the UI to enable easier refactoring later
- NEVER set default values - if you expect something, fail explicitly if it's missing

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.

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
- Whenever you create a new feature in backend. Write a test for it.
- Whenever you move something that is checked into git. use git mv