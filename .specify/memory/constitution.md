<!--
SYNC IMPACT REPORT
==================
Version Change: 0.0.0 → 1.0.0 (MAJOR - initial constitution creation)

Added Principles:
- I. Clean Architecture & Action Pattern
- II. Code Clarity Over Comments
- III. Best Practices Enforcement
- IV. Complete Feature Implementation
- V. Type Safety & Generation
- VI. Testing Discipline
- VII. No Deprecated Code

Added Sections:
- Development Standards
- Quality Gates

Templates Review:
- ✅ plan-template.md - Constitution Check section aligns with principles
- ✅ spec-template.md - Requirements structure aligns with testing discipline
- ✅ tasks-template.md - Task organization supports complete implementation

Follow-up TODOs: None
==================
-->

# Terraforming Mars Constitution

## Core Principles

### I. Clean Architecture & Action Pattern

The codebase follows Clean Architecture with strict separation of concerns. Business logic
resides exclusively in action classes within `/internal/action/`. Each action performs exactly
ONE operation with clear inputs and outputs (~100-200 lines). State mutation MUST only occur
through actions calling Game state methods. Domain types live in `/internal/game/` with private
fields and public accessors. Event-driven architecture ensures actions update state, events
propagate automatically, and the Broadcaster handles WebSocket updates. No manual polling or
effect checking in services.

**Non-negotiable rules:**
- Actions extend BaseAction with injected dependencies (GameRepository, CardRegistry, logger)
- HTTP and WebSocket handlers delegate to actions; handlers contain no business logic
- Game methods publish domain events automatically; services do only what the action says
- State mutation is forbidden outside `/internal/action/`

### II. Code Clarity Over Comments

Code MUST be self-documenting through clear naming, logical structure, and appropriate
abstractions. Comments are a code smell indicating the code itself is unclear. Instead of
explaining what code does via comments, refactor the code to be obvious. Variable and function
names MUST communicate intent without requiring explanation. Comments are only permitted when
documenting non-obvious business rules that cannot be expressed in code, or public API
contracts for exported symbols.

**Non-negotiable rules:**
- Remove all explanatory comments; refactor code to be self-evident instead
- Delete commented-out code; version control preserves history
- Function and variable names MUST describe purpose without abbreviations
- Extract complex conditionals into well-named helper functions
- Prefer explicit over clever; readable code over compact code

### III. Best Practices Enforcement

All code MUST follow established language idioms and community standards. For Go: follow
Effective Go, Go Code Review Comments, and Google's Go Style Guide. For TypeScript/React:
use functional components, proper hooks patterns, and generated types from backend. Run
`make format` and `make lint` after every change. Fix all lint ERRORS immediately with no
exceptions. Use proper error handling patterns: check errors immediately, return early, keep
the happy path left-aligned.

**Non-negotiable rules:**
- Format and lint passes are mandatory before any commit
- Each Go file MUST have exactly ONE package declaration
- Use `gofmt` and `goimports` for Go; Prettier and oxlint for TypeScript
- Never use timeouts, sleeps, or delays as solutions to state management issues
- Use event listeners, promises, proper synchronization (channels, mutexes) instead

### IV. Complete Feature Implementation

Features MUST be implemented comprehensively even when that requires significant code volume.
Under-engineering and shortcuts are forbidden. When implementing a feature, include all
necessary validation, error handling, edge cases, and user feedback. Do not defer obvious
requirements to "future work." However, avoid over-engineering: do not add features, refactor
code, or make improvements beyond what was explicitly requested. Do precisely what was asked;
nothing more, nothing less.

**Non-negotiable rules:**
- Implement features fully; partial implementations are rejected
- Include proper validation at system boundaries (user input, external APIs)
- Handle all error cases explicitly; never silently fail
- Avoid YAGNI violations: no speculative features or unnecessary abstractions
- A bug fix does not require surrounding code cleanup
- Do not add docstrings, comments, or type annotations to code you did not change

### V. Type Safety & Generation

TypeScript types MUST be generated from Go backend structs using `make generate`. Never
manually create duplicate types. All Go structs requiring frontend access MUST have `json:`
and `ts:` tags. Frontend code MUST import types from `src/types/generated/api-types.ts`.
When updating domain types, verify corresponding DTOs and mappers remain synchronized.

**Non-negotiable rules:**
- Run `make generate` after any Go type changes
- Never manually duplicate types between Go and TypeScript
- DTOs in `/internal/delivery/dto/` MUST stay synchronized with domain types
- Frontend MUST use generated types; no local type redefinitions
- Fail explicitly when expected data is missing; no default fallback values

### VI. Testing Discipline

All new backend features MUST include tests. Test files reside in `test/` directory, mirroring
`internal/` structure. Use table-driven tests for scenarios with multiple inputs. Mock external
dependencies via interfaces. Test business logic in isolation. Tests MUST be deterministic:
no timeouts, sleeps, or timing-dependent assertions.

**Non-negotiable rules:**
- Test files go in `test/` directory (e.g., `test/action/my_action_test.go`)
- Run `make test` before any commit touching backend code
- Table-driven tests for multiple scenarios
- No arbitrary timing or polling in tests; use proper synchronization
- Integration tests for contract changes and inter-service communication

### VII. No Deprecated Code

Remove deprecated code completely. Do not keep deprecated fields, functions, methods, or
comments for backwards compatibility. When something is deprecated, delete it entirely and
update all usages. Dead code, unused exports, and legacy patterns have no place in the
codebase.

**Non-negotiable rules:**
- Delete deprecated code; do not comment it out or mark as unused
- Remove all `// deprecated` comments along with the code they reference
- Update all call sites when removing deprecated APIs
- No backwards-compatibility shims, re-exports, or `_unused` variable renames
- Version control preserves history; the codebase shows only current implementation

## Development Standards

### Git Workflow

All changes MUST go through pull requests. Never push directly to main. Create feature
branches with descriptive names. Commits MUST pass format and lint checks. Pre-commit hooks
MUST NOT be skipped (no `--no-verify`).

### Code Quality Gates

Before any commit:
1. `make format` - Format all code
2. `make lint` - Pass all linters without errors
3. `make test` - All tests pass (when backend changes)
4. `make generate` - Types synchronized (when Go types change)

### Component Standards (Frontend)

- Use GameIcon component for all icon display; never use direct `<img>` tags
- Use Tailwind CSS utilities only; CSS Modules are deprecated and forbidden
- Use `void <function>()` to discard promises in event handlers
- Use generated types from backend; no manual type definitions
- Inspect existing components for design patterns before creating new ones

### State Management

- Backend is single source of truth for game state
- No client-side game logic that could cause desync
- All state changes propagate through WebSocket events
- No localStorage for game state; only for session persistence (gameId, playerId, playerName)

## Quality Gates

### Pre-Commit

- All code formatted (`make format`)
- All lint errors resolved (`make lint`)
- All tests passing (`make test`)
- Types synchronized if Go changed (`make generate`)

### Pre-PR

- Feature branch created (never commit directly to main)
- Clear commit messages describing changes
- No deprecated code introduced
- No commented-out code
- No unnecessary comments added

### Architecture Compliance

- Business logic only in `/internal/action/`
- State mutation only through actions
- Events published by Game methods, not manually in handlers
- Handlers delegate to actions without business logic
- DTOs synchronized with domain types

## Governance

This constitution supersedes all other development practices. Violations discovered during
review MUST be fixed before merge. When constitution and expedience conflict, constitution
wins. Amendments require documentation of the change, rationale, and migration plan for
affected code.

All PRs and reviews MUST verify compliance with these principles. Complexity beyond what
is specified here requires explicit justification in the PR description, including why
simpler alternatives are insufficient.

For runtime development guidance, refer to `CLAUDE.md` files throughout the repository.

**Version**: 1.0.0 | **Ratified**: 2026-01-22 | **Last Amended**: 2026-01-22
