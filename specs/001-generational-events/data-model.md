# Data Model: Generational Events System

## New Types

### GenerationalEvent (Enum)

Backend: `internal/game/shared/generational_events.go`

```go
type GenerationalEvent string

const (
    GenerationalEventTRRaise          GenerationalEvent = "tr-raise"
    GenerationalEventOceanPlacement   GenerationalEvent = "ocean-placement"
    GenerationalEventCityPlacement    GenerationalEvent = "city-placement"
    GenerationalEventGreeneryPlacement GenerationalEvent = "greenery-placement"
)
```

TypeScript (generated):
```typescript
type GenerationalEvent = "tr-raise" | "ocean-placement" | "city-placement" | "greenery-placement";
```

### MinMax

Backend: `internal/game/shared/generational_events.go`

```go
type MinMax struct {
    Min *int `json:"min,omitempty" ts:"number | undefined"`
    Max *int `json:"max,omitempty" ts:"number | undefined"`
}
```

TypeScript (generated):
```typescript
interface MinMax {
    min?: number;
    max?: number;
}
```

### GenerationalEventRequirement

Backend: `internal/game/shared/generational_events.go`

```go
type GenerationalEventRequirement struct {
    Event  GenerationalEvent `json:"event" ts:"GenerationalEvent"`
    Count  *MinMax           `json:"count,omitempty" ts:"MinMax | undefined"`
    Target *string           `json:"target,omitempty" ts:"string | undefined"`
}
```

TypeScript (generated):
```typescript
interface GenerationalEventRequirement {
    event: GenerationalEvent;
    count?: MinMax;
    target?: string;
}
```

### PlayerGenerationalEventEntry

Backend: `internal/game/shared/generational_events.go`

```go
type PlayerGenerationalEventEntry struct {
    Event GenerationalEvent `json:"event" ts:"GenerationalEvent"`
    Count int               `json:"count" ts:"number"`
}
```

TypeScript (generated):
```typescript
interface PlayerGenerationalEventEntry {
    event: GenerationalEvent;
    count: number;
}
```

## Modified Types

### CardBehavior

Backend: `internal/game/shared/behavior.go`

```go
type CardBehavior struct {
    Triggers                      []Trigger                       `json:"triggers,omitempty"`
    Inputs                        []ResourceCondition             `json:"inputs,omitempty"`
    Outputs                       []ResourceCondition             `json:"outputs,omitempty"`
    Choices                       []Choice                        `json:"choices,omitempty"`
    GenerationalEventRequirements []GenerationalEventRequirement  `json:"generationalEventRequirements,omitempty" ts:"GenerationalEventRequirement[] | undefined"`
}
```

### Player (Internal State)

Backend: `internal/game/player/player.go`

```go
type Player struct {
    // ... existing fields ...
    generationalEvents *GenerationalEvents  // NEW: tracks events within current generation
}
```

### GenerationalEvents (Component)

Backend: `internal/game/player/generational_events.go`

```go
type GenerationalEvents struct {
    mu     sync.RWMutex
    events map[GenerationalEvent]int
}

func (ge *GenerationalEvents) Increment(event GenerationalEvent) {
    ge.mu.Lock()
    defer ge.mu.Unlock()
    ge.events[event]++
}

func (ge *GenerationalEvents) GetCount(event GenerationalEvent) int {
    ge.mu.RLock()
    defer ge.mu.RUnlock()
    return ge.events[event]
}

func (ge *GenerationalEvents) GetAll() []PlayerGenerationalEventEntry {
    ge.mu.RLock()
    defer ge.mu.RUnlock()
    result := make([]PlayerGenerationalEventEntry, 0, len(ge.events))
    for event, count := range ge.events {
        if count > 0 {
            result = append(result, PlayerGenerationalEventEntry{Event: event, Count: count})
        }
    }
    return result
}

func (ge *GenerationalEvents) Clear() {
    ge.mu.Lock()
    defer ge.mu.Unlock()
    ge.events = make(map[GenerationalEvent]int)
}
```

### PlayerDto

Backend: `internal/delivery/dto/game_dto.go`

```go
type PlayerDto struct {
    // ... existing fields ...
    GenerationalEvents []PlayerGenerationalEventEntryDto `json:"generationalEvents" ts:"PlayerGenerationalEventEntry[]"`
}

type PlayerGenerationalEventEntryDto struct {
    Event string `json:"event" ts:"string"`
    Count int    `json:"count" ts:"number"`
}
```

### StateError (New Error Code)

Backend: `internal/game/player/state.go`

```go
const (
    // ... existing error codes ...
    ErrorCodeGenerationalEventNotMet StateErrorCode = "generational_event_not_met"
)
```

## State Transitions

### Event Tracking Flow

```
Player Action (e.g., place ocean tile)
    ↓
Game State Method (e.g., Board.PlaceTile())
    ↓
EventBus.Publish(TilePlacedEvent)
    ↓
GenerationalEventsSubscriber.OnTilePlaced()
    ↓
Player.GenerationalEvents().Increment(OceanPlacement)
```

### Generation Reset Flow

```
Game.AdvanceGeneration()
    ↓
EventBus.Publish(GenerationAdvancedEvent)
    ↓
GenerationalEventsSubscriber.OnGenerationAdvanced()
    ↓
For each player: Player.GenerationalEvents().Clear()
```

### Action Validation Flow

```
Frontend: User clicks card action
    ↓
Backend: CalculatePlayerCardActionState() called
    ↓
validateGenerationalEventRequirements(behavior, player)
    ↓
Check: player.GenerationalEvents().GetCount(req.Event) >= req.Count.Min
    ↓
If not met: Return StateError with ErrorCodeGenerationalEventNotMet
    ↓
Frontend: Action disabled based on EntityState.Errors
```

## Validation Rules

1. **Non-zero tracking only**: Only events with count > 0 are stored/returned
2. **Increment-only**: Events can only be incremented, never decremented within a generation
3. **Per-player isolation**: Each player has independent generational event tracking
4. **Generation-scoped**: All events are cleared when generation advances
5. **Public visibility**: GenerationalEvents is NOT hidden in OtherPlayersDto (it's public info)

## Card JSON Example: UNMI Corporation

```json
{
  "id": "B10",
  "name": "United Nations Mars Initiative",
  "type": "corporation",
  "cost": 0,
  "description": "Action: If your Terraform Rating was raised this generation, you may pay 3 M€ to raise it 1 step more. You start with 40 M€.",
  "pack": "base-game",
  "tags": ["earth"],
  "behaviors": [
    {
      "triggers": [{ "type": "auto-corporation-start" }],
      "outputs": [{ "type": "credit", "amount": 40, "target": "self-player" }]
    },
    {
      "triggers": [{ "type": "manual" }],
      "inputs": [{ "type": "credit", "amount": 3, "target": "self-player" }],
      "outputs": [{ "type": "tr", "amount": 1, "target": "self-player" }],
      "generationalEventRequirements": [
        { "event": "tr-raise", "count": { "min": 1 } }
      ]
    }
  ]
}
```
