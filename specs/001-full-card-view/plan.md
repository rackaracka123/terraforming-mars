# Implementation Plan: Expandable Card Descriptions

**Branch**: `001-full-card-view` | **Date**: 2026-02-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-full-card-view/spec.md`

## Summary

Enhance existing BehaviorSection and SimpleGameCard (renamed to GameCard) with hover-to-reveal descriptions. Each behavior row becomes hoverable, showing its description as a floating overlay below the row. VP icon and resource storage indicator become hoverable, showing descriptions at the bottom of the card. No new component files — all changes within existing components.

## Technical Context

**Language/Version**: TypeScript 5.x (React 18)
**Primary Dependencies**: React 18, Tailwind CSS v4, existing BehaviorSection component system
**Storage**: N/A (no persistence needed)
**Testing**: Playwright CLI (headless browser validation)
**Target Platform**: Web (desktop-first, responsive)
**Project Type**: Web application (frontend only)
**Performance Goals**: Description appears within 0.5 seconds of hover
**Constraints**: No new component files, descriptions overlay without affecting card layout
**Scale/Scope**: 0 new files, ~12 modified files (3 BehaviorSection + 1 VictoryPointIcon + 1 SimpleGameCard + 8 import rename sites), 208+ cards in database

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Clean Architecture & Action Pattern | PASS | Frontend-only; no action layer changes |
| II. Code Clarity Over Comments | PASS | Self-documenting hover state names; no explanatory comments |
| III. Best Practices Enforcement | PASS | Functional components, Tailwind CSS, generated types, hooks patterns |
| IV. Complete Feature Implementation | PASS | All behavior types, VP, resource storage covered; edge cases handled |
| V. Type Safety & Generation | PASS | Uses existing generated types; ClassifiedBehavior type extended minimally |
| VI. Testing Discipline | PASS | No backend changes, so no backend tests. Playwright for frontend validation. |
| VII. No Deprecated Code | PASS | SimpleGameCard fully renamed to GameCard; old name completely removed |
| Component Standards | PASS | GameIcon for icons, Tailwind CSS only, generated types |
| State Management | PASS | Hover state is local UI state only; no game logic affected |

**Post-Phase 1 Re-check**: All gates pass. No violations.

## Project Structure

### Documentation (this feature)

```text
specs/001-full-card-view/
├── plan.md              # This file
├── spec.md              # Feature specification (refactored)
├── research.md          # Phase 0 research decisions
├── data-model.md        # Frontend state model
├── quickstart.md        # Development quickstart
├── contracts/           # No API contracts (frontend-only)
│   └── README.md
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
frontend/src/
├── components/
│   ├── ui/cards/
│   │   ├── SimpleGameCard.tsx → GameCard.tsx    # RENAMED + MODIFIED: bottom description area, storage indicator, hover state
│   │   └── BehaviorSection/
│   │       ├── BehaviorSection.tsx              # MODIFIED: thread description, manage hover state
│   │       ├── components/
│   │       │   └── BehaviorContainer.tsx        # MODIFIED: add hover handler + description overlay
│   │       └── types.ts                         # MODIFIED: add description to ClassifiedBehavior
│   ├── ui/display/
│   │   └── VictoryPointIcon.tsx                 # MODIFIED: add hover callback prop
│   ├── ui/overlay/
│   │   ├── CardFanOverlay.tsx                   # MODIFIED: import rename
│   │   ├── StartingCardSelectionOverlay.tsx     # MODIFIED: import rename
│   │   ├── PendingCardSelectionOverlay.tsx      # MODIFIED: import rename
│   │   ├── ProductionCardSelectionOverlay.tsx   # MODIFIED: import rename
│   │   ├── CardDrawSelectionOverlay.tsx         # MODIFIED: import rename
│   │   └── DemoSetupOverlay.tsx                 # MODIFIED: import rename
│   ├── ui/modals/
│   │   └── CardsPlayedModal.tsx                 # MODIFIED: import rename
│   └── pages/
│       └── CardsPage.tsx                        # MODIFIED: import rename
```

**Structure Decision**: Zero new files. All changes within existing frontend components. The rename from SimpleGameCard to GameCard changes the file name and all 8 import sites.

## Implementation Design

### Change 1: ClassifiedBehavior Type Extension

**File**: `BehaviorSection/types.ts`

Add `description?: string` to the `ClassifiedBehavior` interface. This carries through the description from `CardBehaviorDto` after classification.

### Change 2: BehaviorSection Description Threading

**File**: `BehaviorSection/BehaviorSection.tsx`

In the `classifyBehaviors()` call pipeline, ensure `description` is preserved from the original `CardBehaviorDto` into the resulting `ClassifiedBehavior`. This may require updating `classifyBehaviors()` in `behaviorClassifier.ts` to copy the description field.

Add hover state: `hoveredBehaviorIndex: number | null` managed via `useState`. Pass the index and a setter callback to each `BehaviorContainer`.

### Change 3: BehaviorContainer Hover + Description Overlay

**File**: `BehaviorSection/components/BehaviorContainer.tsx`

Add props:
- `description?: string` — the behavior's description text
- `isHovered: boolean` — whether this row is currently hovered
- `onHover: (index: number | null) => void` — callback to set/clear hover state
- `index: number` — this behavior's index

Add `onMouseEnter` and `onMouseLeave` handlers to the container div. When `isHovered` and `description` exist, render a description overlay:
- `position: relative` on the container (it may already have this)
- Description element: `position: absolute`, `top: 100%`, `left: 0`, `width: 100%`, `z-index` high enough to float over content below
- Styled as a small text block with semi-transparent background, matching card theme
- Text is the behavior's `description` string

### Change 4: VictoryPointIcon Hover Callback

**File**: `VictoryPointIcon.tsx`

Add optional prop `onHoverDescription?: (description: string | null) => void`. On `mouseEnter`, call with the appropriate VP condition description string. On `mouseLeave`, call with `null`. The component already processes `vpConditions` to determine display — extend to also extract the description.

### Change 5: Card Bottom Description Area + Resource Storage Indicator

**File**: `SimpleGameCard.tsx` (before rename)

Add state: `bottomDescription: string | null` managed via `useState`.

**VP hover integration**: Pass `onHoverDescription` callback to `VictoryPointIcon` that sets `bottomDescription`.

**Resource storage indicator**: If `card.resourceStorage` exists, render a small hoverable icon (using GameIcon with the storage resource type) near the card bottom or behavior section. On hover, set `bottomDescription` to `card.resourceStorage.description`.

**Bottom description area**: When `bottomDescription` is not null, render a text element at the bottom of the card (below behaviors, above checkbox). Positioned absolutely, styled with semi-transparent background.

### Change 6: Rename SimpleGameCard → GameCard

**Files**: SimpleGameCard.tsx → GameCard.tsx + 8 import sites

1. Rename file
2. Update component name: `SimpleGameCard` → `GameCard`
3. Update interface name: `SimpleGameCardProps` → `GameCardProps`
4. Update all 8 import sites to reference `GameCard` from the new path

## Complexity Tracking

No constitution violations. No complexity justifications needed.
