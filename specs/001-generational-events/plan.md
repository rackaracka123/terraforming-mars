# Implementation Plan: Generational Events System

**Branch**: `001-generational-events` | **Date**: 2026-01-24 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-generational-events/spec.md`

## Summary

Implement a generational events tracking system that enables conditional card behaviors based on player activities within a generation. The primary use case is the United Nations Mars Initiative corporation, whose action (pay 3 MC for +1 TR) requires the player to have already raised TR this generation. The system uses event-driven architecture with the existing EventBus pattern.

## Technical Context

**Language/Version**: Go 1.21+ (backend), TypeScript 5.x (frontend)
**Primary Dependencies**: gorilla/websocket, chi router, React 18, Tailwind CSS v4
**Storage**: In-memory game state (no persistence required for generational events)
**Testing**: go test (backend), Playwright (frontend integration)
**Target Platform**: Web application (localhost:3000 frontend, localhost:3001 backend)
**Project Type**: Web application with Go backend + React frontend
**Performance Goals**: Real-time updates via WebSocket, no noticeable latency
**Constraints**: Event-driven architecture, no polling, state calculator pattern for frontend validation
**Scale/Scope**: Single feature affecting card system, player state, and UI display

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Clean Architecture & Action Pattern | PASS | Actions in `/internal/action/`, state via Game methods, events via EventBus |
| II. Code Clarity Over Comments | PASS | Self-documenting code, no explanatory comments |
| III. Best Practices Enforcement | PASS | Follow Go idioms, use `make format` and `make lint` |
| IV. Complete Feature Implementation | PASS | Full implementation including tests, validation, UI |
| V. Type Safety & Generation | PASS | New types with `json:` and `ts:` tags, run `make generate` |
| VI. Testing Discipline | PASS | Tests in `test/` directory, table-driven tests |
| VII. No Deprecated Code | PASS | No deprecated code introduced |

## Project Structure

### Documentation (this feature)

```text
specs/001-generational-events/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (N/A - WebSocket-based)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── action/
│   │   └── state_calculator.go     # Add generational event validation
│   ├── events/
│   │   └── domain_events.go        # Already has TerraformRatingChangedEvent, TilePlacedEvent
│   ├── game/
│   │   ├── game.go                 # Add generational events reset on generation advance
│   │   ├── player/
│   │   │   └── player.go           # Add GenerationalEvents field and methods
│   │   └── shared/
│   │       ├── behavior.go         # Add GenerationalEventRequirement field
│   │       └── generational_events.go  # New: GenerationalEvent enum and types
│   └── delivery/
│       └── dto/
│           └── game_dto.go         # Add GenerationalEvents to PlayerDto
└── test/
    └── action/
        └── generational_events_test.go  # New test file

frontend/
├── src/
│   ├── components/
│   │   └── ui/cards/BehaviorSection/
│   │       └── components/
│   │           ├── ManualActionLayout.tsx   # Add asterisk for conditional behaviors
│   │           └── TriggeredEffectLayout.tsx # Add asterisk for conditional behaviors
│   └── types/
│       └── generated/
│           └── api-types.ts         # Auto-generated with new types
```

**Structure Decision**: Web application structure with Go backend and React frontend. Changes span both layers following existing patterns.

## Complexity Tracking

No violations requiring justification.

## Key Implementation Details

### 1. Event-Driven Architecture (Backend)

The system subscribes to existing domain events to track generational events:

```go
// Subscribe to TerraformRatingChangedEvent
events.Subscribe(eventBus, func(e events.TerraformRatingChangedEvent) {
    if e.NewRating > e.OldRating {
        player.GenerationalEvents().Increment(GenerationalEventTRRaise)
    }
})

// Subscribe to TilePlacedEvent
events.Subscribe(eventBus, func(e events.TilePlacedEvent) {
    switch e.TileType {
    case "ocean":
        player.GenerationalEvents().Increment(GenerationalEventOceanPlacement)
    case "city":
        player.GenerationalEvents().Increment(GenerationalEventCityPlacement)
    case "greenery":
        player.GenerationalEvents().Increment(GenerationalEventGreeneryPlacement)
    }
})
```

### 2. State Calculator Integration (Backend)

The `CalculatePlayerCardActionState` function in `state_calculator.go` will validate generational event requirements:

```go
// In CalculatePlayerCardActionState, add validation for generational event requirements
if len(behavior.GenerationalEventRequirements) > 0 {
    for _, req := range behavior.GenerationalEventRequirements {
        playerEvents := p.GenerationalEvents()
        count := playerEvents.GetCount(req.Event)

        if req.Count != nil {
            if req.Count.Min != nil && count < *req.Count.Min {
                errors = append(errors, player.StateError{
                    Code:     player.ErrorCodeGenerationalEventNotMet,
                    Category: player.ErrorCategoryRequirement,
                    Message:  formatGenerationalEventError(req.Event),
                })
            }
        }
    }
}
```

### 3. Card Behavior JSON Extension

The UNMI corporation card behavior will be extended:

```json
{
  "triggers": [{ "type": "manual" }],
  "inputs": [{ "type": "credit", "amount": 3, "target": "self-player" }],
  "outputs": [{ "type": "tr", "amount": 1, "target": "self-player" }],
  "generationalEventRequirements": [
    { "event": "tr-raise", "count": { "min": 1 } }
  ]
}
```

### 4. Frontend Asterisk Indicator

For behaviors with generational event requirements, add an asterisk on the output side:

```tsx
// In ManualActionLayout.tsx
{behavior.generationalEventRequirements?.length > 0 && (
  <span className="text-white font-bold text-sm ml-1">*</span>
)}
```

### 5. Generation Reset

On `GenerationAdvancedEvent`, clear all player generational events:

```go
events.Subscribe(eventBus, func(e events.GenerationAdvancedEvent) {
    for _, player := range game.Players() {
        player.GenerationalEvents().Clear()
    }
})
```
