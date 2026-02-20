# Tasks: Expandable Card Descriptions

**Input**: Design documents from `/specs/001-full-card-view/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md

**Tests**: No test tasks generated (not explicitly requested). Frontend verified via Playwright CLI per verification plan.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: No project initialization needed — existing codebase. Verify branch is ready.

- [x] T001 Verify branch `001-full-card-view` is checked out and clean with `git status`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Type extension that all user stories depend on — adding `description` to the classification pipeline.

**CRITICAL**: No user story work can begin until this phase is complete.

- [x] T002 Add `description?: string` field to the `ClassifiedBehavior` interface in `frontend/src/components/ui/cards/BehaviorSection/types.ts` — This carries the description from `CardBehaviorDto` through the classification pipeline so BehaviorContainer can access it.
- [x] T003 Update `classifyBehaviors()` in `frontend/src/components/ui/cards/BehaviorSection/utils/behaviorClassifier.ts` to copy the `description` field from `CardBehaviorDto` into each resulting `ClassifiedBehavior` object. Ensure all classification paths (manual-action, triggered-effect, immediate-production, immediate-effect, auto-no-background, discount, payment-substitute, value-modifier, defense) preserve the description.

**Checkpoint**: ClassifiedBehavior objects now carry description data through the pipeline. No visual changes yet.

---

## Phase 3: User Story 1 — Hoverable Behavior Descriptions (Priority: P1)

**Goal**: Each behavior row in BehaviorSection becomes hoverable, revealing a description text overlay below the row on hover.

**Independent Test**: Navigate to /cards, hover over any behavior row on a card — a description should appear underneath that row.

### Implementation for User Story 1

- [x] T004 [US1] Add hover state management to `BehaviorSection` in `frontend/src/components/ui/cards/BehaviorSection/BehaviorSection.tsx` — Add `hoveredBehaviorIndex: number | null` state via `useState`. Create `handleBehaviorHover` callback `(index: number | null) => void` that sets the state. Pass `description`, `index`, `isHovered` (computed from `hoveredBehaviorIndex === index`), and `onHover` (the callback) as new props to each `BehaviorContainer` rendered in `renderBehavior()`.
- [x] T005 [US1] Add hover handler and description overlay to `BehaviorContainer` in `frontend/src/components/ui/cards/BehaviorSection/components/BehaviorContainer.tsx` — Add props: `description?: string`, `isHovered?: boolean`, `onHover?: (index: number | null) => void`, `index?: number`. Add `onMouseEnter` handler that calls `onHover(index)` and `onMouseLeave` handler that calls `onHover(null)`. Ensure the container div has `position: relative`. When `isHovered === true` and `description` exists, render a description overlay element: `position: absolute`, `top: 100%`, `left: 0`, `width: 100%`, `z-index: 50`, styled with semi-transparent dark background, small text, padding, matching the card's visual theme. The overlay floats over content below without affecting card layout.

**Checkpoint**: Hover over any behavior row with a description → text appears as overlay below that row. Move mouse away → disappears. Only one description visible at a time.

---

## Phase 4: User Story 2 — VP and Resource Storage Hover Descriptions (Priority: P2)

**Goal**: Hovering the VP icon or resource storage indicator on a card shows a description at the bottom of the card.

**Independent Test**: On /cards, find a card with VP icon, hover over it — description appears at the bottom of the card.

### Implementation for User Story 2

- [x] T006 [P] [US2] Add hover callback prop to `VictoryPointIcon` in `frontend/src/components/ui/display/VictoryPointIcon.tsx` — Add optional prop `onHoverDescription?: (description: string | null) => void`. Add `onMouseEnter` handler that extracts the description from `vpConditions` and calls `onHoverDescription(description)`. Add `onMouseLeave` handler that calls `onHoverDescription(null)`. The component already processes `vpConditions` to determine display — extend to also extract the description string.
- [x] T007 [P] [US2] Add bottom description area, VP hover integration, and resource storage indicator to `SimpleGameCard` in `frontend/src/components/ui/cards/SimpleGameCard.tsx` — (1) Add state: `bottomDescription: string | null` via `useState`. (2) Pass `onHoverDescription` callback to `VictoryPointIcon` that sets `bottomDescription`. (3) If `card.resourceStorage` exists, render a small hoverable icon (using GameIcon with the storage resource type) near the card bottom or behavior section area. On `mouseEnter`, set `bottomDescription` to `card.resourceStorage.description`. On `mouseLeave`, set to `null`. (4) When `bottomDescription` is not null, render a text element at the bottom of the card (below behaviors, above checkbox): positioned absolutely, semi-transparent dark background, small text, matching card theme. When null, render nothing.

**Checkpoint**: Hover VP icon → description at card bottom. Hover resource storage indicator → description at card bottom. Stop hovering → description disappears. No description area if VP/storage has no description.

---

## Phase 5: User Story 3 — Rename SimpleGameCard to GameCard (Priority: P3)

**Goal**: Rename SimpleGameCard to GameCard throughout the codebase. No functional changes.

**Independent Test**: Search codebase for "SimpleGameCard" — zero results. All card rendering works identically.

### Implementation for User Story 3

- [x] T008 [US3] Rename file `frontend/src/components/ui/cards/SimpleGameCard.tsx` to `frontend/src/components/ui/cards/GameCard.tsx` — Rename the file via `git mv`. Inside the file, rename the component function from `SimpleGameCard` to `GameCard`, rename the interface from `SimpleGameCardProps` to `GameCardProps`, and update the default export.
- [x] T009 [P] [US3] Update import in `frontend/src/components/pages/CardsPage.tsx` — Change import from `SimpleGameCard` to `GameCard` from the new path `../ui/cards/GameCard`. Update all JSX usage of `<SimpleGameCard` to `<GameCard`.
- [x] T010 [P] [US3] Update import in `frontend/src/components/ui/overlay/CardFanOverlay.tsx` — Change import from `SimpleGameCard` to `GameCard` from the new path `../cards/GameCard`. Update all JSX usage.
- [x] T011 [P] [US3] Update import in `frontend/src/components/ui/overlay/StartingCardSelectionOverlay.tsx` — Change import from `SimpleGameCard` to `GameCard` from the new path `../cards/GameCard`. Update all JSX usage.
- [x] T012 [P] [US3] Update import in `frontend/src/components/ui/overlay/PendingCardSelectionOverlay.tsx` — Change import from `SimpleGameCard` to `GameCard` from the new path `../cards/GameCard`. Update all JSX usage.
- [x] T013 [P] [US3] Update import in `frontend/src/components/ui/overlay/ProductionCardSelectionOverlay.tsx` — Change import from `SimpleGameCard` to `GameCard` from the new path `../cards/GameCard`. Update all JSX usage.
- [x] T014 [P] [US3] Update import in `frontend/src/components/ui/overlay/CardDrawSelectionOverlay.tsx` — Change import from `SimpleGameCard` to `GameCard` from the new path `../cards/GameCard`. Update all JSX usage.
- [x] T015 [P] [US3] Update import in `frontend/src/components/ui/overlay/DemoSetupOverlay.tsx` — Change import from `SimpleGameCard` to `GameCard` from the new path `../cards/GameCard`. Update all JSX usage.
- [x] T016 [P] [US3] Update import in `frontend/src/components/ui/modals/CardsPlayedModal.tsx` — Change import from `SimpleGameCard` to `GameCard` from the new path `../../ui/cards/GameCard`. Update all JSX usage.

**Checkpoint**: `grep -r "SimpleGameCard" frontend/src/` returns zero results. All card rendering works identically to before.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Code quality, formatting, and final validation.

- [x] T017 Run `make format` and `make lint` to ensure all code passes formatting and linting
- [x] T018 Run `make prepare-for-commit` to verify everything is ready for commit

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — branch verification only
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
  - T002 → T003 (T003 depends on the type defined in T002)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (T003)
  - T004 → T005 (T005 depends on props passed from T004)
- **User Story 2 (Phase 4)**: Depends on Phase 2 (T003) — no dependency on US1
  - T006 + T007 can run in parallel (different files)
- **User Story 3 (Phase 5)**: Depends on US1 + US2 being complete (rename after all functional changes)
  - T008 must be first (file rename)
  - T009–T016 can all run in parallel (different files, same pattern)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends only on Foundational (Phase 2) — no cross-story dependencies
- **User Story 2 (P2)**: Depends only on Foundational (Phase 2) — no dependency on US1
- **User Story 3 (P3)**: Depends on US1 + US2 — rename after all functional changes to minimize conflicts

### Parallel Opportunities

- T006 + T007: VictoryPointIcon + SimpleGameCard changes (different files)
- T009–T016: All 8 import rename sites (different files, same pattern)
- US1 and US2 can proceed in parallel after Phase 2

---

## Parallel Example: User Story 2

```
# Parallel component modifications:
T006: Add hover callback to VictoryPointIcon.tsx
T007: Add bottom description area to SimpleGameCard.tsx
```

## Parallel Example: User Story 3

```
# Sequential first:
T008: Rename SimpleGameCard.tsx → GameCard.tsx (git mv + internal rename)

# Then all in parallel:
T009: CardsPage.tsx
T010: CardFanOverlay.tsx
T011: StartingCardSelectionOverlay.tsx
T012: PendingCardSelectionOverlay.tsx
T013: ProductionCardSelectionOverlay.tsx
T014: CardDrawSelectionOverlay.tsx
T015: DemoSetupOverlay.tsx
T016: CardsPlayedModal.tsx
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (verify branch)
2. Complete Phase 2: Foundational (ClassifiedBehavior type + classifier update)
3. Complete Phase 3: User Story 1 (BehaviorSection + BehaviorContainer hover)
4. **STOP and VALIDATE**: Navigate to /cards, hover behavior rows — descriptions appear/disappear
5. This alone provides the core value — players can see behavior descriptions inline

### Incremental Delivery

1. Phase 2: Foundational → Description flows through classification pipeline
2. + User Story 1 → Behavior hover descriptions work → **MVP!**
3. + User Story 2 → VP and resource storage descriptions at card bottom
4. + User Story 3 → Rename SimpleGameCard → GameCard (clean codebase)
5. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable after Phase 2
- No test tasks generated (tests not explicitly requested) — use Playwright CLI verification per spec
- No new component files — all changes within existing components (FR-010)
- No backend changes in this task list — DTO updates handled in a separate task
- Commit after each phase checkpoint to maintain incremental progress
