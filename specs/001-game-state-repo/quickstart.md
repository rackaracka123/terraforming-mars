# Quickstart: Game State Repository

**Feature Branch**: `001-game-state-repo`
**Created**: 2026-01-30

## Overview

The Game State Repository provides three operations for managing game state with automatic diff computation:

1. **Write** - Store state and compute diff from previous state
2. **Get** - Retrieve current state
3. **GetDiff** - Retrieve chronological log of all state changes

## Usage Examples

### Basic Usage

```go
// Create repository
repo := NewInMemoryGameStateRepository()

// Write initial state (computes diff from empty)
diff1, err := repo.Write(ctx, "game-123", initialState)
// diff1.SequenceNumber = 1, diff1.Changes contains full initial state

// Write updated state (computes diff from previous)
diff2, err := repo.Write(ctx, "game-123", updatedState)
// diff2.SequenceNumber = 2, diff2.Changes contains only what changed

// Get current state
state, err := repo.Get(ctx, "game-123")

// Get full diff history
diffs, err := repo.GetDiff(ctx, "game-123")
// diffs = [diff1, diff2] in chronological order
```

### Integration with Existing Actions

```go
// In an action that modifies game state
func (a *PlayCardAction) Execute(ctx context.Context, gameID, playerID, cardID string) error {
    // Get game and modify state
    game, _ := a.gameRepo.Get(ctx, gameID)
    // ... perform game logic ...

    // Convert to DTO and write to state repository
    gameDto := dto.ToGameDto(game, a.cardRegistry, playerID)
    diff, err := a.stateRepo.Write(ctx, gameID, gameDto)
    if err != nil {
        return err
    }

    // diff now contains exactly what changed
    return nil
}
```

### Accessing Diff Details

```go
diff, _ := repo.Write(ctx, gameID, newState)

// Check what changed
if diff.Changes.Temperature != nil {
    fmt.Printf("Temperature changed from %d to %d\n",
        diff.Changes.Temperature.Old,
        diff.Changes.Temperature.New)
}

if diff.Changes.PlayerChanges != nil {
    for playerID, changes := range diff.Changes.PlayerChanges {
        if changes.Credits != nil {
            fmt.Printf("Player %s credits: %d → %d\n",
                playerID,
                changes.Credits.Old,
                changes.Credits.New)
        }
    }
}
```

## File Locations

| File | Purpose |
|------|---------|
| `internal/game/state_repository.go` | Interface and in-memory implementation |
| `internal/game/state_diff.go` | Diff types and computation logic |
| `internal/delivery/dto/state_diff.go` | DTO types for diff serialization |
| `test/game/state_repository_test.go` | Unit tests |

## Key Design Decisions

1. **Custom diff structs** over generic JSON-patch for type safety
2. **GameDto as diffable state** (already serializable, has json/ts tags)
3. **In-memory storage** (no persistence required per spec)
4. **Thread-safe** with RWMutex (consistent with existing patterns)
5. **Separate repository** (doesn't modify existing GameRepository)

## Testing

```bash
# Run state repository tests
make test

# Or specifically
go test ./test/game/state_repository_test.go -v
```

## Generated TypeScript Types

After running `make generate`, the following types will be available in frontend:

```typescript
// frontend/src/types/generated/api-types.ts
interface StateDiff {
    sequenceNumber: number;
    timestamp: string;
    gameId: string;
    changes: GameChanges;
}

interface GameChanges {
    status?: DiffValueString;
    phase?: DiffValueString;
    generation?: DiffValueInt;
    temperature?: DiffValueInt;
    oxygen?: DiffValueInt;
    oceans?: DiffValueInt;
    playerChanges?: Record<string, PlayerChanges>;
    boardChanges?: BoardChanges;
}
```
