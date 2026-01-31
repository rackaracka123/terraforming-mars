# Research: Game State Repository with Diff Logging

**Feature Branch**: `001-game-state-repo`
**Created**: 2026-01-30

## Research Tasks

### 1. Existing Repository Pattern Analysis

**Question**: How is game state currently stored and accessed?

**Findings**:
- `InMemoryGameRepository` in `backend/internal/game/repository.go` manages active games
- Uses `sync.RWMutex` for thread-safe access
- Interface: `Get`, `Create`, `Delete`, `List`, `Exists`
- Each `Game` contains all state (players, board, deck, global parameters)
- Game has private fields with public accessor methods
- State mutations publish domain events via per-game `EventBus`

**Decision**: Follow existing repository pattern - create a new repository interface with `Write`, `Get`, `GetDiff` methods.

**Rationale**: Maintains architectural consistency with existing codebase patterns.

---

### 2. Diff Computation Library Selection

**Question**: Which Go library should be used for computing diffs between game states?

**Alternatives Considered**:

| Library | Performance | Features | Maintenance |
|---------|-------------|----------|-------------|
| viant/godiff | 815.9 ns/op, 768 B | Reflection-based, struct tags | Active |
| r3labs/diff | 4,531 ns/op, 2,262 B | Full diff/patch/merge | Active |
| wI2L/jsondiff | 1,300-11,800 ns | RFC 6902 JSON Patch | Active |
| Custom struct | <500 ns | Tailored to GameDto | N/A |

**Decision**: Use **custom diff struct** tailored to the game domain.

**Rationale**:
1. Game state is well-defined with known structure (GameDto)
2. Custom diff struct avoids reflection overhead
3. Type-safe diffs are easier to work with in both Go and TypeScript
4. No external dependency needed
5. Better performance for real-time game use case
6. Aligns with constitution principle: "avoid over-engineering"

---

### 3. Diff Structure Design

**Question**: How should diffs be structured for complex nested state?

**Findings**:
- Game state has nested objects: Players (with resources, cards), Board (with tiles), GlobalParameters
- JSON-patch style (RFC 6902) uses path-based operations: `[{op: "replace", path: "/players/0/credits", value: 150}]`
- Domain-specific approach uses typed structs: `PlayerDiff{Credits: &DiffValue{Old: 100, New: 150}}`

**Decision**: Use **domain-specific typed diff structs** with optional fields.

**Rationale**:
1. Type-safe (compile-time checks)
2. Self-documenting (field names describe what changed)
3. Easy to generate TypeScript types via `make generate`
4. Aligns with existing codebase patterns (DTOs with json/ts tags)
5. Frontend can render changes with proper UI (not generic JSON paths)

**Structure**:
```go
type StateDiff struct {
    SequenceNumber int64                `json:"sequenceNumber" ts:"number"`
    Timestamp      time.Time            `json:"timestamp" ts:"string"`
    GameID         string               `json:"gameId" ts:"string"`
    Changes        *GameChanges         `json:"changes" ts:"GameChanges"`
}

type GameChanges struct {
    Phase            *DiffValue[string]           `json:"phase,omitempty" ts:"DiffValue<string> | undefined"`
    Generation       *DiffValue[int]              `json:"generation,omitempty" ts:"DiffValue<number> | undefined"`
    Temperature      *DiffValue[int]              `json:"temperature,omitempty" ts:"DiffValue<number> | undefined"`
    Oxygen           *DiffValue[int]              `json:"oxygen,omitempty" ts:"DiffValue<number> | undefined"`
    Oceans           *DiffValue[int]              `json:"oceans,omitempty" ts:"DiffValue<number> | undefined"`
    PlayerChanges    map[string]*PlayerChanges    `json:"playerChanges,omitempty" ts:"Record<string, PlayerChanges> | undefined"`
    BoardChanges     *BoardChanges                `json:"boardChanges,omitempty" ts:"BoardChanges | undefined"`
}
```

---

### 4. Thread Safety Approach

**Question**: How should concurrent access be handled?

**Findings**:
- Existing `InMemoryGameRepository` uses `sync.RWMutex`
- Game struct uses `sync.RWMutex` internally
- Pattern: Read lock for getters, write lock for setters
- Events published AFTER lock release to prevent deadlocks

**Decision**: Use `sync.RWMutex` consistent with existing patterns.

**Rationale**: Follows established codebase patterns, proven safe for existing game operations.

---

### 5. Integration with Existing Architecture

**Question**: How should the new repository integrate with existing components?

**Findings**:
- Current flow: Action → Game state method → EventBus → Broadcaster
- GameRepository is injected via BaseAction
- State is accessed via `gameRepo.Get(gameID)` then `game.GetPlayer(playerID)`

**Decision**: Create a **separate `GameStateRepository`** that wraps or works alongside `GameRepository`.

**Rationale**:
1. Separation of concerns: `GameRepository` manages game lifecycle, `GameStateRepository` manages state history
2. Non-invasive: doesn't modify existing working code
3. Can be composed with existing repository in dependency injection
4. Aligns with constitution: "Do precisely what was asked; nothing more, nothing less"

**Integration Pattern**:
```go
type GameStateRepository interface {
    Write(ctx context.Context, gameID string, state *Game) (*StateDiff, error)
    Get(ctx context.Context, gameID string) (*Game, error)
    GetDiff(ctx context.Context, gameID string) ([]StateDiff, error)
}
```

---

### 6. State Serialization for Diffing

**Question**: How should game state be captured for diff computation?

**Findings**:
- Game has private fields, accessible only via methods
- GameDto in `internal/delivery/dto/` is the serializable representation
- Mapper functions convert Game → GameDto for WebSocket broadcasts

**Decision**: Use **GameDto** as the diffable state representation.

**Rationale**:
1. GameDto is already designed for external consumption
2. Contains all relevant game state in serializable form
3. Already has json/ts tags for type generation
4. Diffs can be computed on DTOs, not internal Game struct

---

## Summary

| Decision | Choice | Key Reason |
|----------|--------|------------|
| Repository pattern | New `GameStateRepository` interface | Separation of concerns |
| Diff library | Custom domain-specific structs | Type safety, performance |
| Diff format | Typed `GameChanges` struct | Frontend-friendly, type-safe |
| Thread safety | `sync.RWMutex` | Consistent with codebase |
| State format | Diff on GameDto | Already serializable |
| Storage | In-memory slice per game | Session-scoped, no persistence needed |
