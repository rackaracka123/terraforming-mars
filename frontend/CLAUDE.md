# Frontend - Terraforming Mars Game UI

This document provides guidance for working with the React frontend application.

## Overview

React-based frontend with 3D Mars visualization using Three.js and React Three Fiber. Provides real-time multiplayer game interface with WebSocket state synchronization from Go backend.

## Architecture

### Application Structure

```
frontend/
├── src/
│   ├── components/        # React components
│   │   ├── cards/         # Card display components
│   │   ├── game/          # Core game UI (board, resources, etc.)
│   │   ├── pages/         # Top-level page components
│   │   ├── three/         # Three.js/R3F 3D rendering
│   │   └── ui/            # Reusable UI components
│   ├── contexts/          # React contexts for state management
│   ├── hooks/             # Custom React hooks
│   ├── services/          # API and WebSocket services
│   ├── types/             # TypeScript type definitions
│   │   └── generated/     # Auto-generated from Go backend
│   ├── utils/             # Utility functions and helpers
│   ├── index.css          # Global styles and Tailwind config
│   └── main.tsx           # Application entry point
├── public/
│   ├── assets/            # Static game assets (images, icons)
│   └── models/            # 3D models for Three.js
└── playwright.config.ts   # End-to-end testing configuration
```

### Component Organization

**Pages** (`components/pages/`)
- Top-level route components
- Compose multiple features
- Handle page-level state and routing

**Game Components** (`components/game/`)
- Core gameplay UI elements
- Resource displays, action panels, player boards
- Direct game state visualization

**Card Components** (`components/cards/`)
- Card display and interaction
- Corporation selection
- Card effect visualization

**Three.js Components** (`components/three/`)
- 3D Mars visualization with React Three Fiber
- Hex grid rendering
- Custom camera controls (pan + zoom, no orbit)
- Background celestial objects

**UI Components** (`components/ui/`)
- Reusable design system components
- Buttons, panels, modals, icons
- Consistent styling and behavior

### State Management

**WebSocket-Driven State**
- **No local game state**: Backend is source of truth
- **Real-time synchronization**: All state changes via WebSocket events
- **Unidirectional flow**: Server → WebSocket → React state → UI

**React Contexts**
- Game state from WebSocket
- Player connection status
- UI state (modals, selections)

**localStorage**
- Game ID and player ID persistence
- Automatic reconnection on page reload
- Session recovery after browser close

## Development Workflow

### Running the Frontend

```bash
# From project root
make frontend             # Vite dev server (port 3000)
make run                  # Run both frontend and backend

# From frontend/
npm run dev               # Start dev server
npm run build             # Production build
npm run preview           # Preview production build
```

### Code Quality

```bash
# From project root
make lint-frontend        # Run ESLint
make format               # Format all code

# From frontend/
npm run lint              # ESLint check
npm run format:write      # Prettier format
npm run format:check      # Check formatting
```

**CRITICAL**: Always run `make format` and `make lint` after code changes. Fix all lint ERRORS immediately.

### Type Generation

Frontend consumes TypeScript types generated from Go backend:

```bash
# From project root
make generate             # Generate types from Go structs
```

Types appear in `src/types/generated/api-types.ts`. Import and use them:

```typescript
import { Player, GameState, Corporation } from '../types/generated/api-types';
```

## Key Development Patterns

### Component Development

**UI Component Standards**

1. **Inspect existing design language** before creating new components
2. **Reuse over creation**: Check for existing components first
3. **Consistent styling**: Use Tailwind utilities and theme classes
4. **No emojis in UI**: Use GameIcon component or assets instead

**Component Template**

```tsx
import { FC } from 'react';
import GameIcon from '../ui/display/GameIcon';

interface MyComponentProps {
    data: SomeType;
    onAction: (id: string) => void;
}

const MyComponent: FC<MyComponentProps> = ({ data, onAction }) => {
    return (
        <div className="flex items-center gap-2 p-4 bg-space-black rounded-lg">
            <GameIcon iconType="credits" amount={data.amount} size="medium" />
            <button
                onClick={() => void onAction(data.id)}
                className="px-4 py-2 bg-space-blue-600 hover:bg-space-blue-500 rounded"
            >
                Action
            </button>
        </div>
    );
};

export default MyComponent;
```

**Note**: Use `void <function>()` to explicitly discard promises in event handlers.

### Icon Display - GameIcon Component

**CRITICAL**: ALWAYS use GameIcon component for ANY game icon. NEVER use direct `<img>` tags.

```tsx
import GameIcon from '../ui/display/GameIcon';

// Basic resource icons
<GameIcon iconType="steel" size="medium" />
<GameIcon iconType="plants" size="small" />

// Icons with amounts
<GameIcon iconType="credits" amount={25} size="large" />

// Production resources (automatic brown background)
<GameIcon iconType="energy-production" amount={3} size="medium" />

// Card tags
<GameIcon iconType="space" size="medium" />

// Tiles and global parameters
<GameIcon iconType="ocean-tile" size="small" />
<GameIcon iconType="temperature" size="medium" />
```

**Sizes**: 'small' (24px), 'medium' (32px), 'large' (40px)

**Adding New Icons**: Add to `src/utils/iconStore.ts` in appropriate category (RESOURCE_ICONS, TAG_ICONS, SPECIAL_ICONS), then use via GameIcon.

### Styling with Tailwind CSS v4

**CRITICAL**: This project uses Tailwind CSS v4 with CSS-based configuration.

**Configuration**: `src/index.css` contains `@theme {}` block
**NO CSS Modules**: NEVER create `.module.css` files
**NO JavaScript Config**: `tailwind.config.js` is ignored

**Custom Theme Utilities** (in `index.css`):

```css
@theme {
  /* Colors */
  --color-space-black: #0a0a0f;
  --color-space-blue-600: rgba(30, 60, 150, 0.8);

  /* Typography */
  --font-family-orbitron: Orbitron, sans-serif;

  /* Shadows */
  --shadow-glow: 0 0 20px rgba(59, 130, 246, 0.5);
}
```

**Usage**:

```tsx
<div className="bg-space-black font-orbitron shadow-glow">
    <h1 className="text-shadow-glow-strong tracking-wider-2xl">
        Terraforming Mars
    </h1>
</div>
```

**Arbitrary Values**:

```tsx
<div className="bg-[rgba(10,20,40,0.95)] backdrop-blur-[10px]">
```

**Adding New Theme Values**: Add to `@theme {}` in `index.css`.

### 3D Rendering with Three.js

**React Three Fiber Pattern**

```tsx
import { Canvas } from '@react-three/fiber';
import { OrbitControls } from '@react-three/drei';

const Game3DView: FC = () => {
    return (
        <Canvas camera={{ position: [0, 10, 10], fov: 60 }}>
            <ambientLight intensity={0.5} />
            <pointLight position={[10, 10, 10]} />

            {/* 3D content */}
            <HexGrid />
            <PanControls />
            <BackgroundCelestials />
        </Canvas>
    );
};
```

**Hex Coordinate System**

Uses cube coordinates (q, r, s) where q + r + s = 0.

```typescript
import { HexMath } from '../utils/hexMath';

// Convert hex to pixel position
const position = HexMath.hexToPixel(q, r, hexSize);

// Get hex neighbors
const neighbors = HexMath.getNeighbors(q, r);
```

**Custom Controls**

- **PanControls**: Mouse/touch panning and zoom
- **No orbital rotation**: Camera locked to top-down view
- **Parallax background**: Celestials move with camera for depth

### WebSocket Communication

**WebSocket Service** (`services/websocketService.ts`)

Handles real-time game state synchronization.

```typescript
import { websocketService } from '../services/websocketService';

// Connect to game
websocketService.connect(gameId, playerId, playerName);

// Send action
websocketService.send({
    type: 'raise-temperature',
    data: { amount: 1 }
});

// Listen for updates
websocketService.onMessage((message) => {
    if (message.type === 'game-updated') {
        updateGameState(message.data);
    }
});

// Disconnect
websocketService.disconnect();
```

**Message Types**

**Outbound (Client → Server)**:
- `join-game`: Join or create game session
- `player-reconnect`: Reconnect existing player
- `select-corporation`: Choose corporation
- `raise-temperature`: Increase global temperature
- `skip-action`: Pass turn
- `start-game`: Host starts game from lobby

**Inbound (Server → Client)**:
- `game-updated`: Full game state synchronization
- `player-connected`: New player joined
- `player-reconnected`: Player reconnected
- `player-disconnected`: Player lost connection

### Session Persistence

**localStorage Integration**

```typescript
// Save game session
localStorage.setItem('terraformingMarsGameData', JSON.stringify({
    gameId,
    playerId,
    playerName
}));

// Restore session
const savedData = localStorage.getItem('terraformingMarsGameData');
if (savedData) {
    const { gameId, playerId, playerName } = JSON.parse(savedData);
    // Reconnect
}

// Clear session
localStorage.removeItem('terraformingMarsGameData');
```

**Automatic Reconnection**

1. User reloads page or reopens browser
2. Component checks localStorage for game data
3. Fetches current game state via HTTP API
4. Reconnects WebSocket with `player-reconnect` message
5. Full state restored seamlessly

## Data Flow

### Game State Synchronization

```
Backend State Change → EventBus → WebSocket Hub → Broadcaster
                                                        ↓
                                                   Client receives 'game-updated'
                                                        ↓
                                                   WebSocket service updates React state
                                                        ↓
                                                   Components re-render
```

### User Action Flow

```
User clicks button → Component handler → WebSocket send message
                                              ↓
                                         Backend processes action
                                              ↓
                                         Repository updates state
                                              ↓
                                         EventBus publishes events
                                              ↓
                                         All clients receive 'game-updated'
                                              ↓
                                         UI updates automatically
```

## Testing with Playwright

**Interactive Debugging**

Use Playwright MCP server for live debugging:

```bash
# Ensure servers are running
make run

# Use Playwright MCP tools to:
# - Navigate to http://localhost:3000
# - Interact with UI elements
# - Take snapshots and screenshots
# - Monitor console messages
# - Test user flows
```

**Playwright MCP Tools**:
- `mcp__playwright__browser_navigate`: Navigate to URLs
- `mcp__playwright__browser_snapshot`: Capture page state
- `mcp__playwright__browser_click`: Click elements
- `mcp__playwright__browser_type`: Type into forms
- `mcp__playwright__browser_take_screenshot`: Capture visuals
- `mcp__playwright__browser_evaluate`: Run JavaScript

**Debugging Use Cases**:
- UI rendering issues
- State synchronization problems
- User flow testing
- WebSocket communication
- Performance bottlenecks
- Visual regressions

## Common Tasks

### Adding a New Page

1. Create page component in `components/pages/`
2. Add route in main router
3. Implement page layout and features
4. Connect to WebSocket service if needed

### Creating a New Game Feature

1. **Check generated types**: Import from `types/generated/api-types.ts`
2. **Design UI components**: Follow existing design patterns
3. **Connect WebSocket**: Send actions and listen for updates
4. **Handle state updates**: React to `game-updated` events
5. **Test with Playwright**: Verify user flows work correctly

### Adding a New Icon

1. Place asset in `public/assets/` (resources/, tags/, tiles/, etc.)
2. Add path to `src/utils/iconStore.ts` in appropriate category
3. Use via GameIcon component: `<GameIcon iconType="newIcon" size="medium" />`

### Updating After Backend Changes

1. Backend updates Go structs with `ts:` tags
2. Run `make generate` from project root
3. Check `src/types/generated/api-types.ts` for new types
4. Update components to use new types
5. Run `make format` and `make lint`

## Important Notes

### State Management Rules

**CRITICAL**: No timeouts or arbitrary delays for state synchronization.

```tsx
// ❌ WRONG: Don't add delays for state updates
useEffect(() => {
    setTimeout(() => {
        // Wait for state to update
    }, 1000);
}, []);

// ✅ CORRECT: Use proper event handling
useEffect(() => {
    const handleUpdate = (message: WebSocketMessage) => {
        if (message.type === 'game-updated') {
            setGameState(message.data);
        }
    };

    websocketService.onMessage(handleUpdate);
    return () => websocketService.offMessage(handleUpdate);
}, []);
```

### Design Principles

- **No emojis in UI**: Use GameIcon or assets
- **GameIcon first**: Never use direct `<img src="/assets/...">` tags
- **Tailwind CSS only**: No CSS Modules or custom `.module.css` files
- **Consistent components**: Reuse existing UI components
- **Type safety**: Always use generated types from backend

### Mock Data

- Abstract mock data from UI components
- Enable easier refactoring later
- Never set default values - fail explicitly if data is missing

### Development Best Practices

- **Promise handling**: Use `void <function>()` in event handlers
- **Type imports**: Import from generated types
- **Responsive design**: Support different screen sizes
- **Accessibility**: Use semantic HTML and ARIA labels
- **Performance**: Optimize re-renders with React.memo when needed

## Dependencies

### Core Libraries

- **React 18**: UI framework
- **React Router**: Client-side routing
- **Three.js**: 3D rendering engine
- **React Three Fiber**: React renderer for Three.js
- **@react-three/drei**: Useful helpers for R3F
- **Tailwind CSS v4**: Utility-first CSS framework

### Development Tools

- **Vite**: Build tool and dev server
- **TypeScript**: Type safety
- **ESLint**: Code linting
- **Prettier**: Code formatting
- **Playwright**: End-to-end testing

## Related Documentation

- **Project Root CLAUDE.md**: Full-stack architecture and workflows
- **backend/CLAUDE.md**: Backend API architecture and patterns
- **backend/assets/terraforming_mars_cards.json**: Authoritative card definitions (manually edited)
- **TERRAFORMING_MARS_RULES.md**: Complete game rules reference
