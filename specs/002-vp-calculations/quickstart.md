# Quickstart: VP Calculations

**Feature Branch**: `002-vp-calculations`
**Date**: 2026-01-30

## Prerequisites

```bash
git checkout 002-vp-calculations
make dev-setup
```

## Development Flow

### 1. Backend Changes

Files to create/modify:

```
backend/internal/game/player/player_vp_granters.go  # NEW: VPGranters component
backend/internal/game/player/player.go               # MODIFY: Add vpGranters field
backend/internal/game/game.go                        # MODIFY: Add subscribeToVPEvents()
backend/internal/delivery/dto/game_dto.go            # MODIFY: Add VPGranterDto
backend/internal/delivery/dto/mapper_game.go         # MODIFY: Map VP granters to DTO
```

### 2. Generate Types

```bash
make generate
```

### 3. Frontend Changes

Files to create/modify:

```
frontend/src/components/ui/modals/VictoryPointsModal.tsx  # REWRITE: New stacked bar design
```

### 4. Tests

```
backend/test/game/player/player_vp_granters_test.go  # NEW: VPGranters unit tests
backend/test/game/cards/vp_calculation_test.go       # NEW: VP calculation integration tests
```

### 5. Validate

```bash
make test
make format
make lint
make prepare-for-commit
```

## Key Architecture Decisions

- VP sources stored as ordered list on Player (VPGranters component)
- Event-driven recalculation via existing EventBus subscriptions
- Tile adjacency VP stays in existing VP breakdown (not duplicated)
- VP granter data sent to self-player only via WebSocket state sync
- Complete rewrite of VictoryPointsModal with horizontal stacked bar
