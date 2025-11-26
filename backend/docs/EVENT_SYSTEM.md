# Event System Architecture

This document describes the event-driven architecture used in the Terraforming Mars backend.

## Overview

The backend uses an event bus to decouple game actions from their side effects, particularly for passive card effects. When game state changes (e.g., temperature increases), the system publishes domain events that trigger subscribed effects automatically.

## Key Principles

1. **Actions do what they say, nothing more**: When an action raises temperature, it only raises temperature - it doesn't check for passive effects
2. **Events trigger effects**: Domain events are published by repositories and trigger passive card effects via subscribers
3. **SessionManager auto-broadcasts**: SessionManager subscribes to domain events and automatically broadcasts state updates
4. **No manual polling**: Effects are event-driven, not checked manually in action code

## Architecture Components

### EventBus

Central pub/sub system for domain events:

```go
type EventBus interface {
    Subscribe(eventType string, handler EventHandler)
    Publish(event Event)
}
```

**Location**: `internal/events/event_bus.go`

### Domain Events

Typed events representing state changes:

```go
type ResourcesChangedEvent struct {
    GameID    string
    PlayerID  string
    Changes   map[string]int  // resource type -> delta
    Timestamp time.Time
}

type TemperatureChangedEvent struct {
    GameID        string
    OldValue      int
    NewValue      int
    Timestamp     time.Time
}

type TilePlacedEvent struct {
    GameID        string
    PlayerID      string
    TileType      string
    HexID         string
    Timestamp     time.Time
}
```

**Location**: `internal/events/events.go`

**Available Events**:
- `ResourcesChangedEvent` - Player resources modified
- `TemperatureChangedEvent` - Global temperature changed
- `OxygenChangedEvent` - Global oxygen changed
- `OceanPlacedEvent` - Ocean tile placed
- `TilePlacedEvent` - Any tile placed
- `CardPlayedEvent` - Card played by player
- `TerraformRatingChangedEvent` - Player TR modified

### CardEffectSubscriber

Subscribes to domain events and triggers passive card effects:

```go
type CardEffectSubscriber struct {
    eventBus   *EventBusImpl
    cardRepo   card.Repository
    // ... other dependencies
}

func (s *CardEffectSubscriber) OnTemperatureChanged(event TemperatureChangedEvent) {
    // Find all players with cards that have temperature triggers
    // Execute their passive effects
}
```

**Location**: `internal/events/card_effect_subscriber.go`

**How it works**:
1. Subscribes to all relevant domain events on initialization
2. When event fires, queries all players for cards with matching triggers
3. Executes passive effects for triggered cards
4. Updates player state via repositories (which publish their own events)

### SessionManager as Event Subscriber

SessionManager subscribes to domain events to automatically broadcast updates:

```go
func (sm *SessionManager) OnGameStateChanged(event Event) {
    gameID := event.GetGameID()
    sm.Broadcast(gameID) // Fetch state and broadcast to all clients
}
```

**Pattern**: Repositories publish events → SessionManager receives events → SessionManager broadcasts state

## Data Flow

### Standard Action Flow

```
1. Client sends action (e.g., "convert heat to temperature")
   ↓
2. WebSocket handler routes to action
   ↓
3. Action.Execute() validates and calls repository
   ↓
4. Repository.UpdateTemperature() changes state
   ↓
5. Repository publishes TemperatureChangedEvent
   ↓
6. CardEffectSubscriber receives event
   ↓
7. CardEffectSubscriber triggers passive effects
   ↓
8. Passive effects update state via repositories (publish more events)
   ↓
9. SessionManager receives domain events
   ↓
10. SessionManager broadcasts final state to all clients
```

### Example: Converting Heat to Temperature

```go
// Action code (simplified)
func (a *ConvertHeatToTemperatureAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
    // 1. Validate player has enough heat
    player, _ := sess.GetPlayer(playerID)
    if player.Resources().Heat < 8 {
        return errors.New("insufficient heat")
    }

    // 2. Deduct heat
    resources := player.Resources()
    resources.Heat -= 8
    player.SetResources(ctx, resources)

    // 3. Raise temperature (publishes TemperatureChangedEvent)
    gameRepo.UpdateTemperature(ctx, gameID, currentTemp + 1)

    // Done! CardEffectSubscriber handles passive effects automatically
    return nil
}
```

**Passive effect example** (from card JSON):
```json
{
  "id": "arctic_algae",
  "name": "Arctic Algae",
  "effects": [{
    "trigger": {
      "type": "temperature-raised",
      "min_temperature": -12
    },
    "output": {
      "type": "resources",
      "resources": [{"type": "plants", "amount": 1}]
    }
  }]
}
```

When `UpdateTemperature()` publishes `TemperatureChangedEvent`:
1. CardEffectSubscriber receives event
2. Finds all players with "Arctic Algae" played
3. Checks if new temperature >= -12
4. Awards 1 plant to each qualifying player (publishes ResourcesChangedEvent)
5. SessionManager receives ResourcesChangedEvent and broadcasts final state

## Repository Event Publishing

All repositories that modify state publish events:

### Game Repository

```go
func (r *Repository) UpdateTemperature(ctx context.Context, gameID string, newValue int) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    game := r.games[gameID]
    oldValue := game.Temperature
    game.Temperature = newValue

    // Publish event
    if r.eventBus != nil {
        events.Publish(r.eventBus, events.TemperatureChangedEvent{
            GameID:    gameID,
            OldValue:  oldValue,
            NewValue:  newValue,
            Timestamp: time.Now(),
        })
    }

    return nil
}
```

### Player Repository

```go
func (p *Player) SetResources(ctx context.Context, resources types.Resources) error {
    p.mu.Lock()
    defer p.mu.Unlock()

    oldResources := p.resources
    p.resources = resources

    // Calculate delta and publish event
    changes := make(map[string]int)
    if oldResources.Credits != resources.Credits {
        changes["credits"] = resources.Credits - oldResources.Credits
    }
    // ... other resources

    if p.eventBus != nil && len(changes) > 0 {
        events.Publish(p.eventBus, events.ResourcesChangedEvent{
            GameID:    p.gameID,
            PlayerID:  p.id,
            Changes:   changes,
            Timestamp: time.Now(),
        })
    }

    return nil
}
```

## Event Bus Initialization

EventBus is created at application startup and injected into all components:

```go
// cmd/server/main.go (simplified)
func main() {
    eventBus := events.NewEventBus()

    // Create repositories with event bus
    gameRepo := game.NewRepository(eventBus)
    // playerRepo already has event bus internally

    // Create subscribers
    cardEffectSubscriber := events.NewCardEffectSubscriber(eventBus, gameRepo, cardRepo)
    cardEffectSubscriber.Start()

    // SessionManager subscribes to events for broadcasting
    sessionMgr := session.NewSessionManager(eventBus, gameRepo, ...)

    // ... rest of initialization
}
```

## Adding New Domain Events

To add a new domain event:

1. **Define event struct** in `internal/events/events.go`:
```go
type CardDiscardedEvent struct {
    GameID    string
    PlayerID  string
    CardID    string
    Timestamp time.Time
}

func (e CardDiscardedEvent) GetGameID() string { return e.GameID }
func (e CardDiscardedEvent) GetEventType() string { return "card_discarded" }
```

2. **Publish event** from repository:
```go
func (p *Player) RemoveCardFromHand(ctx context.Context, cardID string) error {
    // ... remove card logic

    if p.eventBus != nil {
        events.Publish(p.eventBus, events.CardDiscardedEvent{
            GameID:    p.gameID,
            PlayerID:  p.id,
            CardID:    cardID,
            Timestamp: time.Now(),
        })
    }

    return nil
}
```

3. **Subscribe in CardEffectSubscriber** (if needed for card effects):
```go
func (s *CardEffectSubscriber) Start() {
    s.eventBus.Subscribe("card_discarded", func(e events.Event) {
        event := e.(events.CardDiscardedEvent)
        s.OnCardDiscarded(event)
    })
}

func (s *CardEffectSubscriber) OnCardDiscarded(event CardDiscardedEvent) {
    // Check for passive effects triggered by card discard
    // Execute effects if conditions match
}
```

## Testing Event-Driven Code

### Unit Testing Actions

Actions should be tested without worrying about events:

```go
func TestConvertHeatToTemperature(t *testing.T) {
    // Setup mocks
    mockGameRepo := &MockGameRepository{}
    mockSessionMgr := &MockSessionManager{}

    action := NewConvertHeatToTemperatureAction(mockGameRepo, mockSessionMgr)

    // Execute action
    err := action.Execute(ctx, sess, playerID)

    // Assert direct effects only
    assert.NoError(t, err)
    assert.Equal(t, 7, player.Resources().Heat) // Started with 15, spent 8
    assert.Equal(t, -28, game.Temperature)      // Started at -30, raised by 2

    // Don't test passive effects here - that's CardEffectSubscriber's job
}
```

### Integration Testing Event Flow

Test complete event flow in integration tests:

```go
func TestPassiveEffectTriggered(t *testing.T) {
    // Setup real event bus
    eventBus := events.NewEventBus()

    // Setup repositories with event bus
    gameRepo := game.NewRepository(eventBus)

    // Setup subscriber
    subscriber := events.NewCardEffectSubscriber(eventBus, gameRepo, cardRepo)
    subscriber.Start()

    // Execute action
    action.Execute(ctx, sess, playerID)

    // Wait for async event processing
    time.Sleep(10 * time.Millisecond)

    // Assert passive effects were applied
    assert.Equal(t, 1, player.Resources().Plants) // Arctic Algae effect
}
```

## Common Patterns

### Pattern: Resource Triggers

Many cards trigger on resource changes:

```json
{
  "trigger": {
    "type": "resources-changed",
    "resource_type": "steel",
    "condition": "increased"
  },
  "output": {
    "type": "resources",
    "resources": [{"type": "credits", "amount": 1}]
  }
}
```

Implementation:
1. `Player.SetResources()` publishes `ResourcesChangedEvent` with delta map
2. `CardEffectSubscriber.OnResourcesChanged()` receives event
3. Checks all played cards for matching triggers
4. Awards resources via `Player.SetResources()` (publishes new event)

### Pattern: Global Parameter Triggers

Cards trigger when global parameters change:

```json
{
  "trigger": {
    "type": "temperature-raised"
  },
  "output": {
    "type": "terraform-rating",
    "amount": 1
  }
}
```

Implementation:
1. `GameRepository.UpdateTemperature()` publishes `TemperatureChangedEvent`
2. `CardEffectSubscriber.OnTemperatureChanged()` receives event
3. Awards TR to players with matching effects
4. `Player.SetTerraformRating()` publishes `TerraformRatingChangedEvent`

### Pattern: Tile Placement Triggers

Cards trigger when tiles are placed:

```json
{
  "trigger": {
    "type": "tile-placed",
    "tile_types": ["city"]
  },
  "output": {
    "type": "resources",
    "resources": [{"type": "credits", "amount": 2}]
  }
}
```

Implementation:
1. `BoardRepository.PlaceTile()` publishes `TilePlacedEvent`
2. `CardEffectSubscriber.OnTilePlaced()` receives event
3. Filters by tile type and awards resources

## Benefits of Event-Driven Architecture

1. **Separation of Concerns**: Actions focus on their primary responsibility, effects handled separately
2. **Extensibility**: New card effects don't require changing action code
3. **Testability**: Actions can be tested independently of effects
4. **Auditability**: All state changes produce events that can be logged/traced
5. **Real-time Updates**: Automatic broadcasting when state changes
6. **Consistency**: Single source of truth for "what happened" (the event)

## Anti-Patterns to Avoid

### ❌ Manual Effect Checking in Actions

```go
// WRONG - Don't do this
func (a *Action) Execute(...) {
    gameRepo.UpdateTemperature(...)

    // Don't manually check for effects
    for _, player := range players {
        for _, card := range player.Cards() {
            if card.HasTemperatureTrigger() {
                // Apply effect...
            }
        }
    }
}
```

### ❌ Direct State Mutation Without Events

```go
// WRONG - Don't bypass event publishing
func (r *Repository) UpdateTemperature(...) {
    game.Temperature = newValue
    // Forgot to publish event! Subscribers won't trigger.
}
```

### ❌ Synchronous Blocking in Event Handlers

```go
// WRONG - Event handlers should be fast
func (s *CardEffectSubscriber) OnEvent(event Event) {
    time.Sleep(5 * time.Second) // Blocks event bus!
    // Do processing...
}
```

## Future Enhancements

Potential improvements to the event system:

1. **Event Replay**: Store events for debugging/replay
2. **Event Sourcing**: Reconstruct game state from event log
3. **Async Event Processing**: Process effects in background workers
4. **Event Priority**: Order effects by priority when multiple trigger
5. **Event Batching**: Batch broadcasts when multiple events fire rapidly
6. **Event History**: UI showing history of game events for players

## Related Documentation

- **backend/CLAUDE.md**: Overall backend architecture
- **internal/events/README.md**: Event bus implementation details
- **internal/cards/README.md**: Card effect system
