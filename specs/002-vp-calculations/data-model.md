# Data Model: VP Calculations

**Feature Branch**: `002-vp-calculations`
**Date**: 2026-01-30

## Domain Entities

### VPGranter (New)

Represents a single VP source registered on a player from a played card or corporation.

| Field | Type | Description |
|-------|------|-------------|
| CardID | string | ID of the source card (project or corporation) |
| CardName | string | Display name of the source card |
| Description | string | Card description for tooltip display |
| VPConditions | []VictoryPointCondition | The VP conditions from the card's `vpConditions` array |
| ComputedValue | int | Current total VP this granter provides (sum of all conditions) |

**Relationships**:
- Belongs to a Player (stored in player's VPGranters list)
- References a Card via CardID (looked up in CardRegistry)
- Contains VictoryPointCondition structs (existing type from `storage_victory.go`)

**Validation**:
- CardID must reference a valid card in CardRegistry
- VPConditions must be non-empty (only cards with VP conditions create VPGranters)
- ComputedValue is always >= 0

### VPGranters (New Player Component)

Manages the ordered list of VPGranter instances for a player.

| Field | Type | Description |
|-------|------|-------------|
| granters | []VPGranter | Ordered list of VP sources (corporation first, then play order) |
| eventBus | *EventBusImpl | For publishing VictoryPointsChangedEvent on recalculation |
| gameID | string | Game context |
| playerID | string | Player context |

**Methods**:
- `Add(granter VPGranter)` — Append a VP source to the list
- `Prepend(granter VPGranter)` — Add corporation VP source at the beginning
- `GetAll() []VPGranter` — Return copy of all VP sources
- `TotalComputedVP() int` — Sum of all granters' ComputedValue
- `UpdateComputedValues(player, board, cardRegistry)` — Recalculate all ComputedValues

### Existing Entities (Modified)

#### Player (Modified)

Add field:
| Field | Type | Description |
|-------|------|-------------|
| vpGranters | *VPGranters | VP source tracking component |

New accessor:
- `VPGranters() *VPGranters`

#### Game (Modified — Event Subscriptions)

New subscription method:
- `subscribeToVPEvents()` — Called from constructor, subscribes to CardPlayedEvent, ResourceStorageChangedEvent, TilePlacedEvent, TagPlayedEvent, CorporationSelectedEvent

### Existing Entities (Unchanged)

#### VictoryPointCondition (Existing)

Already defined in `storage_victory.go`:

| Field | Type | Description |
|-------|------|-------------|
| Amount | int | VP value per trigger |
| Condition | VPConditionType | "fixed", "per", or "once" |
| MaxTrigger | *int | Cap on triggers (nil = unlimited) |
| Per | *PerCondition | What to count for "per" conditions |

#### PerCondition (Existing)

Already defined in `card_behavior_types.go`:

| Field | Type | Description |
|-------|------|-------------|
| ResourceType | ResourceType | What to count |
| Amount | int | Count per N items |
| Location | *string | Where to look (always "anywhere" currently) |
| Target | *string | Where resource lives (e.g., "self-card") |
| Tag | *CardTag | Which tag to count |

## DTO Entities

### VPGranterDto (New)

Transfer object for frontend consumption.

| Field | JSON Key | TypeScript Type | Description |
|-------|----------|----------------|-------------|
| CardID | cardId | string | Source card ID |
| CardName | cardName | string | Source card display name |
| Description | description | string | Card description |
| ComputedValue | computedValue | number | Current VP from this source |
| Conditions | conditions | VPGranterConditionDto[] | Per-condition breakdown |

### VPGranterConditionDto (New)

| Field | JSON Key | TypeScript Type | Description |
|-------|----------|----------------|-------------|
| Amount | amount | number | VP per trigger |
| ConditionType | conditionType | string | "fixed", "per", "once" |
| PerType | perType | string \| null | Resource/tag/tile type being counted |
| PerAmount | perAmount | number \| null | Count divisor |
| Count | count | number | Current count of items |
| ComputedVP | computedVP | number | VP from this condition |
| Explanation | explanation | string | Human-readable description |

### PlayerDto (Modified)

Add field:
| Field | JSON Key | TypeScript Type | Description |
|-------|----------|----------------|-------------|
| VPGranters | vpGranters | VPGranterDto[] | VP sources for this player |

## State Transitions

### VP Source Registration

```
Card Played (with vpConditions) → VPGranter created → Added to player's VPGranters list → VP recalculated
Corporation Selected (with vpConditions) → VPGranter created → Prepended to player's VPGranters list → VP recalculated
```

### VP Recalculation

```
Event fires (CardPlayed/ResourceStorageChanged/TilePlaced/TagPlayed) →
  For each VPGranter on player:
    For each VPCondition on granter:
      Evaluate condition (fixed/per/once) →
      Sum condition VPs →
    Update granter.ComputedValue →
  Publish VictoryPointsChangedEvent
```

## Entity Relationship Diagram

```
Player 1──* VPGranters (component)
  └── VPGranters 1──* VPGranter
        └── VPGranter 1──* VictoryPointCondition (existing)
              └── VictoryPointCondition 0..1── PerCondition (existing)

Card (via CardRegistry) ──referenced by── VPGranter.CardID
Board ──read during── VP recalculation (tile counting)
```
