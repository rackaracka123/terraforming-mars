# Quickstart: Expandable Card Descriptions

**Branch**: `001-full-card-view`

## Prerequisites

```bash
make dev-setup    # Install dependencies
make run          # Start frontend (3000) + backend (3001)
```

## Key Files to Modify

### BehaviorSection (hover descriptions)
- `frontend/src/components/ui/cards/BehaviorSection/BehaviorSection.tsx` — Thread description through, manage hover state
- `frontend/src/components/ui/cards/BehaviorSection/components/BehaviorContainer.tsx` — Add hover handler + description overlay
- `frontend/src/components/ui/cards/BehaviorSection/types.ts` — Add description to ClassifiedBehavior type

### Card-Level Hover (VP + resource storage)
- `frontend/src/components/ui/cards/SimpleGameCard.tsx` — Add bottom description area, hover state, resource storage indicator
- `frontend/src/components/ui/display/VictoryPointIcon.tsx` — Add hover callback prop

### Rename
- `frontend/src/components/ui/cards/SimpleGameCard.tsx` → `GameCard.tsx`
- 8 import sites (CardsPage, CardFanOverlay, StartingCardSelectionOverlay, PendingCardSelectionOverlay, ProductionCardSelectionOverlay, CardDrawSelectionOverlay, DemoSetupOverlay, CardsPlayedModal)

## Development Flow

1. Add `description` to ClassifiedBehavior type
2. Thread description through classification pipeline in BehaviorSection
3. Add hover state + overlay to BehaviorContainer
4. Add VP hover callback to VictoryPointIcon
5. Add bottom description area + resource storage indicator to SimpleGameCard
6. Rename SimpleGameCard → GameCard + update all imports

## Verification

```bash
# Behavior hover
# Navigate to http://localhost:3000/cards
# Hover over behavior rows — descriptions should appear as overlays

# VP hover
# Find a card with VP icon, hover over it — description at card bottom

# Resource storage hover
# Find a card with resource storage, hover over indicator — description at card bottom

# Selection modals
# Start a game, reach starting card selection, hover behavior rows

# Rename
# Search codebase for "SimpleGameCard" — zero results
```

## Quality Checks

```bash
make format       # Format all code
make lint         # Run all linters
```
