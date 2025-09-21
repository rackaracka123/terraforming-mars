# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terraforming Mars 3D Edition: A digital implementation of the board game with 3D visualization and real-time multiplayer. The project consists of:

- **Frontend**: React/TypeScript with Three.js 3D graphics (port 3000)
- **Backend**: Kotlin/Ktor server with WebSocket multiplayer (port 3001)

## Development Commands

### Frontend (React/Vite)
```bash
cd frontend
npm run dev          # Start development server (port 3000)
npm run build        # Production build
npm run lint         # ESLint checking
npm run lint:fix     # Auto-fix ESLint issues
npm run typecheck    # TypeScript type checking
npm run format:write # Format code with Prettier
npm test             # Run tests
```

### Backend (Kotlin/Gradle)
```bash
cd backend
gradle run           # Start backend server (port 3001)
gradle run --continuous # Run with auto-restart on file changes (RECOMMENDED)
gradle build         # Build project
gradle test          # Run tests
gradle clean         # Clean build artifacts
```

### Both Servers
The README mentions a Makefile for unified commands, but it's currently deleted. Individual commands work in their respective directories.

## Architecture

### Backend (Kotlin/Ktor)
- **Clean Architecture**: Domain-driven design with clear separation of concerns
- **Technology Stack**: Ktor 3.0.1, Kotlin 2.2.0, Kotlinx Serialization, WebSockets
- **Structure**:
  - `dto/` - Data transfer objects for API/WebSocket communication
  - `models/` - Domain models and game entities
  - `services/` - Business logic and use cases
  - `repositories/` - Data access layer
  - `routes/` - HTTP endpoints
  - `websocket/` - Real-time multiplayer WebSocket handling
  - `actions/` - Game actions and commands

### Frontend (React/TypeScript)
- **Technology Stack**: React 19, TypeScript, Vite, Three.js/React Three Fiber
- **Structure**:
  - `src/components/game/` - 3D game view components (Mars board, hex tiles)
  - `src/components/layout/` - Main layout and panels
  - `src/components/pages/` - Route pages (Create, Landing, Join)
  - `src/components/ui/` - Reusable UI components
  - `src/services/` - API and WebSocket services
  - `src/types/generated/` - Auto-generated TypeScript types from backend
  - `src/contexts/` - React contexts for state management

### Type Safety & Code Generation
- Backend Kotlin structs generate frontend TypeScript types
- Generation commands mentioned but specific tooling needs verification
- Frontend imports generated types for full type safety

### Real-time Architecture
- **WebSocket Integration**: Backend broadcasts game state changes
- **Event-Driven**: Domain events for real-time synchronization
- **Connection Management**: Auto-reconnection and persistence via localStorage

## Key Development Patterns

### 3D Visualization
- **React Three Fiber**: Declarative Three.js in React components
- **Hex Coordinate System**: Cube coordinates (q, r, s) for 42-hex Mars board
- **Performance**: Optimized for 60fps with manual chunk splitting for Three.js vendors
- **Controls**: Pan/zoom only (no orbit rotation)

### State Management
- **Backend**: Redux-style state management with unified GameReducer
- **Frontend**: React Context + WebSocket-driven state updates
- **Persistence**: localStorage for game reconnection

### API Design
- **REST**: Game CRUD operations via HTTP
- **WebSocket**: Real-time game state and multiplayer actions
- **CORS**: Configured for development cross-origin requests
- **Proxy**: Vite dev server proxies `/api` and `/socket.io` to backend

## File Locations

### Configuration Files
- `backend/build.gradle.kts` - Kotlin/Gradle build configuration
- `backend/gradle.properties` - Gradle properties (versions)
- `backend/src/main/resources/application.conf` - Ktor server configuration
- `frontend/package.json` - NPM dependencies and scripts
- `frontend/vite.config.ts` - Vite development and build configuration
- `frontend/tsconfig.json` - TypeScript configuration

### Game Assets
- `assets/` - Game data files (cards.csv, behaviors.csv, etc.)
- `backend/src/main/resources/assets/` - Backend asset copies
- `frontend/public/assets/` - Frontend static assets

### Testing
- `backend/src/test/kotlin/` - Kotlin test files
- `frontend/tests/` - Frontend test files (Playwright configuration exists)

## Port Configuration
- **Frontend**: 3000 (Vite dev server)
- **Backend**: 3001 (Ktor server)
- **Proxy**: Frontend proxies API calls to backend during development

## Build and Deployment
- **Frontend**: Vite builds to `frontend/build/` directory
- **Backend**: Gradle builds JVM application with main class `com.terraformingmars.ApplicationKt`
- **Java Version**: JVM 22 target compatibility

## Development Workflow Notes
- You can debug frontend with playwright mcp server.
- Do not mix backend models with dto. models should have no dependencies outside of the model folder.
- Do not create mutable lists. unless its a private varaible and is not used anywhere else.