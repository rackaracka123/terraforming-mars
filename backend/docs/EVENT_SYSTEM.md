# Event-Driven Effect System

## Overview

The event-driven effect system decouples game actions from their side effects using a publish-subscribe pattern. This architecture ensures that **services only do what the action explicitly says** - all passive effects and bonuses are handled automatically through domain events.

### Core Principle

> **Services execute actions. Events trigger effects.**

When a player performs an action (e.g., places a tile), the service:
1. ✅ Updates the game state (place the tile)
2. ✅ Awards immediate bonuses (tile placement bonuses from the board)
3. ❌ Does NOT manually check for card effects
4. ❌ Does NOT trigger passive abilities directly

Instead, the **repository publishes an event** when state changes, and **CardEffectSubscriber listens for events** to trigger passive card effects automatically.

## Architecture Components

### 1. EventBus (`internal/events/event_bus.go`)

Type-safe event publishing and subscription system.

```go
// Subscribe to an event type
subID := events.Subscribe(eventBus, func(event repository.TilePlacedEvent) {
    // Handle event
})

// Publish an event
events.Publish(eventBus, repository.TilePlacedEvent{
    GameID:   gameID,
    PlayerID: playerID,
    TileType: "city-tile",
    Q: 0, R: 0, S: 0,
})

// Unsubscribe when done
eventBus.Unsubscribe(subID)
```

### 2. Domain Events (`internal/repository/*_events.go`)

Events published by repositories when game state changes:

**Game Events:**
- `TemperatureChangedEvent` - Global temperature parameter changed
- `OxygenChangedEvent` - Global oxygen parameter changed
- `OceansChangedEvent` - Global ocean count changed
- `TilePlacedEvent` - Tile placed on the Mars board
- `GamePhaseChangedEvent` - Game phase transition
- `GenerationAdvancedEvent` - New generation started

**Player Events:**
- `ResourcesChangedEvent` - Player resource amounts changed
- `ProductionChangedEvent` - Player production changed
- `TerraformRatingChangedEvent` - Player TR changed
- `CardPlayedEvent` - Card played by player
- `CorporationSelectedEvent` - Corporation selected

### 3. CardEffectSubscriber (`internal/cards/effect_subscriber.go`)

Manages passive card effect subscriptions. When a card is played:

```go
// Subscribe card's passive effects
effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, cardID, card)

// CardEffectSubscriber automatically:
// 1. Analyzes card behaviors for auto-triggers with conditions
// 2. Subscribes to appropriate domain events
// 3. Executes behavior outputs when events fire
// 4. Filters by Target (self-player vs any-player)
```

## Event Flow Example

### Example: Ocean Adjacency Bonus

**Scenario:** Player places a greenery tile adjacent to an ocean and should gain 2 MC per adjacent ocean.

#### Traditional Approach (❌ Old System)
```go
// PlayerService.placeTile() - WRONG
func (s *PlayerServiceImpl) placeTile(...) {
    gameRepo.UpdateTileOccupancy(...)

    // ❌ BAD: Service manually checks for effects
    adjacentOceans := s.calculateAdjacentOceans(...)
    if adjacentOceans > 0 {
        player.Resources.Credits += adjacentOceans * 2
        playerRepo.UpdateResources(...)
    }
}
```

**Problems:**
- Service has too many responsibilities
- Hard to add new card effects
- Business logic scattered across services

#### Event-Driven Approach (✅ New System)

```go
// PlayerService.placeTile() - CORRECT
func (s *PlayerServiceImpl) placeTile(...) {
    // 1. Do ONLY what the action says: place the tile
    gameRepo.UpdateTileOccupancy(ctx, gameID, coordinate, occupant, &playerID)

    // 2. Award immediate board bonuses (defined by board, not cards)
    s.awardTilePlacementBonuses(ctx, gameID, playerID, coordinate)

    // 3. That's it! Event system handles the rest
}
```

```go
// GameRepository.UpdateTileOccupancy() - Publishes event
func (r *GameRepositoryImpl) UpdateTileOccupancy(...) error {
    // Update game state
    tile.Occupant = occupant

    // Publish domain event
    if r.eventBus != nil && occupant != nil {
        events.Publish(r.eventBus, TilePlacedEvent{
            GameID:    gameID,
            PlayerID:  playerID,
            TileType:  string(occupant.Type),
            Q: coord.Q, R: coord.R, S: coord.S,
        })
    }
    return nil
}
```

```go
// CardEffectSubscriber - Automatically handles ocean adjacency
// When greenery is placed, TilePlacedEvent fires
// Subscriber calculates adjacent oceans and awards bonus
// No manual service logic needed!
```

## Supported Trigger Types

CardEffectSubscriber currently supports these trigger conditions:

| Trigger Type | Domain Event | Description |
|--------------|--------------|-------------|
| `TriggerTemperatureRaise` | `TemperatureChangedEvent` | Temperature increased |
| `TriggerOxygenRaise` | `OxygenChangedEvent` | Oxygen increased |
| `TriggerOceanPlaced` | `OceansChangedEvent` | Ocean tile placed |
| `TriggerCityPlaced` | `TilePlacedEvent` | City tile placed (any player or self) |
| `TriggerGreeneryPlaced` | `TilePlacedEvent` | Greenery tile placed (any player or self) |
| `TriggerTilePlaced` | `TilePlacedEvent` | Any tile placed |

### Target Filtering

Effects can target different players:

```go
// Effect that only triggers for the card owner
output.Target = model.TargetSelfPlayer
// Example: Tharsis Republic +3 MC when YOU place a city

// Effect that triggers for any player
output.Target = model.TargetAnyPlayer
// Example: Rover Construction +2 MC when ANY city is placed
```

CardEffectSubscriber automatically filters events by target:
- **TargetSelfPlayer**: Only applies if `event.PlayerID == cardOwnerID`
- **TargetAnyPlayer**: Applies for any player's action

## Adding New Card Effects

### Step 1: Define Card Behavior in JSON

```json
{
  "id": "038",
  "name": "Rover Construction",
  "behaviors": [{
    "triggers": [{
      "type": "auto",
      "condition": {"type": "city-placed"}
    }],
    "outputs": [{
      "type": "credits",
      "amount": 2,
      "target": "self-player"
    }]
  }]
}
```

### Step 2: Ensure Event is Published

Check that the relevant repository publishes the event:

```go
// Example: GameRepository.UpdateTileOccupancy already publishes TilePlacedEvent
// No additional code needed for city-placed trigger!
```

### Step 3: Subscribe Effects When Card is Played

```go
// CardService.OnPlayCard() already calls:
effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, cardID, card)

// CardEffectSubscriber automatically:
// - Detects the city-placed trigger
// - Subscribes to TilePlacedEvent
// - Filters by target
// - Applies credits when triggered
```

### Step 4: Test with Integration Test

```go
func TestRoverConstruction_PassiveEffect(t *testing.T) {
    // 1. Play Rover Construction card
    cardService.OnPlayCard(ctx, gameID, playerID, "038", nil, nil)

    // 2. Place a city tile
    gameRepo.UpdateTileOccupancy(ctx, gameID, coord, cityTile, &playerID)

    // 3. Verify credits increased by 2
    player, _ := playerRepo.GetByID(ctx, gameID, playerID)
    assert.Equal(t, expectedCredits+2, player.Resources.Credits)
}
```

## Adding New Domain Events

To add support for a new trigger type:

### 1. Define the Event Struct

```go
// internal/repository/new_events.go
type CardPlayedWithTagEvent struct {
    GameID    string
    PlayerID  string
    CardID    string
    Tags      []string
    Timestamp time.Time
}
```

### 2. Publish Event from Repository

```go
// PlayerRepository.MarkCardAsPlayed()
func (r *PlayerRepositoryImpl) MarkCardAsPlayed(...) error {
    // Update state
    player.PlayedCards = append(player.PlayedCards, cardID)

    // Publish event
    if r.eventBus != nil {
        events.Publish(r.eventBus, CardPlayedWithTagEvent{
            GameID:   gameID,
            PlayerID: playerID,
            CardID:   cardID,
            Tags:     card.Tags,
        })
    }
    return nil
}
```

### 3. Add Handler to CardEffectSubscriber

```go
// internal/cards/effect_subscriber.go
func (ces *CardEffectSubscriberImpl) subscribeEffectByTriggerType(...) {
    switch triggerType {
    case model.TriggerTagPlayed:
        subID := events.Subscribe(ces.eventBus, func(event repository.CardPlayedWithTagEvent) {
            // Check if event matches card's trigger condition
            if event.GameID == gameID && hasRequiredTag(event.Tags, trigger.Condition) {
                ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
            }
        })
        return subID, nil
    }
}
```

## Best Practices

### ✅ DO

1. **Publish events in repositories** when state changes
2. **Use CardEffectSubscriber** for passive card effects
3. **Keep services focused** on their primary action
4. **Filter by target** when effects should only apply to specific players
5. **Subscribe on card play**, unsubscribe on card removal

### ❌ DON'T

1. **Don't trigger effects manually** in service layer
2. **Don't check player.Effects** directly (removed in migration)
3. **Don't add business logic** to event handlers (keep them thin)
4. **Don't forget to unsubscribe** when cards are removed
5. **Don't publish events** from service layer (only repositories)

## Debugging Events

Enable debug logging to see event flow:

```go
// Logs show event lifecycle
DEBUG: 📬 Event handler subscribed (subscription_id: sub-1, event_type: TilePlacedEvent)
DEBUG: 📢 Publishing event to subscribers (event_type: TilePlacedEvent, subscriber_count: 2)
INFO:  🌟 Passive effect triggered (card_name: Rover Construction)
INFO:  ✨ Passive effect applied (resource_type: credits, amount: 2)
```

## Migration from Old System

The old system used `player.Effects` field to store passive effects and manually polled them. This has been completely replaced:

### Old System (Removed)
- ❌ `player.Effects []PlayerEffect` - Removed from model
- ❌ `extractAndAddEffects()` - Removed from CardProcessor
- ❌ `triggerTilePlacementEffects()` - Removed from PlayerService
- ❌ `UpdateEffects()` - Removed from PlayerRepository

### New System
- ✅ `CardEffectSubscriber` - Event-driven subscription manager
- ✅ Domain Events - Published by repositories
- ✅ EventBus - Type-safe pub/sub system
- ✅ Target Filtering - Self vs any player

## Common Patterns

### Pattern 1: Global Parameter Effects

```go
// Card: "Gain 1 heat production when temperature is raised"
{
  "triggers": [{"type": "auto", "condition": {"type": "temperature-raise"}}],
  "outputs": [{"type": "heat-production", "amount": 1}]
}

// EventBus: TemperatureChangedEvent published when temperature changes
// CardEffectSubscriber: Automatically subscribes and applies heat production
```

### Pattern 2: Tile Placement Effects

```go
// Card: "Gain 2 MC when any city is placed"
{
  "triggers": [{"type": "auto", "condition": {"type": "city-placed"}}],
  "outputs": [{"type": "credits", "amount": 2, "target": "any-player"}]
}

// EventBus: TilePlacedEvent published when tile placed
// CardEffectSubscriber: Filters for city tiles, applies credits
```

### Pattern 3: Self-Only Effects

```go
// Card: "Gain 3 MC when YOU place a city"
{
  "triggers": [{"type": "auto", "condition": {"type": "city-placed"}}],
  "outputs": [{"type": "credits", "amount": 3, "target": "self-player"}]
}

// EventBus: TilePlacedEvent includes event.PlayerID
// CardEffectSubscriber: Checks event.PlayerID == cardOwnerID before applying
```

## Future Enhancements

Potential improvements to the event system:

1. **Async Event Handling** - Execute handlers in goroutines for performance
2. **Event Replay** - Store and replay events for undo/redo
3. **Event Sourcing** - Build game state from event log
4. **More Trigger Types** - card-played, tag-played, production-increased, etc.
5. **Complex Conditions** - Per-resource modifiers, conditional multipliers

## Summary

The event-driven effect system provides:
- ✅ **Separation of Concerns** - Services focus on actions, effects handled separately
- ✅ **Extensibility** - Add new cards without modifying service layer
- ✅ **Testability** - Easy to test events and effects in isolation
- ✅ **Maintainability** - Business logic centralized in CardEffectSubscriber
- ✅ **Type Safety** - Compiler-checked event types

Remember: **Services do what the action says. Events trigger the rest.**
