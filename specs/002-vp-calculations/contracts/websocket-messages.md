# WebSocket Contract: VP Calculations

**Feature Branch**: `002-vp-calculations`
**Date**: 2026-01-30

## Overview

VP data flows through the existing WebSocket game state synchronization. No new WebSocket message types are needed. VP granters are included in the `game-updated` message's player data.

## Modified Message: `game-updated`

The existing `game-updated` message already contains `PlayerDto`. The VP granter data is added to the self-player DTO.

### PlayerDto Addition

```json
{
  "vpGranters": [
    {
      "cardId": "birds",
      "cardName": "Birds",
      "description": "Decrease any plant production 1 step. This card stores animals.",
      "computedValue": 3,
      "conditions": [
        {
          "amount": 1,
          "conditionType": "per",
          "perType": "animal",
          "perAmount": 1,
          "count": 3,
          "computedVP": 3,
          "explanation": "1 VP per animal (3 animals = 3 VP)"
        }
      ]
    },
    {
      "cardId": "dust-seals",
      "cardName": "Dust Seals",
      "description": "Requires 3% or less oxygen.",
      "computedValue": 1,
      "conditions": [
        {
          "amount": 1,
          "conditionType": "fixed",
          "perType": null,
          "perAmount": null,
          "count": 0,
          "computedVP": 1,
          "explanation": "1 VP"
        }
      ]
    }
  ]
}
```

### OtherPlayerDto

VP granters are NOT included in other players' DTOs during gameplay (strategy information). They may be included in the endgame state if needed.

## Event Flow

No new WebSocket events. VP updates are delivered through the standard `game-updated` broadcast triggered by the existing `BroadcastEvent` system.

```
Backend Event → VP Recalculation → Player State Updated → BroadcastEvent → game-updated → Frontend
```
