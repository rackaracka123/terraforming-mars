# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Digital implementation of Terraforming Mars board game with real-time multiplayer and 3D game view. The game features drag-to-pan 3D Mars visualization, hexagonal tile system, comprehensive card effects engine, and Socket.io multiplayer.

## Development Commands

### Quick Start (Recommended)
```bash
npm start        # Starts both backend and frontend servers concurrently
npm run dev      # Alias for npm start
```

### Individual Servers
```bash
npm run backend  # Backend only (port 3001)
npm run frontend # Frontend only (port 3000)
```

### Backend (Port 3001)
```bash
cd backend
npm run dev      # Development server with nodemon auto-reload
npm run build    # Compile TypeScript to dist/
npm start        # Run compiled production build
```

### Frontend (Port 3000) 
```bash
cd frontend  
npm start        # React development server
npm run build    # Production build for deployment
npm test         # Run Jest tests
```

### Both Servers
Frontend connects to backend at http://localhost:3001. Use root-level `npm start` to run both servers automatically.


## Core Architecture

### Three-Layer Game State System
1. **UI Layer** (`GameInterface.tsx`): React components with Socket.io client
2. **Network Layer** (`index.ts`): Socket.io server with event handlers  
3. **Logic Layer** (`EffectEngine`): Pure game logic for card effects and state changes

### Card System Architecture
The card system uses a compositional effect-based design:

- **Card Definitions** (`/data/cards/`): JSON-like TypeScript objects defining all cards
- **Effect Engine** (`engine/effectEngine.ts`): Interprets and executes card effects
- **Type System** (`types/cards.ts`): Complete TypeScript definitions for game entities

Key insight: Every card ability is expressed as composable `Effect` objects with `trigger`, `condition`, and `action` properties. This allows complex interactions while maintaining type safety.

### 3D Rendering System
- **Game3DView.tsx**: Main Three.js Canvas with React Three Fiber
- **HexGrid.tsx**: Hexagonal tile system for Mars board (42 hexes currently)
- **PanControls.tsx**: Custom mouse/touch controls (pan + zoom, no orbit rotation)
- **BackgroundCelestials**: Parallax layers for space environment

## Game State Flow

### Socket.io Event Architecture
```
Client -> 'join-game' -> Server creates Player -> Broadcasts 'game-updated'
Client -> 'play-card' -> EffectEngine.executeEffect -> Broadcasts updated state
Client -> 'raise-temperature' -> Parameter change -> Visual feedback
```

### Effect Engine Execution
1. Card played -> Find CardDefinition -> Create Card instance
2. EffectEngine.executeEffect() for each Effect in card
3. Check conditions -> Execute actions -> Update game state
4. Triggered effects processed via event system

## Type System Overview

### Core Entities
- `GameState`: Root state containing players, parameters, cards
- `Player`: Resources, production, played cards, terraform rating
- `CardDefinition`: Static card data with effects array
- `Effect`: Composable abilities (immediate/ongoing/triggered)
- `GlobalParameters`: Temperature (-30 to +8°C), Oxygen (0-14%), Oceans (0-9)

### Resource System
Six resource types with production tracks: Credits, Steel, Titanium, Plants, Energy, Heat. Heat converts to temperature raises (8 heat = 1 step).

## Key Development Patterns

### Adding New Cards
1. Define in appropriate `/data/cards/*.ts` file
2. Use existing Effect types or add custom functions
3. Effects are automatically processed by EffectEngine
4. Add to server's availableCards array

### Custom Card Effects
For unique abilities, use `customFunction` in Effect.action and implement in EffectEngine.executeCustomFunction().

### 3D Scene Modifications
- HexGrid positions calculated via hex-to-pixel coordinate conversion
- Mars visual state driven by GameState.globalParameters
- Custom materials respond to terraforming progress (color changes)

## Important Implementation Details

### Hex Coordinate System
Uses cube coordinates (q, r, s) where q + r + s = 0. Utilities in `HexMath` class handle conversions and neighbor calculations for tile-based game mechanics.

### Multiplayer State Synchronization
Game state is authoritative on server. Clients receive full state updates via 'game-updated' events. No client-side game logic to prevent desync.

### Effect Timing and Triggers
- `immediate`: Execute when card played
- `ongoing`: Passive effects (discounts, bonuses)
- `triggered`: Respond to game events ('card_played', 'turn_start', etc.)

### TypeScript Import Patterns
Frontend uses `.tsx` extensions in imports due to mixed JS/TS setup from Create React App template. Backend uses standard TypeScript imports.

## Current Implementation Status

### Working Systems
- Real-time multiplayer with Socket.io
- 3D game view with hexagonal Mars board
- Card effect engine with 15+ test cards
- Resource management and production
- Global parameter tracking with visual feedback
- Custom pan controls (no orbital rotation)

### Architecture Ready for Extension
- Card database (currently test cards, ready for full 200+ deck)
- Tile placement system (hex grid implemented, placement logic partial)
- Corporation asymmetry (type system ready, 10 corporations defined)
- Turn phases (structure in place, needs state machine)

### Key Missing Pieces
- Milestones and awards system
- Full tile placement with adjacency bonuses
- Card drawing and hand management
- Victory condition checking
- Game lobby and matchmaking

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

## Development Notes

When working with this codebase:
- Both servers must be running for full functionality
- Backend TypeScript compiles to `dist/` directory
- Frontend uses React 19 with Three.js for 3D rendering
- Game state changes should go through EffectEngine for consistency
- Card effects are composable - combine existing patterns rather than creating new ones
- When creating mock. make sure to abstract it from the UI so it does not know, to ensure easier refactoring to real data later.
- NEVER set a default value, if you expect something, crash if you dont have it.

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.

## Resource Display Instructions (UPDATED)
- **Megacredits**: ALWAYS use the CostDisplay component instead of raw assets
  ```tsx
  import CostDisplay from '../display/CostDisplay.tsx';
  <CostDisplay cost={amount} size="medium" />
  ```
- **Production**: When displaying production, use `/assets/misc/production.png` with number overlay (create ProductionDisplay component if needed)