# BehaviorSection - Card Behavior Display System

This document provides guidance for understanding and extending the BehaviorSection component system.

## Overview

BehaviorSection is a sophisticated card behavior rendering system that transforms raw card behavior data from the backend into visually organized, context-aware UI presentations. It handles the complex task of displaying card effects, resource costs/gains, triggered conditions, production chains, and player choices while optimizing for limited card space.

**Core Responsibility**: Convert abstract CardBehaviorDto data structures into human-readable visual representations that communicate game mechanics clearly to players.

## Architecture Principles

### Behavior Classification Pipeline

The system operates through a classification pipeline that determines how each behavior should be rendered:

1. **Classification**: Raw behaviors are analyzed and assigned semantic types (discount, manual-action, triggered-effect, immediate-production, etc.)
2. **Merging**: Related behaviors are consolidated to reduce visual clutter (e.g., multiple production effects merged into single display)
3. **Space Analysis**: System estimates total space requirements and forces compact mode if needed
4. **Layout Routing**: Each classified behavior is routed to an appropriate specialized layout handler
5. **Rendering**: Layout handlers coordinate with primitive components to render final display

### Key Design Patterns

**Semantic Classification**

- Behavior type drives all downstream decisions (layout, styling, affordability)
- Classification based on triggers, inputs, outputs, and combinations
- Examples: "discount" (affects costs), "manual-action" (player-activated), "triggered-effect" (passive)

**Space Optimization Strategy**

- System calculates layout plans before rendering
- If behaviors exceed MAX_CARD_ROWS (4 rows), forces compact mode
- Compact mode converts resources to "NxIcon" format instead of individual icons
- Maintains readability while reducing visual footprint

**Context-Aware Rendering**

- Same resource renders differently based on context (input vs. output, production vs. immediate)
- Icon sizing adapts to standalone vs. action vs. production contexts
- Affordability integrated seamlessly (grayscale for unaffordable resources)

**Display Mode Coordination**

- Resources display as either individual icons or "number + icon" format
- Consistency rule: If ANY resource in a group uses number mode, ALL use it (except amount=1)
- Decision based on available space and configured thresholds

## Component Hierarchy

### Container Level

- **BehaviorSection**: Main orchestrator coordinating the entire rendering pipeline
- **BehaviorContainer**: Wraps behaviors in type-specific styling (gradients, borders, backgrounds)

### Layout Handlers

Specialized components for specific behavior types:

- **ManualActionLayout**: Player-activated actions with choice support (OR alternatives)
- **TriggeredEffectLayout**: Passive effects with trigger conditions (displays trigger : effect)
- **ImmediateResourceLayout**: Production and immediate effects (most complex with extensive special cases)

### Rendering Primitives

- **ResourceDisplay**: Individual resource/production renderer with affordability and display mode logic
- **BehaviorIcon**: Icon wrapper with context-aware sizing and styling
- **CardIcon**: Specialized display for card operations (draw, peek, take, buy)

## Data Flow

```
CardBehaviorDto[] from backend
    ↓
Classify behaviors → Assign semantic types
    ↓
Merge auto-production behaviors → Reduce clutter
    ↓
Detect tile placement scales → Determine icon sizing
    ↓
Analyze card layout → Estimate space requirements
    ↓
Optimize for space if needed → Force compact mode
    ↓
For each behavior:
  Route to appropriate layout handler
    ↓
  Handler uses ResourceDisplay for resources
    ↓
  ResourceDisplay uses BehaviorIcon for icons
    ↓
BehaviorContainer wraps with type-specific styling
    ↓
Rendered in scrollable column
```

## Extending the System

### Adding New Trigger Types

1. **Update Classification Logic**
   - Add detection for new trigger type in behavior classifier
   - Assign to existing ClassifiedBehavior type or create new type

2. **Create Layout Handler**
   - Create new component following existing layout handler patterns
   - Receive behavior data and layout plan
   - Render trigger indicators and effects
   - Handle special spacing/grouping for this trigger type

3. **Register in Router**
   - Add case in BehaviorSection switch statement
   - Route classified behavior to new layout handler

4. **Define Visual Styling**
   - Add type-specific gradient/border in BehaviorContainer
   - Unique background for visual distinction

**Example**: The "placement-bonus-gained" trigger displays affected resources (steel/titanium icons) followed by colon separator, then outputs.

### Adding New Display Modes

1. **Update Display Analysis**
   - Extend resource display analysis logic
   - Add conditions determining when new mode applies

2. **Update ResourceDisplay Component**
   - Add new rendering branch for mode
   - Handle affordability and context-specific styling

### Adding New Resource Types

1. **Centralize Icon Path**
   - Add to iconStore.ts in appropriate category (RESOURCE_ICONS, TAG_ICONS, SPECIAL_ICONS)

2. **Update Classification** (if needed)
   - May need updates in behavior classifier for special display rules
   - May need special handling in layout components

3. **Update ResourceDisplay** (if needed)
   - Add branch for new resource type if special affordability/styling needed

### Handling Complex Layout Scenarios

**Production Grouping**

- Negative production (costs) separated from positive (gains)
- Visual indicators: red "-" for costs, green "+" for gains
- Brown box styling wraps all production for visual cohesion
- Per-condition production uses special format (e.g., "2 MC/planet")

**Multi-Output Layouts**

- Two-row layout: Resources on top, global parameters/tiles on bottom
- Two-column layout: Resources stacked vertically left, global params right
- Priority sorting: TR (terraform rating) always last in global parameter group
- Attack resources separated and displayed first with red pulsing glow

## Key Concepts

### Affordability Integration

- System checks if player can afford action costs
- Unaffordable actions rendered in grayscale
- Seamlessly integrated without UI clutter

### Tile Scaling

- Tiles scale dynamically based on context:
  - 2x if completely alone
  - 1.5x if single tile in behavior
  - 1.25x if pair of tiles
  - 1x normal size otherwise

### Behavior Merging

- Multiple "auto-no-background" behaviors (production/immediate) merged into single display
- Preserves visual clarity by reducing container count
- Does NOT merge behaviors with trigger conditions (maintains clarity about when effects trigger)

### Display Contexts

Different contexts affect icon sizing:

- **Standalone**: Large icons for tiles/cards (36px)
- **Action**: Medium icons with inputs (26px)
- **Production**: Brown box context (26px)
- **Default**: Standard size (26px)

## Development Guidelines

### When to Update This System

- **New Card Mechanics**: If new card introduces trigger or effect type not currently supported
- **Layout Issues**: If existing behaviors don't render well (overlapping, truncation, poor spacing)
- **Space Optimization**: If cards consistently exceed available space
- **Visual Clarity**: If player feedback indicates confusion about card effects

### Testing Considerations

- Test with cards having maximum complexity (multiple inputs, outputs, triggers, choices)
- Verify space optimization kicks in appropriately
- Check affordability indicators work correctly
- Ensure tile scaling applies correctly in different contexts
- Test choice-based behaviors (OR alternatives) display clearly

### Performance Notes

- Classification and space analysis happen once during initial render
- Layout plans cached to avoid recalculation
- Re-renders only when behavior data or affordability changes

## Integration Points

### Backend Data Contract

Expects CardBehaviorDto structures with:

- `triggers`: Array of trigger conditions (manual, auto, or event-based)
- `inputs`: Required resources (costs)
- `outputs`: Produced resources or effects
- `choices`: Alternative action paths (A OR B)
- `per`: Conditional multipliers (per tag, per resource)

### GameIcon Component

All icon rendering goes through centralized GameIcon component:

- NEVER use direct `<img src="/assets/...">` tags
- GameIcon handles sizing, affordability, production backgrounds automatically
- Add new icons to iconStore.ts, use via GameIcon

### Affordability System

Receives `isResourceAffordable` function that checks player resources:

- Called for each input resource
- Returns boolean determining grayscale filter
- Integrates with broader game state management

## Common Patterns

### Separator Usage

- **Arrow (→)**: Between inputs and outputs in manual actions
- **Colon (:)**: Between trigger condition and effect in triggered effects
- **Slash (/)**: Between alternative resources (e.g., steel/titanium)
- **Minus/Plus (-/+)**: For negative/positive production values

### Visual Hierarchy

- Gradients distinguish behavior types (blue for actions, brown for production, transparent for triggered)
- Spacing and grouping communicate relationships between resources
- Color coding indicates affordability, attack indicators, production vs. immediate

### Graceful Degradation

- Falls back to text display if icons unavailable
- Handles missing or malformed data without crashing
- Provides reasonable defaults for edge cases

## Related Documentation

- **frontend/CLAUDE.md**: Frontend architecture and component standards
- **TERRAFORMING_MARS_RULES.md**: Game rules reference for understanding card mechanics
- **backend/assets/terraforming_mars_cards.json**: Card behavior data structures
