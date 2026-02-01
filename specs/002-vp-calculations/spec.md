# Feature Specification: VP Calculations

**Feature Branch**: `002-vp-calculations`
**Created**: 2026-01-30
**Status**: Draft
**Input**: User description: "VP calculations with event-driven recomputation and interactive UI breakdown"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Live VP Tracking During Gameplay (Priority: P1)

A player plays cards throughout the game and wants to see their current VP total update in real time. When a card with VP conditions is played (e.g., "Birds" awarding 1 VP per animal), the system registers that card as a VP source on the player. As the game progresses and resources accumulate or tags are played, the VP total recalculates automatically via event handlers. The player sees their updated VP total on the VP button at all times without manual refresh.

**Why this priority**: This is the core mechanic — without live VP tracking, players have no visibility into their score during the game. It delivers the fundamental value of the feature.

**Independent Test**: Can be tested by playing a card with VP conditions and verifying the VP total updates on the player's VP button. Playing additional cards or gaining resources triggers recalculation.

**Acceptance Scenarios**:

1. **Given** a player has no played cards, **When** they play a card with a fixed VP condition (e.g., "Dust Seals" = 1 VP), **Then** the VP button total increases by 1 and the card appears as a VP source on the player.
2. **Given** a player has played "Birds" (1 VP per animal on self-card) with 0 animals, **When** an effect adds 2 animals to "Birds", **Then** the VP total increases by 2 and the computed value for "Birds" updates to 2.
3. **Given** a player selects a corporation with VP conditions (e.g., "Arklight" = 1 VP per 2 animals), **When** the corporation is selected, **Then** it is registered as a VP source and its computed value reflects the current animal count.
4. **Given** a player has multiple VP sources, **When** any game event triggers recalculation (card played, action played, effect triggered), **Then** all VP sources are recalculated and the total reflects the sum of all computed values plus terraform rating.

---

### User Story 2 - Interactive VP Breakdown Modal (Priority: P2)

A player clicks the VP button to open a detailed breakdown modal. The modal displays a horizontal stacked bar where each segment represents a different VP source (individual card or base categories like TR). Each segment is color-coded with different shades that match the modal's design. Hovering over a segment highlights it subtly and shows a tooltip with the card name, its description, and the VP it contributes. Card names appear in small text right-aligned above the bar for each segment.

**Why this priority**: The breakdown modal gives players strategic insight into their scoring composition. It builds on the live VP tracking (P1) and is the primary way players interact with VP data.

**Independent Test**: Can be tested by opening the VP modal and verifying the bar renders with correct segments, hover interactions work, and card information displays correctly.

**Acceptance Scenarios**:

1. **Given** a player has 3 VP sources (TR, "Dust Seals" fixed 1 VP, "Birds" with 3 animals = 3 VP), **When** they open the VP modal, **Then** the horizontal bar shows 3 distinct colored segments proportional to their VP values.
2. **Given** the VP modal is open, **When** the player hovers over the "Birds" segment, **Then** the segment subtly brightens, a tooltip appears showing "Birds", its card description, and "3 VP (1 per animal)", and the card name label above the bar is highlighted.
3. **Given** a player has 10 VP sources, **When** they open the VP modal, **Then** all 10 segments are visible in the bar with their card name labels positioned right-aligned above each segment, and the bar remains readable.
4. **Given** the VP modal is open, **When** the player hovers off a segment, **Then** the highlight and tooltip disappear smoothly.

---

### User Story 3 - VP Sources from Corporation Cards (Priority: P2)

A player selects a corporation that grants VP (e.g., "Arklight" = 1 VP per 2 animals, "Celestic" = 1 VP per 3 floaters). The corporation's VP condition is registered as a VP source alongside regular played cards. The VP breakdown modal treats corporation VP sources identically to card VP sources, showing them in the bar with appropriate labels and tooltips.

**Why this priority**: Corporations with VP conditions are a subset of VP sources that must be handled to ensure completeness. This shares priority with the modal since it extends the same system.

**Independent Test**: Can be tested by selecting a corporation with VP conditions and verifying it appears in the VP breakdown with correct computed values.

**Acceptance Scenarios**:

1. **Given** a player selects "Arklight" corporation, **When** the corporation is processed, **Then** "Arklight" appears as a VP source with its VP condition (1 VP per 2 animals) and a computed value based on current animals on the corporation card.
2. **Given** a player has "Celestic" as corporation with 9 floaters, **When** the VP modal is opened, **Then** "Celestic" shows as a segment contributing 3 VP (1 per 3 floaters) with appropriate tooltip.

---

### User Story 4 - Event-Driven VP Recalculation (Priority: P1)

The system recalculates VP totals whenever a relevant game event occurs. Specific event types trigger recalculation: card played, action played, and effect triggered. This ensures VP values stay accurate as the game state evolves. The recalculation iterates over all registered VP sources for a player, evaluating each condition (fixed or per-based) against the current game state.

**Why this priority**: The event-driven recalculation is the engine behind live VP tracking. Without it, VP values would be stale. It is architecturally critical.

**Independent Test**: Can be tested by triggering each event type and verifying VP values update correctly.

**Acceptance Scenarios**:

1. **Given** a player has "Ganymede Colony" (1 VP per jovian tag) and 2 jovian tags, **When** they play a card with a jovian tag, **Then** the "card played" event triggers recalculation and "Ganymede Colony" VP updates from 2 to 3.
2. **Given** a player has "Predators" (1 VP per animal) with 3 animals, **When** an action adds an animal to "Predators", **Then** the "action played" event triggers recalculation and "Predators" VP updates from 3 to 4.
3. **Given** a player has "Search For Life" (3 VP per 3 science, maxTrigger: 1), **When** the third science resource is added via an effect, **Then** the "effect triggered" event triggers recalculation and "Search For Life" VP updates from 0 to 3.

---

### Edge Cases

- What happens when a card has VP conditions but the "per" resource count is 0? The computed value is 0 VP, and the card still appears in the VP source list.
- What happens when a card has `maxTrigger: 1` and the condition is not met? The computed value remains 0 until the condition threshold is reached, then caps at the defined amount.
- What happens when a player has no VP-granting cards? The VP source list is empty, and only base categories (TR, milestones, awards, tiles) contribute to total VP.
- What happens with cards that have "per" conditions counting resources on "self-card" but the card has been played with 0 resources? The segment appears in the modal with 0 VP and updates dynamically as resources are added.
- What happens with "per tag" conditions when the player has 0 of that tag? The computed value is 0 VP; the segment still appears in the breakdown to show the potential.
- How are VP segments ordered in the bar? They follow the order cards were played (corporation first), matching the order of the VP sources list on the player.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST register a VP source on the player whenever a card with VP conditions is played, preserving insertion order in a list.
- **FR-002**: System MUST register a VP source when a corporation with VP conditions is selected, placing it at the beginning of the VP source list.
- **FR-003**: Each VP source MUST contain: the card ID (source reference), the VP condition definition (amount, condition type, per details if applicable), and a computed value representing current VP contribution.
- **FR-004**: System MUST recalculate all VP source computed values when a "card played", "action played", or "effect triggered" event fires.
- **FR-005**: For fixed VP conditions, the computed value MUST equal the condition's amount.
- **FR-006**: For "per" conditions with resource target "self-card", the system MUST count the specified resource type on the source card and compute VP as `floor(resourceCount / per.amount) * vpAmount`, respecting `maxTrigger` caps.
- **FR-007**: For "per" conditions with type "tag", the system MUST count the player's played tags of the specified type and compute VP as `floor(tagCount / per.amount) * vpAmount`.
- **FR-008**: For "per" conditions with tile types (city-tile, ocean-tile, colony-tile), the system MUST count the relevant tiles and compute VP accordingly.
- **FR-009**: When `maxTrigger` is a positive number, the computed VP MUST NOT exceed `maxTrigger * vpAmount`.
- **FR-010**: The VP button in the UI MUST display the total VP (sum of all VP source computed values plus terraform rating, milestones, awards, and tile VP).
- **FR-011**: The VP modal MUST display a horizontal stacked bar where each segment represents a VP source, sized proportionally to its computed value.
- **FR-012**: Each bar segment MUST use a distinct color shade that harmonizes with the modal's design palette.
- **FR-013**: Hovering a bar segment MUST subtly brighten the segment and display a tooltip showing the source card name, card description, and VP contribution details.
- **FR-014**: Each bar segment MUST have a small card name label positioned right-aligned and above the bar.
- **FR-015**: The VP modal MUST be a complete rewrite of the existing VP modal component, replacing all previous logic.
- **FR-016**: VP conditions from card `behaviors` inputs/outputs MUST NOT be evaluated — only conditions in the card's `vpConditions` array are VP sources.

### Key Entities

- **VPGranter**: A VP source registered on a player. Contains: source card ID, VP condition definition (amount, condition type, per-details), and a computed value representing the current VP contribution of that source.
- **VPCondition**: The condition definition from a card's `vpConditions` array. Defines whether VP is fixed or calculated per some resource/tag/tile count, with optional max trigger cap.
- **VP Breakdown**: The full scoring summary combining VP granter totals with terraform rating, milestone VP, award VP, greenery VP, and city VP.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: When a card with VP conditions is played, the player's VP total updates within the same game turn with no stale values visible to the player.
- **SC-002**: The VP modal correctly displays all VP sources as distinct bar segments with proportional sizing for any number of VP-granting cards (tested with 1, 5, and 10+ sources).
- **SC-003**: Hovering any bar segment shows accurate tooltip information (card name, description, VP details) matching the underlying VP condition data.
- **SC-004**: All VP condition types found in the card database are correctly computed: fixed (105 cards), per-animal (13), per-microbe (5), per-floater (4), per-tag (3), per-science (2), per-city-tile (2), per-asteroid (1), per-ocean-tile (1), per-colony-tile (1).
- **SC-005**: VP recalculation triggers on card played, action played, and effect triggered events without requiring page refresh or manual action.
- **SC-006**: Corporation VP conditions (e.g., Arklight, Celestic) are tracked identically to regular card VP conditions with correct computed values.

## Assumptions

- The existing event bus system supports subscribing to "card played", "action played", and "effect triggered" events. If any of these event types do not yet exist, they will need to be added.
- The existing `VictoryPointCondition` struct and card `vpConditions` field provide all necessary data for VP computation without schema changes.
- Tile-based VP conditions (adjacent city/ocean tiles) can access board state during recalculation.
- The VP modal redesign replaces the existing endgame VP components entirely — no backward compatibility with the current modal components is needed.
- The VP button and modal are accessible during gameplay (not just at endgame), providing live feedback.
