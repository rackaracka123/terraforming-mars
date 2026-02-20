# Research: Expandable Card Descriptions

**Date**: 2026-02-18 | **Branch**: `001-full-card-view`

## Decision 1: Behavior Description Overlay Approach

**Decision**: Add hover state management to BehaviorContainer. When a behavior row is hovered and has a `description`, render a positioned overlay element below the container using `position: relative` on the container and `position: absolute` on the description. The description floats over content below without affecting card layout.

**Rationale**: BehaviorContainer already wraps every behavior row and applies type-specific styling. Adding hover state here is the minimal-touch approach — each layout handler (ManualActionLayout, TriggeredEffectLayout, etc.) remains unchanged. The overlay approach avoids disrupting card grid layouts in modals and the /cards virtual scroll.

**Alternatives considered**:
- Adding hover logic inside each layout handler — rejected because it would require touching 7+ layout files when BehaviorContainer is the single wrapper
- Using CSS-only `:hover` pseudo-class with adjacent sibling — rejected because description text needs to come from the behavior data prop, not just CSS visibility
- Expanding the card on hover — rejected per clarification (overlay approach chosen)

## Decision 2: Description Data Access in BehaviorContainer

**Decision**: Thread the `description` field from `CardBehaviorDto` through the classification pipeline into `ClassifiedBehavior`, then pass it as a prop to `BehaviorContainer`. BehaviorContainer renders the description overlay when hovered.

**Rationale**: The classification pipeline (`classifyBehaviors`) already maps from `CardBehaviorDto[]` to `ClassifiedBehavior[]`. The `ClassifiedBehavior` type needs to carry through the `description` field so BehaviorContainer has access to it. This is a minimal type extension.

**Alternatives considered**:
- Passing a parallel descriptions array alongside behaviors — rejected as fragile and error-prone
- Looking up descriptions from a context/store — rejected as over-engineering

## Decision 3: VP and Resource Storage Description at Card Bottom

**Decision**: Add a description area at the bottom of the card (below the behaviors content section, above the checkbox). Managed by hover state on VictoryPointIcon and a new resource storage indicator. When either is hovered, the description text renders in this bottom area. When neither is hovered, the area is hidden.

**Rationale**: The VP icon is already positioned at the left side of the title bar. Making it hoverable and showing a description at the card bottom is non-disruptive. For resource storage, there is currently no visual indicator — a small icon needs to be added that's hoverable.

**Alternatives considered**:
- Tooltip floating near the VP icon — rejected because the user specified "at the BOTTOM of the card"
- Integrating VP/storage descriptions into BehaviorSection rows — rejected because VP and storage are card-level attributes, not behaviors

## Decision 4: Resource Storage Visual Indicator

**Decision**: Add a small resource storage icon to the card (near the VP icon area or in the behavior section area) that indicates the card stores resources. This icon is hoverable and shows the storage description at the card bottom.

**Rationale**: Currently there is no visual representation of resource storage on cards. A small icon (using GameIcon with the storage resource type) provides the hover target needed for this feature.

**Alternatives considered**:
- Text label "Stores: [type]" — rejected as too verbose for compact card layout
- No indicator, only show in expanded view — rejected because the old approach of separate expanded view was removed

## Decision 5: Rename Strategy

**Decision**: Rename the file `SimpleGameCard.tsx` to `GameCard.tsx`, update the component name and interface name, and update all 8 import sites. Do this as the last task to avoid merge conflicts with other changes.

**Rationale**: The rename touches 8+ files (CardsPage, CardFanOverlay, StartingCardSelectionOverlay, PendingCardSelectionOverlay, ProductionCardSelectionOverlay, CardDrawSelectionOverlay, DemoSetupOverlay, CardsPlayedModal). Doing it last minimizes conflict risk with hover changes.

**Alternatives considered**:
- Rename first — rejected because it would cause conflicts with every subsequent task
- Keep the name SimpleGameCard — rejected per user requirement
