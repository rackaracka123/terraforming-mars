# Terraforming Mars: 3D Edition

A digital implementation of the board game Terraforming Mars with a unique 3D game view and comprehensive multiplayer support.

## Features

- **3D Game View**: Drag and rotate your view of Mars in a beautiful 3D space
- **Parallax Background**: Earth, moons, and stars move at different depths
- **Complete Card System**: All 200+ project cards with unique abilities
- **Real-time Multiplayer**: Play with friends online
- **Visual Terraforming**: Watch Mars transform as you play

## Project Structure

- `frontend/` - React/TypeScript client with Three.js 3D graphics
- `backend/` - Node.js/Express server with Socket.io for multiplayer

## Quick Start

### Backend
```bash
cd backend
npm install
npm run dev
```

### Frontend
```bash
cd frontend
npm install
npm start
```

## Technology Stack

- **Frontend**: React, TypeScript, Three.js, React Three Fiber, Redux Toolkit
- **Backend**: Node.js, Express, Socket.io, TypeScript
- **3D Graphics**: Three.js with custom 3D game view controls
- **State Management**: Redux Toolkit for complex game state