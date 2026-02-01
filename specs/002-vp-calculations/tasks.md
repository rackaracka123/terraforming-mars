# Tasks: VP Calculations

**Input**: Design documents from `/specs/002-vp-calculations/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Comprehensive tests are required (explicitly requested in spec).

**Organization**: Tasks are grouped by user story. US1 and US4 are merged (event-driven recalculation is the engine for live VP tracking). US3 extends US1 to corporation cards.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: No new project setup needed — existing codebase. This phase ensures the branch is ready.

- [x] T001 Verify branch `002-vp-calculations` is checked out and `make dev-setup` runs cleanly

---

## Phase 2: Foundational (VPGranters Component + Player Integration)

**Purpose**: Core domain types and player integration that ALL user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Create VPGranter struct and VPGranters component in `backend/internal/game/player/player_vp_granters.go`. Define VPGranter struct with fields: CardID (string), CardName (string), Description (string), VPConditions ([]cards.VictoryPointCondition), ComputedValue (int). Define VPGranters struct with mutex-protected granters slice, eventBus, gameID, playerID. Implement NewVPGranters constructor, Add (append), Prepend (prepend for corporation), GetAll (return copy), TotalComputedVP (sum ComputedValues) methods.
- [x] T003 Implement RecalculateAll method on VPGranters in `backend/internal/game/player/player_vp_granters.go`. For each granter, evaluate each VPCondition: fixed conditions return Amount directly; per-conditions with target "self-card" count card storage via player.Resources().GetCardStorage(cardID) and compute floor(count/per.Amount)*vpAmount; per-conditions with tag type use cards.CountPlayerTagsByType; per-conditions with tile types use cards.CountAllTilesOfType or adjacency helpers. Respect maxTrigger caps. Update each granter's ComputedValue. Publish VictoryPointsChangedEvent with old and new totals.
- [x] T004 Add vpGranters field to Player struct in `backend/internal/game/player/player.go`. Add `vpGranters *VPGranters` private field, initialize `NewVPGranters(eventBus, gameID, playerID)` in NewPlayer constructor, add `VPGranters() *VPGranters` accessor method.
- [x] T005 Add VPGranterDto and VPGranterConditionDto to `backend/internal/delivery/dto/game_dto.go`. VPGranterDto with fields: CardID (string, json:"cardId", ts:"string"), CardName (string, json:"cardName", ts:"string"), Description (string, json:"description", ts:"string"), ComputedValue (int, json:"computedValue", ts:"number"), Conditions ([]VPGranterConditionDto, json:"conditions", ts:"VPGranterConditionDto[]"). VPGranterConditionDto with fields: Amount (int), ConditionType (string), PerType (*string), PerAmount (*int), Count (int), ComputedVP (int), Explanation (string). All with json and ts tags.
- [x] T006 Add VPGranters field to PlayerDto in `backend/internal/delivery/dto/game_dto.go` — add `VPGranters []VPGranterDto` field with json:"vpGranters" and ts:"VPGranterDto[]" tags.
- [x] T007 Implement VP granter mapping in `backend/internal/delivery/dto/mapper_game.go`. Add toVPGranterDtos mapper function that converts []player.VPGranter to []VPGranterDto, including per-condition breakdown with human-readable explanation strings. Wire into toPlayerDto to include VP granters in self-player DTO only.
- [x] T008 Run `make generate` to produce TypeScript types for VPGranterDto and VPGranterConditionDto

**Checkpoint**: Foundation ready — VPGranters component exists on Player, DTOs defined, types generated

---

## Phase 3: User Story 1+4 — Live VP Tracking & Event-Driven Recalculation (Priority: P1)

**Goal**: When cards are played, VP sources are registered and recalculated via events. VP total updates live.

**Independent Test**: Play a card with VP conditions → VP total updates on player. Add resources to a card → VP recalculates.

### Tests for User Story 1+4

- [x] T009 [P] [US1] Write table-driven unit tests for fixed VP conditions in `backend/test/game/player/player_vp_granters_test.go`. Test cases: Dust Seals (1 VP fixed), Farming (2 VP fixed), Anti-Gravity Technology (3 VP fixed), Earth Elevator (4 VP fixed). Each test: create VPGranters, add VPGranter with card's vpConditions, call RecalculateAll, assert ComputedValue matches expected.
- [x] T010 [P] [US1] Write table-driven unit tests for per-resource-on-self-card VP conditions in `backend/test/game/player/player_vp_granters_test.go`. Test cases: Birds (1 VP per 1 animal, 5 animals = 5 VP), Small Animals (1 VP per 2 animals, 5 animals = 2 VP), Ants (1 VP per 2 microbes, 6 microbes = 3 VP), Decomposers (1 VP per 3 microbes, 9 microbes = 3 VP), Tardigrades (1 VP per 4 microbes, 8 microbes = 2 VP), Floating Habs (1 VP per 2 floaters, 6 floaters = 3 VP), Physics Complex (2 VP per 2 science, 4 science = 4 VP), Security Fleet (1 VP per 1 asteroid, 3 asteroids = 3 VP). Setup card storage on player before RecalculateAll.
- [x] T011 [P] [US1] Write table-driven unit tests for per-tag VP conditions in `backend/test/game/player/player_vp_granters_test.go`. Test cases: Ganymede Colony (1 VP per 1 jovian tag, 3 jovian tags = 3 VP), Water Import From Europa (1 VP per 1 jovian tag, 0 tags = 0 VP). Setup played cards with jovian tags before RecalculateAll.
- [x] T012 [P] [US1] Write table-driven unit tests for per-tile VP conditions in `backend/test/game/player/player_vp_granters_test.go`. Test cases: Immigration Shuttles (1 VP per 3 city tiles anywhere, 6 cities = 2 VP), Space Port Colony (1 VP per 2 colony tiles, 4 colonies = 2 VP). Setup board tiles before RecalculateAll.
- [x] T013 [P] [US1] Write unit tests for maxTrigger cap in `backend/test/game/player/player_vp_granters_test.go`. Test cases: Search For Life (3 VP per 3 science, maxTrigger:1) with 3+ science = 3 VP (capped), with 2 science = 0 VP (not met). Also test zero resources on per-condition cards = 0 VP.
- [x] T014 [P] [US1] Write unit tests for VP ordering and multiple granters in `backend/test/game/player/player_vp_granters_test.go`. Test: add granter A then B, GetAll returns [A, B]. Test: Prepend corp then Add card, GetAll returns [corp, card]. Test: multiple granters with mixed conditions, TotalComputedVP equals sum.

### Implementation for User Story 1+4

- [x] T015 [US1] Implement subscribeToVPEvents method in `backend/internal/game/game.go`. Call from NewGame constructor (same pattern as subscribeToGenerationalEvents). Subscribe to CardPlayedEvent: look up card in CardRegistry, if card.VPConditions is non-empty, create VPGranter with card info and Add to player's VPGranters, then call RecalculateAll with player, board, cardRegistry. Subscribe to ResourceStorageChangedEvent: get player, call RecalculateAll on player's VPGranters. Subscribe to TagPlayedEvent: get player, call RecalculateAll. Subscribe to TilePlacedEvent: call RecalculateAll on ALL players' VPGranters (tile changes affect all players).
- [x] T016 [US1] Write integration tests for event-driven recalculation in `backend/test/game/player/player_vp_granters_test.go`. Test: play Birds card → VPGranter registered with 0 VP → add 3 animals to Birds storage → ResourceStorageChangedEvent fires → VP recalculates to 3. Test: play Ganymede Colony → 2 jovian tags → VP = 2 → play card with jovian tag (TagPlayedEvent) → VP = 3. Test: tile placed → TilePlacedEvent → all players recalculate.
- [x] T017 [US1] Run `make test` to verify all VP tests pass

**Checkpoint**: Live VP tracking works — cards register VP sources, events trigger recalculation, all condition types compute correctly

---

## Phase 4: User Story 3 — VP Sources from Corporation Cards (Priority: P2)

**Goal**: Corporation cards with VP conditions are registered as VP sources (prepended to list).

**Independent Test**: Select a corporation with VP conditions → it appears first in VP granter list with correct computed value.

### Tests for User Story 3

- [x] T018 [P] [US3] Write unit tests for corporation VP in `backend/test/game/player/player_vp_granters_test.go`. Test cases: Arklight (1 VP per 2 animals on self-card, 4 animals = 2 VP), Celestic (1 VP per 3 floaters on self-card, 9 floaters = 3 VP). Test: corporation VP granter is prepended (first in list when cards are also present).

### Implementation for User Story 3

- [x] T019 [US3] Add CorporationSelectedEvent subscription in subscribeToVPEvents in `backend/internal/game/game.go`. On CorporationSelectedEvent: look up corporation card in CardRegistry, if card.VPConditions is non-empty, create VPGranter and Prepend to player's VPGranters, call RecalculateAll.
- [x] T020 [US3] Write integration test for corporation VP event flow in `backend/test/game/player/player_vp_granters_test.go`. Test: select Arklight corporation → VPGranter prepended → add animals → RecalculateAll → VP updates correctly.
- [x] T021 [US3] Run `make test` to verify corporation VP tests pass

**Checkpoint**: Corporation VP sources work identically to card VP sources, prepended to list

---

## Phase 5: User Story 2 — Interactive VP Breakdown Modal (Priority: P2)

**Goal**: Rewrite VictoryPointsModal with horizontal stacked bar, hover tooltips, and card name labels.

**Independent Test**: Open VP modal → see colored bar segments for each VP source → hover shows tooltip with card info.

### Implementation for User Story 2

- [x] T022 [US2] Rewrite VictoryPointsModal in `frontend/src/components/ui/modals/VictoryPointsModal.tsx`. Replace all existing logic. New props: vpGranters (VPGranterDto[]) and totalVP (number). Use GameModal with theme="victoryPoints". Header: GameModalHeader with title "Victory Points" and VictoryPointsDisplay showing totalVP. Body: horizontal stacked bar with segments proportional to each granter's computedValue. Define color palette array of 12+ curated colors matching space theme (shades of amber, teal, indigo, rose, emerald, violet, cyan, orange, fuchsia, lime, sky, pink). Each granter gets color by index (cycling if > 12). Segments with 0 computedValue are hidden from the bar.
- [x] T023 [US2] Implement card name labels above bar in `frontend/src/components/ui/modals/VictoryPointsModal.tsx`. For each visible segment, render a small text label (text-[10px] font-medium text-white/70) positioned right-aligned above the segment. Labels should align with the right edge of their corresponding segment using absolute positioning relative to the bar container.
- [x] T024 [US2] Implement hover interaction and tooltip in `frontend/src/components/ui/modals/VictoryPointsModal.tsx`. Track hovered segment index in state. On hover: apply brightness(1.2) CSS filter to segment, render tooltip div positioned above the hovered segment. Tooltip content: card name (bold), card description, VP breakdown text from condition explanation (e.g., "3 VP — 1 per animal (3 animals)"). On mouse leave: clear hovered state, tooltip disappears. Transitions should be smooth (transition-all duration-150).
- [x] T025 [US2] Update the VP button/component that opens VictoryPointsModal to pass vpGranters data from the game state DTO. Find where VictoryPointsModal is currently rendered, update the props to use the new vpGranters field from PlayerDto and player's total VP.
- [x] T026 [US2] Run `make format` and `make lint` to verify frontend code passes quality checks

**Checkpoint**: VP modal shows interactive stacked bar with per-card breakdown, hover tooltips, and card labels

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all stories

- [x] T027 Run `make generate` to ensure TypeScript types are in sync
- [x] T028 Run `make test` to verify all backend tests pass
- [x] T029 Run `make format` and `make lint` to verify all code passes quality checks
- [x] T030 Run `make prepare-for-commit` for final validation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **US1+4 (Phase 3)**: Depends on Phase 2 — core VP tracking and event system
- **US3 (Phase 4)**: Depends on Phase 2 (can run in parallel with Phase 3, but logically extends it)
- **US2 (Phase 5)**: Depends on Phase 2 (needs generated TypeScript types from T008)
- **Polish (Phase 6)**: Depends on all phases complete

### User Story Dependencies

- **US1+4 (P1)**: Depends on Foundational only. Core VP engine.
- **US3 (P2)**: Technically independent from US1+4 at the code level (just adds another event subscription), but logically extends it. Can run in parallel with US1+4 if needed.
- **US2 (P2)**: Independent from US1+4 and US3 at the code level (frontend only). Depends only on generated types from Foundational phase. Can run in parallel with backend phases.

### Within Each User Story

- Tests written FIRST, then implementation
- Domain types before DTOs
- DTOs before frontend consumption
- Integration tests after unit tests

### Parallel Opportunities

**Phase 2** (Foundational):
- T005 and T006 can run in parallel (both modify game_dto.go but different sections)
- T002, T003 are sequential (same file)
- T004 can run in parallel with T002/T003 (different file)

**Phase 3** (US1+4 Tests):
- T009, T010, T011, T012, T013, T014 — all test functions in the same file but can be written in parallel

**Cross-phase parallelism**:
- Phase 5 (US2 frontend) can start as soon as T008 (type generation) completes
- Phase 4 (US3) can run in parallel with Phase 3 (US1+4) after Foundational

---

## Parallel Example: User Story 1+4

```bash
# Launch all test tasks in parallel (same file, different test functions):
Task: "T009 - Fixed VP condition tests"
Task: "T010 - Per-resource VP condition tests"
Task: "T011 - Per-tag VP condition tests"
Task: "T012 - Per-tile VP condition tests"
Task: "T013 - MaxTrigger cap tests"
Task: "T014 - Ordering tests"

# Then implementation:
Task: "T015 - Event subscriptions in game.go"
Task: "T016 - Integration tests"
Task: "T017 - Run make test"
```

---

## Implementation Strategy

### MVP First (User Story 1+4 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (VPGranters component + DTOs + type generation)
3. Complete Phase 3: US1+4 (event subscriptions + all VP condition tests)
4. **STOP and VALIDATE**: Run `make test` — all VP calculations correct
5. Backend is fully functional for live VP tracking

### Incremental Delivery

1. Complete Foundational → VPGranters component ready
2. Add US1+4 → All VP conditions compute correctly (MVP!)
3. Add US3 → Corporation VP sources work
4. Add US2 → Frontend modal with interactive bar
5. Polish → Format, lint, final validation

### Parallel Team Strategy

With two developers:
1. Both complete Foundational together
2. Developer A: US1+4 (backend event system + tests) + US3 (corporation extension)
3. Developer B: US2 (frontend modal rewrite) — starts after T008 (type generation)
4. Both converge at Polish phase
