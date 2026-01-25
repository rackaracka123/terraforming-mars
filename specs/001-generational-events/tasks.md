# Tasks: Generational Events System

**Input**: Design documents from `/specs/001-generational-events/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Tests**: Backend tests included as specified in plan.md (Testing Discipline principle).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Backend**: `backend/internal/`, `backend/test/`
- **Frontend**: `frontend/src/`
- **Assets**: `backend/assets/cards/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create new type definitions and shared infrastructure

- [x] T001 [P] Create GenerationalEvent enum and types in `backend/internal/game/shared/generational_events.go`
- [x] T002 [P] Add MinMax struct with json/ts tags in `backend/internal/game/shared/generational_events.go`
- [x] T003 [P] Add GenerationalEventRequirement struct in `backend/internal/game/shared/generational_events.go`
- [x] T004 [P] Add PlayerGenerationalEventEntry struct in `backend/internal/game/shared/generational_events.go`
- [x] T005 Add GenerationalEventRequirements field to CardBehavior in `backend/internal/game/shared/behavior.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core player component and DTO infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T006 Create GenerationalEvents component in `backend/internal/game/player/generational_events.go` with Increment, GetCount, GetAll, Clear methods
- [x] T007 Add generationalEvents field and GenerationalEvents() accessor to Player in `backend/internal/game/player/player.go`
- [x] T008 Initialize generationalEvents in newPlayer constructor in `backend/internal/game/player/player.go`
- [x] T009 [P] Add ErrorCodeGenerationalEventNotMet constant in `backend/internal/game/player/state.go`
- [x] T010 [P] Add PlayerGenerationalEventEntryDto struct in `backend/internal/delivery/dto/game_dto.go`
- [x] T011 [P] Add GenerationalEvents field to PlayerDto in `backend/internal/delivery/dto/game_dto.go`
- [x] T012 Add GenerationalEvents mapping in ToPlayerDto function in `backend/internal/delivery/dto/mapper.go`
- [x] T013 Run `make generate` to create TypeScript types from new Go structs

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 2 - Generational Event Tracking (Priority: P1) 🎯 MVP

**Goal**: Track player activities (TR raises, tile placements) within each generation via event-driven architecture

**Independent Test**: Verify events are tracked when player actions occur and reset when generation advances

**Note**: User Story 2 (tracking system) must be implemented before User Story 1 (corporation card) because the corporation depends on the tracking system

### Tests for User Story 2

- [x] T014 [P] [US2] Create test file with table-driven tests for generational event tracking in `backend/test/action/generational_events_test.go`
- [x] T015 [P] [US2] Add test case: TR raise increments tr-raise count in `backend/test/action/generational_events_test.go`
- [x] T016 [P] [US2] Add test case: Ocean placement increments ocean-placement count in `backend/test/action/generational_events_test.go`
- [x] T017 [P] [US2] Add test case: Generation advance clears all player events in `backend/test/action/generational_events_test.go`

### Implementation for User Story 2

- [x] T018 [US2] Subscribe to TerraformRatingChangedEvent in Game initialization to track TR raises in `backend/internal/game/game.go`
- [x] T019 [US2] Subscribe to TilePlacedEvent in Game initialization to track tile placements (ocean, city, greenery) in `backend/internal/game/game.go`
- [x] T020 [US2] Subscribe to GenerationAdvancedEvent in Game initialization to clear all player generational events in `backend/internal/game/game.go`
- [x] T021 [US2] Run tests to verify event tracking works correctly

**Checkpoint**: Generational events are tracked via event bus and reset on generation advance

---

## Phase 4: User Story 3 - Card Behavior Conditional Requirements (Priority: P2)

**Goal**: Validate generational event requirements in state calculator so card actions/effects are properly enabled/disabled

**Independent Test**: Card actions with generational event requirements are unavailable when requirements not met

### Tests for User Story 3

- [x] T022 [P] [US3] Add test case: Action with tr-raise requirement is unavailable when TR not raised in `backend/test/action/generational_events_test.go`
- [x] T023 [P] [US3] Add test case: Action with tr-raise requirement is available when TR was raised in `backend/test/action/generational_events_test.go`
- [x] T024 [P] [US3] Add test case: Action with ocean-placement min 2 is unavailable with only 1 placement in `backend/test/action/generational_events_test.go`

### Implementation for User Story 3

- [x] T025 [US3] Add validateGenerationalEventRequirements function in `backend/internal/action/state_calculator.go`
- [x] T026 [US3] Add formatGenerationalEventError helper function in `backend/internal/action/state_calculator.go`
- [x] T027 [US3] Call validateGenerationalEventRequirements from CalculatePlayerCardActionState in `backend/internal/action/state_calculator.go`
- [x] T028 [US3] Run tests to verify state calculator validates requirements correctly

**Checkpoint**: State calculator properly validates generational event requirements

---

## Phase 5: User Story 1 - UNMI Corporation Card (Priority: P1)

**Goal**: Implement the United Nations Mars Initiative corporation card with conditional action

**Independent Test**: Select UNMI, raise TR via ocean placement, use corporation action to pay 3 MC for +1 TR

**Note**: This user story depends on US2 (tracking) and US3 (validation) being complete

### Implementation for User Story 1

- [x] T029 [US1] Update UNMI corporation card in `backend/assets/cards/corporations.json` with generationalEventRequirements field
- [x] T030 [US1] Verify UNMI card has: starting credits 40 MC, Earth tag, manual action with tr-raise min 1 requirement
- [x] T031 [US1] Run `make test` to verify UNMI corporation behavior works correctly

**Checkpoint**: UNMI corporation action is only available when TR has been raised this generation

---

## Phase 6: User Story 4 - Frontend Conditional Display (Priority: P2)

**Goal**: Display asterisk indicator on card behaviors with generational event requirements

**Independent Test**: View UNMI corporation card and verify asterisk appears on the action's output side

### Implementation for User Story 4

- [x] T032 [P] [US4] Add asterisk indicator for conditional behaviors in `frontend/src/components/ui/cards/BehaviorSection/components/ManualActionLayout.tsx`
- [x] T033 [P] [US4] Add asterisk indicator for conditional behaviors in `frontend/src/components/ui/cards/BehaviorSection/components/TriggeredEffectLayout.tsx`
- [x] T034 [US4] Verify asterisk appears on UNMI action output side in browser

**Checkpoint**: Cards with generational event requirements show asterisk indicators

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, cleanup, and validation

- [x] T035 [P] Update `backend/CLAUDE.md` with event-driven architecture section documenting EventBus usage for "do this when that happens" patterns
- [x] T036 [P] Update `backend/CLAUDE.md` with state calculator section documenting how to add requirements for cards/actions/standard projects
- [x] T037 Run `make format` to format all code
- [x] T038 Run `make lint` to verify no lint errors
- [x] T039 Run `make test` to verify all existing tests still pass
- [x] T040 Run `make prepare-for-commit` for final validation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 2 (Phase 3)**: Depends on Foundational - tracking system
- **User Story 3 (Phase 4)**: Depends on Foundational - can parallel with US2
- **User Story 1 (Phase 5)**: Depends on US2 (tracking) and US3 (validation)
- **User Story 4 (Phase 6)**: Depends on Foundational - can parallel with backend stories
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 2 (Tracking)**: Foundation only - can start first
- **User Story 3 (Validation)**: Foundation only - can parallel with US2
- **User Story 1 (UNMI Card)**: Requires US2 and US3 complete
- **User Story 4 (Frontend)**: Foundation only - can parallel with backend work

### Within Each User Story

- Tests written first to understand expected behavior
- Models/types before business logic
- Backend before frontend (for API types)
- Story complete before moving to dependent stories

### Parallel Opportunities

- T001-T004: All type definitions can run in parallel
- T009-T011: All DTO additions can run in parallel
- T014-T017: All US2 tests can run in parallel
- T022-T024: All US3 tests can run in parallel
- T032-T033: Frontend layout updates can run in parallel
- T035-T036: Documentation updates can run in parallel

---

## Parallel Example: Phase 1 Setup

```bash
# Launch all type definitions together:
Task: "Create GenerationalEvent enum in backend/internal/game/shared/generational_events.go"
Task: "Add MinMax struct in backend/internal/game/shared/generational_events.go"
Task: "Add GenerationalEventRequirement struct in backend/internal/game/shared/generational_events.go"
Task: "Add PlayerGenerationalEventEntry struct in backend/internal/game/shared/generational_events.go"
```

---

## Parallel Example: User Story 4 Frontend

```bash
# Launch all frontend layout updates together:
Task: "Add asterisk indicator in ManualActionLayout.tsx"
Task: "Add asterisk indicator in TriggeredEffectLayout.tsx"
```

---

## Implementation Strategy

### MVP First (User Stories 2 + 3 + 1)

1. Complete Phase 1: Setup (types)
2. Complete Phase 2: Foundational (player component, DTOs)
3. Complete Phase 3: User Story 2 (event tracking)
4. Complete Phase 4: User Story 3 (validation)
5. Complete Phase 5: User Story 1 (UNMI card)
6. **STOP and VALIDATE**: Test UNMI corporation action works correctly
7. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Types and infrastructure ready
2. Add US2 (tracking) → Events tracked via EventBus
3. Add US3 (validation) → State calculator validates requirements
4. Add US1 (UNMI card) → Corporation card fully functional (Backend MVP!)
5. Add US4 (frontend) → Visual indicators complete
6. Polish → Documentation and cleanup

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- US2 (tracking) and US3 (validation) must precede US1 (UNMI card)
- US4 (frontend) can proceed in parallel with backend work after T013
- Backend tests use table-driven test pattern per constitution
- Run `make generate` after Go type changes to sync TypeScript types
- Run `make prepare-for-commit` before any git operations
