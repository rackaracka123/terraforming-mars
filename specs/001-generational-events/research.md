# Research: Generational Events System

## 1. Event Bus Pattern

**Decision**: Use existing EventBus pattern with type-safe generics for tracking generational events.

**Rationale**: The codebase already has a robust event system in `internal/events/event_bus.go`. Key events are already
published:

- `TerraformRatingChangedEvent` - published by `player_resources.go` when TR changes
- `TilePlacedEvent` - published by `board.go` when tiles are placed (includes `TileType` field)
- `GenerationAdvancedEvent` - published when generation increments

**Alternatives considered**:

- Direct method calls in actions: Rejected - violates event-driven architecture principle
- Polling-based tracking: Rejected - Constitution explicitly forbids polling

**Key files**:

- `backend/internal/events/event_bus.go` - Generic Subscribe/Publish functions
- `backend/internal/events/domain_events.go` - Event type definitions

## 2. State Calculator Pattern

**Decision**: Extend `CalculatePlayerCardActionState` to validate generational event requirements.

**Rationale**: The state calculator is the central validation point for card actions. It already validates:

- Turn ownership (`ErrorCodeNotYourTurn`)
- Active tile selection (`ErrorCodeActiveTileSelection`)
- Resource inputs (`ErrorCodeInsufficientResources`)
- Tile output availability (`validateBehaviorTileOutputs`)

Adding generational event validation here ensures the frontend receives proper error states.

**Key file**: `backend/internal/action/state_calculator.go` (lines 48-92 for card action state)

**Integration pattern**:

```go
errors = append(errors, validateGenerationalEventRequirements(behavior, p)...)
```

## 3. Player State Structure

**Decision**: Add `GenerationalEvents` component to Player, following the delegation pattern.

**Rationale**: Player already uses delegated components for:

- `hand` (*Hand)
- `playedCards` (*PlayedCards)
- `resources` (*PlayerResources)
- `actions` (*Actions)
- `effects` (*Effects)

Adding `generationalEvents` (*GenerationalEvents) follows the same pattern.

**Key file**: `backend/internal/game/player/player.go`

## 4. Card Behavior Extension

**Decision**: Add `GenerationalEventRequirements` field to `CardBehavior` struct.

**Rationale**: CardBehavior already has:

- `Triggers` - when to activate
- `Inputs` - what player pays
- `Outputs` - what player receives
- `Choices` - player alternatives

Adding `GenerationalEventRequirements` fits naturally as another condition field.

**Key file**: `backend/internal/game/shared/behavior.go`

## 5. MinMax Type

**Decision**: Create backend `MinMax` type in shared package.

**Rationale**: Currently only exists as DTO (`MinMaxValueDto`). Backend needs its own type for:

- JSON parsing of card behaviors
- Validation in state calculator

**Key files**:

- Create: `backend/internal/game/shared/generational_events.go`
- Reference: `backend/internal/delivery/dto/game_dto.go` (MinMaxValueDto pattern)

## 6. TargetType

**Decision**: Reuse existing `TargetType` from `internal/game/cards/resource_condition.go`.

**Rationale**: Already defines all needed target types:

- `TargetSelfPlayer` ("self-player")
- `TargetAnyPlayer` ("any-player")
- `TargetOpponent` ("opponent")
- etc.

**Key file**: `backend/internal/game/cards/resource_condition.go`

## 7. Generation End Handling

**Decision**: Subscribe to `GenerationAdvancedEvent` to clear generational events.

**Rationale**: The event is already published in `game.go.AdvanceGeneration()`. No new event needed.

**Key file**: `backend/internal/game/game.go` (AdvanceGeneration method)

## 8. Frontend Display

**Decision**: Add asterisk indicator to ManualActionLayout and TriggeredEffectLayout.

**Rationale**:

- Asterisk is minimal visual indicator that doesn't bloat the UI
- Placed on output side (right side of arrow) to indicate conditional output
- Consistent with user's specification

**Key files**:

- `frontend/src/components/ui/cards/BehaviorSection/components/ManualActionLayout.tsx`
- `frontend/src/components/ui/cards/BehaviorSection/components/TriggeredEffectLayout.tsx`

## 9. Backend CLAUDE.md Update

**Decision**: Add event-driven architecture section to backend CLAUDE.md.

**Rationale**: Spec requires documenting the pattern with examples for future development.

**Key file**: `backend/CLAUDE.md`

## Summary of Required Changes

### Backend (Go)

| File                                          | Change                                                             |
|-----------------------------------------------|--------------------------------------------------------------------|
| `internal/game/shared/generational_events.go` | NEW: GenerationalEvent enum, MinMax type, requirement types        |
| `internal/game/shared/behavior.go`            | ADD: GenerationalEventRequirements field                           |
| `internal/game/player/generational_events.go` | NEW: Player GenerationalEvents component                           |
| `internal/game/player/player.go`              | ADD: generationalEvents field and accessor                         |
| `internal/game/game.go`                       | ADD: Event subscriptions for tracking, reset on generation advance |
| `internal/action/state_calculator.go`         | ADD: validateGenerationalEventRequirements function                |
| `internal/delivery/dto/game_dto.go`           | ADD: GenerationalEvents to PlayerDto                               |
| `internal/delivery/dto/mapper.go`             | ADD: mapping for GenerationalEvents                                |
| `backend/CLAUDE.md`                           | ADD: Event-driven architecture documentation                       |

### Frontend (TypeScript/React)

| File                                                   | Change                                  |
|--------------------------------------------------------|-----------------------------------------|
| `src/types/generated/api-types.ts`                     | AUTO: Generated types from backend      |
| `BehaviorSection/components/ManualActionLayout.tsx`    | ADD: Asterisk for conditional behaviors |
| `BehaviorSection/components/TriggeredEffectLayout.tsx` | ADD: Asterisk for conditional behaviors |

### Card Data (JSON)

| File                                     | Change                                               |
|------------------------------------------|------------------------------------------------------|
| `backend/assets/cards/corporations.json` | UPDATE: UNMI card with generationalEventRequirements |
