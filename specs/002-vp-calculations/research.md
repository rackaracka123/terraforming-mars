# Research: VP Calculations

**Feature Branch**: `002-vp-calculations`
**Date**: 2026-01-30

## Research Task 1: Event Types for VP Recalculation Triggers

**Question**: Which existing domain events should trigger VP recalculation, and do we need new events?

**Decision**: Subscribe to these existing events for VP recalculation:
- `CardPlayedEvent` — when a card with VP conditions is played, register VP source and recalculate
- `ResourceStorageChangedEvent` — covers animals, microbes, floaters, science, asteroids added to cards
- `TilePlacedEvent` — covers tile-based VP conditions (city-tile, ocean-tile, greenery-tile adjacent counting)
- `TagPlayedEvent` — covers per-tag VP conditions (e.g., jovian tags)
- `CorporationSelectedEvent` — register corporation VP source

**Rationale**:
- `ResourceStorageChangedEvent` is more precise than `ResourcesChangedEvent` for card-stored resources. The `ResourcesChangedEvent` covers player-level resources (credits, steel, etc.) which don't affect VP conditions. Card storage (animals, microbes, etc.) fires `ResourceStorageChangedEvent`.
- `TilePlacedEvent` covers board state changes needed for Capital (adjacent ocean tiles), Commercial District (adjacent city tiles), Immigration Shuttles (all city tiles), and Space Port Colony (colony tiles).
- `TagPlayedEvent` fires once per tag when a card is played, covering per-jovian-tag VP conditions.
- No new event types needed — the existing event system is sufficient.

**Alternatives considered**:
- Subscribing to `ResourcesChangedEvent` (player-level resources) — rejected because VP conditions only count card storage resources and tags, not player resources.
- Creating a new `VPRecalculationNeededEvent` — rejected as over-engineering; direct subscription to source events is cleaner.
- Subscribing to `GameStateChangedEvent` (generic) — rejected as too broad, would cause unnecessary recalculations.

## Research Task 2: VP Source Registration Pattern

**Question**: How should VP sources (VPGranters) be stored on the player and when should they be registered?

**Decision**: Add a new `VPGranters` component to the Player struct, following the same pattern as `PlayedCards`, `Effects`, and `GenerationalEvents`. The component maintains an ordered list of VPGranter structs. Registration happens:
1. In the `CardPlayedEvent` handler — when a card with non-empty `vpConditions` is played
2. In the `CorporationSelectedEvent` handler — when a corporation with VP conditions is selected

**Rationale**: The existing Player struct uses delegated components (Hand, PlayedCards, Resources, Effects, GenerationalEvents). Adding VPGranters as another component follows this established pattern. The ordered list preserves play order for UI display.

**Alternatives considered**:
- Storing VP sources directly on PlayerResources — rejected because VPGranters are a distinct concern from resources/production.
- Computing VP on-demand without storing sources — rejected because the user explicitly wants a VP source list on the player with computed values, and on-demand computation every render cycle would be wasteful.
- Storing on Game instead of Player — rejected because VP sources are per-player state.

## Research Task 3: Tile-Based VP Adjacency Handling

**Question**: How to handle VP conditions that depend on board adjacency (Capital: 1 VP per adjacent ocean, Commercial District: 1 VP per adjacent city)?

**Decision**: During VP recalculation, for cards with tile-based "per" conditions that need adjacency (ocean-tile, city-tile with location context), use the existing board adjacency functions. The card's tile placement position needs to be known — this is tracked by the board system (tiles have coordinates and owner IDs). For cards like Capital and Commercial District, we need to:
1. Find the tile the player placed for that card (match by card association or tile type + owner)
2. Get its neighbors via `HexPosition.GetNeighbors()`
3. Count matching adjacent tiles

**Challenge**: The current board system doesn't directly associate a placed tile with the card that caused its placement. Capital and Commercial District place city tiles as part of their card effects.

**Decision**: For the initial implementation, tile-based adjacency VP calculations (Capital, Commercial District) will continue to use the existing `calculateCityVPDetailed` and tile counting helpers in `vp_calculator.go`. The VPGranter system will focus on card-stored VP conditions (fixed, per-resource, per-tag). Tile VP remains in the existing VP breakdown categories (greeneryVP, cityVP). This avoids the tile-card association problem and keeps the implementation focused.

**Rationale**: The existing `vp_calculator.go` already handles tile VP correctly with adjacency. Duplicating that logic in VPGranters would be redundant. Capital and Commercial District's VP are already counted in `cityVP` via `calculateCityVPDetailed`.

**Alternatives considered**:
- Adding a `sourceCardID` field to board tiles — would require changes to tile placement flow, adds complexity beyond the scope of this feature.
- Computing adjacency VP in VPGranters — rejected because it duplicates existing tile VP calculation logic.

## Research Task 4: VP DTO and Frontend Data Flow

**Question**: How should VP granter data flow from backend to frontend for the new modal display?

**Decision**:
1. Add a `VPGranterDto` type with `json:` and `ts:` tags to the DTO layer
2. Include `vpGranters []VPGranterDto` in the `PlayerDto` (self-player view only — other players' VP sources should be hidden during gameplay for strategy)
3. The frontend reads `vpGranters` from the game state and renders the breakdown bar
4. Each VPGranterDto contains: `cardID`, `cardName`, `description`, `vpConditions` (the condition details), `computedValue` (current VP)

**Rationale**: The existing DTO pattern maps domain objects to transfer objects with `json:` and `ts:` tags. Adding VP granter data to the player DTO follows this pattern and makes it available through the existing WebSocket game state sync.

**Alternatives considered**:
- Separate API endpoint for VP data — rejected because the existing WebSocket state sync already provides player data.
- Including VP granters in other players' DTOs — rejected for gameplay strategy reasons (don't reveal opponents' VP breakdown during game).

## Research Task 5: Existing VP Modal Component Strategy

**Question**: What is the best approach for the VP modal rewrite?

**Decision**: Completely rewrite `VictoryPointsModal.tsx` with the new horizontal stacked bar design. The existing component uses a filter/sort list layout which will be replaced by the interactive bar visualization. The new component will:
1. Use `GameModal` with `theme="victoryPoints"` (existing theme)
2. Display a single horizontal stacked bar with segments for each VP granter
3. Add card name labels right-aligned above each segment
4. Implement hover tooltips with card info
5. Keep the total VP display from `VictoryPointsDisplay`

**Rationale**: The user explicitly requested scrapping all existing logic in the VP modal. The new design is fundamentally different (bar visualization vs. list), so a rewrite is cleaner than modification.

**Alternatives considered**:
- Incremental modification of existing modal — rejected per user instruction to scrap existing logic.
- Creating a new component alongside the old one — rejected because the old one would be dead code.

## Research Task 6: Color Scheme for VP Bar Segments

**Question**: How to generate distinct colors for potentially 10+ VP source segments that harmonize with the modal theme?

**Decision**: Use a palette of predefined shades that match the space/dark theme of the modal. The `victoryPoints` theme uses golden accents. Generate colors from a curated palette of dark-to-medium saturated tones:
- Base palette: shades of amber, teal, indigo, rose, emerald, violet, cyan, orange, fuchsia, lime
- Each VP source gets a color from this palette by index (cycling if > 10 sources)
- The segments use `opacity: 0.7` normally, brightening to `opacity: 1.0` on hover

**Rationale**: A curated palette ensures visual harmony with the space theme. Random HSL generation could produce clashing colors. The existing VP color scheme (blue for TR, purple for cards, etc.) serves as inspiration for the palette.

## Summary

All research questions have been resolved. Key architectural decisions:
1. Subscribe to 5 existing event types (no new events needed)
2. VPGranters as a new Player component following existing patterns
3. Tile adjacency VP stays in existing VP breakdown (not duplicated in VPGranters)
4. VP granter data flows through PlayerDto via WebSocket state sync
5. Complete rewrite of VictoryPointsModal with stacked bar design
6. Curated color palette for bar segments
