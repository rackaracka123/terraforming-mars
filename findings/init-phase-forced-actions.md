# Init phase advancement after forced actions

Notes on the corp/prelude application phase (`GamePhaseInitApplyCorp`,
`GamePhaseInitApplyPrelude`) and how forced first actions interact with it.

- 2026-06-30: Issue #568 — a forced colony placement (Poseidon's forced first
  action) during the init-apply phase left the turn stuck. The frontend never
  sent a `ConfirmInitAdvance` after the placement, and the placement itself did
  not advance the phase, so the game hung waiting for confirm. Fix: backend-side
  auto-advance. `ConfirmColonyPlacementAction.Execute` now, when
  `CurrentPhase()` is an init-apply phase, calls the new exported
  `turn_management.AdvanceInitPhaseAfterForcedAction(ctx, g, log)`. The
  `CurrentPhase` guard keeps normal in-game colony placement unaffected.
- 2026-06-30: The already-applied advance path is the single source of truth in
  `AdvanceInitPhaseAfterForcedAction`. It is a probe-style helper: returns
  `(false, nil)` WITHOUT mutating state when advancement is not applicable (wrong
  phase, not waiting for confirm, index out of range, or the current init player
  still has a pending selection / pending tile queue / incomplete forced first
  action); returns `(true, err)` after clearing the waiting flag and advancing.
  Contrast with `ConfirmInitAdvanceAction.Execute`, which performs the SAME gate
  checks but returns errors — because it is a user-initiated action that should
  fail loudly. Different semantics (probe vs assertion), so this is not true
  duplication; do not naively merge the two gate blocks.
- 2026-06-30: Coupling direction is `confirmation` → `turn_management` only
  (one-directional, acyclic). `turn_management` must NOT import
  `action/confirmation`. `advanceToNextPlayer` was converted from a method to a
  package function so both the confirm flow and the auto-advance helper share it
  (no copy-paste of the next-player/phase-transition logic).

# Aggregate action availability (#571)

- 2026-06-30: Issue #571 — no aggregate "can this player act at all?" helper
  exists. Per-surface validators live in `internal/action/state_calculator.go`
  (`CalculatePlayerCardState`, `CalculatePlayerCardActionState`,
  `CalculatePlayerStandardProjectState`, `CalculateMilestoneState`,
  `CalculateAwardState`). The DTO mapper (`mapper_player.go`) already iterates
  every surface to build the per-entity `.Available` arrays — a new
  `HasAvailableActions` MUST reuse these calculators, not re-derive affordability.
- 2026-06-30: Colony trade is the ONLY action surface with NO state-calculator
  function. Affordability is validated inline in `internal/action/colony/trade.go`
  (TradeCreditsCost/EnergyCost/TitaniumCost minus `CalculateActionDiscounts`).
  `mapPlayerActionCosts` duplicates the same numbers. Any trade-availability check
  in the aggregator should reuse `RequirementModifierCalculator.CalculateActionDiscounts`
  + `Colonies().GetTradeableIDs()` + `GetTradeFleetAvailable`, not invent new logic.
- 2026-06-30: PENDING-SELECTION NUANCE. `validateNoPendingSelection` (via
  `g.HasAnyPendingSelection`) makes EVERY surface unavailable while a pending
  tile/card/choice selection or forced first action is active. That state is NOT
  "stuck" — the player has a forced thing to do. So a true zero-available-actions
  / "stuck" signal must EXCLUDE the pending-selection case (treat pending/forced
  as "has action"), else the indicator fires wrongly mid-selection.
- 2026-06-30: sell-patents (free standard project) is available iff
  `Hand().CardCount() > 0`. An empty hand is REQUIRED for the stuck state. The
  per-project calculator already encodes this, so reusing it is correct.
- 2026-06-30: #571 import-cycle confirmed safe. `internal/colonies` (registry) and
  `internal/game/colonies` (state) do NOT import `internal/action`, so
  `package action` can import both. `package action` is the cycle-free home for
  `CalculateColonyTradeState` + the canonical `Trade*Cost` constants; `action/colony`
  already imports `action` as `baseaction`, so it references them as
  `baseaction.TradeCreditsCost` etc. `mapPlayerActionCosts` (dto) already imports
  `internal/action`, so it can call the new helper too.
- 2026-06-30: #568 corp test lives at
  `backend/test/action/card_packs/corporations_colonies_test.go` (NOT
  `test/action/corp/...` as a stale reference implied). The three #568 files that
  must NOT be modified are: confirm_colony_placement.go,
  turn_management/confirm_init_advance.go, and that card_packs test.
- 2026-06-30: #571 repro test currently lives ONLY on branch `repro/state-bugs`
  at `backend/test/action/core/no_available_actions_repro571_test.go`; it is NOT
  on `fix/in-game-issues-batch`. "Extend the committed repro test" => cherry-pick /
  re-create that file onto the batch branch first, then add a HasAvailableActions
  assertion. The frontend already has ad-hoc client-side affordability logic in
  PlayerList.handleSkipAction (cards.some(available) / actions.some(available)) that
  is NOT the source of truth; the new backend *bool is authoritative for the indicator.
- 2026-06-30: #571 c2 — `action.HasAvailableActions(g, p, cardRegistry,
  stdProjRegistry, milestoneRegistry, awardRegistry)` lives in
  `internal/action/has_available_actions.go`. It short-circuits on the first
  available surface and is built ONLY from the existing per-surface calculators so
  it can never diverge from the DTO `.Available` arrays. Two gotchas: (1) the two
  resource-conversion projects (heat->temp, plants->greenery) are NOT in the
  standard-project registry (they are resource buttons), so the aggregator iterates
  the pack-filtered registry AND explicitly checks `conversionStandardProjects`;
  (2) `g.HasAnyPendingSelection` does NOT cover `ForcedFirstAction` (stored at the
  Game level, gated on `!Completed`) — both must be checked separately to treat
  pending/forced as "not stuck".
- 2026-06-30: #571 c2 SELECTION-FILTER GOTCHA (recurred in review). The
  milestone/award action surfaces are NOT just "every definition the calculator
  says is claimable". Every game exposes only the SELECTED subset
  (`g.SelectedMilestones()` / `g.SelectedAwards()`, 5 of 16 milestones), and
  `CalculateMilestoneState` / `CalculateAwardState` do NOT check selection
  themselves. The DTO mapper filters via `filterMilestones` / `filterAwards`
  BEFORE calling the calculators, so the aggregator MUST filter identically or it
  reports false "actions available" (e.g. `generalist` has no requirement, so any
  player with 8 credits "can claim" it; awards need only credits, no
  qualification). Fix: extracted the filter into the cycle-free `package action`
  as exported `FilterMilestones` / `FilterAwards`; both the aggregator and the DTO
  mapper now call the SAME function (single source of truth, cannot diverge). The
  old per-package `filterMilestones`/`filterAwards` in `dto` were deleted.
  DETERMINISTIC GUARD: `TestHasAvailableActions_IgnoresUnselectedMilestone`
  asserts a claimable-but-unselected `terraformer` keeps `HasAvailableActions`
  false, then flips true once selected.
- 2026-06-30: #571 c2 TEST-SETUP GOTCHA. Standard projects carry
  `pack == "base-game"`, and the pack filter (`def.Pack != "" &&
  !enabledPacks[def.Pack]`) drops them unless `CardPacks` includes `base-game`.
  `EnabledPacks()` does NOT implicitly add base-game. A test game built with empty
  `CardPacks` therefore has NO standard-project surface — an earlier repro test
  only "passed" because the unfiltered milestone surface false-positived on
  `generalist`. Real games always set `CardPacks: ["base-game"]`; the repro
  fixture now does too, so PowerPlant is a genuine surface.
- 2026-06-30: CI-TRIGGER GOTCHA. Every workflow under `.github/workflows/`
  (backend-test, frontend, format) fires ONLY on `pull_request` to `main` or
  `push` to `main` — never on a push to a feature branch like
  `fix/in-game-issues-batch`. Pushing the combined batch branch is a no-op for CI:
  `gh run list --branch fix/in-game-issues-batch` stays empty. The only ways to
  exercise CI on the batch work are to open the PR-to-main (deferred to the
  combined-PR step) or `workflow_dispatch`. So a CI-aware push on a non-main branch
  with "do not open a PR" cannot turn CI green by itself — that is expected, not a
  failure. Local `make` (build + go test ./test/... + format-check + lint +
  typecheck) is the authoritative pre-PR gate in this situation.
- 2026-06-30: GENERATE-NON-IDEMPOTENCE GOTCHA. `make generate` runs ONLY
  `tygo generate`, whose `frontend/src/types/generated/api-types.ts` output is
  UNFORMATTED (e.g. `int */}` not `int */ }`, and long literals on one line). The
  committed file is the POST-`make format` (prettier/oxlint) version, so a bare
  `make generate` always shows a large whitespace-only diff vs the committed file
  even when the Go DTOs are unchanged. Do NOT commit that raw drift: after
  `make generate`, run `make format` (or just verify the only diff is whitespace
  and `git checkout` the file if the DTO content is identical). Concretely on this
  run, `hasAvailableActions?: boolean;` was already present and content-correct;
  the 163-line diff was pure formatting and was reverted. ENFORCED-BY: rule line
  added to backend/CLAUDE.md "Go to TypeScript" section.

# Player-name length cap (#558)

- 2026-06-30: The `binding:"...,max=50"` struct tags on `JoinGameRequest`
  (http_dto.go) and `CreateGameRequest` are DECORATIVE ONLY. There is NO
  go-playground/validator dependency in backend/go.mod and no `.Struct()` /
  `ShouldBindWith` validation call anywhere; `JoinGameRequest` is not even
  referenced by any HTTP handler. Changing the tag value enforces nothing at
  runtime. The REAL player-name entry chokepoint is
  `internal/action/game/join_game.go` `JoinGameAction.Execute(ctx, gameID,
  playerName, playerID)` — both the WS join handler and the create-game first
  player route through it (CreateGameAction takes no name). So the backend cap
  MUST be an explicit length guard in JoinGameAction.Execute (single source of
  truth via one Go const), and the tag value is updated only to keep the
  documented contract honest.
- 2026-06-30: Refinement — the `binding:` tag on `JoinGameRequest` was REMOVED
  entirely (not just retargeted), because the struct is never bound/validated
  and the tag advertised an unenforced 45-char limit (an unenforced decorative
  tag). tygo only reads `json:`/`tstype:` tags, so dropping `binding:` leaves
  the generated `frontend/src/types/generated/api-types.ts` byte-identical — no
  `make generate` needed, no DTO shape change. This also removes the last
  scattered `45` literal on the backend, leaving `game.MaxPlayerNameLength` as
  the single source of truth referenced by the real guard in
  JoinGameAction.Execute.
