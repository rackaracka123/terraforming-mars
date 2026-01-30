---
name: tm-rules
description: Terraforming Mars game rules reference. This skill should be used whenever a task involves Terraforming Mars game mechanics, rules, card effects, resource management, global parameters, tile placement, scoring, or any feature implementation that touches game logic. Automatically consult the rules document to ensure correctness before implementing or modifying game features.
---

# Terraforming Mars Rules Reference

## Overview

This skill provides access to the authoritative Terraforming Mars rules document. Consult it whenever implementing, modifying, or debugging any game feature to ensure rule accuracy and correctness.

## When to Use

Trigger this skill for any task involving:

- Implementing or modifying game rules and logic
- Adding or changing card effects and interactions
- Working with resource management (MC, steel, titanium, plants, energy, heat)
- Modifying global parameters (temperature, oxygen, oceans)
- Tile placement rules and adjacency bonuses
- Production phase or generation flow
- Scoring, milestones, or awards
- Corporation abilities and setup
- Victory conditions and game-end triggers
- Validating game state transitions
- Debugging unexpected game behavior

## How to Use

### Step 1: Read the Rules Document

Before making any changes, read the full rules reference:

```
Read docs/TERRAFORMING_MARS_RULES.md
```

### Step 2: Find the Relevant Section

The rules document is organized by topic. Use grep to find specific rules quickly:

```
Grep pattern="temperature" path="docs/TERRAFORMING_MARS_RULES.md"
Grep pattern="ocean" path="docs/TERRAFORMING_MARS_RULES.md"
Grep pattern="production" path="docs/TERRAFORMING_MARS_RULES.md"
```

### Step 3: Verify Implementation Against Rules

Cross-reference the implementation with the rules document to ensure correctness. Pay special attention to:

- Resource conversion rates (e.g., 8 plants = 1 greenery, 8 heat = +1 temperature)
- TR bonus triggers (each global parameter step = +1 TR)
- Tile placement constraints (cities cannot be adjacent, greenery must be adjacent to own tiles)
- Production phase ordering (energy converts to heat first, then production generates)
- Card requirement validation (checked at play time, not at effect time)

## Examples

### Example 1: Implementing heat-to-temperature conversion

A task asks to implement the "convert heat" standard action.

1. Read the rules document
2. Find the relevant section: "Convert Heat: 8 heat -> +1 temperature -> +1 TR"
3. Also check the temperature track: range -30 to +8, bonuses at -24 and -20
4. Implement accordingly, ensuring TR is incremented and temperature bonuses are triggered

### Example 2: Adding a new project card with a building tag

A task asks to add a card that costs 15 MC with a building tag.

1. Read the rules document
2. Check "Steel Discount: Each steel reduces building costs by 2 MC"
3. Ensure the card implementation allows steel to be used for payment
4. Check if the card has any requirements (temperature, oxygen, tags) and validate them

### Example 3: Debugging why scoring is wrong

A bug report says final scores are incorrect.

1. Read the rules document, specifically the "Victory Conditions" and "Scoring" sections
2. Verify: Final Score = TR + VP from cards + Milestones + Awards
3. Check milestone scoring (5 VP each, max 3 per game)
4. Check award scoring rules
5. Verify tiebreakers: most MC first, then most cards played

### Example 4: Implementing the production phase

A task asks to implement end-of-generation production.

1. Read the rules document
2. Find "Production Sequence": energy converts to heat first, then resources are produced
3. Ensure the ordering is correct - energy-to-heat conversion happens BEFORE production
4. Check that MC production adds to TR (MC production = TR + MC production level)

## References

The full rules document is available at `docs/TERRAFORMING_MARS_RULES.md`. Key sections to grep for:

| Topic | Search Pattern |
|-------|---------------|
| Resource costs | `Standard Projects` |
| Temperature rules | `Temperature Track` |
| Oxygen rules | `Oxygen Track` |
| Ocean rules | `Ocean Tiles` |
| Card types | `Card Colors and Types` |
| Tile placement | `Tile Placement Rules` |
| Scoring | `Victory Conditions` |
| Production | `Production Phase` |
| Milestones | `Milestones` |
| Awards | `Awards` |
| Card tags | `Card Categories` |
| Resource conversion | `Resource Conversion` |
