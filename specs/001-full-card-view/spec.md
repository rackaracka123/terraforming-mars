# Feature Specification: Expandable Card Descriptions

**Feature Branch**: `001-full-card-view`
**Created**: 2026-02-18
**Status**: Draft (Refactored)
**Input**: Enhance existing card and behavior components with hover-to-reveal description text. Rename SimpleGameCard to GameCard. No new view components or overlays.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Hoverable Behavior Descriptions (Priority: P1)

A player is viewing cards in a selection modal (starting cards, draw-peek-buy, production) or on the /cards browser page. Each behavior row in the card's behavior section is hoverable. When the player hovers over a behavior row, a description text appears directly underneath that row, explaining what the behavior does. The description spans the full width of the behavior row regardless of how many items are side by side in that row. This helps players understand complex card effects without needing external references.

**Why this priority**: This is the core value of the feature — making card behavior descriptions accessible inline. It directly helps players make informed decisions during card selection and browsing.

**Independent Test**: Navigate to /cards, hover over any behavior row on a card — a description should appear underneath that row.

**Acceptance Scenarios**:

1. **Given** a card with behaviors is displayed (in any context), **When** the player hovers over a behavior row in the BehaviorSection, **Then** a description text appears below that entire row.
2. **Given** a behavior row contains multiple side-by-side elements (e.g., `[input] [arrow] [output]`), **When** the player hovers anywhere on that row, **Then** the description spans underneath all elements as a single block.
3. **Given** a behavior row is hovered and its description is showing, **When** the player moves the mouse away from that row, **Then** the description text disappears.
4. **Given** a behavior has no description in the card data, **When** the player hovers over that row, **Then** no description area appears (graceful fallback).
5. **Given** the card is displayed in the hand fan during gameplay, **Then** hover behavior descriptions still work but are not expected to be visually polished for the compact fan layout.

---

### User Story 2 - VP and Resource Storage Hover Descriptions (Priority: P2)

A player sees a card with a Victory Point icon or resource storage indicator. When the player hovers over the VP icon (shown on the left side of the card title), a description appears at the bottom of the card explaining the VP scoring rule. Similarly, hovering over a resource storage indicator shows a description at the bottom of the card explaining what resources the card stores.

**Why this priority**: VP conditions and resource storage are important card attributes that benefit from explanation, but they are secondary to behavior descriptions since behaviors are more complex and numerous.

**Independent Test**: On the /cards page, find a card with a VP icon, hover over it — description should appear at the bottom of the card.

**Acceptance Scenarios**:

1. **Given** a card has VP conditions, **When** the player hovers over the VP icon on the card, **Then** a description appears at the bottom of the card explaining the VP scoring rule.
2. **Given** a card has resource storage, **When** the player hovers over the resource storage indicator, **Then** a description appears at the bottom of the card explaining the storage.
3. **Given** the player stops hovering the VP icon or storage indicator, **Then** the description at the bottom of the card disappears.
4. **Given** a VP condition or resource storage has no description in the data, **Then** no description area appears on hover.

---

### User Story 3 - Rename SimpleGameCard to GameCard (Priority: P3)

The existing SimpleGameCard component is renamed to GameCard throughout the codebase. This is a single unified card component that supports expandable properties (hover descriptions) rather than maintaining separate "simple" and "full" variants. All references across the codebase are updated.

**Why this priority**: This is a refactoring task that improves code clarity. It depends on the other stories being complete to avoid merge conflicts during the rename.

**Independent Test**: Search the codebase for "SimpleGameCard" — zero results should remain. All card rendering should continue to work as before.

**Acceptance Scenarios**:

1. **Given** the rename is complete, **Then** no references to "SimpleGameCard" exist in the codebase.
2. **Given** the rename is complete, **Then** all card rendering in game view, selection modals, and /cards page works identically to before.

---

### Edge Cases

- What happens when a card has no behaviors? The card renders as it does today — no behavior rows, no hover descriptions.
- What happens when a behavior description is missing from the card data? The hover row still works but no description area expands — the behavior row looks and behaves exactly as it does today.
- What happens on mobile/touch devices where hover is not available? Behavior rows could respond to tap instead of hover, but this is acceptable as a best-effort enhancement. The card remains fully functional without hover descriptions.
- What happens when multiple behavior rows are close together? Only the hovered row shows its description; hovering a different row replaces the previous description.
- What happens in the card hand fan (compact layout)? Hover descriptions work but may not be visually ideal — this is acceptable per requirements.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Each behavior row in BehaviorSection MUST become hoverable, revealing a description text block below the row on hover.
- **FR-002**: The description text block MUST span the full width of the behavior row, regardless of how many elements are displayed side-by-side within that row. The description MUST overlay below the row without changing the card's layout dimensions (floating over content below like an attached tooltip).
- **FR-003**: Only one behavior description MUST be visible at a time within a card — hovering a different row replaces the previous description.
- **FR-004**: When the hover leaves a behavior row (and does not enter another), the description MUST disappear.
- **FR-005**: Hovering the VP icon on a card MUST show the VP condition description at the bottom of the card.
- **FR-006**: Hovering the resource storage indicator on a card MUST show the storage description at the bottom of the card.
- **FR-007**: VP and resource storage descriptions MUST appear at the bottom of the card, separate from behavior descriptions which appear inline below each row.
- **FR-008**: If a behavior, VP condition, or resource storage has no description in the data, hovering MUST NOT show an empty description area.
- **FR-009**: SimpleGameCard MUST be renamed to GameCard throughout the codebase with no functional changes.
- **FR-010**: No new card view components, overlays, or side panels MUST be created. All description functionality MUST be added to the existing card and behavior components using expandable properties.

### Key Entities

- **GameCard** (renamed from SimpleGameCard): The single card display component that renders card data with hoverable behavior descriptions.
- **BehaviorSection**: The existing behavior rendering system, enhanced with per-row hover interaction and description reveal.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Players can see a behavior description within 0.5 seconds of hovering over a behavior row.
- **SC-002**: 100% of behavior rows with descriptions in the card data show their description on hover.
- **SC-003**: VP and resource storage descriptions are accessible via hover on all cards that have them.
- **SC-004**: Zero new component files are created for card display — all changes are within existing components.
- **SC-005**: The "SimpleGameCard" name is fully removed from the codebase after rename.

## Clarifications

### Session 2026-02-18

- Q: When a behavior description appears below a hovered row, how should the card handle vertical space? → A: The description overlays below the row without changing card dimensions (floats over content below, like an attached tooltip), avoiding disruption to grid layouts.

## Assumptions

- The card database has `description` fields populated on `behaviors[]`. Missing descriptions are handled gracefully (no hover effect).
- VP condition and resource storage descriptions are available from the backend (the DTO changes for passing description fields through are handled separately).
- The hover interaction is mouse-based. Touch/tap fallback is best-effort, not a hard requirement.
- The BehaviorSection's existing classification and merging pipeline is preserved — hover descriptions are layered on top of the current rendering, not a replacement.
- The /cards page and card selection modals are the primary contexts for this feature. The hand fan in-game is acceptable as-is.

## Verification Plan

- **Behavior hover**: Navigate to /cards, hover over behavior rows on cards with varying complexity (automated, active, event types) — descriptions should appear/disappear correctly.
- **VP hover**: Find a card with VP conditions on /cards, hover over the VP icon — description should appear at the bottom of the card.
- **Resource storage hover**: Find a card with resource storage on /cards, hover over the storage indicator — description should appear at the bottom of the card.
- **Selection modals**: Start a game, reach starting card selection, hover behavior rows — descriptions work within the modal context.
- **Rename verification**: Search codebase for "SimpleGameCard" — zero results.
