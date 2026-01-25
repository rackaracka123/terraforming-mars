# Feature Specification: Generational Events System

**Feature Branch**: `001-generational-events`
**Created**: 2026-01-24
**Status**: Draft
**Input**: Implement the "United Nations Mars Initiative" corporation card with a generational events tracking system that enables conditional actions based on player activities within a generation.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Corporation Card Action with Generational Condition (Priority: P1)

A player who has chosen the United Nations Mars Initiative corporation wants to use their corporation's special action to gain additional Terraform Rating. The action requires that the player has already raised their TR this generation, and costs 3 MC to gain 1 TR.

**Why this priority**: This is the core feature request - implementing the UNMI corporation card with its conditional action mechanic.

**Independent Test**: Can be fully tested by selecting UNMI as corporation, raising TR through any means (playing a card, placing an ocean, etc.), then using the corporation action to pay 3 MC for +1 TR.

**Acceptance Scenarios**:

1. **Given** a player has UNMI corporation and has raised TR this generation, **When** they activate the corporation action and pay 3 MC, **Then** they gain 1 TR.
2. **Given** a player has UNMI corporation but has NOT raised TR this generation, **When** they attempt to activate the corporation action, **Then** the action is unavailable/disabled.
3. **Given** a player has UNMI corporation and has raised TR this generation but has less than 3 MC, **When** they view available actions, **Then** the corporation action is unavailable/disabled.

---

### User Story 2 - Generational Event Tracking (Priority: P1)

The game system needs to track specific player activities within each generation to enable conditional card behaviors. These tracked events include: TR raises, ocean placements, city placements, and greenery placements.

**Why this priority**: This is the foundational system that enables conditional actions based on generation-scoped player activities.

**Independent Test**: Can be tested by having a player perform various actions (place ocean, raise TR, etc.) and verifying the generational events are tracked correctly and reset at generation end.

**Acceptance Scenarios**:

1. **Given** a player places an ocean tile, **When** the placement completes, **Then** the player's generational events include "ocean-placement" with count 1.
2. **Given** a player places a second ocean tile in the same generation, **When** the placement completes, **Then** the "ocean-placement" count increments to 2.
3. **Given** a player has generational events recorded, **When** the generation ends (production phase completes), **Then** all generational events are cleared for all players.
4. **Given** a player has performed no tracked events, **When** viewing their generational events, **Then** the list is empty (no zero-count entries stored).

---

### User Story 3 - Card Behavior Conditional Requirements (Priority: P2)

Card behaviors (actions, effects, triggers) can specify generational event requirements that must be met before the behavior can execute.

**Why this priority**: Extends the generational events system to be usable by any card, not just UNMI.

**Independent Test**: Can be tested by creating test cards with generational event requirements and verifying the behavior is only available when requirements are met.

**Acceptance Scenarios**:

1. **Given** a card has a manual action with a generational event requirement of "tr-raise" min 1, **When** the player has not raised TR this generation, **Then** the action is not available.
2. **Given** a card has a manual action with a generational event requirement of "ocean-placement" min 2, **When** the player has placed only 1 ocean this generation, **Then** the action is not available.
3. **Given** a card has an effect trigger with a generational event requirement, **When** the requirement is not met, **Then** the effect does not trigger even if other conditions are met.

---

### User Story 4 - Frontend Conditional Display (Priority: P2)

Players viewing card behaviors with generational event conditions see a visual indicator (asterisk) showing that additional conditions apply to the behavior.

**Why this priority**: Users need clear visual feedback about conditional behaviors to understand why actions may or may not be available.

**Independent Test**: Can be tested by viewing the UNMI corporation card and verifying the asterisk appears on the action's output side.

**Acceptance Scenarios**:

1. **Given** a card with a manual action that has a generational event requirement, **When** viewing the card's behavior section, **Then** an asterisk appears on the right side of the action exchange (output side).
2. **Given** a card with an effect that has a generational event requirement, **When** viewing the card's behavior section, **Then** an asterisk appears on the output side of the effect.
3. **Given** a card with multiple behaviors where only some have requirements, **When** viewing the card, **Then** only the behaviors with requirements show asterisks.

---

### Edge Cases

- What happens when a player gains and loses TR in the same generation (net zero)? The count still reflects the number of raises, not net change.
- How does system handle TR raises from passive effects vs active card plays? All TR raises count regardless of source.
- What happens if generation ends mid-action? Generational events are cleared only after production phase completes.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST define a GenerationalEvent enum with values: "tr-raise", "ocean-placement", "city-placement", "greenery-placement"
- **FR-002**: System MUST track generational events per player with a count for each event type that occurred
- **FR-003**: System MUST NOT store zero-count events (only events that actually occurred are tracked)
- **FR-004**: System MUST clear all generational events for all players when a generation ends (after production phase)
- **FR-005**: System MUST publish events via the event bus when players perform trackable actions (TR raise, ocean/city/greenery placement)
- **FR-006**: System MUST validate generational event requirements when checking if card actions are available
- **FR-007**: System MUST validate generational event requirements when checking if card effects should trigger
- **FR-008**: System MUST validate generational event requirements when checking if cards can be played
- **FR-009**: Card behaviors MUST be able to specify generational event requirements with: event type (required), count range (optional, MinMax type), target (optional, TargetType)
- **FR-010**: The UNMI corporation card MUST have: starting credits of 40 MC, Earth tag, manual action requiring "tr-raise" min 1 that costs 3 MC and grants 1 TR
- **FR-011**: Frontend MUST display an asterisk indicator for behaviors that have generational event requirements
- **FR-012**: Backend CLAUDE.md MUST be updated to document the event-driven architecture pattern with examples

### Key Entities

- **GenerationalEvent**: Enum type representing trackable player actions within a generation (tr-raise, ocean-placement, city-placement, greenery-placement)
- **GenerationalEventRequirement**: A requirement specification for card behaviors containing: event type (required), count range (optional MinMax), target (optional TargetType)
- **PlayerGenerationalEvents**: Per-player tracking of events that occurred in the current generation, stored as a list of {event, count} pairs where count > 0

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: UNMI corporation action is only available when the player has raised TR at least once this generation and has at least 3 MC
- **SC-002**: All TR raises, ocean placements, city placements, and greenery placements are tracked in player generational events
- **SC-003**: Generational events reset to empty for all players at generation end
- **SC-004**: Cards with generational event requirements show asterisk indicators in the UI
- **SC-005**: Backend uses event-driven architecture for tracking generational events (no direct coupling to action implementations)
- **SC-006**: All existing tests continue to pass after implementation
