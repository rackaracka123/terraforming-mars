# Terraforming Mars: 3D Edition

A digital implementation of the board game Terraforming Mars with a unique 3D game view and comprehensive multiplayer support.

## 🚀 Features

- **3D Game View**: Interactive Mars board with hex-based tile system
- **Real-time Multiplayer**: WebSocket-based multiplayer with Go backend
- **Clean Architecture**: Backend follows domain-driven design principles  
- **Type Safety**: Automatic TypeScript generation from Go structs
- **Visual Terraforming**: Watch Mars transform as global parameters change

## 📁 Project Structure

- `frontend/` - React/TypeScript client with Three.js 3D graphics
- `backend/` - Go server with clean architecture and WebSocket multiplayer
- `Makefile` - Unified development commands

## ⚡ Quick Start

### Run Both Servers
```bash
make run
```
This starts both frontend (port 3000) and backend (port 3001) servers.

### Individual Servers
```bash
make frontend    # React development server
make backend     # Go backend server
```

## 🛠️ Development Commands

### Main Commands
- `make run` - Start both servers
- `make frontend` - Start frontend only
- `make backend` - Start backend only  
- `make tm` - Launch CLI tool

### Testing
- `make test` - Run all tests
- `make test-backend` - Backend tests only
- `make test-verbose` - Verbose backend tests
- `make test-coverage` - Coverage report

### Code Quality
- `make lint` - Run all linters
- `make format` - Format all code
- `make generate` - Generate TypeScript types

### Build & Deploy
- `make build` - Production builds
- `make clean` - Clean artifacts

## 🏗️ Technology Stack

- **Frontend**: React, TypeScript, Three.js, React Three Fiber
- **Backend**: Go, Gorilla WebSocket, Clean Architecture, Redux-style State Management
- **3D Graphics**: Three.js with custom hex coordinate system
- **Type Safety**: Go-to-TypeScript code generation
- **Multiplayer**: WebSocket real-time communication
- **State Management**: Unified reducer pattern with action dispatching

## 🎮 Game Architecture

### Backend (Go)
- **Domain Layer**: Core game entities and business logic
- **Service Layer**: Use cases and application logic
- **Store Layer**: Redux-style state management with unified GameReducer
- **Delivery Layer**: HTTP handlers and WebSocket hub
- **Repository Layer**: In-memory game state storage

### Frontend (React)
- **3D Rendering**: React Three Fiber with custom pan controls
- **Game State**: WebSocket-driven state updates with generated TypeScript types
- **Real-time Updates**: WebSocket client receives state changes from backend store
- **Component Architecture**: Modular UI with asset integration

### State Management Flow
1. **Actions**: Typed actions dispatched through WebSocket or HTTP
2. **GameReducer**: Single reducer handles all game and player state changes
3. **Store**: Immutable state updates with middleware pipeline
4. **Events**: Domain events published for real-time synchronization
5. **WebSocket**: State changes broadcast to all connected clients

## 📋 Development Workflow

1. **Make changes** to Go domain models with `ts:` tags
2. **Run `make generate`** to update TypeScript types  
3. **Implement frontend** using generated types
4. **Test with CLI** using `make tm` for backend interaction
5. **Run `make lint`** before committing changes