# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Digital implementation of Terraforming Mars board game with real-time multiplayer and 3D game view. The game features drag-to-pan 3D Mars visualization, hexagonal tile system, comprehensive card effects engine, and WebSocket multiplayer with a Go backend and React frontend.

## Development Commands

### Quick Start (Recommended)
```bash
make run         # Alternative: Run both servers using Makefile (must be in root directory)
```

### Individual Servers
```bash
npm run backend  # Go backend only (port 3001)
npm run frontend # React frontend only (port 3000)
```

### Backend (Go - Port 3001)
```bash
cd backend
go run cmd/server/main.go     # Run development server directly
go build -o bin/server cmd/server/main.go  # Build production binary
./bin/server                  # Run production binary
make test                     # Run all tests
go generate                   # Generate TypeScript types and Swagger docs
```

### Frontend (React - Port 3000) 
```bash
cd frontend  
npm start        # React development server
npm run build    # Production build for deployment
npm test         # Run Jest tests
```

### Code Generation
```bash
npm run generate-docs    # Generate Swagger API documentation
```

### Both Servers
Frontend connects to backend at http://localhost:3001. Use root-level `npm start` to run both servers automatically.


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
- **Code Generation**: `go generate` creates TypeScript interfaces
- **Frontend Import**: React components use generated types for full type safety

### 3D Rendering System
- **Game3DView.tsx**: Main Three.js Canvas with React Three Fiber
- **HexGrid.tsx**: Hexagonal tile system for Mars board (42 hexes currently)
- **PanControls.tsx**: Custom mouse/touch controls (pan + zoom, no orbit rotation)
- **BackgroundCelestials**: Parallax layers for space environment

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

## Key Development Patterns

### Adding New Game Features
1. **Define domain entities** in `internal/domain/` with proper `ts:` tags
2. **Implement service logic** in `internal/service/`
3. **Add WebSocket handlers** in `internal/delivery/websocket/`
4. **Generate types**: Run `go generate` to update frontend types
5. **Frontend integration**: Import generated types and implement UI

### Backend Development Flow
1. Modify Go structs -> Add business logic -> Update handlers
2. Run `go generate` for type sync
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
- **Reuse over creation**: Always check for existing components before creating new ones
- **Consistent styling**: Use established components to maintain visual consistency
- **Asset integration**: Prefer official game assets over text/CSS styling
- **Responsive sizing**: Components should support multiple sizes for different contexts

## Code Quality Requirements

**CRITICAL**: Always run these commands before completing any feature or pushing code:

```bash
npm run format:write    # Format code with Prettier  
npm run lint           # Check for ESLint errors
```

**Lint Error Policy**: 
- All lint ERRORS must be fixed immediately - no exceptions
- Lint warnings should be addressed when practical
- Never commit code with lint errors
- Run these commands after any significant code changes

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