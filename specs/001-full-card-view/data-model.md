# Data Model: Expandable Card Descriptions

**Date**: 2026-02-18 | **Branch**: `001-full-card-view`

## Entities

### No New Backend Entities

This feature is entirely frontend. All data uses existing generated types. VP condition and resource storage description fields in the DTOs are handled in a separate backend task.

### Existing Types Used

- **CardDto**: Card data with `behaviors?: CardBehaviorDto[]`, `vpConditions?: VPConditionDto[]`, `resourceStorage?: ResourceStorageDto`
- **PlayerCardDto**: Extends CardDto with playability state
- **CardBehaviorDto**: Behavior with `description?: string` — the key field for hover descriptions
- **VPConditionDto**: VP condition with `description?: string` (after DTO update)
- **ResourceStorageDto**: Storage info with `description?: string` (after DTO update)

### Frontend State (hover management)

#### BehaviorSection Hover State

| Field | Type | Description |
|-------|------|-------------|
| hoveredBehaviorIndex | number \| null | Index of currently hovered behavior row |

#### Card-Level Hover Description State

| Field | Type | Description |
|-------|------|-------------|
| bottomDescription | string \| null | Description text to show at card bottom (from VP or storage hover) |

## Relationships

```
CardDto / PlayerCardDto
  ├── behaviors?: CardBehaviorDto[]
  │     └── description?: string    ← Shown as overlay below hovered behavior row
  ├── vpConditions?: VPConditionDto[]
  │     └── description?: string    ← Shown at card bottom on VP icon hover
  └── resourceStorage?: ResourceStorageDto
        └── description?: string    ← Shown at card bottom on storage icon hover
```

## State Transitions

### Behavior Row Hover
```
no hover → behavior row hovered (description overlay appears) → mouse leaves (overlay disappears)
         → different row hovered (previous overlay disappears, new one appears)
```

### VP / Resource Storage Hover
```
no hover → VP icon hovered (description at card bottom) → mouse leaves (disappears)
         → storage icon hovered (description at card bottom) → mouse leaves (disappears)
```
