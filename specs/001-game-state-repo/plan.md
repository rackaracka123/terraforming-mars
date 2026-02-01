# Implementation Plan: Game State Repository with Diff Logging

**Branch**: `001-game-state-repo` | **Date**: 2026-01-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-game-state-repo/spec.md`

## Summary

Implement a Game State Repository with three functions:
1. **Write** - Store game state and compute diff from previous state
2. **Get** - Retrieve current game state
3. **GetDiff** - Retrieve chronological log of all state diffs

Uses custom domain-specific diff structs (not generic JSON-patch) for type safety and frontend compatibility. In-memory storage only (no persistence required).

## Technical Context

**Language/Version**: Go 1.21+, TypeScript 5.x
**Primary Dependencies**: None (custom diff computation)
**Storage**: In-memory (map-based, per-game isolation)
**Testing**: Go test with table-driven tests in `test/` directory
**Target Platform**: Linux server (backend), Browser (frontend)
**Project Type**: Web application (Go backend + React frontend)
**Performance Goals**: Diff computation < 1ms for typical game state
**Constraints**: Thread-safe, no external dependencies
**Scale/Scope**: Single game session, ~50-200 state changes per game

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Clean Architecture & Action Pattern | ✅ Pass | Repository in `/internal/game/`, follows existing patterns |
| II. Code Clarity Over Comments | ✅ Pass | Self-documenting struct names, no explanatory comments |
| III. Best Practices Enforcement | ✅ Pass | Uses `sync.RWMutex`, follows Go idioms |
| IV. Complete Feature Implementation | ✅ Pass | All 3 functions implemented with tests |
| V. Type Safety & Generation | ✅ Pass | DTOs with `json:` and `ts:` tags, `make generate` |
| VI. Testing Discipline | ✅ Pass | Tests in `test/game/` directory |
| VII. No Deprecated Code | ✅ Pass | New code only, no deprecated patterns |

## Project Structure

### Documentation (this feature)

```text
specs/001-game-state-repo/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Research findings
├── data-model.md        # Entity definitions
├── quickstart.md        # Usage guide
├── contracts/           # Interface contract
│   └── game-state-repository.go
└── checklists/
    └── requirements.md  # Quality checklist
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── game/
│   │   ├── state_repository.go      # NEW: Interface + InMemoryGameStateRepository
│   │   └── state_diff.go            # NEW: Diff types and computation
│   └── delivery/
│       └── dto/
│           └── state_diff.go        # NEW: DTO types for serialization
└── test/
    └── game/
        └── state_repository_test.go # NEW: Unit tests
```

**Structure Decision**: Follows existing web application structure. New files added to existing directories following established patterns.

## Implementation Tasks

### Task 1: Define Diff Types (Priority: P1)

Create `backend/internal/game/state_diff.go`:
- `StateDiff` struct with sequence, timestamp, gameID, changes
- `GameChanges` struct with optional fields for all diffable state
- `DiffValue[T]` helper types for old/new value pairs
- `PlayerChanges`, `BoardChanges`, `TilePlacement` structs

### Task 2: Implement Diff Computation (Priority: P1)

Add diff computation function in `state_diff.go`:
- `ComputeDiff(old, new *dto.GameDto) *GameChanges`
- Compare each field, only populate when different
- Handle nested structures (players, board)
- Handle array diffs (cards added/removed, tiles placed)

### Task 3: Implement Repository Interface (Priority: P1)

Create `backend/internal/game/state_repository.go`:
- `GameStateRepository` interface with `Write`, `Get`, `GetDiff`
- `InMemoryGameStateRepository` implementation
- Thread-safe with `sync.RWMutex`
- Per-game state and diff log storage

### Task 4: Create DTO Types (Priority: P2)

Create `backend/internal/delivery/dto/state_diff.go`:
- Mirror diff types with `json:` and `ts:` tags
- Mapper functions for domain → DTO conversion

### Task 5: Write Tests (Priority: P1)

Create `backend/test/game/state_repository_test.go`:
- Test Write with initial state
- Test Write with state changes
- Test Get returns current state
- Test GetDiff returns chronological history
- Test thread safety with concurrent access
- Test error cases (game not found)

### Task 6: Generate TypeScript Types (Priority: P2)

- Run `make generate` to create frontend types
- Verify types in `frontend/src/types/generated/api-types.ts`

## Verification

1. **Unit Tests**: `make test` passes with new tests
2. **Lint Check**: `make lint` passes
3. **Type Generation**: `make generate` produces correct TypeScript types
4. **Manual Test**:
   - Create game, write initial state, verify diff
   - Modify state, write again, verify incremental diff
   - Call GetDiff, verify chronological order
   - Call Get, verify current state returned

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `backend/internal/game/state_diff.go` | Create | Diff types and computation |
| `backend/internal/game/state_repository.go` | Create | Repository interface and implementation |
| `backend/internal/delivery/dto/state_diff.go` | Create | DTO types with json/ts tags |
| `backend/test/game/state_repository_test.go` | Create | Unit tests |
