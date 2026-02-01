# Data Model: Game State Repository

**Feature Branch**: `001-game-state-repo`
**Created**: 2026-01-30

## Entities

### StateDiff

Represents the difference between two consecutive game states.

| Field | Type | Description |
|-------|------|-------------|
| SequenceNumber | int64 | Monotonically increasing per game, starting at 1 |
| Timestamp | time.Time | When the state change occurred |
| GameID | string | The game this diff belongs to |
| Changes | *GameChanges | Structured representation of what changed |

**Validation Rules**:
- SequenceNumber must be positive and sequential
- Timestamp must be monotonically increasing within a game
- GameID must be non-empty
- Changes may be empty (for audit trail of no-op writes)

---

### GameChanges

Top-level container for all changes in a state transition.

| Field | Type | Description |
|-------|------|-------------|
| Status | *DiffValue[string] | Game status change (lobby, in_progress, etc.) |
| Phase | *DiffValue[string] | Game phase change |
| Generation | *DiffValue[int] | Generation number change |
| CurrentTurnPlayerID | *DiffValue[string] | Active player change |
| Temperature | *DiffValue[int] | Global temperature change |
| Oxygen | *DiffValue[int] | Global oxygen percentage change |
| Oceans | *DiffValue[int] | Placed oceans count change |
| PlayerChanges | map[string]*PlayerChanges | Per-player changes, keyed by player ID |
| BoardChanges | *BoardChanges | Board/tile changes |

**Notes**:
- All fields are optional (nil means no change)
- Only populated fields are serialized to JSON

---

### DiffValue[T]

Generic container for old/new value pairs.

| Field | Type | Description |
|-------|------|-------------|
| Old | T | Previous value |
| New | T | Current value |

**Notes**:
- Used for primitive value changes
- Type parameter T: string, int, bool

---

### PlayerChanges

Changes to a single player's state.

| Field | Type | Description |
|-------|------|-------------|
| Credits | *DiffValue[int] | M€ credits change |
| Steel | *DiffValue[int] | Steel resource change |
| Titanium | *DiffValue[int] | Titanium resource change |
| Plants | *DiffValue[int] | Plants resource change |
| Energy | *DiffValue[int] | Energy resource change |
| Heat | *DiffValue[int] | Heat resource change |
| TerraformRating | *DiffValue[int] | TR change |
| CreditsProduction | *DiffValue[int] | M€ production change |
| SteelProduction | *DiffValue[int] | Steel production change |
| TitaniumProduction | *DiffValue[int] | Titanium production change |
| PlantsProduction | *DiffValue[int] | Plants production change |
| EnergyProduction | *DiffValue[int] | Energy production change |
| HeatProduction | *DiffValue[int] | Heat production change |
| CardsAdded | []string | Card IDs added to hand |
| CardsRemoved | []string | Card IDs removed from hand |
| CardsPlayed | []string | Card IDs played this transition |
| Corporation | *DiffValue[string] | Corporation selection/change |
| Passed | *DiffValue[bool] | Pass status change |

---

### BoardChanges

Changes to the game board.

| Field | Type | Description |
|-------|------|-------------|
| TilesPlaced | []TilePlacement | New tiles placed on the board |

---

### TilePlacement

Record of a single tile placement.

| Field | Type | Description |
|-------|------|-------------|
| HexID | string | Hex coordinate string (e.g., "0,0,0") |
| TileType | string | Type of tile (city, greenery, ocean, etc.) |
| OwnerID | string | Player who placed the tile (may be empty) |

---

### DiffLog

Container for the complete history of a game.

| Field | Type | Description |
|-------|------|-------------|
| GameID | string | The game this log belongs to |
| Diffs | []StateDiff | Ordered list of all state diffs |
| CurrentSequence | int64 | Latest sequence number |

**Notes**:
- In-memory only, not persisted
- Cleared when game is deleted

---

## State Transitions

```
Initial State (empty)
       │
       ▼ Write(state1)
   StateDiff #1 (full state as "additions")
       │
       ▼ Write(state2)
   StateDiff #2 (changes from state1 to state2)
       │
       ▼ Write(state3)
   StateDiff #3 (changes from state2 to state3)
       │
       ▼ GetDiff()
   Returns [StateDiff #1, #2, #3]
```

---

## Relationships

```
GameStateRepository
       │
       ├── stores ──► DiffLog (one per game)
       │                  │
       │                  └── contains ──► []StateDiff
       │                                        │
       │                                        └── contains ──► GameChanges
       │                                                              │
       │                                                              ├── PlayerChanges
       │                                                              └── BoardChanges
       │
       └── stores ──► currentState (GameDto per game)
```

---

## Storage Design

```go
type InMemoryGameStateRepository struct {
    mu            sync.RWMutex
    currentStates map[string]*dto.GameDto  // gameID → current state
    diffLogs      map[string]*DiffLog      // gameID → diff history
}
```

**Notes**:
- Stores serializable GameDto, not internal Game struct
- Each game has independent diff log
- Thread-safe with RWMutex
