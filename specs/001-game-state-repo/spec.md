# Feature Specification: Game State Repository with Diff Logging

**Feature Branch**: `001-game-state-repo`
**Created**: 2026-01-30
**Status**: Draft
**Input**: User description: "I would like a monorepo for the entire game. this repo has only 3 functions. write and get, get diff. Write will compute a diff which will later be used to retrieve a log of the game."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Write Game State with Diff Computation (Priority: P1)

A game action modifies the game state and writes it to the repository. The repository automatically computes the difference between the previous state and the new state, storing this diff for later retrieval.

**Why this priority**: This is the foundational operation that enables all other functionality. Without write with diff computation, there is no history to retrieve.

**Independent Test**: Can be fully tested by writing a state, then writing a modified state, and verifying that the diff is correctly computed and stored.

**Acceptance Scenarios**:

1. **Given** a game with initial state, **When** the state is written to the repository, **Then** the repository stores the state and computes a diff representing the initial state (from empty).
2. **Given** an existing game state in the repository, **When** a modified state is written, **Then** the repository computes and stores the diff showing only what changed between the two states.
3. **Given** multiple state writes occur in sequence, **When** each write completes, **Then** each diff is stored with a sequential identifier and timestamp.

---

### User Story 2 - Get Current Game State (Priority: P1)

A client requests the current game state and receives the complete, up-to-date state of the game.

**Why this priority**: Reading state is equally essential to writing - clients need to retrieve the current state to render the game.

**Independent Test**: Can be fully tested by writing a known state, then calling get and verifying the returned state matches exactly.

**Acceptance Scenarios**:

1. **Given** a game with state stored in the repository, **When** get is called with the game ID, **Then** the complete current game state is returned.
2. **Given** no game exists with the specified ID, **When** get is called, **Then** an appropriate error is returned indicating the game was not found.
3. **Given** a game has had multiple writes, **When** get is called, **Then** the most recent state is returned (not historical states).

---

### User Story 3 - Get Diff Log (Priority: P2)

A client retrieves the history of all diffs for a game, allowing them to see a log of how the game state evolved over time.

**Why this priority**: This is the value-add feature that justifies computing diffs - without retrieval, the diffs serve no purpose. Slightly lower priority than basic operations.

**Independent Test**: Can be fully tested by writing multiple states, then calling get diff and verifying all diffs are returned in chronological order.

**Acceptance Scenarios**:

1. **Given** a game with multiple state writes, **When** get diff is called with the game ID, **Then** all diffs are returned in chronological order (oldest first).
2. **Given** a newly created game with only one write, **When** get diff is called, **Then** a single diff representing the initial state is returned.
3. **Given** no game exists with the specified ID, **When** get diff is called, **Then** an appropriate error is returned indicating the game was not found.

---

### Edge Cases

- What happens when write is called with a state identical to the current state? (Assumption: Store empty diff to preserve action audit trail)
- How does the system handle concurrent writes to the same game? (Assumption: Thread-safe with mutex protection)
- What is the maximum number of diffs stored per game? (Assumption: Unlimited for game session duration, in-memory only)
- How are diffs structured for complex nested state objects? (Assumption: JSON-patch style with path-based changes)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Repository MUST provide a `Write` function that accepts a game ID and complete game state
- **FR-002**: Repository MUST compute a diff between the previous state and new state on every write
- **FR-003**: Repository MUST store each computed diff with a sequential identifier and timestamp
- **FR-004**: Repository MUST provide a `Get` function that returns the current game state for a given game ID
- **FR-005**: Repository MUST provide a `GetDiff` function that returns all diffs for a given game ID in chronological order
- **FR-006**: Repository MUST return appropriate errors when operations are attempted on non-existent games
- **FR-007**: Repository MUST handle the first write to a new game by computing a diff from an empty/nil state
- **FR-008**: Repository MUST ensure thread-safe access for concurrent read and write operations

### Key Entities

- **GameState**: The complete state of a game at a point in time (existing Game struct)
- **Diff**: A record of changes between two consecutive states, containing:
  - Game ID
  - Sequence number (monotonically increasing per game)
  - Timestamp
  - Changes (structured representation of what changed)
- **DiffLog**: An ordered collection of all diffs for a game, representing the complete history of state changes

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All three repository functions (Write, Get, GetDiff) are implemented and accessible
- **SC-002**: Write operations complete and return within acceptable response time for real-time gameplay
- **SC-003**: Diffs accurately capture all state changes between consecutive writes with no data loss
- **SC-004**: GetDiff returns the complete, ordered history of all state changes for a game
- **SC-005**: The repository handles concurrent access without data corruption or race conditions
- **SC-006**: 100% of state changes are captured in the diff log (no silent mutations)

## Assumptions

- Repository is in-memory only (no persistence required beyond game session)
- Diff computation uses a path-based approach similar to JSON-patch for nested structures
- Empty diffs are stored when write is called with identical state (preserves audit trail)
- The existing `Game` struct from the codebase serves as the GameState entity
