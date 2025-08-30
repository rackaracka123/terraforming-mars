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

### Testing & Development Tools
```bash
./capture_display.sh   # Capture screenshot for UI testing (saves to ~/tmp-feedback-loop/)
```
**Note**: Before running capture_display.sh for testing, ensure you're ready as it will take a screenshot of your current display.

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

## Development Notes

When working with this codebase:
- Both servers must be running for full functionality
- Backend TypeScript compiles to `dist/` directory
- Frontend uses React 19 with Three.js for 3D rendering
- Game state changes should go through EffectEngine for consistency
- Card effects are composable - combine existing patterns rather than creating new ones
- Update your testing and development tools to make it mandatory if you use a refrence image to create some design. to use this eval method of screenshoting a thing and compare it to the desired outcome