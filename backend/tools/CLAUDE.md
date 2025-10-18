# Card Parser Tool

This document provides guidance for working with the card parser tool.

## Overview

Go-based command-line tool that converts Terraforming Mars card data from CSV format to structured JSON. This is a **temporary solution** until `backend/assets/terraforming_mars_cards.json` becomes the complete source of truth.

## Purpose

The card parser serves as an interim data pipeline:

1. **Input**: CSV files containing card definitions in `backend/assets/`
2. **Processing**: Parse CSV data into structured card objects
3. **Merging**: Combine with card pack metadata by card ID
4. **Output**: Generate JSON file for game engine consumption

**Important**: This tool will be **deprecated** once the JSON source of truth is complete and comprehensive.

## Current Status

**Limitations**:
- **Not all cards parse successfully yet**: Some card effects are too complex for current parser
- **Incomplete coverage**: Advanced card mechanics may require manual JSON authoring
- **Temporary architecture**: Designed for migration period, not long-term use

**Success Criteria**:
- When `terraforming_mars_cards.json` contains all cards with complete effect definitions
- When JSON source is validated and tested in game engine
- When CSV files are no longer needed

## Data Sources

### CSV Files

Located in `backend/assets/`:

- **Card definitions**: Base card data (name, cost, type, requirements, etc.)
- **Effect descriptions**: Text-based effect descriptions to parse
- **Tag information**: Card tags for gameplay mechanics

### Card Pack Metadata

**Card pack file** in `backend/assets/`:
- Maps card IDs to expansion packs
- Provides pack-specific metadata
- Merged with CSV data during parsing by card ID

**Merging Process**:
1. Parse CSV card data
2. Load card pack metadata
3. Match cards by unique ID
4. Combine card definition with pack info
5. Output unified JSON structure

## Architecture

### Directory Structure

```
tools/
├── cmd/              # Command-line entry points
│   └── parser/       # Card parser executable
├── internal/         # Parser-specific logic
│   ├── csv/          # CSV reading and parsing
│   ├── models/       # Card data structures
│   └── json/         # JSON generation and formatting
└── CLAUDE.md         # This file
```

### Data Flow

```
CSV Files → CSV Parser → Card Objects → Pack Merger → JSON Generator → Output File
     ↑                                        ↑
     |                                        |
backend/assets/               Card Pack Metadata (by ID)
```

## Development Workflow

### Running the Parser

**IMPORTANT**: Always output to either `/tmp/test_generation.json` (for testing) or directly to `assets/terraforming_mars_cards.json` (for permanent updates). **NEVER** output to other files in the repo as they create annoying artifacts that need cleanup.

```bash
# From backend/ directory

# For testing/validation (preferred for development)
go run tools/parse_cards.go /tmp/test_generation.json

# For permanent updates to the card database
go run tools/parse_cards.go assets/terraforming_mars_cards.json
```

### Expected Output

Generates structured JSON file with card definitions:

```json
{
  "cards": [
    {
      "id": "card-001",
      "name": "Asteroid",
      "cost": 14,
      "type": "event",
      "pack": "base",
      "tags": ["space"],
      "requirements": [],
      "effects": {
        "immediate": [
          { "type": "temperature", "amount": 1 },
          { "type": "damage-plants", "target": "any-player", "amount": 3 }
        ]
      }
    }
  ]
}
```

### Parsing Strategy

**Simple Cards** (currently supported):
- Basic resource production changes
- Standard global parameter modifications
- Simple immediate effects

**Complex Cards** (may fail parsing):
- Conditional effects based on game state
- Multi-step effect chains
- Complex player choice mechanics
- Cards with passive ongoing effects

**When Parsing Fails**:
1. Parser logs error with card ID and reason
2. Card is skipped in output
3. Manual JSON authoring required for that card
4. Add to `terraforming_mars_cards.json` directly

## Output Format

### Card Structure

Generated JSON follows the game engine's expected schema:

```json
{
  "id": "unique-card-id",
  "name": "Card Name",
  "cost": 10,
  "type": "automated|active|event",
  "pack": "base|corporate-era|prelude|etc",
  "tags": ["building", "science", "space"],
  "requirements": [
    {
      "type": "temperature|oxygen|oceans|tag-count",
      "value": 5,
      "comparison": "min|max|exact"
    }
  ],
  "effects": {
    "immediate": [/* effects on play */],
    "ongoing": [/* passive effects */],
    "action": [/* activated abilities */]
  },
  "victoryPoints": 2,
  "description": "Text description"
}
```

### Effect Types

**Immediate Effects** (on play):
- Resource production changes
- Global parameter modifications
- Resource transfers
- Tile placements

**Ongoing Effects** (passive):
- Triggered by game events
- Continuous modifiers
- Conditional bonuses

**Action Effects** (activated):
- Player-activated abilities
- Cost and benefit structure
- Timing restrictions

## Merging with Card Packs

### Card Pack File Format

```json
{
  "packs": {
    "base": {
      "name": "Base Game",
      "cards": ["card-001", "card-002"]
    },
    "corporate-era": {
      "name": "Corporate Era",
      "cards": ["card-200", "card-201"]
    }
  }
}
```

### Merge Logic

1. **Load card pack metadata**: Parse card pack JSON
2. **Build ID-to-pack map**: Create lookup table
3. **Parse CSV cards**: Convert each card to object
4. **Match by ID**: Find pack for each card ID
5. **Combine data**: Add pack info to card object
6. **Output merged JSON**: Write complete card definitions

## Parser Improvement Guidelines

### Adding New Effect Parsing

1. **Identify effect pattern**: Find common text patterns in CSV
2. **Define effect structure**: Create JSON representation
3. **Implement parser**: Add parsing logic for new effect type
4. **Test with sample cards**: Verify correct parsing
5. **Handle edge cases**: Account for variations in description

### Handling Parse Failures

1. **Log detailed error**: Card ID, description, failure reason
2. **Continue processing**: Don't halt on individual card failure
3. **Generate report**: List all failed cards after parsing
4. **Manual authoring**: Add complex cards to JSON directly

### Testing Parser Changes

```bash
# From backend/ directory

# Run parser to temporary file for testing
go run tools/parse_cards.go /tmp/test_generation.json

# Check output
jq 'length' /tmp/test_generation.json

# Validate JSON structure - inspect specific cards
jq '.[] | select(.name == "Ecoline")' /tmp/test_generation.json

# Count cards parsed
jq 'length' /tmp/test_generation.json

# Check for parsing errors in logs
```

## Migration to JSON Source of Truth

### Current State

- **CSV files**: Primary data source (temporary)
- **Parser**: Conversion tool (temporary)
- **Generated JSON**: Output for game engine
- **Manual JSON**: Complex cards authored by hand

### Target State

- **terraforming_mars_cards.json**: Single source of truth
- **No parser needed**: Direct JSON authoring
- **Version control**: Track card changes in Git
- **Validation**: JSON schema validation for correctness

### Migration Steps

1. **Identify all cards**: Catalog complete card list from all expansions
2. **Parse what works**: Use parser for simple cards
3. **Author complex cards**: Manually write JSON for complex effects
4. **Validate completeness**: Ensure all cards are represented
5. **Test in game engine**: Verify all cards work correctly
6. **Deprecate parser**: Remove CSV files and parsing tool
7. **Update documentation**: Reflect JSON-only workflow

## Future Deprecation Plan

**When to deprecate**:
- ✅ All cards from all expansions in `terraforming_mars_cards.json`
- ✅ All card effects fully defined and tested
- ✅ Game engine successfully loads and uses JSON source
- ✅ JSON schema validation in place
- ✅ Documentation updated for JSON-only workflow

**Deprecation process**:
1. Mark parser tool as deprecated in documentation
2. Move CSV files to `backend/assets/archive/`
3. Remove parser from build process
4. Delete parser code after grace period
5. Update onboarding docs to reference JSON only

## Common Tasks

### Adding a New Card Manually

For cards that don't parse well:

1. Add card definition to `backend/assets/terraforming_mars_cards.json`:

```json
{
  "id": "new-card-id",
  "name": "Complex Card",
  "cost": 15,
  "type": "automated",
  "pack": "base",
  "tags": ["science", "building"],
  "effects": {
    "immediate": [
      {
        "type": "custom-logic",
        "description": "Complex effect requiring special handling"
      }
    ]
  }
}
```

2. Implement effect handler in backend card system
3. Test card in game engine
4. Validate JSON structure

### Running Parser for Subset of Cards

```bash
# From backend/ directory

# Test run - output to temporary file
go run tools/parse_cards.go /tmp/test_generation.json

# Production run - update permanent card database
go run tools/parse_cards.go assets/terraforming_mars_cards.json

# Note: Parser currently processes all cards. To parse specific packs,
# modify the source code or filter the output JSON afterwards
```

### Validating Parser Output

```bash
# From backend/ directory

# Check JSON is valid
jq empty /tmp/test_generation.json

# Count cards
jq 'length' /tmp/test_generation.json

# Inspect specific card
jq '.[] | select(.id == "B02")' /tmp/test_generation.json

# List all corporations
jq '.[] | select(.type == "corporation") | .name' /tmp/test_generation.json
```

## Important Notes

### Temporary Nature

**This tool is intentionally temporary**:
- Built for migration convenience, not long-term maintenance
- Code quality prioritizes functionality over elegance
- No extensive testing or edge case handling
- Will be removed when JSON source is complete

### Parser Limitations

- **Text parsing is fragile**: CSV descriptions have inconsistent formatting
- **Context-dependent effects**: Some effects need full game state context
- **Human verification required**: Always review parser output
- **Not production-critical**: Game engine uses output, not parser itself

### Best Practices

- **Don't over-engineer**: This is a temporary tool
- **Focus on coverage**: Get as many cards as possible
- **Manual authoring OK**: Complex cards can be written by hand
- **Validate output**: Always check generated JSON before use

## Related Documentation

- **Project Root CLAUDE.md**: Full-stack architecture and workflows
- **backend/CLAUDE.md**: Backend API architecture and card system
- **frontend/CLAUDE.md**: Frontend architecture and patterns
- **TERRAFORMING_MARS_RULES.md**: Complete game rules reference
- **backend/assets/**: Card data sources (CSV and JSON)
