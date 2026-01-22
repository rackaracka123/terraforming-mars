# Assets - Static Game Data

Contains card definitions and static game data loaded by the backend at startup.

## Card Database

Authoritative source for all card definitions: corporations, project cards, and prelude cards with behaviors, costs, requirements, and effects.

### Card Structure

```json
{
  "id": "B07",
  "name": "PhoboLog",
  "type": "corporation",
  "cost": 0,
  "description": "Effect description text",
  "pack": "base-game",
  "tags": ["space"],
  "behaviors": [...]
}
```

**Card Types:** `corporation`, `active`, `automated`, `event`, `prelude`

**Card Packs:** `base-game`, `corporate-era`, `prelude`, `venus-next`, `colonies`, `turmoil`

### Behavior System

Each card has a `behaviors` array. Each behavior contains:
- `triggers`: When the behavior activates
- `inputs`: Resources consumed (costs)
- `outputs`: Resources produced (effects)
- `choices`: Alternative options (A OR B)

### Trigger Types

| Trigger | Description |
|---------|-------------|
| `auto` | Applies immediately when card is played |
| `auto-corporation-start` | Applies once when corporation is selected |
| `manual` | Player-activated action (blue cards) |
| `auto` + `condition` | Passive effect triggered by game events |

### Output Types

**Basic Resources:** `credit`, `steel`, `titanium`, `plant`, `energy`, `heat`

**Production:** `credit-production`, `steel-production`, `titanium-production`, `plant-production`, `energy-production`, `heat-production`

**Global Parameters:** `oxygen`, `temperature`, `ocean-placement`

**Tile Placements:** `city-placement`, `greenery-placement`, `ocean-placement`

**Special:**
- `discount` - Reduces card costs (uses `affectedTags` or `affectedStandardProjects`)
- `payment-substitute` - Allows resource as credit payment (uses `affectedResources` for resource type, `amount` for conversion rate)
- `value-modifier` - Increases resource payment value (uses `affectedResources` for steel/titanium, `amount` for bonus)
- `tr` - Terraform rating

### Value Modifier Example

```json
{
  "type": "value-modifier",
  "amount": 1,
  "target": "self-player",
  "affectedResources": ["titanium"]
}
```

Makes each titanium worth 1 additional credit when paying for cards.

## Adding New Cards

1. Add card JSON to this directory's card database
2. Use existing trigger and output types where possible
3. For new effect types, implement handler in `internal/game/cards/`
4. Run `make test` to validate card loading

Most cards (90%+) can be added via JSON only without Go code changes.
