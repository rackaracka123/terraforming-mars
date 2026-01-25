# Quickstart: Generational Events System

## Overview

This feature adds a generational events tracking system that enables conditional card behaviors based on player activities within a generation. The primary use case is the UNMI corporation's action that requires TR to have been raised this generation.

## Key Files to Create/Modify

### Backend (Priority Order)

1. **`internal/game/shared/generational_events.go`** (NEW)
   - Define `GenerationalEvent` enum
   - Define `MinMax` type
   - Define `GenerationalEventRequirement` type
   - Define `PlayerGenerationalEventEntry` type

2. **`internal/game/shared/behavior.go`** (MODIFY)
   - Add `GenerationalEventRequirements` field to `CardBehavior`

3. **`internal/game/player/generational_events.go`** (NEW)
   - Create `GenerationalEvents` component
   - Implement `Increment()`, `GetCount()`, `GetAll()`, `Clear()`

4. **`internal/game/player/player.go`** (MODIFY)
   - Add `generationalEvents` field
   - Add `GenerationalEvents()` accessor
   - Initialize in constructor

5. **`internal/game/game.go`** (MODIFY)
   - Subscribe to `TerraformRatingChangedEvent`, `TilePlacedEvent`
   - Subscribe to `GenerationAdvancedEvent` for reset

6. **`internal/action/state_calculator.go`** (MODIFY)
   - Add `validateGenerationalEventRequirements()` function
   - Call from `CalculatePlayerCardActionState()`

7. **`internal/game/player/state.go`** (MODIFY)
   - Add `ErrorCodeGenerationalEventNotMet` constant

8. **`internal/delivery/dto/game_dto.go`** (MODIFY)
   - Add `GenerationalEvents` to `PlayerDto`

9. **`internal/delivery/dto/mapper.go`** (MODIFY)
   - Add mapping for `GenerationalEvents`

### Frontend

1. **`src/components/ui/cards/BehaviorSection/components/ManualActionLayout.tsx`** (MODIFY)
   - Add asterisk indicator for behaviors with `generationalEventRequirements`

2. **`src/components/ui/cards/BehaviorSection/components/TriggeredEffectLayout.tsx`** (MODIFY)
   - Add asterisk indicator for behaviors with `generationalEventRequirements`

### Card Data

1. **`backend/assets/cards/corporations.json`** (MODIFY)
   - Update UNMI card with `generationalEventRequirements`

### Documentation

1. **`backend/CLAUDE.md`** (MODIFY)
   - Add event-driven architecture section with examples

## Quick Implementation Steps

### Step 1: Define Types

```go
// internal/game/shared/generational_events.go
package shared

type GenerationalEvent string

const (
    GenerationalEventTRRaise          GenerationalEvent = "tr-raise"
    GenerationalEventOceanPlacement   GenerationalEvent = "ocean-placement"
    GenerationalEventCityPlacement    GenerationalEvent = "city-placement"
    GenerationalEventGreeneryPlacement GenerationalEvent = "greenery-placement"
)

type MinMax struct {
    Min *int `json:"min,omitempty" ts:"number | undefined"`
    Max *int `json:"max,omitempty" ts:"number | undefined"`
}

type GenerationalEventRequirement struct {
    Event  GenerationalEvent `json:"event" ts:"GenerationalEvent"`
    Count  *MinMax           `json:"count,omitempty" ts:"MinMax | undefined"`
    Target *string           `json:"target,omitempty" ts:"string | undefined"`
}

type PlayerGenerationalEventEntry struct {
    Event GenerationalEvent `json:"event" ts:"GenerationalEvent"`
    Count int               `json:"count" ts:"number"`
}
```

### Step 2: Add to CardBehavior

```go
// internal/game/shared/behavior.go
type CardBehavior struct {
    // ... existing fields ...
    GenerationalEventRequirements []GenerationalEventRequirement `json:"generationalEventRequirements,omitempty" ts:"GenerationalEventRequirement[] | undefined"`
}
```

### Step 3: Create Player Component

```go
// internal/game/player/generational_events.go
package player

import (
    "sync"
    "terraforming-mars-backend/internal/game/shared"
)

type GenerationalEvents struct {
    mu     sync.RWMutex
    events map[shared.GenerationalEvent]int
}

func newGenerationalEvents() *GenerationalEvents {
    return &GenerationalEvents{
        events: make(map[shared.GenerationalEvent]int),
    }
}

func (ge *GenerationalEvents) Increment(event shared.GenerationalEvent) {
    ge.mu.Lock()
    defer ge.mu.Unlock()
    ge.events[event]++
}

func (ge *GenerationalEvents) GetCount(event shared.GenerationalEvent) int {
    ge.mu.RLock()
    defer ge.mu.RUnlock()
    return ge.events[event]
}

func (ge *GenerationalEvents) GetAll() []shared.PlayerGenerationalEventEntry {
    ge.mu.RLock()
    defer ge.mu.RUnlock()
    result := make([]shared.PlayerGenerationalEventEntry, 0, len(ge.events))
    for event, count := range ge.events {
        if count > 0 {
            result = append(result, shared.PlayerGenerationalEventEntry{Event: event, Count: count})
        }
    }
    return result
}

func (ge *GenerationalEvents) Clear() {
    ge.mu.Lock()
    defer ge.mu.Unlock()
    ge.events = make(map[shared.GenerationalEvent]int)
}
```

### Step 4: Subscribe to Events

```go
// internal/game/game.go - in initialization
events.Subscribe(g.eventBus, func(e events.TerraformRatingChangedEvent) {
    if e.NewRating > e.OldRating {
        if player := g.GetPlayer(e.PlayerID); player != nil {
            player.GenerationalEvents().Increment(shared.GenerationalEventTRRaise)
        }
    }
})

events.Subscribe(g.eventBus, func(e events.TilePlacedEvent) {
    if player := g.GetPlayer(e.PlayerID); player != nil {
        switch e.TileType {
        case "ocean":
            player.GenerationalEvents().Increment(shared.GenerationalEventOceanPlacement)
        case "city":
            player.GenerationalEvents().Increment(shared.GenerationalEventCityPlacement)
        case "greenery":
            player.GenerationalEvents().Increment(shared.GenerationalEventGreeneryPlacement)
        }
    }
})

events.Subscribe(g.eventBus, func(e events.GenerationAdvancedEvent) {
    for _, p := range g.Players() {
        p.GenerationalEvents().Clear()
    }
})
```

### Step 5: Add Validation

```go
// internal/action/state_calculator.go
func validateGenerationalEventRequirements(
    behavior shared.CardBehavior,
    p *player.Player,
) []player.StateError {
    if len(behavior.GenerationalEventRequirements) == 0 {
        return nil
    }

    var errors []player.StateError
    playerEvents := p.GenerationalEvents()

    for _, req := range behavior.GenerationalEventRequirements {
        count := playerEvents.GetCount(req.Event)

        if req.Count != nil && req.Count.Min != nil {
            if count < *req.Count.Min {
                errors = append(errors, player.StateError{
                    Code:     player.ErrorCodeGenerationalEventNotMet,
                    Category: player.ErrorCategoryRequirement,
                    Message:  formatGenerationalEventError(req.Event),
                })
            }
        }
    }

    return errors
}
```

### Step 6: Frontend Asterisk

```tsx
// ManualActionLayout.tsx - at end of output section
{behavior.generationalEventRequirements?.length > 0 && (
  <span className="text-white font-bold text-sm ml-1 [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">*</span>
)}
```

## Testing Commands

```bash
# Run backend tests
make test

# Generate TypeScript types
make generate

# Format and lint
make format && make lint

# Run full check
make prepare-for-commit
```

## Verification Checklist

- [ ] UNMI corporation card action is disabled when TR hasn't been raised
- [ ] UNMI corporation card action is enabled after TR is raised
- [ ] Generational events reset at generation end
- [ ] Asterisk appears on UNMI action in card display
- [ ] All existing tests pass
- [ ] No lint errors
