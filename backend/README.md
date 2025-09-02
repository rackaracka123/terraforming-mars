# Terraforming Mars Go Backend

A high-performance Go backend for the Terraforming Mars digital board game implementation.

## Architecture

This backend follows clean architecture principles with automatic API documentation and TypeScript type generation.

```
backend/
├── cmd/server/              # Application entry point
├── internal/
│   ├── domain/             # Core business logic and entities
│   ├── usecase/            # Application business rules
│   ├── repository/         # Data access layer
│   └── delivery/           # HTTP and WebSocket handlers
├── pkg/                    # Shared utilities
├── tools/                  # Code generation tools
└── docs/                   # Generated documentation
```

## Features

- **Real-time WebSocket Communication**: Game state synchronization
- **RESTful API**: HTTP endpoints for game operations
- **Automatic TypeScript Generation**: Go structs → TypeScript interfaces
- **OpenAPI Documentation**: Auto-generated from Go annotations
- **Clean Architecture**: Separation of concerns with dependency injection
- **In-memory Game Storage**: Fast, concurrent-safe game state management

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Node.js and npm (for frontend integration)

### Installation

1. Install dependencies:
```bash
go mod tidy
```

2. Install Swagger CLI:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Development

1. **Run the server**:
```bash
go run cmd/server/main.go
```

2. **Generate types and docs**:
```bash
go generate
```

3. **Run from root project**:
```bash
npm run backend
```

## API Endpoints

### HTTP Endpoints

- `GET /health` - Health check
- `GET /api/v1/games/{id}` - Get game state
- `GET /api/v1/corporations` - Get available corporations
- `GET /swagger/index.html` - API documentation

### WebSocket Events

Connect to `ws://localhost:3001/ws` and send JSON messages:

- `join-game` - Join a game
- `select-corporation` - Choose corporation
- `raise-temperature` - Spend heat to increase temperature
- `skip-action` - Pass turn

## Code Generation

### TypeScript Types

Generate TypeScript interfaces from Go structs:

```bash
go run tools/generate-types.go
```

Types are output to: `../frontend/src/types/generated/api-types.ts`

### OpenAPI Documentation

Generate Swagger documentation:

```bash
swag init -g cmd/server/main.go -o ./docs/swagger
```

### Automatic Generation

Run both generators:

```bash
go generate
```

## Project Integration

This backend integrates with the existing Terraforming Mars project:

- **Frontend**: React app connects via WebSocket and HTTP
- **Types**: Automatically synced TypeScript interfaces
- **Scripts**: npm scripts handle Go compilation and execution
- **Development**: Hot reload via `concurrently` in root package.json

## Development Workflow

1. **Modify Go structs** in `internal/domain/`
2. **Update business logic** in `internal/usecase/`
3. **Add API endpoints** in `internal/delivery/http/`
4. **Add Swagger comments** for documentation
5. **Run `go generate`** to update types and docs
6. **Frontend automatically** gets new TypeScript types

## Environment Variables

- `PORT` - Server port (default: 3001)

## Production Build

```bash
go build -o bin/server cmd/server/main.go
./bin/server
```

## Testing

```bash
go test ./...
```

## Contributing

1. Follow clean architecture patterns
2. Add Swagger comments to new endpoints
3. Use struct tags for TypeScript generation
4. Run `go generate` after struct changes
5. Test WebSocket functionality with frontend